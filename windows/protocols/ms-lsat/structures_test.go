package mslsat

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in.
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

func mustSID(t *testing.T, s string) *msdtyp.RPC_SID {
	t.Helper()
	sid, err := msdtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return &sid
}

// TestLSAPR_SID_INFORMATION_RoundTrip exercises the single-field wrapper around a
// [unique] pointer to an RPC_SID.
func TestLSAPR_SID_INFORMATION_RoundTrip(t *testing.T) {
	roundTrip(t, "LSAPR_SID_INFORMATION", LSAPR_SID_INFORMATION{
		Sid: mustSID(t, "S-1-5-21-1004336348-1177238915-682003330-512"),
	})
	roundTrip(t, "LSAPR_SID_INFORMATION/well-known", LSAPR_SID_INFORMATION{
		Sid: mustSID(t, "S-1-5-18"),
	})
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
			{Use: SidTypeUser, Name: msdtyp.NewUnicodeString("Administrator"), DomainIndex: 0},
			{Use: SidTypeWellKnownGroup, Name: msdtyp.NewUnicodeString("SYSTEM"), DomainIndex: -1},
		},
	}
	roundTrip(t, "LSAPR_TRANSLATED_NAMES", in)
}

// TestLSAPR_REFERENCED_DOMAIN_LIST_RoundTrip exercises a structure whose middle field is
// a [unique] pointer to a conformant array of structs with embedded string + SID. The
// element type LSAPR_TRUST_INFORMATION is the shared lsarpc type owned by the MS-LSAD
// structures package.
func TestLSAPR_REFERENCED_DOMAIN_LIST_RoundTrip(t *testing.T) {
	in := LSAPR_REFERENCED_DOMAIN_LIST{
		Entries: 2,
		Domains: []mslsad.LSAPR_TRUST_INFORMATION{
			{Name: msdtyp.NewUnicodeString("CONTOSO"), Sid: mustSID(t, "S-1-5-21-1-2-3")},
			{Name: msdtyp.NewUnicodeString("FABRIKAM"), Sid: mustSID(t, "S-1-5-21-4-5-6")},
		},
		MaxEntries: 32,
	}
	roundTrip(t, "LSAPR_REFERENCED_DOMAIN_LIST", in)
}
