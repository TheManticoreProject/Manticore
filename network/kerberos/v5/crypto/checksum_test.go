package kerbcrypto

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// TestChecksumRoundTrip exercises GetChecksum/VerifyChecksum for the checksum
// types that lack published byte-for-byte vectors in the local RFC set
// (hmac-sha1-96-aes* from RFC 3961 and the RC4 KERB_CHECKSUM_HMAC_MD5 from
// RFC 4757), asserting deterministic output, the spec-mandated length, and that
// verification accepts a good checksum and rejects a tampered one. The AES-SHA2
// checksums are separately validated against RFC 8009 vectors in
// TestAES2ChecksumVectors.
func TestChecksumRoundTrip(t *testing.T) {
	msg := []byte("The quick brown fox jumps over the lazy dog")

	cases := []struct {
		name      string
		cksumType int
		keyLen    int
		wantLen   int
	}{
		{"hmac-sha1-96-aes128", iana.CksumTypeHMACSHA196AES128, 16, 12},
		{"hmac-sha1-96-aes256", iana.CksumTypeHMACSHA196AES256, 32, 12},
		{"hmac-md5-rc4", iana.CksumTypeHMACMD5, 16, 16},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key := make([]byte, c.keyLen)
			for i := range key {
				key[i] = byte(i + 1)
			}
			usage := iana.KeyUsageTGSReqAuthCksum

			sum, err := GetChecksum(c.cksumType, key, usage, msg)
			if err != nil {
				t.Fatalf("GetChecksum: %v", err)
			}
			if len(sum) != c.wantLen {
				t.Fatalf("checksum length = %d, want %d", len(sum), c.wantLen)
			}

			// Deterministic.
			sum2, _ := GetChecksum(c.cksumType, key, usage, msg)
			if string(sum) != string(sum2) {
				t.Fatal("checksum is not deterministic")
			}

			if !VerifyChecksum(c.cksumType, key, usage, msg, sum) {
				t.Fatal("VerifyChecksum rejected a valid checksum")
			}

			// Tamper: flip a message byte → verification must fail.
			bad := append([]byte{}, msg...)
			bad[0] ^= 0x01
			if VerifyChecksum(c.cksumType, key, usage, bad, sum) {
				t.Fatal("VerifyChecksum accepted a checksum over tampered data")
			}

			// Wrong usage → different checksum (key-usage separation).
			other, _ := GetChecksum(c.cksumType, key, iana.KeyUsageAPReqAuthCksum, msg)
			if string(other) == string(sum) {
				t.Fatal("checksum did not depend on key usage")
			}
		})
	}
}

// TestChecksumUnsupported confirms an unknown checksum type is reported.
func TestChecksumUnsupported(t *testing.T) {
	if _, err := GetChecksum(0x7fffffff, make([]byte, 16), 1, []byte("x")); err == nil {
		t.Fatal("expected error for unsupported checksum type")
	}
	if VerifyChecksum(0x7fffffff, make([]byte, 16), 1, []byte("x"), []byte("y")) {
		t.Fatal("VerifyChecksum should be false for unsupported type")
	}
}
