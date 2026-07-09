package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	spnego_ntlm_negotiate_flags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// ntStatusMoreProcessingRequired (STATUS_MORE_PROCESSING_REQUIRED) is the status
// the server returns on the first SESSION_SETUP response to request the second
// (AUTHENTICATE) leg of the NTLM exchange.
const ntStatusMoreProcessingRequired = 0xC0000016

// ntStatusBufferOverflow (STATUS_BUFFER_OVERFLOW) is a warning, not an error: the
// server returns it from READ and from an IOCTL pipe transceive together with the
// partial data that did fit, signalling that more remains to be read.
const ntStatusBufferOverflow = 0x80000005

// ntStatusPending (STATUS_PENDING) is the interim status the server returns
// (with SMB2_FLAGS_ASYNC_COMMAND set) for a request it cannot complete
// immediately — a blocking lock, CHANGE_NOTIFY, a long IOCTL. The final response
// follows once the operation completes.
const ntStatusPending = 0x00000103

// ntStatusCancelled (STATUS_CANCELLED) is the final status of a request that was
// cancelled via SMB2 CANCEL (e.g. a blocked CHANGE_NOTIFY).
const ntStatusCancelled = 0xC0000120

// ntStatusNotifyEnumDir (STATUS_NOTIFY_ENUM_DIR) is returned from CHANGE_NOTIFY
// when too many changes occurred to report individually; the client should
// re-enumerate the directory. The output buffer is empty.
const ntStatusNotifyEnumDir = 0x0000010C

