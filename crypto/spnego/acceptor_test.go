package spnego

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/rc4"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

// The identity used throughout. The domain is deliberately mixed-case: NTLMv2
// folds the domain into NTOWFv2 exactly as sent, so a verifier that normalizes it
// rejects every client whose domain is not already upper-case.
const (
	testDomain      = "Lab.Example.Local"
	testUsername    = "alice"
	testPassword    = "Passw0rd!"
	testWorkstation = "CLIENT01"
)

// clientNegotiateFlags are the options the initiator in this repository offers.
const clientNegotiateFlags = flags.NTLMSSP_NEGOTIATE_UNICODE |
	flags.NTLMSSP_NEGOTIATE_NTLM |
	flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
	flags.NTLMSSP_NEGOTIATE_SIGN |
	flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
	flags.NTLMSSP_NEGOTIATE_128 |
	flags.NTLMSSP_NEGOTIATE_56 |
	flags.NTLMSSP_REQUEST_TARGET |
	flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
	flags.NTLMSSP_NEGOTIATE_VERSION

// serverTargetInfo builds the TargetInfo an acceptor advertises. The timestamp is
// omitted by default so the initiator is not obliged to carry a MIC.
func serverTargetInfo(t *testing.T, timestamp []byte) []byte {
	t.Helper()
	info, err := targetinfo.BuildServerTargetInfo("SERVER01", "LAB", "server01.lab.example.local", "lab.example.local", timestamp)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo() error = %v", err)
	}
	return info
}

// lookupFor returns a CredentialLookup that answers for one identity, recording
// what it was asked so a test can assert the acceptor passed the claimed identity
// through unchanged.
func lookupFor(domain, username, password string, asked *[2]string) func(string, string) ([16]byte, bool) {
	return func(gotDomain, gotUsername string) ([16]byte, bool) {
		if asked != nil {
			*asked = [2]string{gotDomain, gotUsername}
		}
		if !strings.EqualFold(gotUsername, username) || gotDomain != domain {
			return [16]byte{}, false
		}
		return nt.NTHash(password), true
	}
}

// runExchange drives the initiator in this repository against the acceptor and
// returns both, so a test can compare what the two derived.
func runExchange(t *testing.T, ctx *AcceptContext) *AuthContext {
	t.Helper()

	client := NewAuthContext(AuthTypeNTLM, testDomain, testUsername, testPassword, testWorkstation, true)
	v := version.DefaultVersion()

	negotiateToken, err := client.CreateNegotiateToken(clientNegotiateFlags, &v)
	if err != nil {
		t.Fatalf("client CreateNegotiateToken() error = %v", err)
	}

	challengeToken, err := ctx.AcceptNegotiateToken(negotiateToken)
	if err != nil {
		t.Fatalf("AcceptNegotiateToken() error = %v", err)
	}

	authenticateToken, err := client.CreateAuthenticateTokenFromChallengeToken(challengeToken)
	if err != nil {
		t.Fatalf("client CreateAuthenticateTokenFromChallengeToken() error = %v", err)
	}

	if err := ctx.AcceptAuthenticateToken(authenticateToken); err != nil {
		t.Fatalf("AcceptAuthenticateToken() error = %v", err)
	}

	return client
}

// TestAcceptorAgreesWithInitiator drives the initiator in this repository against
// the acceptor end to end and asserts both sides derive the identical session key.
// Agreeing on the key is the whole point: it is what a consumer signs and seals
// with, so a mismatch would break every message after authentication.
func TestAcceptorAgreesWithInitiator(t *testing.T) {
	var asked [2]string
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	ctx.CredentialLookup = lookupFor(testDomain, testUsername, testPassword, &asked)
	v := version.DefaultVersion()
	ctx.Version = &v

	client := runExchange(t, ctx)

	if err := ctx.Verify(); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ctx.Verified {
		t.Fatal("Verify() succeeded but Verified is false")
	}

	clientKey := client.GetSessionKey()
	if len(clientKey) == 0 {
		t.Fatal("the initiator derived no session key, so there is nothing to compare")
	}
	if !bytes.Equal(ctx.GetSessionKey(), clientKey) {
		t.Fatalf("session keys differ:\n  acceptor %s\n  initiator %s",
			hex.EncodeToString(ctx.GetSessionKey()), hex.EncodeToString(clientKey))
	}

	// The identity passed to the lookup must be exactly what the client claimed,
	// with the domain's case intact.
	if asked[0] != testDomain {
		t.Fatalf("lookup was asked for domain %q, want %q", asked[0], testDomain)
	}
	if asked[1] != testUsername {
		t.Fatalf("lookup was asked for username %q, want %q", asked[1], testUsername)
	}
}

