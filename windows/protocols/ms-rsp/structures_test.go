package msrsp

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

// TestREG_UNICODE_STRING_RoundTrip exercises the counted Unicode string carried by the
// InitShutdown / WindowsShutdown shutdown-message parameters. REG_UNICODE_STRING is an
// alias of dtyp.RPC_UNICODE_STRING, so this also confirms the byte-count vs char-count
// (Length/MaximumLength are bytes; Buffer max_count is MaximumLength/2) modeling holds
// through the alias.
func TestREG_UNICODE_STRING_RoundTrip(t *testing.T) {
	// Non-empty message string.
	in := dtyp.NewUnicodeString("System going down for maintenance")
	var reg REG_UNICODE_STRING = in
	if reg.Length != uint16(len([]rune("System going down for maintenance"))*2) {
		t.Errorf("Length = %d, want %d", reg.Length, len([]rune("System going down for maintenance"))*2)
	}
	if reg.String() != "System going down for maintenance" {
		t.Errorf("String() = %q, want the original message", reg.String())
	}
	roundTrip(t, "REG_UNICODE_STRING(non-empty)", reg)

	// Empty string yields a zero-length value with a NULL Buffer.
	roundTrip(t, "REG_UNICODE_STRING(empty)", REG_UNICODE_STRING(dtyp.NewUnicodeString("")))

	// Over-allocated buffer: MaximumLength (bytes) > Length (bytes), a shape a server
	// may advertise. Only actual_count (Length/2) chars are transmitted by the varying
	// array, so a strict Buffer round trip is not expected; instead confirm the decoded
	// message and the transmitted char count survive marshalling.
	over := REG_UNICODE_STRING{
		Length:        4, // 2 chars valid
		MaximumLength: 8, // 4 chars advertised as allocated
		Buffer:        []uint16{0x0041, 0x0042, 0x0000, 0x0000},
	}
	if over.String() != "AB" {
		t.Errorf("String() = %q, want \"AB\"", over.String())
	}
	raw, err := ndr.Marshal(&over)
	if err != nil {
		t.Fatalf("over-allocated: Marshal: %v", err)
	}
	var back REG_UNICODE_STRING
	if err := ndr.Unmarshal(raw, &back); err != nil {
		t.Fatalf("over-allocated: Unmarshal: %v", err)
	}
	if back.String() != "AB" {
		t.Errorf("over-allocated decoded String() = %q, want \"AB\"", back.String())
	}
	if len(back.Buffer) != 2 {
		t.Errorf("over-allocated decoded Buffer len = %d, want 2 (only actual_count is on the wire)", len(back.Buffer))
	}
}
