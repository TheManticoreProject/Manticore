package rpcserver

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// victimNTLMFlags is a client's NEGOTIATE flag set: NTLMv2 with signing, sealing and a
// MIC, matching what a coerced Windows client offers. It mirrors the in-repo client's
// own rpcNTLMNegotiateFlags so the relayed exchange carries a MIC the target verifies.
const victimNTLMFlags = flags.NTLMSSP_NEGOTIATE_UNICODE |
	flags.NTLMSSP_REQUEST_TARGET |
	flags.NTLMSSP_NEGOTIATE_SIGN |
	flags.NTLMSSP_NEGOTIATE_SEAL |
	flags.NTLMSSP_NEGOTIATE_NTLM |
	flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
	flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
	flags.NTLMSSP_NEGOTIATE_128 |
	flags.NTLMSSP_NEGOTIATE_56

// buildVictimNegotiate returns a bare NTLM NEGOTIATE token for the given workstation.
func buildVictimNegotiate(t *testing.T, workstation string) []byte {
	t.Helper()
	neg, err := negotiate.CreateNegotiateMessage("", workstation, victimNTLMFlags, nil)
	if err != nil {
		t.Fatal(err)
	}
	negBytes, err := neg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return negBytes
}

// buildVictimAuthenticate answers a CHALLENGE as domain\user, computing the MIC over
// NEGOTIATE||CHALLENGE||AUTHENTICATE when the challenge requires it, exactly as a real
// client does. It returns the bare NTLM AUTHENTICATE token.
func buildVictimAuthenticate(t *testing.T, negotiateBytes, challengeBytes []byte, domain, user, password, workstation string) []byte {
	t.Helper()
	var chal challenge.ChallengeMessage
	if _, err := chal.Unmarshal(challengeBytes); err != nil {
		t.Fatalf("parse challenge: %v", err)
	}
	auth, err := authenticate.CreateAuthenticateMessageWithOptions(&chal, user, password, domain, workstation, targetinfo.BlobOptions{})
	if err != nil {
		t.Fatalf("build authenticate: %v", err)
	}
	if auth.NeedsMIC {
		if err := auth.ComputeMIC(negotiateBytes, challengeBytes); err != nil {
			t.Fatalf("compute MIC: %v", err)
		}
	}
	authBytes, err := auth.Marshal()
	if err != nil {
		t.Fatalf("marshal authenticate: %v", err)
	}
	return authBytes
}

// TestDCERPCRelayClientBindsAsRelayedIdentity drives the outbound relay seam
// (client.BindRelayNegotiate / BindRelayAuthenticate) against a relay-unaware NTLM
// endpoint: a victim's NTLM is shuttled into the bind with no credential of the relay's
// own, and the target then answers a call as the relayed identity — proving it accepted
// the connect-level relayed authentication.
func TestDCERPCRelayClientBindsAsRelayedIdentity(t *testing.T) {
	service := &stubService{answer: []byte("pong")}
	target := ntlmEndpoint(t, service) // knows LAB\alice : Passw0rd!

	relayOut := client.NewClient(&loopTransport{association: target.Open()})

	// Leg 1: forward the victim's NEGOTIATE, receive the target's CHALLENGE.
	victimNeg := buildVictimNegotiate(t, "VICTIM")
	challengeBytes, err := relayOut.BindRelayNegotiate(testSyntax(), victimNeg)
	if err != nil {
		t.Fatalf("BindRelayNegotiate() error = %v", err)
	}
	if len(challengeBytes) == 0 {
		t.Fatal("BindRelayNegotiate returned no challenge")
	}

	// Leg 2: the victim (alice) answers the challenge; forward its AUTHENTICATE.
	victimAuth := buildVictimAuthenticate(t, victimNeg, challengeBytes, "LAB", "alice", "Passw0rd!", "VICTIM")
	if err := relayOut.BindRelayAuthenticate(victimAuth); err != nil {
		t.Fatalf("BindRelayAuthenticate() error = %v", err)
	}

	// The association is bound as the relayed identity: a call is answered.
	stub, err := relayOut.Call(0, []byte("ping"))
	if err != nil {
		t.Fatalf("Call() after relayed bind error = %v", err)
	}
	if string(stub) != "pong" {
		t.Fatalf("Call() = %q, want %q", stub, "pong")
	}
	if got := string(service.lastStub); got != "ping" {
		t.Fatalf("the interface received %q, want %q", got, "ping")
	}
}