// TestAcceptorIdentityAndCapture asserts the claimed identity is decoded and the
// response is rendered as crackable NetNTLMv2 material, which is what an acceptor
// that only harvests responses needs.
func TestAcceptorIdentityAndCapture(t *testing.T) {
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	runExchange(t, ctx)

	domain, username, workstation := ctx.Identity()
	if domain != testDomain {
		t.Fatalf("domain = %q, want %q", domain, testDomain)
	}
	if username != testUsername {
		t.Fatalf("username = %q, want %q", username, testUsername)
	}
	if !strings.EqualFold(workstation, testWorkstation) {
		t.Fatalf("workstation = %q, want %q", workstation, testWorkstation)
	}

	v1, v2 := ctx.CapturedResponse()
	if v1 != nil {
		t.Fatalf("an NTLMv2 exchange yielded an NTLMv1 response: %v", v1)
	}
	if v2 == nil {
		t.Fatal("CapturedResponse() returned no NTLMv2 response")
	}

	line, err := v2.HashcatString()
	if err != nil {
		t.Fatalf("HashcatString() error = %v", err)
	}
	fields := strings.Split(line, ":")
	if len(fields) != 6 {
		t.Fatalf("hashcat line has %d fields, want 6: %q", len(fields), line)
	}
	// The captured challenge must be the one the acceptor issued.
	if !strings.EqualFold(fields[3], hex.EncodeToString(ctx.ServerChallenge[:])) {
		t.Fatalf("captured challenge = %q, want %x", fields[3], ctx.ServerChallenge)
	}
	// The NTProofStr must be the first 16 bytes of what the client sent.
	wantProof := hex.EncodeToString(ctx.Authenticate.NtChallengeResponse[:16])
	if !strings.EqualFold(fields[4], wantProof) {
		t.Fatalf("captured NTProofStr = %q, want %s", fields[4], wantProof)
	}
}

// TestAcceptorCaptureOnly asserts an acceptor with no credential records the
// response and says so, rather than failing in a way that loses it.
func TestAcceptorCaptureOnly(t *testing.T) {
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	runExchange(t, ctx)

	err := ctx.Verify()
	if !errors.Is(err, ErrCaptureOnly) {
		t.Fatalf("Verify() error = %v, want ErrCaptureOnly", err)
	}
	if ctx.Verified {
		t.Fatal("Verified is set although nothing was verified")
	}
	if ctx.GetSessionKey() != nil {
		t.Fatal("a session key was derived without a credential")
	}
	// The capture survives the unverified outcome.
	if _, v2 := ctx.CapturedResponse(); v2 == nil {
		t.Fatal("the response was lost when verification could not run")
	}
}

// TestAcceptorRejectsWrongCredential asserts a response that does not match the
// credential is refused, and separately that an identity with no credential is
// distinguished from one with the wrong password.
func TestAcceptorRejectsWrongCredential(t *testing.T) {
	t.Run("wrong password", func(t *testing.T) {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
		ctx.CredentialLookup = lookupFor(testDomain, testUsername, "not-the-password", nil)
		runExchange(t, ctx)

		if err := ctx.Verify(); !errors.Is(err, ErrBadResponse) {
			t.Fatalf("Verify() error = %v, want ErrBadResponse", err)
		}
		if ctx.GetSessionKey() != nil {
			t.Fatal("a session key was derived from a response that did not verify")
		}
	})

	t.Run("unknown identity", func(t *testing.T) {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
		ctx.CredentialLookup = lookupFor(testDomain, "someone-else", testPassword, nil)
		runExchange(t, ctx)

		if err := ctx.Verify(); !errors.Is(err, ErrUnknownIdentity) {
			t.Fatalf("Verify() error = %v, want ErrUnknownIdentity", err)
		}
	})

	t.Run("tampered NT response", func(t *testing.T) {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
		ctx.CredentialLookup = lookupFor(testDomain, testUsername, testPassword, nil)
		runExchange(t, ctx)

		// Flip a bit in the NTProofStr.
		ctx.Authenticate.NtChallengeResponse[0] ^= 0x01
		if err := ctx.Verify(); !errors.Is(err, ErrBadResponse) {
			t.Fatalf("Verify() error = %v, want ErrBadResponse", err)
		}
	})
}

