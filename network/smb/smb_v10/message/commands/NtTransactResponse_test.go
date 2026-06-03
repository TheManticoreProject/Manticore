package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// Test_NtTransactResponse_RoundTrip verifies that the standard SMB_COM_NT_TRANSACT
// response format defined by MS-CIFS is marshalled and unmarshalled symmetrically,
// including the Setup words and the padded Parameters/Data blocks.
func Test_NtTransactResponse_RoundTrip(t *testing.T) {
	r := commands.NewNtTransactResponse()
	r.TotalParameterCount = 4
	r.TotalDataCount = 5
	r.ParameterCount = 4
	r.DataCount = 5
	r.SetupCount = 2
	r.Setup = []types.USHORT{0x1234, 0xABCD}

	// Parameter words byte size: Reserved1(3) + 8*ULONG(32) + SetupCount(1) + Setup(4) = 40,
	// which is 20 words, so WordCount = 20 and the marshalled SMB_Parameters block is
	// 1 (WordCount) + 40 = 41 bytes. The SMB_Data block then starts after the 2-byte
	// ByteCount field. All *Offset fields are measured from the start of the SMB Header.
	dataBlockStart := header.SMB_HEADER_SIZE + 41 + 2

	r.Pad1 = []types.UCHAR{}
	r.Parameters = []types.UCHAR{0xAA, 0xBB, 0xCC, 0xDD}
	r.ParameterOffset = types.ULONG(dataBlockStart)
	r.Pad2 = []types.UCHAR{0x00, 0x00, 0x00}
	r.Data = []types.UCHAR{1, 2, 3, 4, 5}
	r.DataOffset = types.ULONG(dataBlockStart + len(r.Parameters) + len(r.Pad2))

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// WordCount MUST equal SetupCount + 18 (0x12).
	if got, want := marshalled[0], byte(r.SetupCount)+18; got != want {
		t.Fatalf("WordCount = %d, want %d", got, want)
	}

	parsed := commands.NewNtTransactResponse()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.TotalParameterCount != r.TotalParameterCount ||
		parsed.TotalDataCount != r.TotalDataCount ||
		parsed.ParameterCount != r.ParameterCount ||
		parsed.DataCount != r.DataCount ||
		parsed.ParameterOffset != r.ParameterOffset ||
		parsed.DataOffset != r.DataOffset {
		t.Fatalf("count/offset mismatch: %+v", parsed)
	}
	if parsed.SetupCount != r.SetupCount || len(parsed.Setup) != 2 ||
		parsed.Setup[0] != 0x1234 || parsed.Setup[1] != 0xABCD {
		t.Fatalf("Setup mismatch: %+v", parsed.Setup)
	}
	if !bytes.Equal([]byte(parsed.Parameters), []byte(r.Parameters)) {
		t.Fatalf("Parameters mismatch: got %v", parsed.Parameters)
	}
	if !bytes.Equal([]byte(parsed.Data), []byte(r.Data)) {
		t.Fatalf("Data mismatch: got %v", parsed.Data)
	}
	if len(parsed.Pad2) != len(r.Pad2) {
		t.Fatalf("Pad2 length mismatch: got %d, want %d", len(parsed.Pad2), len(r.Pad2))
	}
}

// Test_NtTransactResponse_InterimEmpty verifies that an interim/error response with an
// empty Parameter and Data section (WordCount and ByteCount both zero) is parsed without
// error and without panicking.
func Test_NtTransactResponse_InterimEmpty(t *testing.T) {
	// WordCount = 0, ByteCount = 0x0000.
	marshalled := []byte{0x00, 0x00, 0x00}

	r := commands.NewNtTransactResponse()
	if _, err := r.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal of interim response failed: %v", err)
	}
	if r.ParameterCount != 0 || r.DataCount != 0 || r.SetupCount != 0 {
		t.Fatalf("expected empty interim response, got %+v", r)
	}
}
