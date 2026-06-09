package commands_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestFlushRequest_RoundTrip(t *testing.T) {
	req := commands.NewFlushRequest()
	req.FileId = types.SMB2_FILEID{Persistent: 0x1122334455667788, Volatile: 0x99AABBCCDDEEFF00}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 24 || binary.LittleEndian.Uint16(wire[0:2]) != commands.FlushRequestStructureSize {
		t.Fatalf("wire len %d / StructureSize %d, want 24 / 24", len(wire), binary.LittleEndian.Uint16(wire[0:2]))
	}

	var decoded commands.FlushRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.FileId != req.FileId {
		t.Errorf("FileId = %+v, want %+v", decoded.FileId, req.FileId)
	}
}

func TestFlushResponse_RoundTrip(t *testing.T) {
	resp := commands.NewFlushResponse()
	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 4 || binary.LittleEndian.Uint16(wire[0:2]) != commands.FlushResponseStructureSize {
		t.Fatalf("wire = % x, want 4-byte StructureSize=4", wire)
	}
	var decoded commands.FlushResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

func TestFlush_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_FLUSH); err != nil {
		t.Errorf("FLUSH request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_FLUSH); err != nil {
		t.Errorf("FLUSH response dispatch: %v", err)
	}
}
