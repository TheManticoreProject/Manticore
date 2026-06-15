package ndr

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// pipeResp mirrors an [out] pipe parameter followed by an NTSTATUS return value (e.g.
// EfsRpcReadFileRaw): a pipe of bytes, then the retval.
type pipeResp struct {
	Pipe   []byte `ndr:"pipe"`
	Status uint32 `ndr:"retval"`
}

// TestNDR64_PipeMultiChunk decodes a multi-chunk NDR64 pipe and round-trips a payload.
// The stub is hand-built to the NDR64 framing — each chunk is [+count(8)][elements]
// [-count(8)], terminated by a 0(8) count — which was verified on the wire against a
// Windows Server 2016 EfsRpcReadFileRaw response (issue #596); the bytes here are
// synthetic so no captured data is embedded.
func TestNDR64_PipeMultiChunk(t *testing.T) {
	var stub []byte
	addChunk := func(data []byte) {
		var c [8]byte
		binary.LittleEndian.PutUint64(c[:], uint64(len(data)))
		stub = append(stub, c[:]...) // +count
		stub = append(stub, data...) // elements
		for len(stub)%8 != 0 {       // pad to 8 before the trailer
			stub = append(stub, 0)
		}
		binary.LittleEndian.PutUint64(c[:], ^uint64(len(data))+1)
		stub = append(stub, c[:]...) // -count trailer
	}
	addChunk([]byte{0x01, 0x02, 0x03, 0x04})
	addChunk([]byte{0x05, 0x06})
	stub = append(stub, make([]byte, 8)...)     // 0 terminator (8 octets)
	stub = append(stub, 0x00, 0x00, 0x00, 0x00) // retval status = 0

	var resp pipeResp
	if err := UnmarshalAs(stub, &resp, NDR64); err != nil {
		t.Fatalf("UnmarshalAs NDR64: %v", err)
	}
	want := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	if !bytes.Equal(resp.Pipe, want) {
		t.Errorf("decoded pipe = %x, want %x", resp.Pipe, want)
	}
	if resp.Status != 0 {
		t.Errorf("status = %#x, want 0", resp.Status)
	}

	// Round-trip a payload (marshal emits a single chunk).
	wire, err := MarshalAs(&pipeResp{Pipe: want}, NDR64)
	if err != nil {
		t.Fatalf("MarshalAs NDR64: %v", err)
	}
	var back pipeResp
	if err := UnmarshalAs(wire, &back, NDR64); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if !bytes.Equal(back.Pipe, want) {
		t.Errorf("pipe round trip = %x, want %x", back.Pipe, want)
	}
}
