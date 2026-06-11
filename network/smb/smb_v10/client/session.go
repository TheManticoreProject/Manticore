package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	spnego_ntlm_negotiate_flags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// ntStatusMoreProcessingRequired (STATUS_MORE_PROCESSING_REQUIRED) is the status
// a server returns on the challenge leg of an extended-security session setup to
// signal that it expects the follow-up AUTHENTICATE message.
const ntStatusMoreProcessingRequired = 0xC0000016

// Session represents an established session between the client and server
type Session struct {
	// The SMB connection associated with this session
	Client *Client

	// The cryptographic session key associated with this session
	SessionKey []byte

	// The 2-byte UID for this session
	SessionUID uint16

	// TreeID is the TID of the tree connect currently selected for outbound
	// commands (file I/O, etc.). It is populated by Client.TreeConnect.
	TreeID uint16

	// The credentials for this session
	Credentials *credentials.Credentials
}

func (s *Session) SessionSetup() error {
	if !s.Client.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	// The NTLM challenge/response path and the second authentication phase both
	// build an authentication context from s.Credentials. Reject nil credentials
	// up front with a clear error instead of dereferencing a nil pointer later.
	if s.Credentials == nil {
		return fmt.Errorf("session setup requires credentials but none were provided")
	}

	// Prepare and send a NTLMSSP NEGOTIATE message =============================================================================================
	requestStep1Msg := message.NewMessage()
	sessionSetupCmd := commands.NewSessionSetupAndxRequest()

	// Here put the common logic for all session setup commands
	requestStep1Msg.Header.Command = codes.SMB_COM_SESSION_SETUP_ANDX
	requestStep1Msg.Header.Flags = flags.Flags(flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE)
	requestStep1Msg.Header.Flags2 = flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_LONG_NAMES_ALLOWED | flags2.FLAGS2_EXTENDED_SECURITY)

	// Add Unicode support if server supports it
	if s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE {
		requestStep1Msg.Header.Flags2 |= flags2.Flags2(flags2.FLAGS2_UNICODE)
	}

	// Set message signing flags based on server security mode
	if s.Client.Connection.Server.SecurityMode.IsSecuritySignatureEnabled() ||
		s.Client.Connection.Server.SecurityMode.IsSecuritySignatureRequired() {
		requestStep1Msg.Header.Flags2 |= flags2.Flags2(flags2.FLAGS2_SECURITY_SIGNATURE)
	}

	// Set process ID and multiplex ID
	requestStep1Msg.Header.SetPID(0)
	requestStep1Msg.Header.MID = 0
	requestStep1Msg.Header.TID = 65535
	requestStep1Msg.Header.UID = 0

	sessionSetupCmd.MaxBufferSize = types.USHORT(s.Client.Connection.Server.MaxBufferSize)
	sessionSetupCmd.MaxMpxCount = s.Client.Connection.MaxMpxCount
	sessionSetupCmd.Capabilities = s.Client.Connection.Server.Capabilities

	// NativeOS / NativeLanMan are informational strings, but they MUST NOT be empty:
	// strict servers (e.g. Windows Server 2016) reject a SESSION_SETUP_ANDX whose
	// NativeOS/NativeLanMan are empty, replying with a DOS-class error and an empty
	// security blob — which then fails SPNEGO parsing with "asn1: sequence truncated".
	// Default them when the caller left them blank so every caller interoperates.
	sessionSetupCmd.NativeOS = s.Client.NativeOS
	sessionSetupCmd.NativeLanMan = s.Client.NativeLanMan
	if sessionSetupCmd.NativeOS == "" {
		sessionSetupCmd.NativeOS = DefaultNativeOS
	}
	if sessionSetupCmd.NativeLanMan == "" {
		sessionSetupCmd.NativeLanMan = DefaultNativeLanMan
	}

	// Track whether step 1 used the extended-security (NTLMSSP) challenge/response
	// path. Only that path performs the second-step authenticate exchange below;
	// the share-level and plaintext paths complete in a single round trip.
	useExtendedSecurity := false

	// Check if we're using share level access control
	if s.Client.Connection.Server.SecurityMode.SupportsShareLevelAccessControl() {
		// Share level access control is required by the server
		// If no authentication has been performed on the SMB connection, use anonymous authentication

		// Parameters
		sessionSetupCmd.VcNumber = types.USHORT(0x0000)
		sessionSetupCmd.SessionKey = s.Client.Connection.Server.SessionKey
		sessionSetupCmd.OEMPasswordLen = types.USHORT(0x0000)
		sessionSetupCmd.UnicodePasswordLen = types.USHORT(0x0000)

		// Data section - for null session, use empty strings
		sessionSetupCmd.OEMPassword = []types.UCHAR{}
		sessionSetupCmd.UnicodePassword = []types.UCHAR{}

	} else {
		// User level access control is required by the server.
		// Reuse of an existing session for the same credentials is handled by
		// Client.SessionSetup, which consults Connection.SessionTable before
		// reaching this authentication path.

		// Handle authentication based on server capabilities
		if s.Client.Connection.Server.SecurityMode.SupportsChallengeResponseAuth() {
			// Server supports challenge/response authentication
			// Determine authentication type based on policies
			useExtendedSecurity = true

			sessionSetupCmd.VcNumber = types.USHORT(0x0000)
			sessionSetupCmd.SessionKey = s.Client.Connection.Server.SessionKey

			useUnicode := s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE

			authCtx := spnego.NewAuthContext(
				spnego.AuthTypeNTLM,
				s.Credentials.Domain,
				s.Credentials.Username,
				s.Credentials.Password,
				s.Client.Workstation,
				useUnicode,
			)

			negotiateFlags := spnego_ntlm_negotiate_flags.NegotiateFlags(
				spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_NTLM |
					// Extended session security (NTLM2) is required to derive a session key:
					// it forces the server's CHALLENGE to set the flag, so the AUTHENTICATE
					// takes the NTLMv2 path and computes the SessionBaseKey ([MS-NLMP] 3.3.2).
					// Without it the NTLMv1 path runs and no key is derived, so SMB message
					// signing (required by hardened servers, e.g. Windows Server 2003+) cannot
					// be activated. We deliberately do NOT request NTLMSSP_NEGOTIATE_KEY_EXCH,
					// so ExportedSessionKey == KeyExchangeKey == SessionBaseKey ([MS-NLMP]
					// KXKEY) and EncryptedRandomSessionKey stays empty.
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_SIGN |
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_128 |
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_56 |
					spnego_ntlm_negotiate_flags.NTLMSSP_REQUEST_TARGET |
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_VERSION,
			)
			if useUnicode && !negotiateFlags.HasFlag(spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_UNICODE) {
				negotiateFlags |= spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_UNICODE
			}

			v := &version.Version{
				ProductMajorVersion: 5,
				ProductMinorVersion: 1,
				ProductBuild:        0,
				Reserved:            [3]byte{0, 0, 0},
				NTLMRevision:        version.NTLMSSP_REVISION_W2K3,
			}
			negotiateToken, err := authCtx.CreateNegotiateToken(negotiateFlags, v)
			if err != nil {
				return fmt.Errorf("failed to create negotiate token: %v", err)
			}
			sessionSetupCmd.SecurityBlob = negotiateToken
		} else {
			// Server doesn't support challenge/response authentication

			// Use plaintext authentication
			sessionSetupCmd.VcNumber = types.USHORT(0x0000)
			sessionSetupCmd.SessionKey = s.Client.Connection.Server.SessionKey

			if s.Credentials == nil {
				return fmt.Errorf("plaintext authentication requires credentials but none were provided")
			}
			password := s.Credentials.Password

			// Check if Unicode is supported
			if s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE {
				// Send password in Unicode
				sessionSetupCmd.UnicodePassword = []types.UCHAR(utf16.EncodeUTF16LE(password))
				sessionSetupCmd.UnicodePasswordLen = types.USHORT(len(sessionSetupCmd.UnicodePassword))
				sessionSetupCmd.OEMPasswordLen = types.USHORT(0x0000)
				sessionSetupCmd.OEMPassword = []types.UCHAR{}
			} else {
				// Send password in OEM format
				sessionSetupCmd.OEMPassword = []types.UCHAR(password)
				sessionSetupCmd.OEMPasswordLen = types.USHORT(len(sessionSetupCmd.OEMPassword))
				sessionSetupCmd.UnicodePasswordLen = types.USHORT(0x0000)
				sessionSetupCmd.UnicodePassword = []types.UCHAR{}
			}
		}
	}
	requestStep1Msg.AddCommand(sessionSetupCmd)

	marshalledMessage, err := requestStep1Msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal negotiate message: %v", err)
	}

	// Send the message
	_, err = s.Client.Transport.Send(marshalledMessage)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Wait for a NTLMSSP CHALLENGE response message =============================================================================================

	rawResponseMessage, err := s.Client.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	responseMsg := message.NewMessage()
	err = responseMsg.Unmarshal(rawResponseMessage)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if responseMsg.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", responseMsg.Header.Command)
	}

	challengeResponse := responseMsg.Command.(*commands.SessionSetupAndxResponse)

	// Share-level and plaintext authentication complete in a single round trip.
	// Only the extended-security (NTLMSSP) path continues to the authenticate
	// step below; for the other paths the step-1 response is the final response,
	// so validate its status and finalize here instead of treating the response
	// as an NTLMSSP challenge.
	if !useExtendedSecurity {
		if responseMsg.Header.Status != 0x00 {
			if name, ok := nt_status.NTStatusToStringName[nt_status.NT_STATUS(responseMsg.Header.Status)]; ok {
				return fmt.Errorf("session setup failed: %s (0x%08x)", name, responseMsg.Header.Status)
			}
			return fmt.Errorf("session setup failed: 0x%08x", responseMsg.Header.Status)
		}
		s.SessionUID = responseMsg.Header.UID
		return nil
	}

	// Prepare and send a NTLMSSP AUTH message ==================================================================================================

	// Server supports challenge/response authentication
	// Determine authentication type based on policies

	useUnicode := s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE

	authCtx := spnego.NewAuthContext(
		spnego.AuthTypeNTLM,
		s.Credentials.Domain,
		s.Credentials.Username,
		s.Credentials.Password,
		s.Client.Workstation,
		useUnicode,
	)

	requestStep2Msg := message.NewMessage()
	requestStep2Msg.Header.Command = codes.SMB_COM_SESSION_SETUP_ANDX
	requestStep2Msg.Header.Flags = requestStep1Msg.Header.Flags
	requestStep2Msg.Header.Flags2 = requestStep1Msg.Header.Flags2
	requestStep2Msg.Header.SetPID(requestStep1Msg.Header.GetPID())
	requestStep2Msg.Header.MID = requestStep1Msg.Header.MID
	requestStep2Msg.Header.TID = requestStep1Msg.Header.TID
	// Here we need to set the UID to the UID of the response message
	requestStep2Msg.Header.UID = responseMsg.Header.UID
	// Save the session UID
	s.SessionUID = requestStep2Msg.Header.UID

	sessionSetupStep2Cmd := commands.NewSessionSetupAndxRequest()
	sessionSetupStep2Cmd.VcNumber = sessionSetupCmd.VcNumber
	sessionSetupStep2Cmd.SessionKey = sessionSetupCmd.SessionKey
	sessionSetupStep2Cmd.Capabilities = sessionSetupCmd.Capabilities
	// The AUTHENTICATE (second) leg is a full SESSION_SETUP_ANDX request and MUST
	// repeat the connection parameters from the first leg. In particular the
	// server validates MaxBufferSize against its minimum on every leg
	// ([MS-CIFS] 2.2.4.6; Samba's smbd/sesssetup.c rejects smb_bufsize <
	// SMB_BUFFER_SIZE_MIN with ERRSRV/ERRerror), so leaving it at the zero value
	// makes a server that already accepted the NTLM credentials still fail the
	// session setup. Carry over MaxBufferSize, MaxMpxCount and the native OS/LanMan
	// strings just as the first leg sent them.
	sessionSetupStep2Cmd.MaxBufferSize = sessionSetupCmd.MaxBufferSize
	sessionSetupStep2Cmd.MaxMpxCount = sessionSetupCmd.MaxMpxCount
	sessionSetupStep2Cmd.NativeOS = sessionSetupCmd.NativeOS
	sessionSetupStep2Cmd.NativeLanMan = sessionSetupCmd.NativeLanMan

	// The extended-security challenge leg MUST come back with
	// STATUS_MORE_PROCESSING_REQUIRED (or, on some servers, STATUS_SUCCESS) and a
	// non-empty NTLMSSP CHALLENGE security blob. If the server instead rejected our
	// NEGOTIATE token it returns an error status and/or an empty blob; surface that
	// directly here rather than feeding an empty buffer to the SPNEGO parser, which
	// would otherwise fail with the opaque "asn1: syntax error: sequence truncated".
	if responseMsg.Header.Status != 0x00000000 && responseMsg.Header.Status != ntStatusMoreProcessingRequired {
		if name, ok := nt_status.NTStatusToStringName[nt_status.NT_STATUS(responseMsg.Header.Status)]; ok {
			return fmt.Errorf("session setup challenge failed: %s (0x%08x)", name, responseMsg.Header.Status)
		}
		return fmt.Errorf("session setup challenge failed: 0x%08x", responseMsg.Header.Status)
	}
	if len(challengeResponse.SecurityBlob) == 0 {
		return fmt.Errorf("server returned an empty security blob in the session setup challenge (status 0x%08x); the server rejected the SPNEGO/NTLMSSP negotiate token", responseMsg.Header.Status)
	}

	authenticateToken, err := authCtx.CreateAuthenticateTokenFromChallengeToken(challengeResponse.SecurityBlob)
	if err != nil {
		return fmt.Errorf("failed to process challenge token: %v", err)
	}

	sessionSetupStep2Cmd.SecurityBlob = authenticateToken

	requestStep2Msg.AddCommand(sessionSetupStep2Cmd)

	// Determine whether to activate SMB message signing. Signing is activated only
	// when the server requires it; the MAC key is the session key derived from the
	// NTLM exchange above. The AUTHENTICATE request is the first signed message
	// (sequence number 0) and bootstraps signing for the connection.
	signingKey := authCtx.GetSessionKey()
	signingRequired := s.Client.Connection.Server.SigningState == SigningStateRequired
	if signingRequired && len(signingKey) == 0 {
		return fmt.Errorf("server requires SMB signing but no session key was derived (signing requires the NTLMv2/extended-security path)")
	}

	marshalledStep2, err := requestStep2Msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal step 2 message: %v", err)
	}

	if signingRequired {
		s.Client.Connection.SigningSessionKey = signingKey
		s.Client.Connection.IsSigningActive = true
		signSMBMessage(signingKey, marshalledStep2, 0)
	}

	// Send the message
	_, err = s.Client.Transport.Send(marshalledStep2)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Wait for a response message =============================================================================================

	rawAuthResponse, err := s.Client.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	authResponseMsg := message.NewMessage()
	err = authResponseMsg.Unmarshal(rawAuthResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if authResponseMsg.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", authResponseMsg.Header.Command)
	}

	_, ok := authResponseMsg.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		return fmt.Errorf("failed to cast response command to SessionSetupAndxResponse")
	}

	if authResponseMsg.Header.Status != 0x00 {
		if _, ok := nt_status.NTStatusToStringName[nt_status.NT_STATUS(authResponseMsg.Header.Status)]; ok {
			return fmt.Errorf("session setup failed: %s (0x%08x)", nt_status.NTStatusToStringName[nt_status.NT_STATUS(authResponseMsg.Header.Status)], authResponseMsg.Header.Status)
		} else {
			return fmt.Errorf("session setup failed: 0x%08x", authResponseMsg.Header.Status)
		}
	}

	// Persist the derived session key and, when signing was activated, verify the
	// signature on the (successful) AUTHENTICATE response (sequence number 1) before
	// arming the per-message sequence counter for subsequent requests.
	s.SessionKey = signingKey
	if s.Client.Connection.IsSigningActive {
		if !verifySMBSignature(signingKey, rawAuthResponse, 1) {
			return fmt.Errorf("session setup response failed SMB signature verification")
		}
		// The AUTHENTICATE request used sequence 0 and its response sequence 1, so the
		// next outbound request uses sequence 2.
		s.Client.Connection.ClientNextSendSequenceNumber = 2
	}

	return nil
}
