package authenticate_test

import (
	"bytes"
	"crypto/rc4"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// sessionFlags are the AUTHENTICATE flags that select how the signing and sealing
// keys are derived, and so only hold when the server's CHALLENGE offered them.
var sessionFlags = []struct {
	name string
	flag flags.NegotiateFlags
}{
	{"EXTENDED_SESSIONSECURITY", flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY},
	{"128", flags.NTLMSSP_NEGOTIATE_128},
	{"56", flags.NTLMSSP_NEGOTIATE_56},
}

// TestAuthenticateDoesNotAssertUnofferedSessionFlags guards against the client
// asserting extended session security (or a sealing key strength) the server's
// CHALLENGE did not offer. The client keys its signing and sealing from the flags
// in its own AUTHENTICATE ([MS-NLMP] 3.4.5), so asserting one the server did not
// negotiate keys the two sides differently and every signature fails.
func TestAuthenticateDoesNotAssertUnofferedSessionFlags(t *testing.T) {
	ch := kxChallenge(t)
	ch.NegotiateFlags &^= flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		flags.NTLMSSP_NEGOTIATE_128 | flags.NTLMSSP_NEGOTIATE_56

	builders := map[string]func() (*authenticate.AuthenticateMessage, error){
		"password": func() (*authenticate.AuthenticateMessage, error) {
			return authenticate.CreateAuthenticateMessage(ch, kxUser, kxPassword, kxDomain, "WORKSTATION")
		},
		"nt hash": func() (*authenticate.AuthenticateMessage, error) {
			return authenticate.CreateAuthenticateMessageWithNTHash(ch, kxUser, nt.NTHash(kxPassword), kxDomain, "WORKSTATION")
		},
		"anonymous": func() (*authenticate.AuthenticateMessage, error) {
			return authenticate.CreateAnonymousAuthenticateMessage(ch, kxDomain, "WORKSTATION")
		},
	}
	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			msg, err := build()
			if err != nil {
				t.Fatalf("build AUTHENTICATE: %v", err)
			}
			for _, f := range sessionFlags {
				if msg.NegotiateFlags&f.flag != 0 {
					t.Errorf("AUTHENTICATE asserts NTLMSSP_NEGOTIATE_%s although the CHALLENGE did not offer it", f.name)
				}
			}
		})
	}
}

// TestAuthenticateAssertsOfferedSessionFlags pins the other branch: a server that
// offers extended session security and both key strengths gets them all back.
func TestAuthenticateAssertsOfferedSessionFlags(t *testing.T) {
	ch := kxChallenge(t)
	ch.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		flags.NTLMSSP_NEGOTIATE_128 | flags.NTLMSSP_NEGOTIATE_56

	msg, err := authenticate.CreateAuthenticateMessage(ch, kxUser, kxPassword, kxDomain, "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}
	for _, f := range sessionFlags {
		if msg.NegotiateFlags&f.flag == 0 {
			t.Errorf("AUTHENTICATE does not assert NTLMSSP_NEGOTIATE_%s although the CHALLENGE offered it", f.name)
		}
	}
}

// TestNoESSAuthenticateKeepsKeyExchangeAndMIC checks that clearing extended session
// security leaves the rest of the message intact: key exchange still wraps the
// exported session key under the SessionBaseKey, and the MIC the server's
// MsvAvTimestamp calls for still verifies over the message as sent.
func TestNoESSAuthenticateKeepsKeyExchangeAndMIC(t *testing.T) {
	ch := kxChallenge(t)
	ch.NegotiateFlags &^= flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY
	ch.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_KEY_EXCH

	msg, err := authenticate.CreateAuthenticateMessage(ch, kxUser, kxPassword, kxDomain, "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}
	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_KEY_EXCH == 0 {
		t.Fatal("AUTHENTICATE does not assert the key exchange the CHALLENGE offered")
	}
	if !msg.NeedsMIC {
		t.Fatal("a CHALLENGE carrying MsvAvTimestamp produced no MIC")
	}

	c, err := rc4.NewCipher(sessionBaseKeyOf(t, msg.NtChallengeResponse))
	if err != nil {
		t.Fatalf("rc4.NewCipher: %v", err)
	}
	unwrapped := make([]byte, 16)
	c.XORKeyStream(unwrapped, msg.EncryptedRandomSessionKey)
	if !bytes.Equal(unwrapped, msg.SessionKey) {
		t.Fatal("EncryptedRandomSessionKey does not unwrap to SessionKey under the SessionBaseKey")
	}

	negotiateBytes := []byte("NTLMSSP\x00\x01\x00\x00\x00pretend-negotiate")
	challengeBytes, err := ch.Marshal()
	if err != nil {
		t.Fatalf("marshal CHALLENGE: %v", err)
	}
	if err := msg.ComputeMIC(negotiateBytes, challengeBytes); err != nil {
		t.Fatalf("ComputeMIC: %v", err)
	}
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var received authenticate.AuthenticateMessage
	if _, err := received.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if received.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY != 0 {
		t.Error("the AUTHENTICATE on the wire asserts extended session security")
	}
	if !received.VerifyMIC(unwrapped, negotiateBytes, challengeBytes, raw) {
		t.Error("the MIC does not verify under the exported session key")
	}
}
