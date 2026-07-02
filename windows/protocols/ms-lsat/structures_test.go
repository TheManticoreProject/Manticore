package mslsat

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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

func mustSID(t *testing.T, s string) *dtyp.RPC_SID {
	t.Helper()
	sid, err := dtyp.ParseSID(s)
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
