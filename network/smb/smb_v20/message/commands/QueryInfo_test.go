package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestQueryInfoRequest_RoundTrip(t *testing.T) {
	req := commands.NewQueryInfoRequest()
	req.InfoType = commands.SMB2_0_INFO_FILE
	req.FileInfoClass = 0x12 // FileAllInformation
	req.OutputBufferLength = 0x1000
	req.FileId = types.SMB2_FILEID{Persistent: 0x7, Volatile: 0x8}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.QueryInfoRequestStructureSize {
		t.Errorf("StructureSize = %d, want 41", got)
	}

	decoded := commands.NewQueryInfoRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.InfoType != req.InfoType || decoded.FileInfoClass != req.FileInfoClass ||
		decoded.OutputBufferLength != req.OutputBufferLength || decoded.FileId != req.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestQueryInfoRequest_WithInput(t *testing.T) {
	req := commands.NewQueryInfoRequest()
	req.InfoType = commands.SMB2_0_INFO_QUOTA
	req.Input = []byte{0x01, 0x02, 0x03, 0x04}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// InputBufferOffset is header-relative: 64 + 40 = 104.
	if got := binary.LittleEndian.Uint16(wire[8:10]); got != 104 {
		t.Errorf("InputBufferOffset = %d, want 104", got)
	}
	if got := binary.LittleEndian.Uint32(wire[12:16]); int(got) != len(req.Input) {
		t.Errorf("InputBufferLength = %d, want %d", got, len(req.Input))
	}

	decoded := commands.NewQueryInfoRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(decoded.Input, req.Input) {
		t.Errorf("Input = % x, want % x", decoded.Input, req.Input)
	}
}

func TestQueryInfoResponse_RoundTrip(t *testing.T) {
	resp := commands.NewQueryInfoResponse()
	resp.OutputBuffer = []byte{0xDE, 0xAD, 0xBE, 0xEF}

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[2:4]); got != 72 {
		t.Errorf("OutputBufferOffset = %d, want 72", got)
	}

	decoded := commands.NewQueryInfoResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(decoded.OutputBuffer, resp.OutputBuffer) {
		t.Errorf("OutputBuffer = % x, want % x", decoded.OutputBuffer, resp.OutputBuffer)
	}
}

func TestQueryInfo_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_QUERY_INFO); err != nil {
		t.Errorf("QUERY_INFO request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_QUERY_INFO); err != nil {
		t.Errorf("QUERY_INFO response dispatch: %v", err)
	}
}
