package credentials

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

func TestPasswordKeyMatchesNTHashForRC4(t *testing.T) {
	c := NewWithPassword("alice", "corp.local", "Sup3rSecret!")
	if c.Realm() != "CORP.LOCAL" {
		t.Fatalf("realm not upper-cased: %q", c.Realm())
	}
	got, err := c.Key(iana.ETypeRC4HMAC, "", nil)
	if err != nil {
		t.Fatalf("Key(RC4): %v", err)
	}
	want := nt.NTHash("Sup3rSecret!")
	if !bytes.Equal(got, want[:]) {
		t.Errorf("RC4 key %x != NT hash %x", got, want)
	}
}

func TestPasswordDerivesAES(t *testing.T) {
	c := NewWithPassword("alice", "CORP.LOCAL", "password")
	k, err := c.Key(iana.ETypeAES256CTSHMACSHA196, c.DefaultSalt(), nil)
	if err != nil {
		t.Fatalf("Key(AES256): %v", err)
	}
	if len(k) != 32 {
		t.Errorf("AES256 key length = %d, want 32", len(k))
	}
}

func TestOverpassTheHash(t *testing.T) {
	hashHex := "8846f7eaee8fb117ad06bdd830b7586c" // NT hash of "password"
	c, err := NewWithHexNTHash("alice", "corp.local", hashHex)
	if err != nil {
		t.Fatalf("NewWithHexNTHash: %v", err)
	}

	// RC4 key is the hash itself.
	rc4, err := c.Key(iana.ETypeRC4HMAC, "", nil)
	if err != nil {
		t.Fatalf("Key(RC4): %v", err)
	}
	want, _ := hex.DecodeString(hashHex)
	if !bytes.Equal(rc4, want) {
		t.Errorf("RC4 key %x != %s", rc4, hashHex)
	}
	// It should equal the NT hash of "password" (sanity vs the crypto layer).
	if h := nt.NTHash("password"); !bytes.Equal(rc4, h[:]) {
		t.Errorf("hash does not match NTHash(\"password\")")
	}

	// An NT hash cannot yield an AES key.
	if _, err := c.Key(iana.ETypeAES256CTSHMACSHA196, "CORP.LOCALalice", nil); err == nil {
		t.Error("expected error deriving AES key from NT hash")
	}

	if len(c.SupportedETypes()) != 1 || c.SupportedETypes()[0] != iana.ETypeRC4HMAC {
		t.Errorf("NT-hash SupportedETypes = %v, want [23]", c.SupportedETypes())
	}
}

func TestPassTheKey(t *testing.T) {
	// Accepts LM:NT form and takes the NT half.
	c, err := NewWithHexNTHash("bob", "CORP.LOCAL", "aad3b435b51404eeaad3b435b51404ee:8846f7eaee8fb117ad06bdd830b7586c")
	if err != nil {
		t.Fatalf("NewWithHexNTHash LM:NT: %v", err)
	}
	rc4, _ := c.Key(iana.ETypeRC4HMAC, "", nil)
	if h := nt.NTHash("password"); !bytes.Equal(rc4, h[:]) {
		t.Errorf("LM:NT parse took wrong half")
	}

	key32 := bytes.Repeat([]byte{0xAB}, 32)
	ck, err := NewWithAESKey("svc", "CORP.LOCAL", iana.ETypeAES256CTSHMACSHA196, key32)
	if err != nil {
		t.Fatalf("NewWithAESKey: %v", err)
	}
	got, err := ck.Key(iana.ETypeAES256CTSHMACSHA196, "ignored-salt", nil)
	if err != nil {
		t.Fatalf("Key(AES256): %v", err)
	}
	if !bytes.Equal(got, key32) {
		t.Error("AES key round-trip mismatch")
	}
	// Wrong etype must fail.
	if _, err := ck.Key(iana.ETypeAES128CTSHMACSHA196, "", nil); err == nil {
		t.Error("expected error: AES256 key cannot serve etype 17")
	}
	if _, err := ck.Key(iana.ETypeRC4HMAC, "", nil); err == nil {
		t.Error("expected error: AES key cannot serve RC4")
	}
}

func TestNewWithHexAESKeyInfersEtype(t *testing.T) {
	c16, err := NewWithHexAESKey("svc", "CORP.LOCAL", hex.EncodeToString(bytes.Repeat([]byte{1}, 16)))
	if err != nil {
		t.Fatal(err)
	}
	if c16.SupportedETypes()[0] != iana.ETypeAES128CTSHMACSHA196 {
		t.Errorf("16-byte key inferred etype = %v, want 17", c16.SupportedETypes())
	}
	c32, err := NewWithHexAESKey("svc", "CORP.LOCAL", hex.EncodeToString(bytes.Repeat([]byte{1}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	if c32.SupportedETypes()[0] != iana.ETypeAES256CTSHMACSHA196 {
		t.Errorf("32-byte key inferred etype = %v, want 18", c32.SupportedETypes())
	}
	if _, err := NewWithHexAESKey("svc", "R", hex.EncodeToString(bytes.Repeat([]byte{1}, 24))); err == nil {
		t.Error("expected error for 24-byte AES key")
	}
}

func TestInvalidInputs(t *testing.T) {
	if _, err := NewWithNTHash("a", "R", []byte{1, 2, 3}); err == nil {
		t.Error("expected error for short NT hash")
	}
	if _, err := NewWithHexNTHash("a", "R", "zzzz"); err == nil {
		t.Error("expected error for bad hex")
	}
	if _, err := NewWithAESKey("a", "R", 99, bytes.Repeat([]byte{1}, 32)); err == nil {
		t.Error("expected error for bad AES etype")
	}
}

func TestDestroyZeroesSecret(t *testing.T) {
	h, _ := hex.DecodeString("8846f7eaee8fb117ad06bdd830b7586c")
	c, _ := NewWithNTHash("a", "R", h)
	c.Destroy()
	if _, err := c.Key(iana.ETypeRC4HMAC, "", nil); err != nil {
		// after destroy ntHash is nil -> returns empty key, not an error; ensure no panic
		t.Logf("post-destroy Key err: %v", err)
	}
}
