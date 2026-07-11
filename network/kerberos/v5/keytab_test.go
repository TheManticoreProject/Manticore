package kerberos

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/keytab"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// keytabBytes marshals a keytab holding the given (etype, key) entries for
// alice@CORP.LOCAL, all at the same kvno so selection turns on enctype strength
// rather than key version.
func keytabBytes(t *testing.T, entries map[int][]byte) []byte {
	t.Helper()
	kt := keytab.New()
	princ := keytab.Principal{
		NameType:   iana.NameTypePrincipal,
		Realm:      testRealm,
		Components: []string{"alice"},
	}
	for etype, key := range entries {
		kt.Add(princ, etype, key, 1)
	}
	data, err := kt.Marshal()
	if err != nil {
		t.Fatalf("keytab.Marshal: %v", err)
	}
	return data
}

// TestWithKeytabBytesPrefersAES256 confirms WithKeytabBytes selects the strongest
// usable enctype (AES256 > AES128 > RC4) and configures the client credential
// with it.
func TestWithKeytabBytesPrefersAES256(t *testing.T) {
	data := keytabBytes(t, map[int][]byte{
		iana.ETypeRC4HMAC:             bytes.Repeat([]byte{0x11}, 16),
		iana.ETypeAES128CTSHMACSHA196: bytes.Repeat([]byte{0x22}, 16),
		iana.ETypeAES256CTSHMACSHA196: bytes.Repeat([]byte{0x33}, 32),
	})

	c := NewClient("alice", "corp.local", "10.0.0.1")
	if err := c.WithKeytabBytes(data); err != nil {
		t.Fatalf("WithKeytabBytes: %v", err)
	}
	et := c.cred.SupportedETypes()
	if len(et) != 1 || et[0] != iana.ETypeAES256CTSHMACSHA196 {
		t.Errorf("selected credential etypes = %v, want [AES256]", et)
	}
}

// TestWithKeytabBytesFallsBackToRC4 confirms that when no AES key is present the
// RC4-HMAC key is adopted as an overpass-the-hash (NT-hash) credential.
func TestWithKeytabBytesFallsBackToRC4(t *testing.T) {
	data := keytabBytes(t, map[int][]byte{
		iana.ETypeRC4HMAC: bytes.Repeat([]byte{0x11}, 16),
	})

	c := NewClient("alice", "corp.local", "10.0.0.1")
	if err := c.WithKeytabBytes(data); err != nil {
		t.Fatalf("WithKeytabBytes: %v", err)
	}
	et := c.cred.SupportedETypes()
	if len(et) != 1 || et[0] != iana.ETypeRC4HMAC {
		t.Errorf("selected credential etypes = %v, want [RC4]", et)
	}
}

// TestWithKeytabForcesEType confirms an explicit etype overrides the
// strongest-first preference (WithKeytab with etype > 0).
func TestWithKeytabForcesEType(t *testing.T) {
	data := keytabBytes(t, map[int][]byte{
		iana.ETypeAES256CTSHMACSHA196: bytes.Repeat([]byte{0x33}, 32),
		iana.ETypeAES128CTSHMACSHA196: bytes.Repeat([]byte{0x22}, 16),
	})
	kt, err := keytab.Unmarshal(data)
	if err != nil {
		t.Fatalf("keytab.Unmarshal: %v", err)
	}

	c := NewClient("alice", "corp.local", "10.0.0.1")
	if err := c.WithKeytab(kt, iana.ETypeAES128CTSHMACSHA196); err != nil {
		t.Fatalf("WithKeytab(AES128): %v", err)
	}
	et := c.cred.SupportedETypes()
	if len(et) != 1 || et[0] != iana.ETypeAES128CTSHMACSHA196 {
		t.Errorf("forced credential etypes = %v, want [AES128]", et)
	}
}

// TestWithKeytabRejectsUnsupportedEType confirms an entry whose only key is an
// AES-SHA2 (or DES) enctype — which the credentials layer cannot key for
// authentication — is rejected rather than silently accepted.
func TestWithKeytabRejectsUnsupportedEType(t *testing.T) {
	for _, etype := range []int{iana.ETypeAES256CTSHMACSHA384, iana.ETypeDESCBCMD5} {
		data := keytabBytes(t, map[int][]byte{etype: bytes.Repeat([]byte{0x44}, 32)})
		c := NewClient("alice", "corp.local", "10.0.0.1")
		if err := c.WithKeytabBytes(data); err == nil {
			t.Errorf("etype %d: expected WithKeytabBytes to reject an unusable enctype", etype)
		}
	}
}

// TestWithKeytabNoMatch confirms a keytab with no entry for the client principal
// is reported as an error.
func TestWithKeytabNoMatch(t *testing.T) {
	data := keytabBytes(t, map[int][]byte{iana.ETypeAES256CTSHMACSHA196: bytes.Repeat([]byte{0x33}, 32)})
	c := NewClient("bob", "corp.local", "10.0.0.1") // keytab holds alice, not bob
	if err := c.WithKeytabBytes(data); err == nil {
		t.Fatal("expected an error selecting a keytab entry for a non-present principal")
	}
}
