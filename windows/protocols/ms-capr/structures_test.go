package mscapr

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mslsat "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsat"
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

// TestLSAPR_WRAPPED_CAPID_SET_RoundTrip exercises a [unique] pointer to a conformant
// array (SidInfo, sized by Entries) of LSAPR_SID_INFORMATION structures that each
// hold a [unique] pointer to an RPC_SID — the wire shape returned by
// LsarGetAvailableCAPIDs.
func TestLSAPR_WRAPPED_CAPID_SET_RoundTrip(t *testing.T) {
	roundTrip(t, "LSAPR_WRAPPED_CAPID_SET/two", LSAPR_WRAPPED_CAPID_SET{
		Entries: 2,
		SidInfo: []mslsat.LSAPR_SID_INFORMATION{
			{Sid: mustSID(t, "S-1-17-1727747406-3436664333-2626952977-3096872999")},
			{Sid: mustSID(t, "S-1-17-991899094-3157369668-3186527664-2135218404")},
		},
	})
}

// TestLSAPR_WRAPPED_CAPID_SET_Empty exercises the empty set (no policies deployed):
// Entries == 0 with a nil SidInfo referent.
func TestLSAPR_WRAPPED_CAPID_SET_Empty(t *testing.T) {
	roundTrip(t, "LSAPR_WRAPPED_CAPID_SET/empty", LSAPR_WRAPPED_CAPID_SET{
		Entries: 0,
		SidInfo: nil,
	})
}
