package functions

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// TestClientRequest_BufferWireShape pins the independent-bounds wire layout of
// ClientRequest's pBuffer ([MS-TRP] 3.2.4.2): a top-level [ref] conformant-varying byte
// array whose maximum_count is lNeededSize (capacity) and actual_count is plUsedSize (the
// valid length), with no referent id. This matches the reference wire encoding used by
// interop implementations.
func TestClientRequest_BufferWireShape(t *testing.T) {
	var ctx mstrp.PCONTEXT_HANDLE_TYPE
	for i := range ctx {
		ctx[i] = 0xAA
	}
	req := &clientRequestRequest{
		PhContext:   ctx,
		PBuffer:     []uint8{0x01, 0x02, 0x03, 0x04},
		LNeededSize: 8, // capacity (maximum_count)
		PlUsedSize:  4, // valid bytes (actual_count)
	}
	raw, err := ndr.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := bytes.Join([][]byte{
		bytes.Repeat([]byte{0xAA}, 20), // phContext (inline, no referent id)
		{0x08, 0, 0, 0},                // maximum_count = lNeededSize = 8
		{0x00, 0, 0, 0},                // offset = 0
		{0x04, 0, 0, 0},                // actual_count = plUsedSize = 4
		{0x01, 0x02, 0x03, 0x04},       // 4 valid bytes
		{0x08, 0, 0, 0},                // lNeededSize field = 8
		{0x04, 0, 0, 0},                // plUsedSize field = 4
	}, nil)
	if !bytes.Equal(raw, want) {
		t.Errorf("wire mismatch:\n got  %x\n want %x", raw, want)
	}
}

// TestClientRequest_ResponseRoundTrip confirms the void-method response (no return value
// on the wire) decodes the returned buffer and used size. The server returns fewer valid
// bytes than the advertised capacity.
func TestClientRequest_ResponseRoundTrip(t *testing.T) {
	// A response wire image: pBuffer max_count=8, offset=0, actual_count=3, 3 bytes, then
	// the [out] plUsedSize=3.
	raw := bytes.Join([][]byte{
		{0x08, 0, 0, 0},    // maximum_count = 8 (advertised capacity)
		{0x00, 0, 0, 0},    // offset = 0
		{0x03, 0, 0, 0},    // actual_count = 3 (valid bytes)
		{0xDE, 0xAD, 0xBE}, // 3 valid bytes
		{0x00},             // pad to 4-align the following DWORD
		{0x03, 0, 0, 0},    // plUsedSize = 3
	}, nil)
	var resp clientRequestResponse
	if err := ndr.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(resp.PBuffer, []uint8{0xDE, 0xAD, 0xBE}) {
		t.Errorf("PBuffer = %x, want deadbe", resp.PBuffer)
	}
	if resp.PlUsedSize != 3 {
		t.Errorf("PlUsedSize = %d, want 3", resp.PlUsedSize)
	}
}

// TestClientDetach_ResponseRoundTrip covers the void ClientDetach response: just the
// nulled-out context handle, no return value.
func TestClientDetach_ResponseRoundTrip(t *testing.T) {
	raw := make([]byte, 20) // all-zero handle
	var resp clientDetachResponse
	if err := ndr.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !resp.PphContext.IsZero() {
		t.Errorf("PphContext = %x, want zero", resp.PphContext)
	}
}
