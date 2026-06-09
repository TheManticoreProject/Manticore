package commands_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
)

func TestCancelRequest_RoundTrip(t *testing.T) {
	req := commands.NewCancelRequest()
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 4 || binary.LittleEndian.Uint16(wire[0:2]) != commands.CancelRequestStructureSize {
		t.Fatalf("wire = % x, want 4-byte StructureSize=4", wire)
	}
	var decoded commands.CancelRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.StructureSize != commands.CancelRequestStructureSize {
		t.Errorf("StructureSize = %d, want 4", decoded.StructureSize)
	}
}

func TestCancel_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_CANCEL); err != nil {
		t.Errorf("CANCEL request dispatch: %v", err)
	}
	// CANCEL has no response; the response side MUST remain unsupported.
	if _, err := commands.CreateResponseCommand(codes.SMB2_CANCEL); err == nil {
		t.Errorf("expected CANCEL response to be unsupported (CANCEL has no response)")
	}
}