// TestAcceptorVerifiesMIC asserts a client that commits to the exchange with a MIC
// is checked, and that a tampered MIC is refused even though the NT response
// itself is genuine.
//
// The initiator in this repository does not emit a MIC, so the AUTHENTICATE is
// built directly here to exercise the path.
func TestAcceptorVerifiesMIC(t *testing.T) {
	build := func(t *testing.T, tamper bool) (*AcceptContext, error) {
		t.Helper()

		// A TargetInfo carrying a timestamp is what obliges a client to send a MIC.
		timestamp := make([]byte, 8)
		for i := range timestamp {
			timestamp[i] = byte(i + 1)
		}
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, timestamp))
		ctx.CredentialLookup = lookupFor(testDomain, testUsername, testPassword, nil)

		client := NewAuthContext(AuthTypeNTLM, testDomain, testUsername, testPassword, testWorkstation, true)
		v := version.DefaultVersion()
		negotiateToken, err := client.CreateNegotiateToken(clientNegotiateFlags, &v)
		if err != nil {
			t.Fatalf("CreateNegotiateToken() error = %v", err)
		}
		if _, err := ctx.AcceptNegotiateToken(negotiateToken); err != nil {
			t.Fatalf("AcceptNegotiateToken() error = %v", err)
		}

		// Build an AUTHENTICATE that carries a MIC over this exchange.
		message, err := authenticate.CreateAuthenticateMessage(ctx.Challenge, testUsername, testPassword, testDomain, testWorkstation)
		if err != nil {
			t.Fatalf("CreateAuthenticateMessage() error = %v", err)
		}
		message.NeedsMIC = true
		if err := message.ComputeMIC(ctx.NegotiateMessageBytes, ctx.ChallengeMessageBytes); err != nil {
			t.Fatalf("ComputeMIC() error = %v", err)
		}
		raw, err := message.Marshal()
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		if tamper {
			// Locate the MIC field by parsing the message back, then flip a bit
			// inside it, leaving the NT response intact.
			parsed := &authenticate.AuthenticateMessage{}
			if _, err := parsed.Unmarshal(raw); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !parsed.NeedsMIC || parsed.MICOffset == 0 {
				t.Fatal("could not locate the MIC field in the marshalled message")
			}
			raw[parsed.MICOffset] ^= 0x01
		}

		if err := ctx.AcceptAuthenticateToken(raw); err != nil {
			t.Fatalf("AcceptAuthenticateToken() error = %v", err)
		}
		if !ctx.Authenticate.NeedsMIC {
			t.Fatal("the accepted AUTHENTICATE does not report a MIC, so Unmarshal did not read one")
		}
		return ctx, ctx.Verify()
	}

	t.Run("valid MIC", func(t *testing.T) {
		ctx, err := build(t, false)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
		if len(ctx.GetSessionKey()) == 0 {
			t.Fatal("no session key after a successful verification")
		}
	})

	t.Run("tampered MIC", func(t *testing.T) {
		_, err := build(t, true)
		if !errors.Is(err, ErrBadMIC) {
			t.Fatalf("Verify() error = %v, want ErrBadMIC", err)
		}
	})
}

