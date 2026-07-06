package msdtyp

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

func TestLUID_RoundTrip(t *testing.T) {
	type rec struct{ ID LUID }
	in := &rec{ID: LUID{LowPart: 0x11223344, HighPart: -2}}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.ID != in.ID {
		t.Errorf("round trip: got %+v want %+v", out.ID, in.ID)
	}
	if got := LUIDFromUint64(in.ID.Uint64()); got != in.ID {
		t.Errorf("Uint64 round trip: got %+v want %+v", got, in.ID)
	}
}

func TestFILETIME_GoldenAndRoundTrip(t *testing.T) {
	type rec struct{ T FILETIME }
	in := &rec{T: FILETIME{DwLowDateTime: 0x11223344, DwHighDateTime: 0x55667788}}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Two 32-bit LE words, low half first ([MS-DTYP] 2.3.3).
	want := []byte{0x44, 0x33, 0x22, 0x11, 0x88, 0x77, 0x66, 0x55}
	if !bytes.Equal(raw, want) {
		t.Errorf("FILETIME wire = % x, want % x", raw, want)
	}
	var out rec
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.T != in.T {
		t.Errorf("FILETIME round-trip: got %+v want %+v", out.T, in.T)
	}
	if got := in.T.Uint64(); got != 0x5566778811223344 {
		t.Errorf("FILETIME.Uint64() = %#x", got)
	}
}

func TestSYSTEMTIME_GoldenAndRoundTrip(t *testing.T) {
	type rec struct{ T SYSTEMTIME }
	in := &rec{T: SYSTEMTIME{
		WYear: 2026, WMonth: 7, WDayOfWeek: 4, WDay: 2,
		WHour: 13, WMinute: 30, WSecond: 15, WMilliseconds: 500,
	}}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Eight 16-bit LE fields in declaration order ([MS-DTYP] 2.3.13).
	want := []byte{
		0xEA, 0x07, 0x07, 0x00, 0x04, 0x00, 0x02, 0x00,
		0x0D, 0x00, 0x1E, 0x00, 0x0F, 0x00, 0xF4, 0x01,
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("SYSTEMTIME wire = % x, want % x", raw, want)
	}
	var out rec
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.T != in.T {
		t.Errorf("SYSTEMTIME round-trip: got %+v want %+v", out.T, in.T)
	}
}

func TestLargeInteger_RoundTrip(t *testing.T) {
	type rec struct {
		Q LARGE_INTEGER
		U ULARGE_INTEGER
	}
	in := &rec{Q: -123456789, U: 0xDEADBEEFCAFE}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip: got %+v want %+v", out, *in)
	}
}

func TestRPC_SID_GoldenAndRoundTrip(t *testing.T) {
	sid, err := ParseSID("S-1-5-21-1-2-3")
	if err != nil {
		t.Fatalf("ParseSID: %v", err)
	}
	raw, err := ndr.Marshal(&sid)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x04, 0, 0, 0, // hoisted maximum_count (SubAuthorityCount)
		0x01,                               // Revision
		0x04,                               // SubAuthorityCount (derived from slice len)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x05, // IdentifierAuthority (6-octet big-endian = 5)
		0x15, 0, 0, 0, // SubAuthority[0] = 21
		0x01, 0, 0, 0, // SubAuthority[1] = 1
		0x02, 0, 0, 0, // SubAuthority[2] = 2
		0x03, 0, 0, 0, // SubAuthority[3] = 3
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("RPC_SID:\n got %x\nwant %x", raw, want)
	}
	var out RPC_SID
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := out.String(); got != "S-1-5-21-1-2-3" {
		t.Errorf("String: got %q want %q", got, "S-1-5-21-1-2-3")
	}
}

func TestRPC_UNICODE_STRING_GoldenAndRoundTrip(t *testing.T) {
	u := NewUnicodeString("Hi")
	if u.Length != 4 || u.MaximumLength != 4 {
		t.Fatalf("NewUnicodeString counts: got Length=%d MaximumLength=%d want 4/4", u.Length, u.MaximumLength)
	}
	raw, err := ndr.Marshal(&u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x04, 0x00, // Length (bytes)
		0x04, 0x00, // MaximumLength (bytes)
		0x00, 0x00, 0x02, 0x00, // Buffer referent id
		0x02, 0, 0, 0, // buffer maximum_count (chars)
		0x00, 0, 0, 0, // offset
		0x02, 0, 0, 0, // actual_count (chars)
		0x48, 0x00, 0x69, 0x00, // 'H', 'i'
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("RPC_UNICODE_STRING:\n got %x\nwant %x", raw, want)
	}
	var out RPC_UNICODE_STRING
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := out.String(); got != "Hi" {
		t.Errorf("String: got %q want %q", got, "Hi")
	}
}

