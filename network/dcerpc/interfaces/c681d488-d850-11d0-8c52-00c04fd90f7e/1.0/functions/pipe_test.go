package functions

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestWriteFileRawRequest_PipeRoundTrip checks that the [in] EFS_EXIM_PIPE is marshalled
// as an NDR pipe after the context handle and round-trips.
func TestWriteFileRawRequest_PipeRoundTrip(t *testing.T) {
	var h structures.PEXIMPORT_CONTEXT_HANDLE
	for i := range h {
		h[i] = byte(i)
	}
	in := &efsRpcWriteFileRawRequest{HContext: h, EfsInPipe: structures.EFS_EXIM_PIPE{0xde, 0xad, 0xbe, 0xef}}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// 20-byte handle, then the pipe: count=4, 4 bytes, terminating 0 count.
	want := append(h[:], 0x04, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef, 0x00, 0x00, 0x00, 0x00)
	if !bytes.Equal(raw, want) {
		t.Errorf("WriteFileRaw request wire:\n got %x\nwant %x", raw, want)
	}
	var out efsRpcWriteFileRawRequest
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.HContext != in.HContext || !bytes.Equal(out.EfsInPipe, in.EfsInPipe) {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}

// TestReadFileRawResponse_PipeRoundTrip checks the [out] EFS_EXIM_PIPE stream precedes
// the return value and round-trips.
func TestReadFileRawResponse_PipeRoundTrip(t *testing.T) {
	in := &efsRpcReadFileRawResponse{EfsOutPipe: structures.EFS_EXIM_PIPE{1, 2, 3, 4, 5}, Status: 0}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out efsRpcReadFileRawResponse
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(out.EfsOutPipe, in.EfsOutPipe) || out.Status != 0 {
		t.Errorf("round trip: got pipe=%x status=%#x", out.EfsOutPipe, out.Status)
	}
}
