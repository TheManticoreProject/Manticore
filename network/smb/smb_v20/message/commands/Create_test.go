package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestCreateRequest_RoundTripWithName(t *testing.T) {
	req := commands.NewCreateRequest()
	req.RequestedOplockLevel = commands.SMB2_OPLOCK_LEVEL_NONE
	req.ImpersonationLevel = 0x00000002 // Impersonation
	req.DesiredAccess = 0x00100081
	req.ShareAccess = 0x00000007 // READ|WRITE|DELETE
	req.CreateDisposition = 0x00000001 // FILE_OPEN
	req.CreateOptions = 0x00000040     // FILE_NON_DIRECTORY_FILE
	req.Name = `dir\file.txt`

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.CreateRequestStructureSize {
		t.Errorf("StructureSize = %d, want 57", got)
	}
	// NameOffset is header-relative: 64 + 56 = 120; NameLength is the UTF-16 byte count.
	if got := binary.LittleEndian.Uint16(wire[44:46]); got != 120 {
		t.Errorf("NameOffset = %d, want 120", got)
	}
	if got := binary.LittleEndian.Uint16(wire[46:48]); int(got) != len(req.Name)*2 {
		t.Errorf("NameLength = %d, want %d", got, len(req.Name)*2)
	}

	decoded := commands.NewCreateRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Name != req.Name || decoded.DesiredAccess != req.DesiredAccess ||
		decoded.CreateDisposition != req.CreateDisposition || decoded.CreateOptions != req.CreateOptions {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestCreateRequest_EmptyNameHasMinimumBuffer(t *testing.T) {
	req := commands.NewCreateRequest()
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Buffer MUST be at least one byte even with no name and no contexts.
	if len(wire) != commands.CreateRequestStructureSize-1+1 {
		t.Errorf("wire length = %d, want %d (56 fixed + 1 buffer byte)", len(wire), 57)
	}
	if got := binary.LittleEndian.Uint16(wire[46:48]); got != 0 {
		t.Errorf("NameLength = %d, want 0", got)
	}
}

func TestCreateRequest_WithCreateContexts(t *testing.T) {
	req := commands.NewCreateRequest()
	req.Name = "f" // 2 UTF-16 bytes -> name ends at body offset 58 (header 122), needs pad to 8
	req.CreateContexts = []byte{0x10, 0x20, 0x30, 0x40}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	ccOffset := binary.LittleEndian.Uint32(wire[48:52])
	ccLength := binary.LittleEndian.Uint32(wire[52:56])
	if ccLength != 4 {
		t.Errorf("CreateContextsLength = %d, want 4", ccLength)
	}
	if ccOffset%8 != 0 {
		t.Errorf("CreateContextsOffset = %d, not 8-byte aligned", ccOffset)
	}

	decoded := commands.NewCreateRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Name != req.Name || !bytes.Equal(decoded.CreateContexts, req.CreateContexts) {
		t.Errorf("round-trip mismatch: name=%q contexts=% x", decoded.Name, decoded.CreateContexts)
	}
}

func TestCreateResponse_RoundTrip(t *testing.T) {
	resp := commands.NewCreateResponse()
	resp.OplockLevel = commands.SMB2_OPLOCK_LEVEL_NONE
	resp.CreateAction = 0x00000001 // FILE_OPENED
	resp.EndOfFile = 0x1234
	resp.FileAttributes = 0x00000080 // FILE_ATTRIBUTE_NORMAL
	resp.FileId = types.SMB2_FILEID{Persistent: 0xCAFE, Volatile: 0xBEEF}

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.CreateResponseStructureSize {
		t.Errorf("StructureSize = %d, want 89", got)
	}
	if len(wire) != 88 {
		t.Errorf("wire length = %d, want 88 (no create contexts)", len(wire))
	}

	decoded := commands.NewCreateResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.CreateAction != resp.CreateAction || decoded.EndOfFile != resp.EndOfFile || decoded.FileId != resp.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestCreate_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_CREATE); err != nil {
		t.Errorf("CREATE request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_CREATE); err != nil {
		t.Errorf("CREATE response dispatch: %v", err)
	}
}