// SessionSetup authenticates a session with the server using NTLM over SPNEGO,
// the SMB2 analog of the SMB 1.0 extended-security session setup. It performs the
// two-leg NEGOTIATE -> CHALLENGE -> AUTHENTICATE exchange, capturing the
// server-assigned SessionId from the first response and the derived session key.
//
// The AUTHENTICATE request itself is not signed (the session is not yet
// established); once the server confirms the session, signing is activated so
// every subsequent request is signed. Live-validated against Windows Server 2016
// with signing required.
func (c *Client) SessionSetup(creds *credentials.Credentials) error {
	if creds == nil {
		return fmt.Errorf("session setup requires credentials but none were provided")
	}
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	// SMB2 always uses Unicode for its strings.
	authCtx := spnego.NewAuthContext(
		spnego.AuthTypeNTLM,
		creds.Domain,
		creds.Username,
		creds.Password,
		c.Workstation,
		true,
	)
	// Pass-the-hash: when an NT hash is supplied, authenticate from it instead of the password.
	authCtx.NTHash = creds.GetNTHash()

	negotiateFlags := spnego_ntlm_negotiate_flags.NegotiateFlags(
		spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_UNICODE |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_NTLM |
			// Extended session security (NTLM2) so the AUTHENTICATE derives a
			// SessionBaseKey, which Phase 6 signing will use.
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_SIGN |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_128 |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_56 |
			spnego_ntlm_negotiate_flags.NTLMSSP_REQUEST_TARGET |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_VERSION,
	)
	v := &version.Version{
		ProductMajorVersion: 6,
		ProductMinorVersion: 1,
		ProductBuild:        0,
		Reserved:            [3]byte{0, 0, 0},
		NTLMRevision:        version.NTLMSSP_REVISION_W2K3,
	}

	negotiateToken, err := authCtx.CreateNegotiateToken(negotiateFlags, v)
	if err != nil {
		return fmt.Errorf("failed to create NTLM negotiate token: %w", err)
	}

	// Leg 1: send the NTLM NEGOTIATE token; the server replies with the assigned
	// SessionId (in the header) and a CHALLENGE token, with status
	// STATUS_MORE_PROCESSING_REQUIRED.
	req1 := commands.NewSessionSetupRequest()
	req1.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req1.SecurityBuffer = negotiateToken

	msg1 := c.newRequest(req1)
	msg1.Header.SessionId = 0

	resp1, err := c.sendReceive(msg1, "SessionSetup(negotiate)")
	if err != nil {
		return err
	}
	if status := statusFromResponse(resp1); status != ntStatusMoreProcessingRequired {
		return fmt.Errorf("unexpected SESSION_SETUP status: %s", formatNTStatus(status))
	}

	// Fold the first SESSION_SETUP request/response into the SMB 3.1.1 pre-auth
	// integrity hash (started from the connection hash after NEGOTIATE).
	sessionHash := append([]byte(nil), c.Connection.PreauthIntegrityHashValue...)
	sessionHash = preauthUpdate(sessionHash, c.lastSentBytes)
	sessionHash = preauthUpdate(sessionHash, c.lastRecvBytes)

	sessionId := resp1.Header.SessionId
	challengeResp, ok := resp1.Command.(*commands.SessionSetupResponse)
	if !ok {
		return fmt.Errorf("unexpected SESSION_SETUP response command: %T", resp1.Command)
	}

	// Leg 2: send the NTLM AUTHENTICATE token computed from the challenge, with the
	// assigned SessionId set in the header.
	authenticateToken, err := authCtx.CreateAuthenticateTokenFromChallengeToken(challengeResp.SecurityBuffer)
	if err != nil {
		return fmt.Errorf("failed to create NTLM authenticate token: %w", err)
	}

	// Install the session (with its signing key) before sending the final leg, so
	// the engine signs the AUTHENTICATE request when signing is in effect. For the
	// SMB 2.0.2/2.1 dialects the signing key is the session key itself. Signing is
	// activated when the server enables or requires it.
	sessionKey := authCtx.GetSessionKey()
	serverMode := c.Connection.Server.SecurityMode
	session := &Session{
		Client:      c,
		SessionId:   sessionId,
		Credentials: creds,
		SessionKey:  sessionKey,
		SigningKey:  sessionKey,
		// The AUTHENTICATE request itself is NOT signed — the session is not yet
		// established. Signing is activated below, after the server confirms the
		// session, so it applies to subsequent requests (tree connect, file I/O).
		SigningActive: false,
	}
	c.Session = session

	req2 := commands.NewSessionSetupRequest()
	req2.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req2.SecurityBuffer = authenticateToken

	msg2 := c.newRequest(req2)
	msg2.Header.SessionId = sessionId

	resp2, err := c.sendReceive(msg2, "SessionSetup(authenticate)")
	if err != nil {
		c.Session = nil
		return err
	}
	if status := statusFromResponse(resp2); status != 0x00000000 {
		c.Session = nil
		return fmt.Errorf("session setup failed: %s", formatNTStatus(status))
	}

	// SMB 3.x: fold the (unsigned) AUTHENTICATE request into the pre-auth hash —
	// the final SUCCESS response is excluded — then derive the key hierarchy and
	// verify the server's signature over that response.
	if isSMB3Dialect(c.Connection.Dialect) {
		sessionHash = preauthUpdate(sessionHash, c.lastSentBytes)
		finalRespBytes := append([]byte(nil), c.lastRecvBytes...)
		session.PreauthHash = sessionHash
		deriveSMB3Keys(session, c.Connection.Dialect, sessionHash)

		if len(finalRespBytes) >= 64 && !verifySignatureForDialect(c.Connection.Dialect, session.SigningKey, finalRespBytes) {
			c.Session = nil
			return fmt.Errorf("session setup: SMB3 signature of final SESSION_SETUP response did not verify (derived signing key mismatch)")
		}

		if resp2Cmd, ok := resp2.Command.(*commands.SessionSetupResponse); ok &&
			resp2Cmd.SessionFlags&commands.SMB2_SESSION_FLAG_ENCRYPT_DATA != 0 {
			session.EncryptData = true
		}
	}

	// Session established: activate signing for all subsequent requests when the
	// server enables or requires it.
	session.SigningActive = serverMode.IsSigningEnabled() || serverMode.IsSigningRequired()

	// Retain the server identity (NetBIOS/DNS names, OS version) advertised in the
	// NTLM CHALLENGE so callers can read it after authentication.
	if id, ok := authCtx.ServerIdentity(); ok {
		srv := c.Connection.Server
		srv.NetBIOSComputerName = id.NetBIOSComputerName
		srv.NetBIOSDomainName = id.NetBIOSDomainName
		srv.DNSComputerName = id.DNSComputerName
		srv.DNSDomainName = id.DNSDomainName
		srv.OSVersionMajor = id.OSVersionMajor
		srv.OSVersionMinor = id.OSVersionMinor
		srv.OSVersionBuild = id.OSVersionBuild
	}
	c.Connection.Server.SupportsNTLMv2 = authCtx.SupportsExtendedSessionSecurity()

	if c.Connection.SessionTable == nil {
		c.Connection.SessionTable = make(map[uint64]*Session)
	}
	c.Connection.SessionTable[sessionId] = session

	return nil
}

// formatNTStatus renders an NT status code with its symbolic name when known.
func formatNTStatus(status uint32) string {
	if name, ok := nt_status.NTStatusToStringName[nt_status.NT_STATUS(status)]; ok {
		return fmt.Sprintf("%s (0x%08x)", name, status)
	}
	return fmt.Sprintf("0x%08x", status)
}
