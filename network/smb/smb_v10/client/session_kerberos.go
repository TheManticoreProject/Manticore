package client

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// SessionSetupKerberos authenticates an SMB1 session using Kerberos over SPNEGO.
// Unlike the two-leg NTLMSSP exchange, Kerberos is a single round: the client
// sends a KRB_AP_REQ inside a SPNEGO NegTokenInit in one SESSION_SETUP_ANDX and
// the server replies with STATUS_SUCCESS, optionally carrying a KRB_AP_REP
// (mutual authentication) which is verified to capture the negotiated key.
//
// kdcHost is the KDC (domain controller) address used to obtain the TGT and
// service ticket. spn is the target server's service principal name; when empty
// it defaults to "cifs/<DNSComputerName>". Mirrors the SMB2 implementation in
// network/smb/smb_v20/client/session_kerberos.go, adapted to SMB1 framing and
// (session-key, non-SMB3) signing.
func (c *Client) SessionSetupKerberos(creds *credentials.Credentials, kdcHost, spn string) error {
	if creds == nil {
		return fmt.Errorf("session setup requires credentials but none were provided")
	}
	if kdcHost == "" {
		return fmt.Errorf("kerberos session setup requires a KDC host")
	}
	kc := kerberos.NewClient(creds.GetUsername(), strings.ToUpper(creds.GetDomain()), kdcHost)
	if nt := creds.GetNTHash(); nt != "" {
		// Overpass-the-hash: obtain the TGT from the NT hash rather than a password.
		if err := kc.WithNTHash(nt); err != nil {
			return fmt.Errorf("kerberos: invalid NT hash: %w", err)
		}
	} else {
		kc.WithPassword(creds.GetPassword())
	}
	return c.SessionSetupKerberosWithClient(kc, spn)
}

// SessionSetupKerberosWithClient authenticates the session using a caller-built
// Kerberos client, targeting service principal spn. It drives pass-the-ticket
// and forged-ticket flows (ccache/kirbi TGT, or a silver service ticket loaded
// with LoadServiceTicket); the client already carries the KDC host and identity.
func (c *Client) SessionSetupKerberosWithClient(kc *kerberos.KerberosClient, spn string) error {
	if kc == nil {
		return fmt.Errorf("session setup requires a Kerberos client but none was provided")
	}
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}
	if spn == "" {
		spn = "cifs/" + c.Connection.Server.DNSComputerName
	}
	if spn == "cifs/" {
		return fmt.Errorf("kerberos session setup requires a service principal name (or a known server name)")
	}

	// Synthesize the credential record the session stores from the client's own
	// identity; the password is unused on the Kerberos path (the mechanism drives
	// the exchange).
	creds, err := credentials.NewCredentials(kc.Realm(), kc.Username(), "", "")
	if err != nil {
		return fmt.Errorf("kerberos session setup: %w", err)
	}

	session := &Session{Client: c, Credentials: creds}
	if err := session.sessionSetupKerberos(kerberos.NewSPNEGOMechanism(kc, spn)); err != nil {
		return err
	}
	c.Session = session

	if c.Connection != nil {
		if c.Connection.SessionTable == nil {
			c.Connection.SessionTable = make(map[uint16]*Session)
		}
		c.Connection.SessionTable[c.Session.SessionUID] = c.Session
	}
	return nil
}

