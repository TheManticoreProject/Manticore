package server_test

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	spnego_ntlm_negotiate_flags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
)

// TestClientRelaySessionSetup drives the outbound relay seam end to end: an
// independent SPNEGO initiator (the stand-in victim, holding the account's
// secret) produces the NTLM NEGOTIATE and AUTHENTICATE, while the client carries
// those opaque tokens to the server with SessionSetupNegotiateBlob /
// SessionSetupAuthenticateBlob and computes nothing itself. A session is
// installed only if the server accepted the relayed AUTHENTICATE.
func TestClientRelaySessionSetup(t *testing.T) {
	_, port := startServer(t, server.Config{})

	c := client.NewClientUsingTCPTransport(net.IPv4(127, 0, 0, 1), port)
	c.Profile = c.Profile.Clone()
	c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0}
	t.Cleanup(func() { c.Disconnect() })
	if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// The victim owns the credential; the relaying client never sees it.
	victim := spnego.NewAuthContext(spnego.AuthTypeNTLM, "LAB", "alice", "Passw0rd!", "VICTIM", true)
	flags := spnego_ntlm_negotiate_flags.NegotiateFlags(
		spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_UNICODE |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_NTLM |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_SIGN |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_128 |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_56 |
			spnego_ntlm_negotiate_flags.NTLMSSP_REQUEST_TARGET |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_KEY_EXCH |
			spnego_ntlm_negotiate_flags.NTLMSSP_NEGOTIATE_VERSION,
	)
	v := version.DefaultVersion()
	negotiateToken, err := victim.CreateNegotiateToken(flags, &v)
	if err != nil {
		t.Fatalf("victim CreateNegotiateToken: %v", err)
	}

	// Leg 1: relay the NEGOTIATE, get back the server's CHALLENGE.
	challengeToken, sessionId, err := c.SessionSetupNegotiateBlob(negotiateToken)
	if err != nil {
		t.Fatalf("SessionSetupNegotiateBlob: %v", err)
	}
	if len(challengeToken) == 0 {
		t.Fatal("no CHALLENGE token returned")
	}

	// The victim answers the relayed CHALLENGE; the client never computes this.
	authenticateToken, err := victim.CreateAuthenticateTokenFromChallengeToken(challengeToken)
	if err != nil {
		t.Fatalf("victim CreateAuthenticateTokenFromChallengeToken: %v", err)
	}

	// Leg 2: relay the AUTHENTICATE. Success proves the server honoured a response
	// the relaying client could not have produced.
	if err := c.SessionSetupAuthenticateBlob(sessionId, authenticateToken); err != nil {
		t.Fatalf("SessionSetupAuthenticateBlob: %v", err)
	}
	if c.Session == nil {
		t.Fatal("no session was installed after a successful relay")
	}
	if c.Session.SessionId != sessionId {
		t.Fatalf("installed session %#x, want %#x", c.Session.SessionId, sessionId)
	}
}