// TestRPC_UNICODE_STRING_OverAllocated covers the spec-legal case where MaximumLength
// (buffer capacity in bytes) exceeds Length (valid chars in bytes). Per [MS-DTYP] 2.3.10
// the wchar array's maximum_count must be MaximumLength/2 and actual_count Length/2,
// independent of len(Buffer); only the valid Length/2 chars are transmitted.
func TestRPC_UNICODE_STRING_OverAllocated(t *testing.T) {
	// "Hi" valid (Length=4), capacity for 5 wchars (MaximumLength=10), buffer carries
	// the two valid chars plus reserved slots.
	u := RPC_UNICODE_STRING{Length: 4, MaximumLength: 10, Buffer: []uint16{'H', 'i', 0, 0, 0}}
	raw, err := ndr.Marshal(&u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x04, 0x00, // Length (bytes)
		0x0a, 0x00, // MaximumLength (bytes)
		0x00, 0x00, 0x02, 0x00, // Buffer referent id
		0x05, 0, 0, 0, // maximum_count = MaximumLength/2 = 5
		0x00, 0, 0, 0, // offset
		0x02, 0, 0, 0, // actual_count = Length/2 = 2
		0x48, 0x00, 0x69, 0x00, // 'H', 'i' — only the valid chars
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("over-allocated RPC_UNICODE_STRING:\n got %x\nwant %x", raw, want)
	}
	var out RPC_UNICODE_STRING
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Length != 4 || out.MaximumLength != 10 {
		t.Errorf("counts: got Length=%d MaximumLength=%d want 4/10", out.Length, out.MaximumLength)
	}
	if got := out.String(); got != "Hi" {
		t.Errorf("String: got %q want %q", got, "Hi")
	}
}

func TestRPC_UNICODE_STRING_Empty(t *testing.T) {
	u := NewUnicodeString("")
	raw, err := ndr.Marshal(&u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Length=0, MaximumLength=0, Buffer=NULL referent.
	want := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	if !bytes.Equal(raw, want) {
		t.Errorf("empty RPC_UNICODE_STRING:\n got %x\nwant %x", raw, want)
	}
	var out RPC_UNICODE_STRING
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := out.String(); got != "" {
		t.Errorf("String: got %q want empty", got)
	}
}

// TestArrayOfUnicodeStrings exercises the #419 interaction: an array of structs each
// carrying a [unique] pointer to a conformant-varying buffer. Each element's buffer
// must be deferred past the whole array, which only round-trips with the array referent
// ordering fix in place.
func TestArrayOfUnicodeStrings(t *testing.T) {
	type rec struct {
		Names []RPC_UNICODE_STRING `ndr:"conformant"`
	}
	in := &rec{Names: []RPC_UNICODE_STRING{
		NewUnicodeString("alice"),
		NewUnicodeString("bob"),
	}}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Names) != 2 || out.Names[0].String() != "alice" || out.Names[1].String() != "bob" {
		t.Errorf("round trip: got %q,%q", out.Names[0].String(), out.Names[1].String())
	}
}

func TestArrayOfUnicodeStringPointers(t *testing.T) {
	type rec struct {
		Names []*RPC_UNICODE_STRING `ndr:"conformant"`
	}
	a := NewUnicodeString("foo")
	b := NewUnicodeString("barbaz")
	in := &rec{Names: []*RPC_UNICODE_STRING{&a, &b}}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Names) != 2 || out.Names[0] == nil || out.Names[1] == nil ||
		out.Names[0].String() != "foo" || out.Names[1].String() != "barbaz" {
		t.Errorf("round trip mismatch: %+v", out.Names)
	}
}

func TestParseSID_Errors(t *testing.T) {
	for _, s := range []string{"", "1-5-21", "S-x-5", "S-1-5-" + "9999999999999999999999"} {
		if _, err := ParseSID(s); err == nil {
			t.Errorf("ParseSID(%q): expected error", s)
		}
	}
	// Round-trip a hex authority.
	sid, err := ParseSID("S-1-0x10000000000-1")
	if err != nil {
		t.Fatalf("ParseSID hex authority: %v", err)
	}
	if got := sid.String(); got != "S-1-0x10000000000-1" {
		t.Errorf("hex authority round trip: got %q", got)
	}
}