// TestAcceptorKeyExchange asserts the acceptor unseals a client-chosen session key
// when NTLMSSP_NEGOTIATE_KEY_EXCH was negotiated. A Windows client requests key
// exchange, so this is the path that matters in the field even though the
// initiator in this repository does not use it.
func TestAcceptorKeyExchange(t *testing.T) {
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	ctx.CredentialLookup = lookupFor(testDomain, testUsername, testPassword, nil)
	runExchange(t, ctx)

	// Recompute the KeyExchangeKey the way the acceptor will, then seal a chosen
	// session key under it exactly as a client does.
	ntHash := nt.NTHash(testPassword)
	v2, err := ntlmv2.NewNTLMv2CtxWithNTHash(testDomain, testUsername, ntHash, ctx.ServerChallenge, [8]byte{})
	if err != nil {
		t.Fatalf("NewNTLMv2CtxWithNTHash() error = %v", err)
	}
	keyExchangeKey := ntlmv2.SessionBaseKeyFromResponse(v2.ResponseKeyNT[:], ctx.Authenticate.NtChallengeResponse)
	if len(keyExchangeKey) != 16 {
		t.Fatalf("KeyExchangeKey is %d bytes, want 16", len(keyExchangeKey))
	}

	chosenKey := []byte("0123456789abcdef")
	cipher, err := rc4.NewRC4WithKey(keyExchangeKey)
	if err != nil {
		t.Fatalf("NewRC4WithKey() error = %v", err)
	}
	sealed := make([]byte, len(chosenKey))
	cipher.XORKeyStream(sealed, chosenKey)

	ctx.Authenticate.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_KEY_EXCH
	ctx.Authenticate.EncryptedRandomSessionKey = sealed

	if err := ctx.Verify(); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !bytes.Equal(ctx.GetSessionKey(), chosenKey) {
		t.Fatalf("unsealed session key = %x, want %x", ctx.GetSessionKey(), chosenKey)
	}

	// A malformed sealed key is refused rather than producing a short key.
	ctx.Verified = false
	ctx.SessionKey = nil
	ctx.Authenticate.EncryptedRandomSessionKey = []byte{0x01, 0x02}
	if err := ctx.Verify(); err == nil {
		t.Fatal("Verify() accepted a truncated EncryptedRandomSessionKey")
	}
}

// TestAcceptorChallengeFlags asserts the CHALLENGE advertises the intersection of
// what the client offered with what the acceptor supports, and asserts the bits an
// acceptor must set for a real client to continue.
func TestAcceptorChallengeFlags(t *testing.T) {
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	v := version.DefaultVersion()
	ctx.Version = &v
	runExchange(t, ctx)

	got := ctx.Challenge.NegotiateFlags

	// TargetInfo was supplied, so both bits that announce it must be set or a
	// Windows client abandons the exchange.
	if !got.HasFlag(flags.NTLMSSP_NEGOTIATE_TARGET_INFO) {
		t.Error("CHALLENGE does not set NTLMSSP_NEGOTIATE_TARGET_INFO although TargetInfo was supplied")
	}
	if !got.HasFlag(flags.NTLMSSP_REQUEST_TARGET) {
		t.Error("CHALLENGE does not set NTLMSSP_REQUEST_TARGET although a TargetName was supplied")
	}
	// The client asked to sign, so the acceptor must agree to it.
	if !got.HasFlag(flags.NTLMSSP_NEGOTIATE_SIGN) {
		t.Error("CHALLENGE does not echo NTLMSSP_NEGOTIATE_SIGN")
	}
	if !got.HasFlag(flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) {
		t.Error("CHALLENGE does not set NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY")
	}
	// A target type describes the TargetName.
	if !got.HasFlag(flags.NTLMSSP_TARGET_TYPE_DOMAIN) {
		t.Error("CHALLENGE does not set NTLMSSP_TARGET_TYPE_DOMAIN")
	}
	// Unicode was offered, so OEM must not also be asserted.
	if !got.HasFlag(flags.NTLMSSP_NEGOTIATE_UNICODE) {
		t.Error("CHALLENGE does not set NTLMSSP_NEGOTIATE_UNICODE")
	}
	if got.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM) {
		t.Error("CHALLENGE sets both NTLMSSP_NEGOTIATE_UNICODE and NTLMSSP_NEGOTIATE_OEM")
	}
	// A bit the client never offered must not appear.
	if got.HasFlag(flags.NTLMSSP_NEGOTIATE_DATAGRAM) {
		t.Error("CHALLENGE asserts NTLMSSP_NEGOTIATE_DATAGRAM, which the client did not offer")
	}
	// The advertised TargetInfo must be exactly what was configured.
	if !bytes.Equal(ctx.Challenge.TargetInfo, ctx.TargetInfo) {
		t.Error("CHALLENGE TargetInfo differs from what was configured")
	}
}

