package structures

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestNetlogonCredentialWire verifies the 8-octet fixed-array credential marshals verbatim
// with no NDR framing when carried as a struct field (its only use on the wire).
func TestNetlogonCredentialWire(t *testing.T) {
	type wrap struct{ C NETLOGON_CREDENTIAL }
	c := NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8}
	raw, err := ndr.Marshal(&wrap{C: c})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(raw, c[:]) {
		t.Fatalf("credential wire = %x, want %x", raw, c[:])
	}
}

// TestNetlogonAuthenticatorWire verifies the authenticator is a 12-octet structure (8-byte
// credential + 4-byte timestamp) with the timestamp little-endian at offset 8.
func TestNetlogonAuthenticatorWire(t *testing.T) {
	a := NETLOGON_AUTHENTICATOR{Credential: NETLOGON_CREDENTIAL{1, 1, 1, 1, 1, 1, 1, 1}, Timestamp: 0x04030201}
	raw, err := ndr.Marshal(&a)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{1, 1, 1, 1, 1, 1, 1, 1, 0x01, 0x02, 0x03, 0x04}
	if !bytes.Equal(raw, want) {
		t.Fatalf("authenticator wire = %x, want %x", raw, want)
	}
}

// TestTrustPasswordWire verifies an all-zero NL_TRUST_PASSWORD marshals to 516 zero octets
// (512-byte buffer + 4-byte length), the empty-password form.
func TestTrustPasswordWire(t *testing.T) {
	raw, err := ndr.Marshal(&NL_TRUST_PASSWORD{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 516 {
		t.Fatalf("trust password length = %d, want 516", len(raw))
	}
	for i, b := range raw {
		if b != 0 {
			t.Fatalf("octet %d = 0x%02x, want 0", i, b)
		}
	}
}
