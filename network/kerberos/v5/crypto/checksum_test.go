package kerbcrypto

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
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

// ---------------------------------------------------------------------------
// Checksum composition known-answer tests (cksumtypes 15, 16, -138)
//
// Neither RFC 3961/3962 (hmac-sha1-96-aes128/256) nor RFC 4757 (KERB_CHECKSUM_
// HMAC_MD5) publishes a byte-for-byte checksum vector the way RFC 8009 Appendix
// A does for cksumtypes 19/20 (see TestAES2ChecksumVectors). There is no wire-
// published composite vector for these three types. Instead, each underlying
// PRIMITIVE is already pinned to its own RFC:
//
//   - HMAC-SHA1 and HMAC-MD5 -> RFC 2202 (TestHMACSHA1RFC2202 / TestHMACMD5RFC2202)
//   - the Kc = DK(base, usage|0x99) derivation -> RFC 3962 App B string-to-key
//     (TestStringToKeyRFC3962AppendixB, which exercises the same DK path)
//
// so what remains untested is the COMPOSITION itself: for cksumtype 15/16 the
// get_mic of RFC 3961 Section 5.4 (Kc purpose octet 0x99, HMAC-SHA1 truncated to
// 96 bits); for cksumtype -138 the RFC 4757 Section 4 layering (the fixed
// "signaturekey\0" derivation, the inner MD5(msg-type || data), and the outer
// HMAC-MD5). These tests recompute the composite from the documented steps with
// independent, test-local primitives and pin GetChecksum byte-for-byte against
// it, catching a purpose-octet, ordering, or truncation regression that the
// round-trip/tamper test in TestChecksumRoundTrip cannot.
// ---------------------------------------------------------------------------

// TestAESSHA1ChecksumComposition pins cksumtypes 15 and 16 against an independent
// recomputation of the RFC 3961 Section 5.4 get_mic: HMAC-SHA1(DK(base,
// usage|0x99), data) truncated to the first 12 bytes (96 bits).
func TestAESSHA1ChecksumComposition(t *testing.T) {
	msg := []byte("The quick brown fox jumps over the lazy dog")
	usage := iana.KeyUsageTGSReqAuthCksum

	cases := []struct {
		name      string
		cksumType int
		keyLen    int
	}{
		{"hmac-sha1-96-aes128", iana.CksumTypeHMACSHA196AES128, 16},
		{"hmac-sha1-96-aes256", iana.CksumTypeHMACSHA196AES256, 32},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base := make([]byte, c.keyLen)
			for i := range base {
				base[i] = byte(i + 1)
			}

			// Independent reference per RFC 3961 Section 5.4.
			kc := deriveKey(base, usageConstant(usage, 0x99), c.keyLen)
			want := hmacSHA1(kc, msg)[:12]

			got, err := GetChecksum(c.cksumType, base, usage, msg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("checksum:\n got  %X\n want %X", got, want)
			}
		})
	}
}

// TestRC4HMACMD5ChecksumComposition pins cksumtype -138 (KERB_CHECKSUM_HMAC_MD5)
// against an independent recomputation of RFC 4757 Section 4:
//
//	Ksign  = HMAC-MD5(key, "signaturekey\0")
//	tmp    = MD5(T || data)   where T is the little-endian mapped MS message type
//	CHKSUM = HMAC-MD5(Ksign, tmp)
//
// T is built here from a test-local usage remapping (RFC 4757 errata) so the
// reference does not share the package's own message-type encoder.
func TestRC4HMACMD5ChecksumComposition(t *testing.T) {
	msg := []byte("The quick brown fox jumps over the lazy dog")
	key := unhex(t, "88 46 f7 ea ee 8f b1 17 ad 06 bd d8 30 b7 58 6c") // NT hash of "password"

	for _, usage := range []int{iana.KeyUsageTGSReqAuthCksum, iana.KeyUsageAPReqAuthCksum} {
		// Independent reference per RFC 4757 Section 4.
		ksign := refHMACMD5(key, []byte("signaturekey\x00"))
		tb := make([]byte, 4)
		binary.LittleEndian.PutUint32(tb, refMapRC4Usage(usage))
		h := md5.New()
		h.Write(tb)
		h.Write(msg)
		want := refHMACMD5(ksign, h.Sum(nil))

		got, err := GetChecksum(iana.CksumTypeHMACMD5, key, usage, msg)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("usage %d checksum:\n got  %X\n want %X", usage, got, want)
		}
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
