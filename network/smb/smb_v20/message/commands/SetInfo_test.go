package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestSetInfoRequest_RoundTrip(t *testing.T) {
	req := commands.NewSetInfoRequest()
	req.InfoType = commands.SMB2_0_INFO_FILE
	req.FileInfoClass = 0x0D // FileDispositionInformation
	req.FileId = types.SMB2_FILEID{Persistent: 0x3, Volatile: 0x4}
	req.Buffer = []byte{0x01} // DeletePending = TRUE

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.SetInfoRequestStructureSize {
		t.Errorf("StructureSize = %d, want 33", got)
	}
	// BufferOffset is header-relative: 64 + 32 = 96; BufferLength matches the payload.
	if got := binary.LittleEndian.Uint16(wire[8:10]); got != 96 {
		t.Errorf("BufferOffset = %d, want 96", got)
	}
	if got := binary.LittleEndian.Uint32(wire[4:8]); int(got) != len(req.Buffer) {
		t.Errorf("BufferLength = %d, want %d", got, len(req.Buffer))
	}

	decoded := commands.NewSetInfoRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.InfoType != req.InfoType || decoded.FileInfoClass != req.FileInfoClass ||
		decoded.FileId != req.FileId || !bytes.Equal(decoded.Buffer, req.Buffer) {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestSetInfoResponse_RoundTrip(t *testing.T) {
	resp := commands.NewSetInfoResponse()
	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 2 || binary.LittleEndian.Uint16(wire[0:2]) != commands.SetInfoResponseStructureSize {
		t.Fatalf("wire = % x, want 2-byte StructureSize=2", wire)
	}
	var decoded commands.SetInfoResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

func TestSetInfo_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_SET_INFO); err != nil {
		t.Errorf("SET_INFO request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_SET_INFO); err != nil {
		t.Errorf("SET_INFO response dispatch: %v", err)
	}
}
