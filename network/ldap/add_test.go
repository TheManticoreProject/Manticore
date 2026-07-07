package ldap_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
)

func TestNewAddRequest(t *testing.T) {
	dn := "cn=John Doe,dc=example,dc=com"
	req := ldap.NewAddRequest(dn)

	if req.DistinguishedName != dn {
		t.Errorf("DistinguishedName = %q; want %q", req.DistinguishedName, dn)
	}
	if req.Attributes == nil {
		t.Errorf("Attributes = nil; want a non-nil empty slice")
	}
	if len(req.Attributes) != 0 {
		t.Errorf("len(Attributes) = %d; want 0", len(req.Attributes))
	}
	if len(req.Controls) != 0 {
		t.Errorf("len(Controls) = %d; want 0", len(req.Controls))
	}
}

func TestAddRequestAttributeOrderAndValues(t *testing.T) {
	req := ldap.NewAddRequest("cn=John Doe,dc=example,dc=com")
	req.Attribute("objectClass", []string{"top", "person"})
	req.Attribute("cn", []string{"John Doe"})
	req.Attribute("sn", []string{"Doe"})

	if len(req.Attributes) != 3 {
		t.Fatalf("len(Attributes) = %d; want 3", len(req.Attributes))
	}

	// Attributes must be preserved in insertion order (objectClass first by convention).
	wantTypes := []string{"objectClass", "cn", "sn"}
	for i, want := range wantTypes {
		if req.Attributes[i].Type != want {
			t.Errorf("Attributes[%d].Type = %q; want %q", i, req.Attributes[i].Type, want)
		}
	}

	if len(req.Attributes[0].Vals) != 2 ||
		req.Attributes[0].Vals[0] != "top" || req.Attributes[0].Vals[1] != "person" {
		t.Errorf("objectClass Vals = %v; want [top person]", req.Attributes[0].Vals)
	}
}

func TestAddRequestBinaryValueRoundTrip(t *testing.T) {
	// Binary attribute values are carried as string(bytes); the bytes must survive verbatim,
	// including embedded NUL and high bytes.
	raw := []byte{0x00, 0x01, 0xFF, 0x7F, 0x80, 0x00, 0xAB}

	req := ldap.NewAddRequest("cn=binary,dc=example,dc=com")
	req.Attribute("objectSid", []string{string(raw)})

	if len(req.Attributes) != 1 || len(req.Attributes[0].Vals) != 1 {
		t.Fatalf("unexpected attribute shape: %+v", req.Attributes)
	}

	got := []byte(req.Attributes[0].Vals[0])
	if len(got) != len(raw) {
		t.Fatalf("round-tripped length = %d; want %d", len(got), len(raw))
	}
	for i := range raw {
		if got[i] != raw[i] {
			t.Errorf("byte %d = 0x%02X; want 0x%02X", i, got[i], raw[i])
		}
	}
}

func TestAddRequestAddControl(t *testing.T) {
	req := ldap.NewAddRequest("cn=John Doe,dc=example,dc=com")

	controls := ldap.NewControlsWithOIDs([]string{ldap.LDAP_SERVER_PERMISSIVE_MODIFY_OID}, false)
	for _, c := range controls {
		req.AddControl(c)
	}

	if len(req.Controls) != len(controls) {
		t.Fatalf("len(Controls) = %d; want %d", len(req.Controls), len(controls))
	}
	if req.Controls[0].GetControlType() != ldap.LDAP_SERVER_PERMISSIVE_MODIFY_OID {
		t.Errorf("Controls[0] type = %q; want %q",
			req.Controls[0].GetControlType(), ldap.LDAP_SERVER_PERMISSIVE_MODIFY_OID)
	}
}
