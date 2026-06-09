package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestChangeNotifyRequest_RoundTrip(t *testing.T) {
	req := commands.NewChangeNotifyRequest()
	req.Flags = commands.SMB2_WATCH_TREE
	req.OutputBufferLength = 0x1000
	req.CompletionFilter = 0x00000017
	req.FileId = types.SMB2_FILEID{Persistent: 0xD, Volatile: 0xE}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 32 || binary.LittleEndian.Uint16(wire[0:2]) != commands.ChangeNotifyRequestStructureSize {
		t.Fatalf("wire len %d / StructureSize %d, want 32 / 32", len(wire), binary.LittleEndian.Uint16(wire[0:2]))
	}

	decoded := commands.NewChangeNotifyRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Flags != req.Flags || decoded.CompletionFilter != req.CompletionFilter || decoded.FileId != req.FileId {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestChangeNotifyResponse_RoundTrip(t *testing.T) {
	resp := commands.NewChangeNotifyResponse()
	resp.OutputBuffer = []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00} // pretend FILE_NOTIFY_INFORMATION

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.ChangeNotifyResponseStructureSize {
		t.Errorf("StructureSize = %d, want 9", got)
	}
	if got := binary.LittleEndian.Uint16(wire[2:4]); got != 72 {
		t.Errorf("OutputBufferOffset = %d, want 72", got)
	}

	decoded := commands.NewChangeNotifyResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(decoded.OutputBuffer, resp.OutputBuffer) {
		t.Errorf("OutputBuffer = % x, want % x", decoded.OutputBuffer, resp.OutputBuffer)
	}
}

func TestChangeNotify_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_CHANGE_NOTIFY); err != nil {
		t.Errorf("CHANGE_NOTIFY request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_CHANGE_NOTIFY); err != nil {
		t.Errorf("CHANGE_NOTIFY response dispatch: %v", err)
	}
}
