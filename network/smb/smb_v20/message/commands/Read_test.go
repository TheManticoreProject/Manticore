package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestReadRequest_RoundTrip(t *testing.T) {
	req := commands.NewReadRequest()
	req.Length = 0x1000
	req.Offset = 0x2000
	req.MinimumCount = 1
	req.FileId = types.SMB2_FILEID{Persistent: 0xAA, Volatile: 0xBB}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.ReadRequestStructureSize {
		t.Errorf("StructureSize = %d, want 49", got)
	}
	// No channel info: Buffer is the single mandatory padding byte.
	if len(wire) != 48+1 {
		t.Errorf("wire length = %d, want 49", len(wire))
	}

	decoded := commands.NewReadRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Length != req.Length || decoded.Offset != req.Offset || decoded.FileId != req.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestReadResponse_RoundTrip(t *testing.T) {
	resp := commands.NewReadResponse()
	resp.Data = []byte("hello smb2")

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.ReadResponseStructureSize {
		t.Errorf("StructureSize = %d, want 17", got)
	}
	// DataOffset is header-relative: 64 + 16 = 80; DataLength matches the payload.
	if got := wire[2]; got != 80 {
		t.Errorf("DataOffset = %d, want 80", got)
	}
	if got := binary.LittleEndian.Uint32(wire[4:8]); int(got) != len(resp.Data) {
		t.Errorf("DataLength = %d, want %d", got, len(resp.Data))
	}

	decoded := commands.NewReadResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(decoded.Data, resp.Data) {
		t.Errorf("Data = %q, want %q", decoded.Data, resp.Data)
	}
}

func TestRead_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_READ); err != nil {
		t.Errorf("READ request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_READ); err != nil {
		t.Errorf("READ response dispatch: %v", err)
	}
}
