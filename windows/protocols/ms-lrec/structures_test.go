package mslrec

import (
	"reflect"
	"testing"

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

// TestEVENT_BUFFER_RoundTrip exercises the [unique] pointer to a conformant byte array:
// BufferLength drives the array maximum_count via size_is.
func TestEVENT_BUFFER_RoundTrip(t *testing.T) {
	roundTrip(t, "EVENT_BUFFER", EVENT_BUFFER{
		BufferLength: 6,
		Buffer:       []uint8{0xde, 0xad, 0xbe, 0xef, 0x01, 0x02},
	})
}

// TestEVENT_BUFFER_Empty exercises the empty-buffer case (BufferLength 0, no payload),
// which the server returns when no events are ready.
func TestEVENT_BUFFER_Empty(t *testing.T) {
	roundTrip(t, "EVENT_BUFFER-empty", EVENT_BUFFER{
		BufferLength: 0,
		Buffer:       []uint8{},
	})
}

// TestPSESSION_HANDLE_RoundTrip exercises the 20-byte context handle (wrapped in a struct,
// since a context handle only ever appears as a field of a request/response), and IsZero
// verifies the nulled-handle predicate used after a session close.
func TestPSESSION_HANDLE_RoundTrip(t *testing.T) {
	h := PSESSION_HANDLE{0x01, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef}
	roundTrip(t, "PSESSION_HANDLE", struct{ H PSESSION_HANDLE }{h})
	if h.IsZero() {
		t.Fatal("IsZero() = true for a non-zero handle")
	}
	if !(PSESSION_HANDLE{}).IsZero() {
		t.Fatal("IsZero() = false for the zero handle")
	}
}
