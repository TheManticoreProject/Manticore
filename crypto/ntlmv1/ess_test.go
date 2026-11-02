package ntlmv1

import (
	"encoding/hex"
	"testing"
)

// The [MS-NLMP] 4.2.3 example ("NTLM v1 with Client Challenge"): user "User",
// domain "Domain", password "Password", ServerChallenge 0123456789abcdef,
// ClientChallenge aaaaaaaaaaaaaaaa. The expected values are the specification's;
// impacket's computeResponseNTLMv1 and KXKEY give the same bytes.
func TestComputeResponseESSMatchesMSNLMP(t *testing.T) {
	serverChallenge := [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	clientChallenge := [8]byte{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa}
	ctx, err := NewNTLMv1CtxWithPassword("Domain", "User", "Password", serverChallenge)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ctx.ComputeResponseESS(clientChallenge)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ name, got, want string }{
		{"NtChallengeResponse", hex.EncodeToString(resp.NtChallengeResponse[:]), "7537f803ae367128ca458204bde7caf81e97ed2683267232"},
		{"LmChallengeResponse", hex.EncodeToString(resp.LmChallengeResponse[:]), "aaaaaaaaaaaaaaaa00000000000000000000000000000000"},
		{"SessionBaseKey", hexOf(ctx.SessionBaseKey()), "d87262b0cde4b1cb7499becccdf10784"},
		{"KeyExchangeKey", hexOf(ctx.KeyExchangeKeyESS(clientChallenge)), "eb93429a8bd952f8b89c55b87f475edc"},
	} {
		if c.got != c.want {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}

	// The plain NTLMv1 response of the same inputs is unchanged by the DESL
	// refactor ([MS-NLMP] 4.2.2).
	plain, err := ctx.ComputeResponse()
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(plain.NtChallengeResponse[:]); got != "67c43011f30298a2ad35ece64f16331c44bdbed927841f94" {
		t.Errorf("plain NtChallengeResponse = %s", got)
	}
	if got := hex.EncodeToString(plain.LmChallengeResponse[:]); got != "98def7b87f88aa5dafe2df779688a172def11c7d5ccdef13" {
		t.Errorf("plain LmChallengeResponse = %s", got)
	}
}

// TestSessionBaseKeyDerivesNTHash checks SessionBaseKey derives the NT hash from
// the password when it has not been computed yet, instead of hashing zeros.
func TestSessionBaseKeyDerivesNTHash(t *testing.T) {
	ctx := &NTLMv1Ctx{Password: "Password"}
	if got := hexOf(ctx.SessionBaseKey()); got != "d87262b0cde4b1cb7499becccdf10784" {
		t.Errorf("SessionBaseKey = %s, want the MS-NLMP value", got)
	}
}

func hexOf(b [16]byte) string { return hex.EncodeToString(b[:]) }
