package mscmpo

import (
	"reflect"
	"testing"

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

// TestBIND_VERSION_SET_RoundTrip exercises the six-DWORD version request block.
func TestBIND_VERSION_SET_RoundTrip(t *testing.T) {
	roundTrip(t, "BIND_VERSION_SET", BIND_VERSION_SET{
		DwMinLevelOne:   1,
		DwMaxLevelOne:   2,
		DwMinLevelTwo:   3,
		DwMaxLevelTwo:   4,
		DwMinLevelThree: 5,
		DwMaxLevelThree: 6,
	})
}

// TestBOUND_VERSION_SET_RoundTrip exercises the three-DWORD negotiated version block.
func TestBOUND_VERSION_SET_RoundTrip(t *testing.T) {
	roundTrip(t, "BOUND_VERSION_SET", BOUND_VERSION_SET{
		DwLevelOneAccepted:   2,
		DwLevelTwoAccepted:   4,
		DwLevelThreeAccepted: 6,
	})
}

// TestBIND_INFO_BLOB_RoundTrip exercises the fixed 8-byte bind-info structure carried in
// the rguchBlob byte arrays of Poke/BuildContext.
func TestBIND_INFO_BLOB_RoundTrip(t *testing.T) {
	roundTrip(t, "BIND_INFO_BLOB", BIND_INFO_BLOB{
		DwcbThisStruct:    8,
		GrbitComProtocols: 0x00000001,
	})
}

// TestPCONTEXT_HANDLE_RoundTrip exercises the 20-byte context handle (wrapped in a struct,
// since a context handle only ever appears as a field of a request/response), and IsZero
// verifies the nulled-handle predicate used after a tear-down.
func TestPCONTEXT_HANDLE_RoundTrip(t *testing.T) {
	h := PCONTEXT_HANDLE{0x01, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef}
	roundTrip(t, "PCONTEXT_HANDLE", struct{ H PCONTEXT_HANDLE }{h})
	if h.IsZero() {
		t.Fatal("IsZero() = true for a non-zero handle")
	}
	if !(PCONTEXT_HANDLE{}).IsZero() {
		t.Fatal("IsZero() = false for the zero handle")
	}
}

// TestEnumWidths verifies the interface enums encode as 16-bit NDR values ([C706] 14.3.6):
// a single enum field marshals to exactly two octets.
func TestEnumWidths(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{"SESSION_RANK", struct {
			V SESSION_RANK `ndr:"enum"`
		}{SRANK_SECONDARY}},
		{"TEARDOWN_TYPE", struct {
			V TEARDOWN_TYPE `ndr:"enum"`
		}{TT_PROBLEM}},
		{"RESOURCE_TYPE", struct {
			V RESOURCE_TYPE `ndr:"enum"`
		}{RT_CONNECTIONS}},
	}
	for _, c := range cases {
		raw, err := ndr.Marshal(c.in)
		if err != nil {
			t.Fatalf("%s: Marshal: %v", c.name, err)
		}
		if len(raw) != 2 {
			t.Errorf("%s: encoded %d octets, want 2 (16-bit NDR enum)", c.name, len(raw))
		}
	}
}