// TestAcceptorGeneratesAChallenge asserts a challenge is generated when none was
// supplied, and that two exchanges do not reuse one. A predictable or repeated
// challenge lets a captured response be replayed.
func TestAcceptorGeneratesAChallenge(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 8; i++ {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
		runExchange(t, ctx)

		if ctx.ServerChallenge == ([8]byte{}) {
			t.Fatal("no challenge was generated")
		}
		key := hex.EncodeToString(ctx.ServerChallenge[:])
		if seen[key] {
			t.Fatalf("challenge %s was issued twice", key)
		}
		seen[key] = true
	}
}

// TestAcceptorHonoursSuppliedChallenge asserts an explicitly configured challenge
// is used unchanged, which is what makes an exchange reproducible in a test or
// against a recorded capture.
func TestAcceptorHonoursSuppliedChallenge(t *testing.T) {
	want := [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	ctx.ServerChallenge = want
	runExchange(t, ctx)

	if ctx.ServerChallenge != want {
		t.Fatalf("challenge = %x, want %x", ctx.ServerChallenge, want)
	}
	if ctx.Challenge.ServerChallenge != want {
		t.Fatalf("CHALLENGE carries %x, want %x", ctx.Challenge.ServerChallenge, want)
	}
}

// TestAcceptorRejectsMalformedTokens asserts each way a token can be wrong is
// refused rather than parsed as something else.
func TestAcceptorRejectsMalformedTokens(t *testing.T) {
	t.Run("negotiate leg", func(t *testing.T) {
		cases := []struct {
			name  string
			token []byte
		}{
			{"empty", nil},
			{"not a token", []byte{0x01, 0x02, 0x03}},
			{"truncated NTLMSSP", []byte("NTLMSSP\x00")},
			{"wrong message type", authenticateShapedToken(t)},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, nil)
				if _, err := ctx.AcceptNegotiateToken(tc.token); err == nil {
					t.Fatal("AcceptNegotiateToken() should fail")
				}
			})
		}
	})

	t.Run("authenticate before challenge", func(t *testing.T) {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, nil)
		if err := ctx.AcceptAuthenticateToken(authenticateShapedToken(t)); err == nil {
			t.Fatal("AcceptAuthenticateToken() should refuse to run before a CHALLENGE was issued")
		}
	})

	t.Run("verify before authenticate", func(t *testing.T) {
		ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, nil)
		if err := ctx.Verify(); !errors.Is(err, ErrNoAuthenticate) {
			t.Fatalf("Verify() error = %v, want ErrNoAuthenticate", err)
		}
	})
}

// TestAcceptorRejectsClientWithNoCharacterSet asserts a client offering neither
// Unicode nor OEM is refused, since there is no encoding both sides could agree
// on for the identity.
func TestAcceptorRejectsClientWithNoCharacterSet(t *testing.T) {
	client := NewAuthContext(AuthTypeNTLM, testDomain, testUsername, testPassword, testWorkstation, false)
	negotiateToken, err := client.CreateNegotiateToken(flags.NTLMSSP_NEGOTIATE_NTLM, nil)
	if err != nil {
		t.Fatalf("CreateNegotiateToken() error = %v", err)
	}

	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	if _, err := ctx.AcceptNegotiateToken(negotiateToken); err == nil {
		t.Fatal("AcceptNegotiateToken() accepted a client offering no character set")
	}
}

