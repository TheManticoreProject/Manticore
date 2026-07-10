package kerbcrypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/aes/cts"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// ---------------------------------------------------------------------------
// RFC 2202 HMAC-MD5 / HMAC-SHA-1 known-answer tests
//
// These pin the keyed-MAC primitives underneath both RC4-HMAC (HMAC-MD5 keys
// K1/K2/K3 and the KERB_CHECKSUM_HMAC_MD5 checksum) and AES-CTS-HMAC-SHA1-96
// (the HMAC-SHA1 integrity tag, truncated to 96 bits). A regression in either
// helper would silently break interop; the RFC 2202 vectors catch it.
// ---------------------------------------------------------------------------

// TestHMACMD5RFC2202 validates hmacMD5 against RFC 2202 Section 2 test cases.
func TestHMACMD5RFC2202(t *testing.T) {
	cases := []struct {
		name   string
		key    []byte
		data   []byte
		digest string
	}{
		{"case1", bytes.Repeat([]byte{0x0b}, 16), []byte("Hi There"),
			"92 94 72 7a 36 38 bb 1c 13 f4 8e f8 15 8b fc 9d"},
		{"case2", []byte("Jefe"), []byte("what do ya want for nothing?"),
			"75 0c 78 3e 6a b0 b5 03 ea a8 6e 31 0a 5d b7 38"},
		{"case3", bytes.Repeat([]byte{0xaa}, 16), bytes.Repeat([]byte{0xdd}, 50),
			"56 be 34 52 1d 14 4c 88 db b8 c7 33 f0 e8 b3 f6"},
		{"case6", bytes.Repeat([]byte{0xaa}, 80),
			[]byte("Test Using Larger Than Block-Size Key - Hash Key First"),
			"6b 1a b7 fe 4b d7 bf 8f 0b 62 e6 ce 61 b9 d0 cd"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := hmacMD5(c.key, c.data)
			if want := unhex(t, c.digest); !bytes.Equal(got, want) {
				t.Errorf("hmacMD5 = %X, want %X", got, want)
			}
		})
	}
}

