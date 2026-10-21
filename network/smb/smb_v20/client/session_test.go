package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	ntlmflags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestSessionSetupNilCredentials(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	if err := c.SessionSetup(nil); err == nil {
		t.Errorf("expected SessionSetup(nil) to error, got nil")
	}
}

func TestFormatNTStatus(t *testing.T) {
	// A success code formats as a bare hex value.
	if got := formatNTStatus(0x00000000); got == "" {
		t.Errorf("formatNTStatus returned empty string")
	}
	// MORE_PROCESSING_REQUIRED should at least include its hex form.
	got := formatNTStatus(ntStatusMoreProcessingRequired)
	if got == "" {
		t.Errorf("formatNTStatus returned empty string for MORE_PROCESSING_REQUIRED")
	}
}

func TestSessionGuestFlagDisablesSigning(t *testing.T) {
	session := &Session{
		SessionKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningActive: true,
	}

	// Simulate what SessionSetup does for guest sessions.
	sessionFlags := uint16(commands.SMB2_SESSION_FLAG_IS_GUEST)
	session.IsGuest = sessionFlags&commands.SMB2_SESSION_FLAG_IS_GUEST != 0

	if !session.IsGuest {
		t.Fatal("IsGuest should be true")
	}

	// Per MS-SMB2 3.2.5.3.1, guest sessions disable signing.
	if session.IsGuest {
		session.SessionKey = make([]byte, 16)
		session.SigningKey = session.SessionKey
		session.SigningActive = false
	}

	if session.SigningActive {
		t.Error("signing should be disabled for guest sessions")
	}
	for _, b := range session.SessionKey {
		if b != 0 {
			t.Error("session key should be zeroed for guest sessions")
			break
		}
	}
}

func TestSessionNullFlagDisablesSigning(t *testing.T) {
	session := &Session{
		SessionKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningActive: true,
	}

	sessionFlags := uint16(commands.SMB2_SESSION_FLAG_IS_NULL)
	session.IsNull = sessionFlags&commands.SMB2_SESSION_FLAG_IS_NULL != 0

	if !session.IsNull {
		t.Fatal("IsNull should be true")
	}

	if session.IsNull {
		session.SessionKey = make([]byte, 16)
		session.SigningKey = session.SessionKey
		session.SigningActive = false
	}

	if session.SigningActive {
		t.Error("signing should be disabled for anonymous sessions")
	}
}

func TestSessionNormalFlagPreservesSigning(t *testing.T) {
	session := &Session{
		SessionKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningKey:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		SigningActive: true,
	}

	sessionFlags := uint16(commands.SMB2_SESSION_FLAG_ENCRYPT_DATA)
	session.IsGuest = sessionFlags&commands.SMB2_SESSION_FLAG_IS_GUEST != 0
	session.IsNull = sessionFlags&commands.SMB2_SESSION_FLAG_IS_NULL != 0

	if session.IsGuest || session.IsNull {
		t.Error("neither guest nor null should be set for an encrypted session")
	}
	if !session.SigningActive {
		t.Error("signing should remain active for normal sessions")
	}
	if session.SessionKey[0] != 0x01 {
		t.Error("session key should be preserved for normal sessions")
	}
}

// anonymousSessionSetupResponses scripts the two SESSION_SETUP responses of an
// NTLM exchange: a CHALLENGE (STATUS_MORE_PROCESSING_REQUIRED) produced by the
// in-tree NTLM acceptor, then a final STATUS_SUCCESS carrying sessionFlags and no
// signature.
func anonymousSessionSetupResponses(t *testing.T, sessionFlags uint16) [][]byte {
	t.Helper()

	initiator := spnego.NewAuthContext(spnego.AuthTypeNTLM, "", "", "", "", true)
	v := version.DefaultVersion()
	negotiateToken, err := initiator.CreateNegotiateToken(
		ntlmflags.NTLMSSP_NEGOTIATE_UNICODE|ntlmflags.NTLMSSP_NEGOTIATE_NTLM|
			ntlmflags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY|ntlmflags.NTLMSSP_NEGOTIATE_KEY_EXCH|
			ntlmflags.NTLMSSP_NEGOTIATE_SIGN|ntlmflags.NTLMSSP_REQUEST_TARGET, &v)
	if err != nil {
		t.Fatalf("CreateNegotiateToken: %v", err)
	}
	acceptor := spnego.NewAcceptContext("SAMBA", challenge.TargetTypeServer, nil)
	challengeToken, err := acceptor.AcceptNegotiateToken(negotiateToken)
	if err != nil {
		t.Fatalf("AcceptNegotiateToken: %v", err)
	}

	challengeResp := commands.NewSessionSetupResponse()
	challengeResp.SecurityBuffer = challengeToken
	finalResp := commands.NewSessionSetupResponse()
	finalResp.SessionFlags = sessionFlags

	return [][]byte{
		cannedResponse(t, challengeResp, ntStatusMoreProcessingRequired, 0, 0x4242),
		cannedResponse(t, finalResp, 0, 0, 0x4242),
	}
}

// TestSessionSetupAnonymousWithoutNullFlag reproduces an anonymous login to Samba
// over SMB 3.1.1. Samba maps anonymous to guest internally, returns neither
// SMB2_SESSION_FLAG_IS_NULL nor IS_GUEST, and does not sign the final response.
// Per MS-SMB2 3.2.5.3.1 the client knows from its own security subsystem that the
// session is anonymous, so it MUST NOT require (or verify) a signature.
func TestSessionSetupAnonymousWithoutNullFlag(t *testing.T) {
	ft := &fakeTransport{responses: anonymousSessionSetupResponses(t, 0)}
	c := newTestClient(ft)
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
	c.Connection.Server.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED

	if err := c.SessionSetup(&credentials.Credentials{}); err != nil {
		t.Fatalf("anonymous SessionSetup: %v", err)
	}
	if !c.Session.IsAnonymous {
		t.Error("Session.IsAnonymous = false, want true")
	}
	if c.Session.IsNull || c.Session.IsGuest {
		t.Errorf("IsNull=%v IsGuest=%v, want both false (server set no flag)", c.Session.IsNull, c.Session.IsGuest)
	}
	if c.Session.SigningActive {
		t.Error("signing is active on an anonymous session")
	}
}

// TestSessionSetupNamedUserStillVerifiesSignature guards the other side of the
// rule: a named user's SMB 3.1.1 session with no guest/null flag must still have
// its final response verified, so an unsigned one is rejected.
func TestSessionSetupNamedUserStillVerifiesSignature(t *testing.T) {
	ft := &fakeTransport{responses: anonymousSessionSetupResponses(t, 0)}
	c := newTestClient(ft)
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
	c.Connection.Server.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED

	creds := &credentials.Credentials{Username: "User", Password: "Password"}
	if err := c.SessionSetup(creds); err == nil {
		t.Fatal("named-user SessionSetup accepted an unsigned final SMB 3.1.1 response")
	}
}
