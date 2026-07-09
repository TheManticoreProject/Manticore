package kerbcrypto

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// TestPRFAESSHA1KnownAnswer checks the RFC 3962 AES-SHA1 PRF against
// independently computed reference values.
func TestPRFAESSHA1KnownAnswer(t *testing.T) {
	cases := []struct {
		name  string
		etype int
		key   string
		input string
		want  string
	}{
		{"aes256", iana.ETypeAES256CTSHMACSHA196,
			"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
			"test-input", "d7b84749812fd1726731ab893c467309"},
		{"aes128", iana.ETypeAES128CTSHMACSHA196,
			"0102030405060708090a0b0c0d0e0f10",
			"test-input", "3744a3b414aa5b406927ec729dc879d6"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PRF(tc.etype, mustHex(t, tc.key), []byte(tc.input))
			if err != nil {
				t.Fatalf("PRF: %v", err)
			}
			if want := mustHex(t, tc.want); !bytes.Equal(got, want) {
				t.Fatalf("PRF mismatch:\n got  %x\n want %x", got, want)
			}
		})
	}
}

// TestPRFRC4KnownAnswer checks the arcfour-hmac PRF (HMAC-SHA1) output.
func TestPRFRC4KnownAnswer(t *testing.T) {
	got, err := PRF(iana.ETypeRC4HMAC, mustHex(t, "0102030405060708090a0b0c0d0e0f10"), []byte("test-input"))
	if err != nil {
		t.Fatalf("PRF: %v", err)
	}
	want := mustHex(t, "84fb47c2594710701cccfd586c17084a0c2180fe")
	if !bytes.Equal(got, want) {
		t.Fatalf("RC4 PRF mismatch:\n got  %x\n want %x", got, want)
	}
}

// TestPRFAESSHA2KnownAnswer checks the RFC 8009 §8 PRF test vectors for the
// AES-SHA2 enctypes.
func TestPRFAESSHA2KnownAnswer(t *testing.T) {
	cases := []struct {
		name  string
		etype int
		key   string
		want  string
	}{
		{"aes128-sha256", iana.ETypeAES128CTSHMACSHA256,
			"3705d96080c17728a0e800eab6e0d23c",
			"9d188616f63852fe86915bb840b4a886ff3e6bb0f819b49b893393d393854295"},
		{"aes256-sha384", iana.ETypeAES256CTSHMACSHA384,
			"6d404d37faf79f9df0d33568d3206698" + "00eb4836472ea8a026d16b7182460c52",
			"9801f69a368c2bf675e59521e177d9a0" + "7f67efe1cfde8d3c8d6f6a0256e3b17d" + "b3c1b62ad1b8553360d17367eb1514d2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PRF(tc.etype, mustHex(t, tc.key), []byte("test"))
			if err != nil {
				t.Fatalf("PRF: %v", err)
			}
			if want := mustHex(t, tc.want); !bytes.Equal(got, want) {
				t.Fatalf("PRF mismatch:\n got  %x\n want %x", got, want)
			}
		})
	}
}

// TestKRBFXCF2KnownAnswer verifies the RFC 6113 KRB-FX-CF2 key combination
// against reference values, including the cross-enctype case (K1 and K2 of
// different enctypes; the result adopts K1's enctype).
func TestKRBFXCF2KnownAnswer(t *testing.T) {
	cases := []struct {
		name             string
		k1               string
		k1Etype          int
		k2               string
		k2Etype          int
		pepper1, pepper2 string
		want             string
	}{
		{
			"aes256-subkeyarmor",
			"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20", iana.ETypeAES256CTSHMACSHA196,
			"2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40", iana.ETypeAES256CTSHMACSHA196,
			"subkeyarmor", "ticketarmor",
			"18b1ecfa39a1bf373bcbc098aae658fe35082a00d7d6660e6cb6fd6138404d21",
		},
		{
			"aes128-subkeyarmor",
			"0102030405060708090a0b0c0d0e0f10", iana.ETypeAES128CTSHMACSHA196,
			"1112131415161718191a1b1c1d1e1f20", iana.ETypeAES128CTSHMACSHA196,
			"subkeyarmor", "ticketarmor",
			"84e6a11aecdec85598b78b8c37df3f1f",
		},
		{
			"rc4-challenge",
			"0102030405060708090a0b0c0d0e0f10", iana.ETypeRC4HMAC,
			"1112131415161718191a1b1c1d1e1f20", iana.ETypeRC4HMAC,
			"clientchallengearmor", "challengelongterm",
			"a7fa55719b5140c8693af782dd1775ef",
		},
		{
			"aes256-strengthenkey",
			"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20", iana.ETypeAES256CTSHMACSHA196,
			"2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40", iana.ETypeAES256CTSHMACSHA196,
			"strengthenkey", "replykey",
			"eb67d71c98f3dfdfe301e7352f8dfcb11516c4fe0661cdb2995b9d67b21bf08a",
		},
		{
			"cross-256-128",
			"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20", iana.ETypeAES256CTSHMACSHA196,
			"0102030405060708090a0b0c0d0e0f10", iana.ETypeAES128CTSHMACSHA196,
			"clientchallengearmor", "challengelongterm",
			"075d94a2a0bf5f0b69a18f216b0b2d7b1679441568eae2cfadbc81d7a400ecea",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, etype, err := KRBFXCF2(mustHex(t, tc.k1), tc.k1Etype, mustHex(t, tc.k2), tc.k2Etype, tc.pepper1, tc.pepper2)
			if err != nil {
				t.Fatalf("KRBFXCF2: %v", err)
			}
			if etype != tc.k1Etype {
				t.Fatalf("result etype = %d, want K1 etype %d", etype, tc.k1Etype)
			}
			if want := mustHex(t, tc.want); !bytes.Equal(got, want) {
				t.Fatalf("KRB-FX-CF2 mismatch:\n got  %x\n want %x", got, want)
			}
		})
	}
}
