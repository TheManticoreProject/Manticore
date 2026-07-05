package msswn

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestRESP_ASYNC_NOTIFY_RoundTrip exercises the notification response: a scalar header
// plus a [size_is(Length)] [unique] PBYTE buffer. Length must be derived from the slice
// on marshal (the size_is reference points at the exported Go field name Length), and the
// opaque MessageBuffer bytes must survive the round trip.
func TestRESP_ASYNC_NOTIFY_RoundTrip(t *testing.T) {
	in := RESP_ASYNC_NOTIFY{
		MessageType:      1, // RESOURCE_CHANGE_NOTIFICATION
		NumberOfMessages: 2,
		MessageBuffer:    []uint8{0x10, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out RESP_ASYNC_NOTIFY
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Length != uint32(len(in.MessageBuffer)) {
		t.Errorf("Length = %d, want %d (must be derived from the slice)", out.Length, len(in.MessageBuffer))
	}
	if out.MessageType != in.MessageType || out.NumberOfMessages != in.NumberOfMessages {
		t.Errorf("header round-trip: got type=%d msgs=%d, want type=%d msgs=%d",
			out.MessageType, out.NumberOfMessages, in.MessageType, in.NumberOfMessages)
	}
	if !reflect.DeepEqual(out.MessageBuffer, in.MessageBuffer) {
		t.Errorf("MessageBuffer round-trip: got %v want %v", out.MessageBuffer, in.MessageBuffer)
	}
}

// TestRESP_ASYNC_NOTIFY_NilBuffer covers a [unique] pointer that is NULL (no pending
// messages): MessageBuffer marshals as a null referent and unmarshals back to nil.
func TestRESP_ASYNC_NOTIFY_NilBuffer(t *testing.T) {
	in := RESP_ASYNC_NOTIFY{MessageType: 0, NumberOfMessages: 0}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out RESP_ASYNC_NOTIFY
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Length != 0 || len(out.MessageBuffer) != 0 {
		t.Errorf("nil buffer round-trip: Length=%d len(MessageBuffer)=%d, want 0/0", out.Length, len(out.MessageBuffer))
	}
}

// TestWITNESS_INTERFACE_INFO_RoundTrip exercises the fixed-size arrays
// (InterfaceGroupName[260], IPV6[8]) and the scalar fields, all transmitted inline.
func TestWITNESS_INTERFACE_INFO_RoundTrip(t *testing.T) {
	in := WITNESS_INTERFACE_INFO{
		Version: 0x00020000,
		State:   0x0001, // AVAILABLE
		IPV4:    0xC0A80101,
		Flags:   0x00000001,
	}
	// "GROUP" in the group name; a couple of IPv6 words.
	for i, c := range []uint16{'G', 'R', 'O', 'U', 'P'} {
		in.InterfaceGroupName[i] = c
	}
	in.IPV6[0] = 0xfe80
	in.IPV6[7] = 0x0001

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out WITNESS_INTERFACE_INFO
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

// TestWITNESS_INTERFACE_LIST_RoundTrip exercises a count field plus a [unique] pointer to
// a conformant array of fixed-array-bearing structs. NumberOfInterfaces must be derived
// from the slice on marshal.
func TestWITNESS_INTERFACE_LIST_RoundTrip(t *testing.T) {
	in := WITNESS_INTERFACE_LIST{
		InterfaceInfo: []WITNESS_INTERFACE_INFO{
			{Version: 0x00010001, State: 1, IPV4: 0x0A000001, Flags: 1},
			{Version: 0x00020000, State: 0xFF, IPV4: 0x0A000002, Flags: 0},
		},
	}
	in.InterfaceInfo[0].InterfaceGroupName[0] = 'A'
	in.InterfaceInfo[1].InterfaceGroupName[0] = 'B'

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out WITNESS_INTERFACE_LIST
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.NumberOfInterfaces != 2 || len(out.InterfaceInfo) != 2 {
		t.Fatalf("NumberOfInterfaces=%d len(InterfaceInfo)=%d, want 2/2", out.NumberOfInterfaces, len(out.InterfaceInfo))
	}
	if !reflect.DeepEqual(in.InterfaceInfo, out.InterfaceInfo) {
		t.Errorf("element round-trip mismatch:\n in=%+v\nout=%+v", in.InterfaceInfo, out.InterfaceInfo)
	}
}

// TestPCONTEXT_HANDLE_RoundTrip covers the 20-octet context handle transmitted inline and
// the IsZero helper.
func TestPCONTEXT_HANDLE_RoundTrip(t *testing.T) {
	var h PCONTEXT_HANDLE
	if !h.IsZero() {
		t.Errorf("zero-valued handle: IsZero() = false, want true")
	}
	h[0] = 0x01
	h[19] = 0xff
	if h.IsZero() {
		t.Errorf("non-zero handle: IsZero() = true, want false")
	}

	type wrap struct{ H PCONTEXT_HANDLE }
	raw, err := ndr.Marshal(&wrap{H: h})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out wrap
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.H != h {
		t.Errorf("handle round-trip: got %v want %v", out.H, h)
	}
}
