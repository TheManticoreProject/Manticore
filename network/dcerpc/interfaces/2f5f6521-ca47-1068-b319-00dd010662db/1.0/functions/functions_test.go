package functions

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// TestRemoteSPEventProc_BufferWireShape pins the wire layout of RemoteSPEventProc's
// pBuffer ([MS-TRP] 3.1.4.2): a top-level conformant-varying byte array transmitted inline
// (no referent id) whose maximum_count and actual_count both equal lSize, followed by the
// lSize field itself. lSize is derived from the slice length so the count and elements
// cannot disagree.
func TestRemoteSPEventProc_BufferWireShape(t *testing.T) {
	var ctx mstrp.PCONTEXT_HANDLE_TYPE2
	for i := range ctx {
		ctx[i] = 0xBB
	}
	req := &remoteSPEventProcRequest{
		PhContext: ctx,
		PBuffer:   []uint8{0x11, 0x22, 0x33, 0x44},
		LSize:     4,
	}
	raw, err := ndr.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := bytes.Join([][]byte{
		bytes.Repeat([]byte{0xBB}, 20), // phContext (inline, no referent id)
		{0x04, 0, 0, 0},                // maximum_count = lSize = 4
		{0x00, 0, 0, 0},                // offset = 0
		{0x04, 0, 0, 0},                // actual_count = lSize = 4
		{0x11, 0x22, 0x33, 0x44},       // 4 bytes
		{0x04, 0, 0, 0},                // lSize field = 4
	}, nil)
	if !bytes.Equal(raw, want) {
		t.Errorf("wire mismatch:\n got  %x\n want %x", raw, want)
	}
}

// TestRemoteSPEventProc_EmptyResponse confirms the void method decodes an empty response
// stub with no error (no return value on the wire).
func TestRemoteSPEventProc_EmptyResponse(t *testing.T) {
	var resp remoteSPEventProcResponse
	if err := ndr.Unmarshal(nil, &resp); err != nil {
		t.Fatalf("Unmarshal empty response: %v", err)
	}
}

// TestRemoteSPDetach_ResponseRoundTrip covers the void RemoteSPDetach response: just the
// nulled-out context handle, no return value.
func TestRemoteSPDetach_ResponseRoundTrip(t *testing.T) {
	raw := make([]byte, 20) // all-zero handle
	var resp remoteSPDetachResponse
	if err := ndr.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !resp.PphContext.IsZero() {
		t.Errorf("PphContext = %x, want zero", resp.PphContext)
	}
}
