package cts

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

// unhex decodes a whitespace-separated hex string (as printed in the RFCs).
func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.Join(strings.Fields(s), ""))
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// TestRFC3962Vectors validates the AES-CTS (CBC ciphertext-stealing) encryption
// against the published known-answer vectors in RFC 3962 Appendix B. This is the
// AES cipher-mode core the Kerberos etype-17/18 (aes128/256-cts-hmac-sha1-96)
// encryption is built on. The 128-bit key is the ASCII "chicken teriyaki", the
// IV is all-zero, and the plaintext is successive prefixes of a fixed sentence.
// Both Encrypt→exact ciphertext and Decrypt→plaintext are checked.
func TestRFC3962Vectors(t *testing.T) {
	key := unhex(t, "63 68 69 63 6b 65 6e 20 74 65 72 69 79 61 6b 69") // "chicken teriyaki"
	iv := make([]byte, 16)

	// RFC 3962 Appendix B plaintext, of which each vector encrypts a prefix.
	const plaintext = "I would like the General Gau's Chicken, please, and wonton soup."
	if len(plaintext) != 64 {
		t.Fatalf("plaintext length = %d, want 64", len(plaintext))
	}

	cases := []struct {
		length int
		out    string
	}{
		{17, "c6 35 35 68 f2 bf 8c b4 d8 a5 80 36 2d a7 ff 7f " +
			"97"},
		{31, "fc 00 78 3e 0e fd b2 c1 d4 45 d4 c8 ef f7 ed 22 " +
			"97 68 72 68 d6 ec cc c0 c0 7b 25 e2 5e cf e5"},
		{32, "39 31 25 23 a7 86 62 d5 be 7f cb cc 98 eb f5 a8 " +
			"97 68 72 68 d6 ec cc c0 c0 7b 25 e2 5e cf e5 84"},
		{47, "97 68 72 68 d6 ec cc c0 c0 7b 25 e2 5e cf e5 84 " +
			"b3 ff fd 94 0c 16 a1 8c 1b 55 49 d2 f8 38 02 9e " +
			"39 31 25 23 a7 86 62 d5 be 7f cb cc 98 eb f5"},
		{48, "97 68 72 68 d6 ec cc c0 c0 7b 25 e2 5e cf e5 84 " +
			"9d ad 8b bb 96 c4 cd c0 3b c1 03 e1 a1 94 bb d8 " +
			"39 31 25 23 a7 86 62 d5 be 7f cb cc 98 eb f5 a8"},
		{64, "97 68 72 68 d6 ec cc c0 c0 7b 25 e2 5e cf e5 84 " +
			"39 31 25 23 a7 86 62 d5 be 7f cb cc 98 eb f5 a8 " +
			"48 07 ef e8 36 ee 89 a5 26 73 0d bc 2f 7b c8 40 " +
			"9d ad 8b bb 96 c4 cd c0 3b c1 03 e1 a1 94 bb d8"},
	}

	for _, c := range cases {
		pt := []byte(plaintext[:c.length])
		wantCT := unhex(t, c.out)
		if len(wantCT) != c.length {
			t.Fatalf("len=%d: vector output is %d bytes, want %d", c.length, len(wantCT), c.length)
		}

		gotCT, err := Encrypt(key, iv, pt)
		if err != nil {
			t.Fatalf("len=%d: Encrypt: %v", c.length, err)
		}
		if !bytes.Equal(gotCT, wantCT) {
			t.Errorf("len=%d ciphertext:\n got  %X\n want %X", c.length, gotCT, wantCT)
		}

		gotPT, err := Decrypt(key, iv, wantCT)
		if err != nil {
			t.Fatalf("len=%d: Decrypt: %v", c.length, err)
		}
		if !bytes.Equal(gotPT, pt) {
			t.Errorf("len=%d plaintext:\n got  %X\n want %X", c.length, gotPT, pt)
		}
	}
}
