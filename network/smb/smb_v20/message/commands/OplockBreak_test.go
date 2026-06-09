package commands_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestOplockBreakRequest_RoundTrip(t *testing.T) {
	req := commands.NewOplockBreakRequest()
	req.OplockLevel = commands.SMB2_OPLOCK_LEVEL_II
	req.FileId = types.SMB2_FILEID{Persistent: 0xF1, Volatile: 0xF2}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 24 || binary.LittleEndian.Uint16(wire[0:2]) != commands.OplockBreakRequestStructureSize {
		t.Fatalf("wire len %d / StructureSize %d, want 24 / 24", len(wire), binary.LittleEndian.Uint16(wire[0:2]))
	}

	decoded := commands.NewOplockBreakRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.OplockLevel != req.OplockLevel || decoded.FileId != req.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestOplockBreakResponse_RoundTrip(t *testing.T) {
	resp := commands.NewOplockBreakResponse()
	resp.OplockLevel = commands.SMB2_OPLOCK_LEVEL_NONE
	resp.FileId = types.SMB2_FILEID{Persistent: 0xA1, Volatile: 0xA2}

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 24 || binary.LittleEndian.Uint16(wire[0:2]) != commands.OplockBreakResponseStructureSize {
		t.Fatalf("wire len %d / StructureSize %d, want 24 / 24", len(wire), binary.LittleEndian.Uint16(wire[0:2]))
	}

	decoded := commands.NewOplockBreakResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.OplockLevel != resp.OplockLevel || decoded.FileId != resp.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestOplockBreak_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_OPLOCK_BREAK); err != nil {
		t.Errorf("OPLOCK_BREAK request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_OPLOCK_BREAK); err != nil {
		t.Errorf("OPLOCK_BREAK response dispatch: %v", err)
	}
}
