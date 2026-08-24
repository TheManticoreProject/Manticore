package ntlmv1

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestNTLMv1ResponseServerChallenge(t *testing.T) {
	serverChallenge := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	lmResponse := [24]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	ntResponse := [24]byte{0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30}

	resp := NewNTLMv1Response("user", "domain", serverChallenge, lmResponse, ntResponse)

	got := resp.GetServerChallenge()
	if got != serverChallenge {
		t.Errorf("GetServerChallenge() = %v, want %v", got, serverChallenge)
	}

	newChallenge := [8]byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	resp.SetServerChallenge(newChallenge)
	got = resp.GetServerChallenge()
	if got != newChallenge {
		t.Errorf("After SetServerChallenge(%v), GetServerChallenge() = %v", newChallenge, got)
	}
}

func TestNTLMv1ResponseLMChallenge(t *testing.T) {
	serverChallenge := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	lmResponse := [24]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	ntResponse := [24]byte{0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30}

	resp := NewNTLMv1Response("user", "domain", serverChallenge, lmResponse, ntResponse)

	got := resp.GetLmChallengeResponse()
	if got != lmResponse {
		t.Errorf("GetLmChallengeResponse() = %v, want %v", got, lmResponse)
	}

	newResponse := [24]byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38,
		0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0x40,
		0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
	resp.SetLmChallengeResponse(newResponse)
	got = resp.GetLmChallengeResponse()
	if got != newResponse {
		t.Errorf("After SetLmChallengeResponse(%v), GetLmChallengeResponse() = %v", newResponse, got)
	}
}

func TestNTLMv1ResponseNTChallenge(t *testing.T) {
	serverChallenge := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	lmResponse := [24]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	ntResponse := [24]byte{0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30}

	resp := NewNTLMv1Response("user", "domain", serverChallenge, lmResponse, ntResponse)

	got := resp.GetNtChallengeResponse()
	if got != ntResponse {
		t.Errorf("GetNtChallengeResponse() = %v, want %v", got, ntResponse)
	}

	newResponse := [24]byte{0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50,
		0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58,
		0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60}
	resp.SetNtChallengeResponse(newResponse)
	got = resp.GetNtChallengeResponse()
	if got != newResponse {
		t.Errorf("After SetNtChallengeResponse(%v), GetNtChallengeResponse() = %v", newResponse, got)
	}
}

func TestNTLMv1ResponseEqual(t *testing.T) {
	serverChallenge := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	lmResponse := [24]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	ntResponse := [24]byte{0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30}

	resp := NewNTLMv1Response("user", "domain", serverChallenge, lmResponse, ntResponse)

	// Test equal responses
	other := NewNTLMv1Response(
		resp.Username,
		resp.Domain,
		resp.GetServerChallenge(),
		resp.GetLmChallengeResponse(),
		resp.GetNtChallengeResponse(),
	)
	if !resp.Equal(other) {
		t.Error("Equal() returned false for identical responses")
	}

	// Test nil comparison
	if resp.Equal(nil) {
		t.Error("Equal() returned true when comparing with nil")
	}

	// Test unequal responses
	other.SetServerChallenge([8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	if resp.Equal(other) {
		t.Error("Equal() returned true for different responses")
	}
}

// TestNTLMv1HashcatStringMatchesPublishedExample asserts the renderer reproduces
// the published hashcat mode-5500 example hash, pinning the field order.
//
// NetNTLMv1 and NetNTLMv2 use different layouts, and confusing them is easy: mode
// 5500 puts the server challenge LAST, after the LM and NT responses, whereas
// mode 5600 puts it BEFORE the NTProofStr and blob. This guards the v1 side of
// that distinction; crypto/ntlmv2 guards the other.
//
// Source: https://hashcat.net/wiki/doku.php?id=example_hashes
func TestNTLMv1HashcatStringMatchesPublishedExample(t *testing.T) {
	const (
		username        = "u4-netntlm"
		domain          = "kNS"
		lmResponse      = "338d08f8e26de93300000000000000000000000000000000"
		ntResponse      = "9526fb8c23a90751cdd619b6cea564742e1e4bf33006ba41"
		serverChallenge = "cb8086049ec4736c"
		exampleHash     = username + "::" + domain + ":" + lmResponse + ":" + ntResponse + ":" + serverChallenge
	)

	mustDecode := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("bad hex %q: %v", s, err)
		}
		return b
	}

	var challenge [8]byte
	var lm, nt [24]byte
	copy(challenge[:], mustDecode(serverChallenge))
	copy(lm[:], mustDecode(lmResponse))
	copy(nt[:], mustDecode(ntResponse))

	response := NewNTLMv1Response(username, domain, challenge, lm, nt)

	got, err := response.HashcatString()
	if err != nil {
		t.Fatalf("HashcatString() error = %v", err)
	}
	// The renderer uppercases hex; hashcat parses it case-insensitively.
	if !strings.EqualFold(got, exampleHash) {
		t.Fatalf("HashcatString() =\n  %s\nwant (any case)\n  %s", got, exampleHash)
	}

	// The server challenge is the last field for mode 5500, not the first.
	fields := strings.Split(got, ":")
	if len(fields) != 6 {
		t.Fatalf("line has %d colon-separated fields, want 6: %q", len(fields), got)
	}
	if !strings.EqualFold(fields[5], serverChallenge) {
		t.Fatalf("last field = %q, want the server challenge %q", fields[5], serverChallenge)
	}
}

// TestNTLMv1HashcatModeIsNetNTLMv1 guards the mode number a caller labels
// captured material with.
func TestNTLMv1HashcatModeIsNetNTLMv1(t *testing.T) {
	if HashcatMode != 5500 {
		t.Fatalf("HashcatMode = %d, want 5500 for NetNTLMv1", HashcatMode)
	}
}
