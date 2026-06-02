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

	// Prepare and send a NTLMSSP NEGOTIATE message =============================================================================================
	request_step1_msg := message.NewMessage()
	session_setup_cmd := commands.NewSessionSetupAndxRequest()

	// Here put the common logic for all session setup commands
	request_step1_msg.Header.Command = codes.SMB_COM_SESSION_SETUP_ANDX
	request_step1_msg.Header.Flags = flags.Flags(flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE)
	request_step1_msg.Header.Flags2 = flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_LONG_NAMES_ALLOWED | flags2.FLAGS2_EXTENDED_SECURITY)

	// Add Unicode support if server supports it
	if s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE {
		request_step1_msg.Header.Flags2 |= flags2.Flags2(flags2.FLAGS2_UNICODE)
	}

	// Set message signing flags based on server security mode
	if s.Client.Connection.Server.SecurityMode.IsSecuritySignatureEnabled() {
		request_step1_msg.Header.Flags2 |= flags2.Flags2(flags2.FLAGS2_SECURITY_SIGNATURE)
	}

	// Set process ID and multiplex ID
	request_step1_msg.Header.SetPID(0)
	request_step1_msg.Header.MID = 0
	request_step1_msg.Header.TID = 65535
	request_step1_msg.Header.UID = 0

	session_setup_cmd.MaxBufferSize = types.USHORT(s.Client.Connection.Server.MaxBufferSize)
	session_setup_cmd.MaxMpxCount = s.Client.Connection.MaxMpxCount
	session_setup_cmd.Capabilities = s.Client.Connection.Server.Capabilities

	// if s.Client.Connection.Server.Capabilities&0x00000004 != 0 { // CAP_UNICODE
	// 	session_setup_cmd.NativeOS = string(utf16.EncodeUTF16LE(s.Client.NativeOS))
	// 	session_setup_cmd.NativeLanMan = string(utf16.EncodeUTF16LE(s.Client.NativeLanMan))
	// } else {
	session_setup_cmd.NativeOS = s.Client.NativeOS
	session_setup_cmd.NativeLanMan = s.Client.NativeLanMan
	// }

	// Check if we're using share level access control
	if s.Client.Connection.Server.SecurityMode.SupportsShareLevelAccessControl() {
		// Share level access control is required by the server
		// If no authentication has been performed on the SMB connection, use anonymous authentication

		// Parameters
		session_setup_cmd.VcNumber = types.USHORT(0x0000)
		session_setup_cmd.SessionKey = s.Client.Connection.Server.SessionKey
		session_setup_cmd.OEMPasswordLen = types.USHORT(0x0000)
		session_setup_cmd.UnicodePasswordLen = types.USHORT(0x0000)

		// Data section - for null session, use empty strings
		session_setup_cmd.OEMPassword = []types.UCHAR{}
		session_setup_cmd.UnicodePassword = []types.UCHAR{}

	} else {
		// User level access control is required by the server
		// TODO: Look up Session from Client.Connection.SessionTable where Session.UserCredentials matches
		// the application-supplied UserCredentials and reuse if found

		// Handle authentication based on server capabilities
		if s.Client.Connection.Server.SecurityMode.SupportsChallengeResponseAuth() {
			// Server supports challenge/response authentication
			// Determine authentication type based on policies

			session_setup_cmd.VcNumber = types.USHORT(0x0000)
			session_setup_cmd.SessionKey = s.Client.Connection.Server.SessionKey

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
					spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
					// spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
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
			session_setup_cmd.SecurityBlob = negotiateToken
		} else {
			// Server doesn't support challenge/response authentication

			// Use plaintext authentication
			session_setup_cmd.VcNumber = types.USHORT(0x0000)
			session_setup_cmd.SessionKey = s.Client.Connection.Server.SessionKey

			if s.Credentials == nil {
				return fmt.Errorf("plaintext authentication requires credentials but none were provided")
			}
			password := s.Credentials.Password

			// Check if Unicode is supported
			if s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE {
				// Send password in Unicode
				session_setup_cmd.UnicodePassword = []types.UCHAR(utf16.EncodeUTF16LE(password))
				session_setup_cmd.UnicodePasswordLen = types.USHORT(len(session_setup_cmd.UnicodePassword))
				session_setup_cmd.OEMPasswordLen = types.USHORT(0x0000)
				session_setup_cmd.OEMPassword = []types.UCHAR{}
			} else {
				// Send password in OEM format
				session_setup_cmd.OEMPassword = []types.UCHAR(password)
				session_setup_cmd.OEMPasswordLen = types.USHORT(len(session_setup_cmd.OEMPassword))
				session_setup_cmd.UnicodePasswordLen = types.USHORT(0x0000)
				session_setup_cmd.UnicodePassword = []types.UCHAR{}
			}
		}
	}
	request_step1_msg.AddCommand(session_setup_cmd)

	marshalled_message, err := request_step1_msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal negotiate message: %v", err)
	}

	// Send the message
	_, err = s.Client.Transport.Send(marshalled_message)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Wait for a NTLMSSP CHALLENGE response message =============================================================================================

	raw_response_message, err := s.Client.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	response_msg := message.NewMessage()
	err = response_msg.Unmarshal(raw_response_message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if response_msg.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", response_msg.Header.Command)
	}

	session_setup_response_challenge := response_msg.Command.(*commands.SessionSetupAndxResponse)

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

	request_step2_msg := message.NewMessage()
	request_step2_msg.Header.Command = codes.SMB_COM_SESSION_SETUP_ANDX
	request_step2_msg.Header.Flags = request_step1_msg.Header.Flags
	request_step2_msg.Header.Flags2 = request_step1_msg.Header.Flags2
	request_step2_msg.Header.SetPID(request_step1_msg.Header.GetPID())
	request_step2_msg.Header.MID = request_step1_msg.Header.MID
	request_step2_msg.Header.TID = request_step1_msg.Header.TID
	// Here we need to set the UID to the UID of the response message
	request_step2_msg.Header.UID = response_msg.Header.UID
	// Save the session UID
	s.SessionUID = request_step2_msg.Header.UID

	session_setup_step2_cmd := commands.NewSessionSetupAndxRequest()
	session_setup_step2_cmd.VcNumber = session_setup_cmd.VcNumber
	session_setup_step2_cmd.SessionKey = session_setup_cmd.SessionKey
	session_setup_step2_cmd.Capabilities = session_setup_cmd.Capabilities

	authenticateToken, err := authCtx.CreateAuthenticateTokenFromChallengeToken(session_setup_response_challenge.SecurityBlob)
	if err != nil {
		return fmt.Errorf("failed to process challenge token: %v", err)
	}

	session_setup_step2_cmd.SecurityBlob = authenticateToken

	request_step2_msg.AddCommand(session_setup_step2_cmd)

	marshalled_message_step2, err := request_step2_msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal step 2 message: %v", err)
	}

	// Send the message
	_, err = s.Client.Transport.Send(marshalled_message_step2)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Wait for a response message =============================================================================================

	raw_response_message_step4, err := s.Client.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	response_msg_step4 := message.NewMessage()
	err = response_msg_step4.Unmarshal(raw_response_message_step4)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if response_msg_step4.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", response_msg_step4.Header.Command)
	}

	_, ok := response_msg_step4.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		return fmt.Errorf("failed to cast response command to SessionSetupAndxResponse")
	}

	if response_msg_step4.Header.Status != 0x00 {
		if _, ok := nt_status.NTStatusToStringName[nt_status.NT_STATUS(response_msg_step4.Header.Status)]; ok {
			return fmt.Errorf("session setup failed: %s (0x%08x)", nt_status.NTStatusToStringName[nt_status.NT_STATUS(response_msg_step4.Header.Status)], response_msg_step4.Header.Status)
		} else {
			return fmt.Errorf("session setup failed: 0x%08x", response_msg_step4.Header.Status)
		}
	}

	return nil
}