// TestSessionBaseKeyMatchesComputed asserts the acceptor's route to the
// SessionBaseKey agrees with the computing side's, since the two derive it from
// different starting points.
func TestSessionBaseKeyMatchesComputed(t *testing.T) {
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))
	runExchange(t, ctx)

	ntHash := nt.NTHash(testPassword)
	v2, err := ntlmv2.NewNTLMv2CtxWithNTHash(testDomain, testUsername, ntHash, ctx.ServerChallenge, [8]byte{})
	if err != nil {
		t.Fatalf("NewNTLMv2CtxWithNTHash() error = %v", err)
	}

	ntResponse := ctx.Authenticate.NtChallengeResponse
	fromResponse := ntlmv2.SessionBaseKeyFromResponse(v2.ResponseKeyNT[:], ntResponse)
	fromComputation := v2.ComputeSessionBaseKey(ntResponse[:16])

	if !bytes.Equal(fromResponse, fromComputation) {
		t.Fatalf("SessionBaseKeyFromResponse = %x, ComputeSessionBaseKey = %x", fromResponse, fromComputation)
	}

	// And the same value, derived independently here.
	mac := hmac.New(md5.New, v2.ResponseKeyNT[:])
	mac.Write(ntResponse[:16])
	if !bytes.Equal(fromResponse, mac.Sum(nil)) {
		t.Fatal("SessionBaseKeyFromResponse does not match HMAC_MD5(ResponseKeyNT, NTProofStr)")
	}
}

// TestTargetInfoBuild asserts the TargetInfo builder produces a list the parser in
// this package reads back, terminated as the wire format requires.
func TestTargetInfoBuild(t *testing.T) {
	info := serverTargetInfo(t, []byte{1, 2, 3, 4, 5, 6, 7, 8})

	pairs, err := targetinfo.ParseTargetInfo(info)
	if err != nil {
		t.Fatalf("ParseTargetInfo() error = %v", err)
	}
	for _, id := range []avpair.AvId{
		avpair.MsvAvNbDomainName,
		avpair.MsvAvNbComputerName,
		avpair.MsvAvDnsDomainName,
		avpair.MsvAvDnsComputerName,
		avpair.MsvAvTimestamp,
	} {
		if _, ok := pairs[id]; !ok {
			t.Errorf("TargetInfo is missing %s", id)
		}
	}
	if !targetinfo.HasTimestamp(info) {
		t.Error("HasTimestamp() is false although a timestamp was supplied")
	}
	// The list ends with the EOL terminator: AvId 0, AvLen 0.
	if len(info) < 4 || !bytes.Equal(info[len(info)-4:], []byte{0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("TargetInfo does not end with the MsvAvEOL terminator: % x", info)
	}

	// An omitted name is not advertised as an empty pair.
	sparse, err := targetinfo.BuildServerTargetInfo("SERVER01", "", "", "", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo() error = %v", err)
	}
	sparsePairs, err := targetinfo.ParseTargetInfo(sparse)
	if err != nil {
		t.Fatalf("ParseTargetInfo() error = %v", err)
	}
	if len(sparsePairs) != 1 {
		t.Fatalf("a single supplied name produced %d pairs, want 1", len(sparsePairs))
	}

	// A timestamp of the wrong width is refused rather than truncated.
	if _, err := targetinfo.BuildServerTargetInfo("S", "", "", "", []byte{1, 2, 3}); err == nil {
		t.Error("BuildServerTargetInfo() accepted a 3-byte timestamp")
	}
}

// authenticateShapedToken returns a bare, well-formed AUTHENTICATE message, used
// where a test needs a token of the wrong type for the leg being exercised.
func authenticateShapedToken(t *testing.T) []byte {
	t.Helper()

	message := &authenticate.AuthenticateMessage{}
	message.NegotiateFlags = flags.NTLMSSP_NEGOTIATE_UNICODE
	message.DomainName = []byte{}
	message.UserName = []byte{}
	message.Workstation = []byte{}
	message.LmChallengeResponse = []byte{}
	message.NtChallengeResponse = []byte{}
	message.EncryptedRandomSessionKey = []byte{}

	raw, err := message.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return raw
}
