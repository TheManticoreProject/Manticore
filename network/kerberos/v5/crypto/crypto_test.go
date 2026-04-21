package kerbcrypto

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/network/kerberos/messages"
)

// ---------------------------------------------------------------------------
// RC4-HMAC
// ---------------------------------------------------------------------------

// TestStringToKeyRC4 verifies that StringToKey for RC4-HMAC returns the NT hash.
//
// Well-known vectors:
//   NT hash of "password" = 8846f7eaee8fb117ad06bdd830b7586c
//   NT hash of "Password" = a4f49c406510bdcab6824ee7c30fd852
func TestStringToKeyRC4(t *testing.T) {
	// Verify that StringToKey returns the same value as nt.NTHash for any password.
	for _, password := range []string{"password", "Password", "abc123"} {
		want := nt.NTHash(password)
		got, err := StringToKey(messages.ETypeRC4HMAC, password, "", nil)
		if err != nil {
			t.Fatalf("StringToKey RC4(%q): unexpected error: %v", password, err)
		}
		if !bytes.Equal(got, want[:]) {
			t.Errorf("StringToKey RC4(%q): got %x, want %x", password, got, want)
		}
	}

	// Cross-check against well-known NT hash value for "password" (lowercase).
	const knownPassword = "password"
	const knownHex = "8846f7eaee8fb117ad06bdd830b7586c"
	got, _ := StringToKey(messages.ETypeRC4HMAC, knownPassword, "", nil)
	knownBytes, _ := hex.DecodeString(knownHex)
	if !bytes.Equal(got, knownBytes) {
		t.Errorf("StringToKey RC4(%q): got %x, want known vector %s", knownPassword, got, knownHex)
	}
}

// TestRC4HMACRoundtrip verifies that encrypting then decrypting recovers the original plaintext.
func TestRC4HMACRoundtrip(t *testing.T) {
	key, err := StringToKey(messages.ETypeRC4HMAC, "Password", "", nil)
	if err != nil {
		t.Fatalf("StringToKey: %v", err)
	}

	tests := []struct {
		name      string
		usage     int
		plaintext string
	}{
		{"pa-enc-timestamp", KeyUsageASReqPAEncTimestamp, "hello kerberos"},
		{"as-rep-enc-part", KeyUsageASRepEncPart, "secret data 1234567890"},
		{"empty", KeyUsageTGSRepEncSessionKey, ""},
		{"long", KeyUsageAPReqAuthen, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := Encrypt(messages.ETypeRC4HMAC, key, tc.usage, []byte(tc.plaintext))
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			plaintext, err := Decrypt(messages.ETypeRC4HMAC, key, tc.usage, ciphertext)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if !bytes.Equal(plaintext, []byte(tc.plaintext)) {
				t.Errorf("roundtrip mismatch: got %q, want %q", plaintext, tc.plaintext)
			}
		})
	}
}