// TestHMACSHA1RFC2202 validates hmacSHA1 against RFC 2202 Section 3 test cases.
// The aes-sha1 enctypes use the first 12 bytes (hmac-sha1-96); the vectors are
// full 20-byte digests, so we compare the whole output.
func TestHMACSHA1RFC2202(t *testing.T) {
	cases := []struct {
		name   string
		key    []byte
		data   []byte
		digest string
	}{
		{"case1", bytes.Repeat([]byte{0x0b}, 20), []byte("Hi There"),
			"b6 17 31 86 55 05 72 64 e2 8b c0 b6 fb 37 8c 8e f1 46 be 00"},
		{"case2", []byte("Jefe"), []byte("what do ya want for nothing?"),
			"ef fc df 6a e5 eb 2f a2 d2 74 16 d5 f1 84 df 9c 25 9a 7c 79"},
		{"case3", bytes.Repeat([]byte{0xaa}, 20), bytes.Repeat([]byte{0xdd}, 50),
			"12 5d 73 42 b9 ac 11 cd 91 a3 9a f4 8a a1 7b 4f 63 f1 75 d3"},
		{"case6", bytes.Repeat([]byte{0xaa}, 80),
			[]byte("Test Using Larger Than Block-Size Key - Hash Key First"),
			"aa 4a e5 e1 52 72 d0 0e 95 70 56 37 ce 8a 3b 55 ed 40 21 12"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := hmacSHA1(c.key, c.data)
			if want := unhex(t, c.digest); !bytes.Equal(got, want) {
				t.Errorf("hmacSHA1 = %X, want %X", got, want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RFC 3962 Appendix B string-to-key known-answer tests (aes128-cts-hmac-sha1-96)
//
// The single iter=1 case is covered in crypto_test.go; this pins the remaining
// Appendix B cases: higher iteration counts, a non-ASCII (UTF-8 g-clef, U+1D11E)
// salt, and a non-ASCII passphrase. The 128-bit reference keys are the ones
// published in RFC 3962.
// ---------------------------------------------------------------------------

func TestAES128StringToKeyRFC3962(t *testing.T) {
	gclef := string([]byte{0xf0, 0x9d, 0x84, 0x9e}) // U+1D11E MUSICAL SYMBOL G CLEF

	cases := []struct {
		name     string
		password string
		salt     string
		iter     int
		want     string
	}{
		{"iter2", "password", "ATHENA.MIT.EDUraeburn", 2, "c6 51 bf 29 e2 30 0a c2 7f a4 69 d6 93 bd da 13"},
		{"iter1200", "password", "ATHENA.MIT.EDUraeburn", 1200, "4c 01 cd 46 d6 32 d0 1e 6d be 23 0a 01 ed 64 2a"},
		{"iter50-gclef-pass", gclef, "EXAMPLE.COMpianist", 50, "f1 49 c1 f2 e1 54 a7 34 52 d4 3e 7f e6 2a 56 e5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params := make([]byte, 4)
			binary.BigEndian.PutUint32(params, uint32(c.iter))
			got, err := StringToKey(iana.ETypeAES128CTSHMACSHA196, c.password, c.salt, params)
			if err != nil {
				t.Fatal(err)
			}
			if want := unhex(t, c.want); !bytes.Equal(got, want) {
				t.Errorf("StringToKey = %X, want %X", got, want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Full-message ciphertext known-answer tests
//
// Neither RFC 4757 (RC4-HMAC) nor RFC 3961/3962 (aes-sha1) publish a full
// confounder+HMAC message vector, so these pin the composition against an
// independent reference computed from the documented primitives with a fixed
// confounder: they catch a divergence in key-usage handling, confounder size,
// block ordering, or MAC truncation that the symmetric round-trip test cannot.
// ---------------------------------------------------------------------------

// TestRC4HMACEncryptReference cross-checks rc4HMACEncrypt against an independent
// RFC 4757 Section 4 reference (K1/K2/K3 HMAC-MD5 chain + RC4), for both an
// unremapped usage (1) and a remapped one (3 -> 8), and confirms Decrypt inverts
// the pinned ciphertext.
func TestRC4HMACEncryptReference(t *testing.T) {
	key := unhex(t, "88 46 f7 ea ee 8f b1 17 ad 06 bd d8 30 b7 58 6c") // NT hash of "password"
	conf := unhex(t, "01 23 45 67 89 ab cd ef")
	pt := []byte("kerberos rc4 vector")

	for _, usage := range []int{1, 3, 9} {
		// Independent reference (RFC 4757 Section 4).
		msgType := make([]byte, 4)
		binary.LittleEndian.PutUint32(msgType, refMapRC4Usage(usage))
		k2 := refHMACMD5(key, msgType)
		data := append(append([]byte{}, conf...), pt...)
		checksum := refHMACMD5(k2, data)
		k3 := refHMACMD5(k2, checksum)
		rc, err := rc4.NewCipher(k3)
		if err != nil {
			t.Fatal(err)
		}
		enc := make([]byte, len(data))
		rc.XORKeyStream(enc, data)
		want := append(append([]byte{}, checksum...), enc...)

		var got []byte
		withConfounder(conf, func() { got, err = Encrypt(iana.ETypeRC4HMAC, key, usage, pt) })
		if err != nil {
			t.Fatalf("usage %d: Encrypt: %v", usage, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("usage %d ciphertext:\n got  %X\n want %X", usage, got, want)
		}

		back, err := Decrypt(iana.ETypeRC4HMAC, key, usage, got)
		if err != nil {
			t.Fatalf("usage %d: Decrypt: %v", usage, err)
		}
		if !bytes.Equal(back, pt) {
			t.Errorf("usage %d: Decrypt = %q, want %q", usage, back, pt)
		}
	}
}

// TestAESEncryptReference cross-checks aesEncrypt for etypes 17 and 18 against an
// independent RFC 3961/3962 reference: Ke = DK(key, usage|0xAA), Ki = DK(key,
// usage|0x55), ciphertext = AES-CTS(Ke, 0, conf||pt) || HMAC-SHA1(Ki, conf||pt)
// truncated to 96 bits. The confounder is fixed so the whole message is
// deterministic, and Decrypt is confirmed to invert it.
func TestAESEncryptReference(t *testing.T) {
	conf := unhex(t, "00 11 22 33 44 55 66 77 88 99 aa bb cc dd ee ff")
	pt := []byte("kerberos aes-cts vector, longer than one block")

	cases := []struct {
		etype  int
		keyHex string
	}{
		{iana.ETypeAES128CTSHMACSHA196, "42 26 3c 6e 89 f4 fc 28 b8 df 68 ee 09 79 9f 15"},
		{iana.ETypeAES256CTSHMACSHA196,
			"fe 69 7b 52 bc 0d 3c e1 44 32 ba 03 6a 92 e6 5b bb 52 28 09 90 a2 fa 27 88 39 98 d7 2a f3 01 61"},
	}

	for _, c := range cases {
		key := unhex(t, c.keyHex)
		keyLen := len(key)
		ke := deriveKey(key, usageConstant(4, 0xAA), keyLen)
		ki := deriveKey(key, usageConstant(4, 0x55), keyLen)
		ptc := append(append([]byte{}, conf...), pt...)
		ctBody, err := cts.Encrypt(ke, make([]byte, 16), ptc)
		if err != nil {
			t.Fatal(err)
		}
		want := append(ctBody, hmacSHA1(ki, ptc)[:12]...)

		var got []byte
		withConfounder(conf, func() { got, err = Encrypt(c.etype, key, 4, pt) })
		if err != nil {
			t.Fatalf("etype %d: Encrypt: %v", c.etype, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("etype %d ciphertext:\n got  %X\n want %X", c.etype, got, want)
		}

		back, err := Decrypt(c.etype, key, 4, got)
		if err != nil {
			t.Fatalf("etype %d: Decrypt: %v", c.etype, err)
		}
		if !bytes.Equal(back, pt) {
			t.Errorf("etype %d: Decrypt = %q, want %q", c.etype, back, pt)
		}
	}
}

// refHMACMD5 is a test-local HMAC-MD5 independent of the package helper.
func refHMACMD5(key, data []byte) []byte {
	m := hmac.New(md5.New, key)
	m.Write(data)
	return m.Sum(nil)
}

// refMapRC4Usage applies the RFC 4757 errata usage remapping independently of
// the package: only usages 3 and 23 are remapped (to 8 and 13).
func refMapRC4Usage(usage int) uint32 {
	switch usage {
	case 3:
		return 8
	case 23:
		return 13
	default:
		return uint32(usage)
	}
}