// TestDCERPCRelayClientRejectsWrongPassword confirms the relayed bind is genuinely
// verified end to end: a victim that answers with the wrong secret is not admitted, so
// the target faults the call rather than answering it.
func TestDCERPCRelayClientRejectsWrongPassword(t *testing.T) {
	service := &stubService{answer: []byte("pong")}
	target := ntlmEndpoint(t, service)

	relayOut := client.NewClient(&loopTransport{association: target.Open()})
	victimNeg := buildVictimNegotiate(t, "VICTIM")
	challengeBytes, err := relayOut.BindRelayNegotiate(testSyntax(), victimNeg)
	if err != nil {
		t.Fatalf("BindRelayNegotiate() error = %v", err)
	}
	victimAuth := buildVictimAuthenticate(t, victimNeg, challengeBytes, "LAB", "alice", "wrong-password", "VICTIM")
	if err := relayOut.BindRelayAuthenticate(victimAuth); err != nil {
		t.Fatalf("BindRelayAuthenticate() error = %v", err)
	}
	if _, err := relayOut.Call(0, []byte("ping")); err == nil {
		t.Fatal("a call succeeded after a relayed bind with the wrong password")
	}
	if service.callCount() != 0 {
		t.Fatal("the interface was called despite a failed relayed authentication")
	}
}

// TestOnAcceptContextRelaysInboundAuth drives the inbound relay seam
// (Dispatcher.OnAcceptContext): the endpoint sources its CHALLENGE from an upstream
// target via the acceptor's ChallengeProvider, never admits the victim, and the
// AUTHENTICATE it captures authenticates as the victim against that upstream target —
// the full inbound half of a relay.
func TestOnAcceptContextRelaysInboundAuth(t *testing.T) {
	info, err := targetinfo.BuildServerTargetInfo("DC", "LAB", "dc.lab.local", "lab.local", nil)
	if err != nil {
		t.Fatal(err)
	}
	// The upstream target the relay would authenticate to (stands in for the relay's
	// outbound leg): it knows LAB\alice and issues the CHALLENGE the victim commits to.
	target := spnego.NewAcceptContext("LAB", challenge.TargetTypeDomain, info)
	target.CredentialLookup = func(domain, username string) ([16]byte, bool) {
		if strings.EqualFold(username, "alice") {
			return nt.NTHash("Passw0rd!"), true
		}
		return [16]byte{}, false
	}

	d := New(&stubService{answer: []byte("x")})
	d.EnableNTLM(NTLMAuth{}) // accept authenticated binds; the challenge comes from the provider
	var hookRan bool
	var captured []byte
	d.OnAcceptContext(func(accept *spnego.AcceptContext) {
		hookRan = true
		accept.ChallengeProvider = func([]byte) ([]byte, error) {
			if _, err := target.AcceptNegotiateToken(accept.RawNegotiateToken); err != nil {
				return nil, err
			}
			return target.ChallengeMessageBytes, nil
		}
		accept.CredentialLookup = func(domain, username string) ([16]byte, bool) {
			captured = accept.RawAuthenticateToken
			return [16]byte{}, false // the relay was the point; never admit the victim
		}
	})

	// The victim (alice) authenticates to the inbound endpoint.
	victim := client.NewClient(&loopTransport{association: d.Open()})
	creds, err := credentials.NewCredentials("LAB", "alice", "Passw0rd!", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := victim.SetAuth(pdu.AuthTypeNTLMSSP, pdu.AuthLevelConnect, creds); err != nil {
		t.Fatal(err)
	}
	// The auth3 is unacknowledged, so the victim's own bind reports success locally even
	// though the endpoint refused admission; the relay has what it needs regardless.
	_ = victim.Bind(testSyntax())

	if !hookRan {
		t.Fatal("OnAcceptContext was not invoked")
	}
	if len(captured) == 0 {
		t.Fatal("the victim's AUTHENTICATE was not captured for relaying")
	}

	// The captured AUTHENTICATE, forwarded to the upstream target, authenticates as alice.
	if err := target.AcceptAuthenticateToken(captured); err != nil {
		t.Fatalf("target rejected the relayed AUTHENTICATE token: %v", err)
	}
	if err := target.Verify(); err != nil {
		t.Fatalf("the upstream target rejected the relayed victim: %v", err)
	}
	if _, user, _ := target.Identity(); !strings.EqualFold(user, "alice") {
		t.Fatalf("the relayed identity is %q, want alice", user)
	}
}
