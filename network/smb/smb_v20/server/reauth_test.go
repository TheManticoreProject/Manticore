package server_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	spnego_ntlm_negotiate_flags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestClientRelayReauthMechanics exercises the client side of the SMB signing
// bypass: after a credentialed SessionSetup, RelayReauthNegotiateBlob must
// re-authenticate the EXISTING session (not open a new one) and must leave that
// session — its id and signing key — untouched.
//
// The full bypass cannot be exercised here: it depends on the target keeping the
// session's signing key across a cross-user re-authentication, which Windows does
// but the Manticore SMB2 server deliberately refuses (NT_STATUS_REQUEST_NOT_ACCEPTED,
// session.go). So this asserts only the client invariants; the end-to-end bypass
// must be validated live against a signing-enforced Windows target.
func TestClientRelayReauthMechanics(t *testing.T) {
	_, port := startServer(t, server.Config{})

	c := client.NewClientUsingTCPTransport(net.IPv4(127, 0, 0, 1), port)
	c.Profile = c.Profile.Clone()
	c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0}
	t.Cleanup(func() { c.Disconnect() })
	if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Seed a real session with a credential this client controls.
	creds, err := credentials.NewCredentials("LAB", "alice", "Passw0rd!", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SessionSetup(creds); err != nil {
		t.Fatalf("seed SessionSetup: %v", err)
	}
	if c.Session == nil {
		t.Fatal("seed session was not installed")
	}
	seedID := c.Session.SessionId
	seedKey := append([]byte(nil), c.Session.SessionKey...)

	// A relayed NEGOTIATE (its origin is irrelevant: the server refuses the
	// re-auth before validating it).
	victim := spnego.NewAuthContext(spnego.AuthTypeNTLM, "LAB", "alice", "Passw0rd!", "VICTIM", true)
	v := version.DefaultVersion()
	negotiateToken, err := victim.CreateNegotiateToken(
		spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_UNICODE|
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_NTLM|
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY|
			spnego_ntlm_negotiate_flags.NTLMSSP_REQUEST_TARGET|
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_TARGET_INFO|
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_VERSION, &v)
	if err != nil {
		t.Fatalf("victim CreateNegotiateToken: %v", err)
	}

	// The re-auth targets the existing session. Against this (non-vulnerable)
	// server it is refused, which is the point of the assertion below: the
	// defensive server rejects a re-authentication outright.
	if _, err := c.RelayReauthNegotiateBlob(negotiateToken); err == nil {
		t.Fatal("expected the Manticore server to refuse re-authentication, got success")
	}

	// Whatever the refusal, the seed session must be preserved: its id and key
	// are what a real bypass keeps signing with.
	if c.Session == nil {
		t.Fatal("re-auth attempt cleared the established session")
	}
	if c.Session.SessionId != seedID {
		t.Fatalf("session id changed: %#x -> %#x", seedID, c.Session.SessionId)
	}
	if !bytes.Equal(c.Session.SessionKey, seedKey) {
		t.Fatal("re-auth attempt altered the session signing key")
	}
}
