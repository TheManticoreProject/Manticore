package commands_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestCloseRequest_RoundTrip(t *testing.T) {
	req := commands.NewCloseRequest()
	req.Flags = commands.SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB
	req.FileId = types.SMB2_FILEID{Persistent: 0xAABB, Volatile: 0xCCDD}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 24 || binary.LittleEndian.Uint16(wire[0:2]) != commands.CloseRequestStructureSize {
		t.Fatalf("wire len %d / StructureSize %d, want 24 / 24", len(wire), binary.LittleEndian.Uint16(wire[0:2]))
	}

	var decoded commands.CloseRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Flags != req.Flags || decoded.FileId != req.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestCloseResponse_RoundTrip(t *testing.T) {
	resp := commands.NewCloseResponse()
	resp.CreationTime = 0x01D0000000000001
	resp.LastWriteTime = 0x01D0000000000002
	resp.AllocationSize = 0x1000
	resp.EndOfFile = 0x0FFF
	resp.FileAttributes = 0x00000020 // FILE_ATTRIBUTE_ARCHIVE

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 60 || binary.LittleEndian.Uint16(wire[0:2]) != commands.CloseResponseStructureSize {
		t.Fatalf("wire len %d / StructureSize %d, want 60 / 60", len(wire), binary.LittleEndian.Uint16(wire[0:2]))
	}

	// Construct via the constructor so the embedded CommandCode (a header field,
	// not part of the body) matches; Unmarshal only populates body fields.
	decoded := commands.NewCloseResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if *decoded != *resp {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", *decoded, *resp)
	}
}

func TestClose_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_CLOSE); err != nil {
		t.Errorf("CLOSE request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_CLOSE); err != nil {
		t.Errorf("CLOSE response dispatch: %v", err)
	}
}