// sessionSetupKerberos runs the single-leg SPNEGO/Kerberos SESSION_SETUP_ANDX
// exchange for a prepared mechanism.
func (s *Session) sessionSetupKerberos(mech *kerberos.SPNEGOMechanism) error {
	if !s.Client.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	useUnicode := s.Client.Connection.Server.Capabilities&capabilities.CAP_UNICODE == capabilities.CAP_UNICODE

	authCtx := spnego.NewAuthContext(
		spnego.AuthTypeKerberos,
		s.Credentials.Domain,
		s.Credentials.Username,
		s.Credentials.Password,
		s.Client.Workstation,
		useUnicode,
	)
	// Wire the Kerberos provider into the SPNEGO context (dependency inversion:
	// crypto/spnego stays free of any Kerberos dependency).
	authCtx.Kerberos = mech

	// Build the SPNEGO NegTokenInit carrying the KRB_AP_REQ. This acquires the
	// TGT/service ticket from the KDC via the mechanism.
	negotiateToken, err := authCtx.CreateNegotiateToken(0, nil)
	if err != nil {
		return fmt.Errorf("failed to create Kerberos negotiate token: %w", err)
	}

	// Build the single SESSION_SETUP_ANDX leg carrying the AP-REQ, with the same
	// extended-security header framing the NTLM path uses.
	reqMsg := message.NewMessage()
	cmd := commands.NewSessionSetupAndxRequest()

	reqMsg.Header.Command = codes.SMB_COM_SESSION_SETUP_ANDX
	reqMsg.Header.Flags = flags.Flags(flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE)
	reqMsg.Header.Flags2 = flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_LONG_NAMES_ALLOWED | flags2.FLAGS2_EXTENDED_SECURITY)
	if useUnicode {
		reqMsg.Header.Flags2 |= flags2.Flags2(flags2.FLAGS2_UNICODE)
	}
	if s.Client.Connection.Server.SecurityMode.IsSecuritySignatureEnabled() ||
		s.Client.Connection.Server.SecurityMode.IsSecuritySignatureRequired() {
		reqMsg.Header.Flags2 |= flags2.Flags2(flags2.FLAGS2_SECURITY_SIGNATURE)
	}
	reqMsg.Header.SetPID(clientProcessID)
	reqMsg.Header.MID = s.Client.Connection.nextMID()
	reqMsg.Header.TID = 65535
	reqMsg.Header.UID = 0

	maxBufferSize := s.Client.MaxBufferSize
	if maxBufferSize == 0 {
		maxBufferSize = DefaultMaxBufferSize
	}
	cmd.MaxBufferSize = types.USHORT(maxBufferSize)
	cmd.MaxMpxCount = s.Client.Connection.MaxMpxCount
	cmd.Capabilities = s.Client.Connection.Server.Capabilities
	cmd.VcNumber = types.USHORT(0x0000)
	cmd.SessionKey = s.Client.Connection.Server.SessionKey
	cmd.NativeOS = s.Client.NativeOS
	cmd.NativeLanMan = s.Client.NativeLanMan
	if cmd.NativeOS == "" {
		cmd.NativeOS = DefaultNativeOS
	}
	if cmd.NativeLanMan == "" {
		cmd.NativeLanMan = DefaultNativeLanMan
	}
	cmd.SecurityBlob = negotiateToken
	reqMsg.AddCommand(cmd)

	marshalled, err := reqMsg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal Kerberos session setup: %w", err)
	}

	// Signing bootstrap. Unlike NTLM (whose NEGOTIATE leg carries no key, so only
	// the AUTHENTICATE leg can be signed), the Kerberos GSS context key is known
	// before this single leg is sent, so the AP-REQ request is itself the first
	// signed message (sequence 0) when the server enables or requires signing —
	// mirroring the SMB1 NTLM AUTHENTICATE bootstrap (session.go). The response is
	// then verified at sequence 1 and subsequent requests continue from sequence 2.
	//
	// NOTE: this signing path has not been validated against a live
	// signing-required server; it reproduces the NTLM sequence bootstrap and is
	// the primary thing to confirm in a lab. Non-signing servers skip it entirely.
	signingAlreadyActive := s.Client.Connection.IsSigningActive
	signingKey := authCtx.GetSessionKey()
	if len(signingKey) == 0 {
		signingKey = mech.SessionKey()
	}
	signingState := s.Client.Connection.Server.SigningState
	if !signingAlreadyActive && signingState == SigningStateRequired && len(signingKey) == 0 {
		return fmt.Errorf("server requires SMB signing but no Kerberos session key was derived")
	}
	signMessages := !signingAlreadyActive && signingState != SigningStateDisabled && len(signingKey) != 0

	// Arm signing before sending; restore the connection's previous signing state
	// if the setup does not complete.
	authenticated := false
	if signMessages {
		previousKey := s.Client.Connection.SigningSessionKey
		previousActive := s.Client.Connection.IsSigningActive
		defer func() {
			if !authenticated {
				s.Client.Connection.SigningSessionKey = previousKey
				s.Client.Connection.IsSigningActive = previousActive
			}
		}()
		s.Client.Connection.SigningSessionKey = signingKey
		s.Client.Connection.IsSigningActive = true
		signing.Sign(signingKey, marshalled, 0)
	}
	respSequence, respSigned := uint32(0), false
	if signingAlreadyActive {
		respSequence, respSigned = s.Client.signOutgoing(marshalled)
	}

	if _, err := s.Client.Transport.Send(marshalled); err != nil {
		return fmt.Errorf("failed to send Kerberos session setup: %w", err)
	}

	raw, err := s.Client.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive Kerberos session setup response: %w", err)
	}
	if respSigned {
		if err := s.Client.verifyIncoming(raw, respSequence); err != nil {
			return fmt.Errorf("kerberos session setup response: %w", err)
		}
	}

	respMsg := message.NewMessage()
	if err := respMsg.Unmarshal(raw); err != nil {
		return fmt.Errorf("failed to unmarshal Kerberos session setup response: %w", err)
	}
	if respMsg.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", respMsg.Header.Command)
	}
	resp, ok := respMsg.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		return fmt.Errorf("failed to cast response command to SessionSetupAndxResponse")
	}

	// Kerberos is single-leg: the server returns STATUS_SUCCESS directly (not the
	// STATUS_MORE_PROCESSING_REQUIRED that the NTLM challenge leg returns).
	if respMsg.Header.Status != 0x00000000 {
		if name := nt_status.NT_STATUS(respMsg.Header.Status).Name(); name != "" {
			return fmt.Errorf("kerberos session setup failed: %s (0x%08x)", name, respMsg.Header.Status)
		}
		return fmt.Errorf("kerberos session setup failed: 0x%08x", respMsg.Header.Status)
	}

	s.SessionUID = respMsg.Header.UID

	// Verify the server's AP-REP (mutual authentication) when present and capture
	// the GSS context key. For Kerberos this produces no follow-up token.
	if len(resp.SecurityBlob) > 0 {
		follow, ferr := authCtx.CreateAuthenticateTokenFromChallengeToken(resp.SecurityBlob)
		if ferr != nil {
			return fmt.Errorf("failed to verify Kerberos server response: %w", ferr)
		}
		if len(follow) > 0 {
			return fmt.Errorf("kerberos session setup: unexpected multi-leg SPNEGO continuation")
		}
	}

	if osName := resp.NativeOSString(); osName != "" {
		s.Client.Connection.Server.OSName = osName
	}

	s.SessionKey = signingKey
	if signMessages {
		if !signing.Verify(signingKey, raw, 1) {
			return fmt.Errorf("kerberos session setup response failed SMB signature verification")
		}
		// The AP-REQ request used sequence 0 and its response sequence 1, so the
		// next outbound request uses sequence 2.
		s.Client.Connection.ClientNextSendSequenceNumber = signing.NextRequestSequenceNumber(0)
	}

	authenticated = true
	return nil
}
