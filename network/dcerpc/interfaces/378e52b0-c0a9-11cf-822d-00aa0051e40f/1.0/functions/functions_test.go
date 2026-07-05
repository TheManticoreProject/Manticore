package functions

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

// roundTrip marshals in, unmarshals into a fresh T, and asserts deep equality. It
// validates the [in,out] size_is(ccBufferSize) conformant wchar buffer used by the
// SAGet* methods, which is not exercised by a live server in CI.
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

// TestSAGetNSAccountInformationRequest exercises the [in]+[in,out] request: a [unique]
// handle string, the count, and the conformant buffer whose maximum_count is the count.
func TestSAGetNSAccountInformationRequest(t *testing.T) {
	roundTrip(t, "SAGetNSAccountInformationRequest", sAGetNSAccountInformationRequest{
		Handle:       wstr("\\\\SERVER"),
		CcBufferSize: 273,
		WszBuffer:    make([]uint16, 273),
	})
}

// TestSAGetNSAccountInformationResponse exercises the [in,out] buffer echoed back: a
// conformant wchar array whose maximum_count is read from the wire (no size_is sibling).
func TestSAGetNSAccountInformationResponse(t *testing.T) {
	buf := make([]uint16, 12)
	for i, r := range []uint16{'L', 'o', 'c', 'a', 'l'} { // "Local\0..."
		buf[i] = r
	}
	roundTrip(t, "SAGetNSAccountInformationResponse", sAGetNSAccountInformationResponse{
		WszBuffer: buf,
		Status:    0,
	})
}

// TestDecodeWideBuffer verifies NUL-terminated UTF-16 decoding of the returned buffer.
func TestDecodeWideBuffer(t *testing.T) {
	buf := make([]uint16, 16)
	for i, r := range []uint16{'D', 'O', 'M', '\\', 'u'} {
		buf[i] = r
	}
	if got := decodeWideBuffer(buf); got != "DOM\\u" {
		t.Fatalf("decodeWideBuffer = %q, want %q", got, "DOM\\u")
	}
	if got := decodeWideBuffer([]uint16{}); got != "" {
		t.Fatalf("decodeWideBuffer(empty) = %q, want empty", got)
	}
}
