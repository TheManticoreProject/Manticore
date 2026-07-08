package functions

// White-box tests for the Netlogon per-message security tokens. They pin the AES sign/seal
// output byte-for-byte against vectors generated from impacket's nrpc.SIGN/SEAL (with a
// round-trip confirmed by nrpc.UNSEAL), which is itself faithful to [MS-NRPC] 3.3.4.2.1.
// Being in package functions lets the tests inject a deterministic confounder source.

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// repeatReader yields the same byte forever, so a fixed confounder can be injected into
// Seal for deterministic output.
type repeatReader struct{ b byte }

func (r repeatReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.b
	}
	return len(p), nil
}

// Vector inputs shared by the pinned tests (see .private/plan_netlogon_rpc.md, Appendix 1).
var (
	vecSessionKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	vecPlaintext  = []byte("NetrLogonGetCapabilities-stub-bytes!!")
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// TestSealVector pins the AES sealing token and encrypted stub to the impacket vector
// (sessionKey 00..0f, confounder 0xAA*8, sequence number 0).
func TestSealVector(t *testing.T) {
	m := NewMessageSecurityAES(vecSessionKey)
	m.confounderSrc = repeatReader{0xAA}

	sealed, token, err := m.Seal(vecPlaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	wantSealed := "af57a4b849dee8df36b76b23dd783d234d68580a36fc7ce2e5d0aa4ec47ef23392481d372b"
	if got := hex.EncodeToString(sealed); got != wantSealed {
		t.Errorf("sealed stub\n got  %s\n want %s", got, wantSealed)
	}
	wantToken := "13001a00ffff0000a8a7c0b370691c1de5b3d4e3ccacd9f4979ba2aab4c9b01e" +
		"000000000000000000000000000000000000000000000000"
	if got := hex.EncodeToString(token); got != wantToken {
		t.Errorf("token\n got  %s\n want %s", got, wantToken)
	}
	if len(token) != 56 {
		t.Errorf("sealing token length = %d, want 56", len(token))
	}
}

// TestSignVector pins the AES integrity-only token to the impacket vector.
func TestSignVector(t *testing.T) {
	m := NewMessageSecurityAES(vecSessionKey)

	token, err := m.Sign(vecPlaintext)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	wantToken := "1300ffffffff00002bf226c4856d2bb7f596b051a3f2891f" +
		"000000000000000000000000000000000000000000000000"
	if got := hex.EncodeToString(token); got != wantToken {
		t.Errorf("token\n got  %s\n want %s", got, wantToken)
	}
	if len(token) != 48 {
		t.Errorf("integrity token length = %d, want 48", len(token))
	}
}

// TestSealUnsealRoundTrip seals with one context and unseals with a fresh context under the
// same key, confirming the token is self-contained.
func TestSealUnsealRoundTrip(t *testing.T) {
	sender := NewMessageSecurityAES(vecSessionKey)
	sender.confounderSrc = repeatReader{0x5a}

	sealed, token, err := sender.Seal(vecPlaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	receiver := NewMessageSecurityAES(vecSessionKey)
	got, err := receiver.Unseal(sealed, token)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}
	if !bytes.Equal(got, vecPlaintext) {
		t.Fatalf("round-trip plaintext = %x, want %x", got, vecPlaintext)
	}
}

// TestUnsealTamperDetected confirms a flipped ciphertext byte fails the checksum.
func TestUnsealTamperDetected(t *testing.T) {
	m := NewMessageSecurityAES(vecSessionKey)
	m.confounderSrc = repeatReader{0x11}
	sealed, token, err := m.Seal(vecPlaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	sealed[0] ^= 0x01
	if _, err := NewMessageSecurityAES(vecSessionKey).Unseal(sealed, token); err == nil {
		t.Fatal("Unseal accepted a tampered stub, want checksum mismatch")
	}
}

// TestVerifySignature confirms a good signature verifies and a tampered stub is rejected.
func TestVerifySignature(t *testing.T) {
	token, err := NewMessageSecurityAES(vecSessionKey).Sign(vecPlaintext)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if err := NewMessageSecurityAES(vecSessionKey).VerifySignature(vecPlaintext, token); err != nil {
		t.Fatalf("VerifySignature (good) = %v, want nil", err)
	}
	tampered := append([]byte(nil), vecPlaintext...)
	tampered[0] ^= 0x01
	if err := NewMessageSecurityAES(vecSessionKey).VerifySignature(tampered, token); err == nil {
		t.Fatal("VerifySignature accepted a tampered stub, want mismatch")
	}
}

// TestSequenceNumberAdvances confirms consecutive messages carry different (encrypted)
// sequence numbers, i.e. the counter advances per call.
func TestSequenceNumberAdvances(t *testing.T) {
	m := NewMessageSecurityAES(vecSessionKey)
	t1, err := m.Sign(vecPlaintext)
	if err != nil {
		t.Fatalf("Sign #1: %v", err)
	}
	t2, err := m.Sign(vecPlaintext)
	if err != nil {
		t.Fatalf("Sign #2: %v", err)
	}
	// SequenceNumber field occupies token bytes [8:16].
	if bytes.Equal(t1[8:16], t2[8:16]) {
		t.Fatal("sequence number did not advance between messages")
	}
}

// TestDeriveSequenceNumber pins the wire rendering of the counter, including the client bit.
func TestDeriveSequenceNumber(t *testing.T) {
	if got := hex.EncodeToString(deriveSequenceNumber(0, true)); got != "0000000080000000" {
		t.Errorf("client seq 0 = %s, want 0000000080000000", got)
	}
	if got := hex.EncodeToString(deriveSequenceNumber(0, false)); got != "0000000000000000" {
		t.Errorf("server seq 0 = %s, want 0000000000000000", got)
	}
	// low = 0x01020304 big-endian, high = 0x80000000 (client bit only).
	want := mustHex(t, "0102030480000000")
	if got := deriveSequenceNumber(0x01020304, true); !bytes.Equal(got, want) {
		t.Errorf("seq 0x01020304 = %x, want %x", got, want)
	}
}
