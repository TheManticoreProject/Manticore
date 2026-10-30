package commands_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// lengthField describes a header-relative (offset, 32-bit length) pair in a
// command body: the offset is offWidth bytes wide at offPos, the length is a
// uint32 at lenPos, and the variable data may start no earlier than fixed bytes
// into the body.
type lengthField struct {
	name      string
	fixed     int
	offPos    int
	offWidth  int
	lenPos    int
	unmarshal func([]byte) (int, error)
}

// TestUnmarshal_HugeLengthOrOffset feeds every body parser that reads a 32-bit
// length a length (or offset) of 2 GiB and above. On 32-bit platforms the
// start+length bounds check used to wrap negative and pass, so the subsequent
// slice or make panicked; every case must instead be rejected with an error.
func TestUnmarshal_HugeLengthOrOffset(t *testing.T) {
	fields := []lengthField{
		{"ReadResponse", 16, 2, 1, 4, func(b []byte) (int, error) { return commands.NewReadResponse().Unmarshal(b) }},
		{"CreateRequest/CreateContexts", 56, 48, 4, 52, func(b []byte) (int, error) { return commands.NewCreateRequest().Unmarshal(b) }},
		{"CreateResponse/CreateContexts", 88, 80, 4, 84, func(b []byte) (int, error) { return commands.NewCreateResponse().Unmarshal(b) }},
		{"IoctlRequest/Input", 56, 24, 4, 28, func(b []byte) (int, error) { return commands.NewIoctlRequest().Unmarshal(b) }},
		{"IoctlResponse/Input", 48, 24, 4, 28, func(b []byte) (int, error) { return commands.NewIoctlResponse().Unmarshal(b) }},
		{"IoctlResponse/Output", 48, 32, 4, 36, func(b []byte) (int, error) { return commands.NewIoctlResponse().Unmarshal(b) }},
		{"QueryDirectoryResponse", 8, 2, 2, 4, func(b []byte) (int, error) { return commands.NewQueryDirectoryResponse().Unmarshal(b) }},
		{"QueryInfoRequest", 40, 8, 2, 12, func(b []byte) (int, error) { return commands.NewQueryInfoRequest().Unmarshal(b) }},
		{"QueryInfoResponse", 8, 2, 2, 4, func(b []byte) (int, error) { return commands.NewQueryInfoResponse().Unmarshal(b) }},
		{"SetInfoRequest", 32, 8, 2, 4, func(b []byte) (int, error) { return commands.NewSetInfoRequest().Unmarshal(b) }},
		{"ChangeNotifyResponse", 8, 2, 2, 4, func(b []byte) (int, error) { return commands.NewChangeNotifyResponse().Unmarshal(b) }},
		{"WriteRequest", 48, 2, 2, 4, func(b []byte) (int, error) { return commands.NewWriteRequest().Unmarshal(b) }},
	}

	type pair struct{ offset, length uint32 }
	for _, f := range fields {
		validOffset := uint32(64 + f.fixed)
		pairs := []pair{
			{validOffset, 0x7FFFFFF8},
			{validOffset, 0x80000000},
			{validOffset, 0xFFFFFFFF},
		}
		if f.offWidth == 4 {
			// A huge 32-bit offset with a small length overflows start+length too.
			pairs = append(pairs, pair{0x7FFFFFFF, 0x10}, pair{0xFFFFFFF0, 0x10})
		}
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/off=%#x/len=%#x", f.name, p.offset, p.length), func(t *testing.T) {
				data := make([]byte, f.fixed+16)
				switch f.offWidth {
				case 1:
					data[f.offPos] = byte(p.offset)
				case 2:
					binary.LittleEndian.PutUint16(data[f.offPos:], uint16(p.offset))
				case 4:
					binary.LittleEndian.PutUint32(data[f.offPos:], p.offset)
				}
				binary.LittleEndian.PutUint32(data[f.lenPos:], p.length)
				if _, err := f.unmarshal(data); err == nil {
					t.Fatal("Unmarshal accepted an out-of-bounds length, want error")
				}
			})
		}
	}
}

// TestErrorResponse_HugeByteCount checks that an ERROR Response whose ByteCount
// is 2 GiB or more is rejected rather than wrapping the bounds check on 32-bit
// platforms.
func TestErrorResponse_HugeByteCount(t *testing.T) {
	for _, n := range []uint32{0x7FFFFFFF, 0x80000000, 0xFFFFFFFF} {
		data := make([]byte, 16)
		binary.LittleEndian.PutUint32(data[4:8], n)
		if _, err := commands.NewErrorResponse(0).Unmarshal(data); err == nil {
			t.Errorf("ByteCount %#x: Unmarshal accepted an out-of-bounds count, want error", n)
		}
	}
}
