package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// SessionSetupKerberos authenticates a session with the server using Kerberos
// over SPNEGO. Unlike the two-leg NTLM exchange, Kerberos is a single round: the
// client sends a KRB_AP_REQ inside a SPNEGO NegTokenInit and the server replies
// with STATUS_SUCCESS, optionally carrying a KRB_AP_REP (mutual authentication)
// which we verify to capture the negotiated context key.
//
// kdcHost is the KDC (domain controller) address used to obtain the TGT and
// service ticket. spn is the service principal name of the target server; when
// empty it defaults to "cifs/<host>", the SMB service class.
//
// The GSS context key established here is used as the SMB signing key (for the
// SMB 2.0.2/2.1 dialects the signing key is the session key itself). The
// NegTokenInit request is not signed — the session is not yet established;
// signing is activated once the server confirms the session.
func (c *Client) SessionSetupKerberos(creds *credentials.Credentials, kdcHost, spn string) error {
	if creds == nil {
		return fmt.Errorf("session setup requires credentials but none were provided")
	}
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}
	if kdcHost == "" {
		return fmt.Errorf("kerberos session setup requires a KDC host")
	}
	if spn == "" {
		spn = "cifs/" + c.Connection.Server.DNSComputerName
	}
	if spn == "cifs/" {
		return fmt.Errorf("kerberos session setup requires a service principal name (or a known server name)")
	}

	// Build the native Kerberos client from the SMB credentials (realm = domain).
	kc := kerberos.NewClient(creds.GetUsername(), creds.GetDomain(), kdcHost)
	if nt := creds.GetNTHash(); nt != "" {
		// Overpass-the-hash: obtain the TGT from the NT hash rather than a password.
		if err := kc.WithNTHash(nt); err != nil {
			return fmt.Errorf("kerberos: invalid NT hash: %w", err)
		}
	} else {
		kc.WithPassword(creds.GetPassword())
	}

	return c.sessionSetupKerberosMechanism(kerberos.NewSPNEGOMechanism(kc, spn), creds)
}

// SessionSetupKerberosWithClient authenticates the session using a caller-built
// native Kerberos client, targeting service principal spn. Unlike
// SessionSetupKerberos it does not construct the client from a password/NT hash,
// so it drives pass-the-ticket and forged-ticket flows: a golden TGT imported
// with LoadTGT (the client then obtains the service ticket from the KDC) or a
// forged silver service ticket wired in with LoadServiceTicket (used directly,
// with no KDC round-trip). The client already carries the KDC host and identity.
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
	return c.sessionSetupKerberosMechanism(kerberos.NewSPNEGOMechanism(kc, spn), creds)
}

// sessionSetupKerberosMechanism runs the single-leg SPNEGO/Kerberos SESSION_SETUP
// exchange for a prepared mechanism, shared by SessionSetupKerberos (client built
// from credentials) and SessionSetupKerberosWithClient (caller-supplied client).
func (c *Client) sessionSetupKerberosMechanism(mech *kerberos.SPNEGOMechanism, creds *credentials.Credentials) error {
	// SMB2 always uses Unicode for its strings.
	authCtx := spnego.NewAuthContext(
		spnego.AuthTypeKerberos,
		creds.Domain,
		creds.Username,
		creds.Password,
		c.Workstation,
		true,
	)
	// Wire the Kerberos provider into the SPNEGO context (dependency inversion:
	// crypto/spnego stays free of any Kerberos dependency).
	authCtx.Kerberos = mech

	// Build the SPNEGO NegTokenInit carrying the KRB_AP_REQ.
	negotiateToken, err := authCtx.CreateNegotiateToken(0, nil)
	if err != nil {
		return fmt.Errorf("failed to create Kerberos negotiate token: %w", err)
	}

	// Single leg: send the AP-REQ; the server replies with the assigned SessionId
	// and STATUS_SUCCESS, optionally carrying an AP-REP.
	req := commands.NewSessionSetupRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.SecurityBuffer = negotiateToken

	msg := c.newRequest(req)
	msg.Header.SessionId = 0

	resp, err := c.sendReceive(msg, "SessionSetup(kerberos)")
	if err != nil {
		return err
	}

	status := statusFromResponse(resp)
	// STATUS_MORE_PROCESSING_REQUIRED can precede the AP-REP on some servers; both
	// it and STATUS_SUCCESS carry the response token we need to verify.
	if status != 0x00000000 && status != ntStatusMoreProcessingRequired {
		return fmt.Errorf("kerberos session setup failed: %s", formatNTStatus(status))
	}

	sessionId := resp.Header.SessionId
	setupResp, ok := resp.Command.(*commands.SessionSetupResponse)
	if !ok {
		return fmt.Errorf("unexpected SESSION_SETUP response command: %T", resp.Command)
	}

	// Verify the server's AP-REP (mutual authentication) when present and capture
	// the GSS context key. For Kerberos this produces no follow-up token.
	if len(setupResp.SecurityBuffer) > 0 {
		follow, ferr := authCtx.CreateAuthenticateTokenFromChallengeToken(setupResp.SecurityBuffer)
		if ferr != nil {
			return fmt.Errorf("failed to verify Kerberos server response: %w", ferr)
		}
		if len(follow) > 0 {
			// A well-behaved AD server completes Kerberos in one round; a follow-up
			// token would indicate an unexpected multi-leg negotiation.
			return fmt.Errorf("kerberos session setup: unexpected multi-leg SPNEGO continuation")
		}
	}

	// The signing key derives from the established GSS context key (subkey if
	// negotiated, else the service-ticket session key). Per MS-SMB2 3.2.5.3.1 the
	// SMB2 Session.SessionKey is the first 16 bytes of the GSS-exported key,
	// right-padded with zeros if shorter — Kerberos AES256 yields a 32-byte key,
	// so it must be truncated (NTLM keys are already 16 bytes). For the 2.0.2
	// dialect this 16-byte value is the HMAC-SHA256 signing key directly.
	gssKey := authCtx.GetSessionKey()
	if len(gssKey) == 0 {
		gssKey = mech.SessionKey()
	}
	sessionKey := make([]byte, 16)
	copy(sessionKey, gssKey)

	serverMode := c.Connection.Server.SecurityMode
	session := &Session{
		Client:        c,
		SessionId:     sessionId,
		Credentials:   creds,
		SessionKey:    sessionKey,
		SigningKey:    sessionKey,
		SigningActive: false,
	}
	c.Session = session

	// If the server still requires another leg (unusual for AD Kerberos), send the
	// final NegTokenResp confirming completion before the session is usable.
	if status == ntStatusMoreProcessingRequired {
		req2 := commands.NewSessionSetupRequest()
		req2.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
		req2.SecurityBuffer = []byte{}

		msg2 := c.newRequest(req2)
		msg2.Header.SessionId = sessionId

		resp2, err := c.sendReceive(msg2, "SessionSetup(kerberos-final)")
		if err != nil {
			c.Session = nil
			return err
		}
		if st := statusFromResponse(resp2); st != 0x00000000 {
			c.Session = nil
			return fmt.Errorf("kerberos session setup failed: %s", formatNTStatus(st))
		}
	}

	// Session established: activate signing for subsequent requests when the server
	// enables or requires it.
	session.SigningActive = serverMode.IsSigningEnabled() || serverMode.IsSigningRequired()

	if c.Connection.SessionTable == nil {
		c.Connection.SessionTable = make(map[uint64]*Session)
	}
	c.Connection.SessionTable[sessionId] = session

	return nil
}