// TestRC4HMACDecryptTamperedMAC verifies that a tampered MAC causes an integrity error.
func TestRC4HMACDecryptTamperedMAC(t *testing.T) {
	key, _ := StringToKey(messages.ETypeRC4HMAC, "Password", "", nil)
	ct, err := Encrypt(messages.ETypeRC4HMAC, key, KeyUsageASReqPAEncTimestamp, []byte("test"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	// Flip a byte in the MAC
	ct[0] ^= 0xFF
	_, err = Decrypt(messages.ETypeRC4HMAC, key, KeyUsageASReqPAEncTimestamp, ct)
	if err == nil {
		t.Error("expected integrity error on tampered MAC, got nil")
	}
}

// TestRC4HMACDecryptTooShort verifies that a too-short ciphertext returns an error.
func TestRC4HMACDecryptTooShort(t *testing.T) {
	key, _ := StringToKey(messages.ETypeRC4HMAC, "Password", "", nil)
	_, err := Decrypt(messages.ETypeRC4HMAC, key, 1, []byte("tooshort"))
	if err == nil {
		t.Error("expected error for short ciphertext, got nil")
	}
}

// ---------------------------------------------------------------------------
// AES-128 and AES-256
// ---------------------------------------------------------------------------

// TestStringToKeyAES128KnownVector tests AES-128 key derivation against the
// RFC 3962 Appendix B test vector:
//   password = "password"
//   salt     = "ATHENA.MIT.EDUraeburn"
//   iter     = 1
//   key      = 42263c6e89f4fc28b8df68ee09799f15
func TestStringToKeyAES128KnownVector(t *testing.T) {
	// S2KParams encoding of iter_count=1 as 4-byte big-endian
	params := []byte{0, 0, 0, 1}
	got, err := StringToKey(messages.ETypeAES128CTSHMACSHA196, "password", "ATHENA.MIT.EDUraeburn", params)
	if err != nil {
		t.Fatalf("StringToKey AES-128: %v", err)
	}
	const want = "42263c6e89f4fc28b8df68ee09799f15"
	if hex.EncodeToString(got) != want {
		t.Errorf("AES-128 StringToKey: got %x, want %s", got, want)
	}
}

// TestStringToKeyAES256KnownVector tests AES-256 key derivation against the
// RFC 3962 Appendix B test vector:
//   password = "password"
//   salt     = "ATHENA.MIT.EDUraeburn"
//   iter     = 1
//   key      = fe697b52bc0d3ce14432ba036a92e65bbb52280990a2fa27883998d72af30161
func TestStringToKeyAES256KnownVector(t *testing.T) {
	params := []byte{0, 0, 0, 1}
	got, err := StringToKey(messages.ETypeAES256CTSHMACSHA196, "password", "ATHENA.MIT.EDUraeburn", params)
	if err != nil {
		t.Fatalf("StringToKey AES-256: %v", err)
	}
	const want = "fe697b52bc0d3ce14432ba036a92e65bbb52280990a2fa27883998d72af30161"
	if hex.EncodeToString(got) != want {
		t.Errorf("AES-256 StringToKey: got %x, want %s", got, want)
	}
}

// TestAES128Roundtrip verifies encrypt/decrypt roundtrip for AES-128.
func TestAES128Roundtrip(t *testing.T) {
	key, err := StringToKey(messages.ETypeAES128CTSHMACSHA196, "Password", "REALM.EXAMPLEuser", nil)
	if err != nil {
		t.Fatalf("StringToKey: %v", err)
	}
	aesRoundtrip(t, messages.ETypeAES128CTSHMACSHA196, key)
}

// TestAES256Roundtrip verifies encrypt/decrypt roundtrip for AES-256.
func TestAES256Roundtrip(t *testing.T) {
	key, err := StringToKey(messages.ETypeAES256CTSHMACSHA196, "Password", "REALM.EXAMPLEuser", nil)
	if err != nil {
		t.Fatalf("StringToKey: %v", err)
	}
	aesRoundtrip(t, messages.ETypeAES256CTSHMACSHA196, key)
}

func aesRoundtrip(t *testing.T, etype int, key []byte) {
	t.Helper()
	tests := []struct {
		name      string
		usage     int
		plaintext string
	}{
		{"pa-enc-timestamp", KeyUsageASReqPAEncTimestamp, "hello kerberos AES"},
		{"as-rep-enc-part", KeyUsageASRepEncPart, "secret AES data 1234567890"},
		{"single-block", KeyUsageTGSRepEncSessionKey, "exactly16bytess!"},
		{"multi-block", KeyUsageAPReqAuthen, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{"non-aligned", KeyUsageTGSRepEncSubSessionKey, "seventeen bytes!!"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := Encrypt(etype, key, tc.usage, []byte(tc.plaintext))
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			plaintext, err := Decrypt(etype, key, tc.usage, ciphertext)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if !bytes.Equal(plaintext, []byte(tc.plaintext)) {
				t.Errorf("roundtrip mismatch: got %q, want %q", plaintext, tc.plaintext)
			}
		})
	}
}

// TestAES128DecryptTamperedMAC verifies that a tampered MAC returns an integrity error.
func TestAES128DecryptTamperedMAC(t *testing.T) {
	key, _ := StringToKey(messages.ETypeAES128CTSHMACSHA196, "Password", "REALM.EXAMPLEuser", nil)
	ct, err := Encrypt(messages.ETypeAES128CTSHMACSHA196, key, KeyUsageASReqPAEncTimestamp, []byte("test data here!"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	// Flip the last byte of the MAC (last 12 bytes of output)
	ct[len(ct)-1] ^= 0xFF
	_, err = Decrypt(messages.ETypeAES128CTSHMACSHA196, key, KeyUsageASReqPAEncTimestamp, ct)
	if err == nil {
		t.Error("expected integrity error on tampered MAC, got nil")
	}
}

// TestAESDecryptTooShort verifies that too-short ciphertext returns an error.
func TestAESDecryptTooShort(t *testing.T) {
	key, _ := StringToKey(messages.ETypeAES128CTSHMACSHA196, "Password", "REALM.EXAMPLEuser", nil)
	_, err := Decrypt(messages.ETypeAES128CTSHMACSHA196, key, 1, []byte("tooshort"))
	if err == nil {
		t.Error("expected error for short ciphertext, got nil")
	}
}

// ---------------------------------------------------------------------------
// KeyLen
// ---------------------------------------------------------------------------

func TestKeyLen(t *testing.T) {
	tests := []struct {
		etype int
		want  int
	}{
		{messages.ETypeRC4HMAC, 16},
		{messages.ETypeAES128CTSHMACSHA196, 16},
		{messages.ETypeAES256CTSHMACSHA196, 32},
		{99, 0}, // unsupported
	}
	for _, tc := range tests {
		got := KeyLen(tc.etype)
		if got != tc.want {
			t.Errorf("KeyLen(%d): got %d, want %d", tc.etype, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Unsupported etype errors
// ---------------------------------------------------------------------------

func TestUnsupportedEType(t *testing.T) {
	_, err := StringToKey(99, "Password", "SALT", nil)
	if err == nil {
		t.Error("StringToKey: expected error for unsupported etype")
	}
	_, err = Encrypt(99, []byte{0}, 1, []byte("x"))
	if err == nil {
		t.Error("Encrypt: expected error for unsupported etype")
	}
	_, err = Decrypt(99, []byte{0}, 1, []byte("x"))
	if err == nil {
		t.Error("Decrypt: expected error for unsupported etype")
	}
}
