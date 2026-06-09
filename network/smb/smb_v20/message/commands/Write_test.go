package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestWriteRequest_RoundTrip(t *testing.T) {
	req := commands.NewWriteRequest()
	req.Offset = 0x4000
	req.FileId = types.SMB2_FILEID{Persistent: 0x11, Volatile: 0x22}
	req.Data = []byte("payload bytes")

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.WriteRequestStructureSize {
		t.Errorf("StructureSize = %d, want 49", got)
	}
	// DataOffset is header-relative: 64 + 48 = 112; Length matches the payload.
	if got := binary.LittleEndian.Uint16(wire[2:4]); got != 112 {
		t.Errorf("DataOffset = %d, want 112", got)
	}
	if got := binary.LittleEndian.Uint32(wire[4:8]); int(got) != len(req.Data) {
		t.Errorf("Length = %d, want %d", got, len(req.Data))
	}

	decoded := commands.NewWriteRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Offset != req.Offset || decoded.FileId != req.FileId || !bytes.Equal(decoded.Data, req.Data) {
		t.Errorf("round-trip mismatch: %+v data=%q", decoded, decoded.Data)
	}
}

func TestWriteResponse_RoundTrip(t *testing.T) {
	resp := commands.NewWriteResponse()
	resp.Count = 13

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.WriteResponseStructureSize {
		t.Errorf("StructureSize = %d, want 17", got)
	}

	decoded := commands.NewWriteResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Count != resp.Count {
		t.Errorf("Count = %d, want %d", decoded.Count, resp.Count)
	}
}

func TestWrite_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_WRITE); err != nil {
		t.Errorf("WRITE request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_WRITE); err != nil {
		t.Errorf("WRITE response dispatch: %v", err)
	}
}
