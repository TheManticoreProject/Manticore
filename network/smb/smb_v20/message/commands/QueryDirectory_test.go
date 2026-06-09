package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestQueryDirectoryRequest_RoundTrip(t *testing.T) {
	req := commands.NewQueryDirectoryRequest()
	req.FileInformationClass = 0x25 // FileIdBothDirectoryInformation
	req.Flags = commands.SMB2_RESTART_SCANS
	req.OutputBufferLength = 0x10000
	req.FileId = types.SMB2_FILEID{Persistent: 0x9, Volatile: 0xA}
	req.FileName = "*"

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.QueryDirectoryRequestStructureSize {
		t.Errorf("StructureSize = %d, want 33", got)
	}
	// FileNameOffset is header-relative: 64 + 32 = 96.
	if got := binary.LittleEndian.Uint16(wire[24:26]); got != 96 {
		t.Errorf("FileNameOffset = %d, want 96", got)
	}

	decoded := commands.NewQueryDirectoryRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.FileInformationClass != req.FileInformationClass || decoded.FileName != req.FileName || decoded.FileId != req.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestQueryDirectoryResponse_RoundTrip(t *testing.T) {
	resp := commands.NewQueryDirectoryResponse()
	resp.OutputBuffer = []byte{0x60, 0x00, 0x00, 0x00, 0x00, 0x00} // pretend one FILE_*_INFORMATION entry

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.QueryDirectoryResponseStructureSize {
		t.Errorf("StructureSize = %d, want 9", got)
	}
	// OutputBufferOffset is header-relative: 64 + 8 = 72.
	if got := binary.LittleEndian.Uint16(wire[2:4]); got != 72 {
		t.Errorf("OutputBufferOffset = %d, want 72", got)
	}

	decoded := commands.NewQueryDirectoryResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(decoded.OutputBuffer, resp.OutputBuffer) {
		t.Errorf("OutputBuffer = % x, want % x", decoded.OutputBuffer, resp.OutputBuffer)
	}
}

func TestQueryDirectory_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_QUERY_DIRECTORY); err != nil {
		t.Errorf("QUERY_DIRECTORY request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_QUERY_DIRECTORY); err != nil {
		t.Errorf("QUERY_DIRECTORY response dispatch: %v", err)
	}
}
