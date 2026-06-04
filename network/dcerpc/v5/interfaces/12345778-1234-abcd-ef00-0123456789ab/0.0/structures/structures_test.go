package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func mustSID(t *testing.T, s string) *dtyp.RPC_SID {
	t.Helper()
	sid, err := dtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return &sid
}

// TestLSAPR_SID_ENUM_BUFFER_RoundTrip exercises a [unique] pointer to a conformant array
// of structs that each hold a [unique] pointer to an RPC_SID.
func TestLSAPR_SID_ENUM_BUFFER_RoundTrip(t *testing.T) {
	in := LSAPR_SID_ENUM_BUFFER{
		Entries: 2,
		SidInfo: []LSAPR_SID_INFORMATION{
			{Sid: mustSID(t, "S-1-5-21-1004336348-1177238915-682003330-512")},
			{Sid: mustSID(t, "S-1-5-18")},
		},
	}
	roundTrip(t, "LSAPR_SID_ENUM_BUFFER", in)
}

// TestLSAPR_TRANSLATED_NAMES_RoundTrip exercises an array of structs each embedding an
// RPC_UNICODE_STRING and an NDR enum.
func TestLSAPR_TRANSLATED_NAMES_RoundTrip(t *testing.T) {
	in := LSAPR_TRANSLATED_NAMES{
		Entries: 2,
		Names: []LSAPR_TRANSLATED_NAME{
			{Use: SidTypeUser, Name: dtyp.NewUnicodeString("Administrator"), DomainIndex: 0},
			{Use: SidTypeWellKnownGroup, Name: dtyp.NewUnicodeString("SYSTEM"), DomainIndex: -1},
		},
	}
	roundTrip(t, "LSAPR_TRANSLATED_NAMES", in)
}

// TestLSAPR_REFERENCED_DOMAIN_LIST_RoundTrip exercises a structure whose middle field is
// a [unique] pointer to a conformant array of structs with embedded string + SID.
func TestLSAPR_REFERENCED_DOMAIN_LIST_RoundTrip(t *testing.T) {
	in := LSAPR_REFERENCED_DOMAIN_LIST{
		Entries: 2,
		Domains: []LSAPR_TRUST_INFORMATION{
			{Name: dtyp.NewUnicodeString("CONTOSO"), Sid: mustSID(t, "S-1-5-21-1-2-3")},
			{Name: dtyp.NewUnicodeString("FABRIKAM"), Sid: mustSID(t, "S-1-5-21-4-5-6")},
		},
		MaxEntries: 32,
	}
	roundTrip(t, "LSAPR_REFERENCED_DOMAIN_LIST", in)
}

// TestLSAPR_CR_CIPHER_VALUE_RoundTrip exercises a conformant-varying counted byte blob.
func TestLSAPR_CR_CIPHER_VALUE_RoundTrip(t *testing.T) {
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02}
	in := LSAPR_CR_CIPHER_VALUE{
		Length:        uint32(len(payload)),
		MaximumLength: uint32(len(payload)),
		Buffer:        payload,
	}
	roundTrip(t, "LSAPR_CR_CIPHER_VALUE", in)
}

// TestLSAPR_PRIVILEGE_SET_RoundTrip exercises an inline conformant array (not a pointer)
// whose maximum_count is hoisted to the front of the structure.
func TestLSAPR_PRIVILEGE_SET_RoundTrip(t *testing.T) {
	in := LSAPR_PRIVILEGE_SET{
		PrivilegeCount: 2,
		Control:        1,
		Privilege: []LSAPR_LUID_AND_ATTRIBUTES{
			{Luid: dtyp.LUID{LowPart: 0x14, HighPart: 0}, Attributes: 3},
			{Luid: dtyp.LUID{LowPart: 0x11, HighPart: 0}, Attributes: 0},
		},
	}
	roundTrip(t, "LSAPR_PRIVILEGE_SET", in)
}
