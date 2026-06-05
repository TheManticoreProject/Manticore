package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTripBase marshals in, unmarshals into a fresh value of the same type, and
// asserts the result is deeply equal to in.
func roundTripBase[T any](t *testing.T, name string, in T) {
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

func mustBaseSID(t *testing.T, s string) *dtyp.RPC_SID {
	t.Helper()
	sid, err := dtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return &sid
}

// TestSAMPR_ENUMERATION_BUFFER_RoundTrip exercises a [unique] pointer to a
// conformant array of structs that each embed an RPC_UNICODE_STRING.
func TestSAMPR_ENUMERATION_BUFFER_RoundTrip(t *testing.T) {
	in := SAMPR_ENUMERATION_BUFFER{
		EntriesRead: 2,
		Buffer: []SAMPR_RID_ENUMERATION{
			{RelativeId: 500, Name: dtyp.NewUnicodeString("Administrator")},
			{RelativeId: 501, Name: dtyp.NewUnicodeString("Guest")},
		},
	}
	roundTripBase(t, "SAMPR_ENUMERATION_BUFFER", in)
}

// TestSAMPR_PSID_ARRAY_RoundTrip exercises a [unique] pointer to a conformant
// array of structs that each hold a [unique] pointer to an RPC_SID.
func TestSAMPR_PSID_ARRAY_RoundTrip(t *testing.T) {
	in := SAMPR_PSID_ARRAY{
		Count: 2,
		Sids: []SAMPR_SID_INFORMATION{
			{SidPointer: mustBaseSID(t, "S-1-5-21-1004336348-1177238915-682003330-512")},
			{SidPointer: mustBaseSID(t, "S-1-5-18")},
		},
	}
	roundTripBase(t, "SAMPR_PSID_ARRAY", in)
}
