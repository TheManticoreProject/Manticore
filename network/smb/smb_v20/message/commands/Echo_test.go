package commands_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
)

func TestEchoRequest_RoundTrip(t *testing.T) {
	req := commands.NewEchoRequest()
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 4 || binary.LittleEndian.Uint16(wire[0:2]) != commands.EchoRequestStructureSize {
		t.Fatalf("wire = % x, want 4-byte StructureSize=4", wire)
	}
	var decoded commands.EchoRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.StructureSize != commands.EchoRequestStructureSize {
		t.Errorf("StructureSize = %d, want 4", decoded.StructureSize)
	}
}

func TestEchoResponse_RoundTrip(t *testing.T) {
	resp := commands.NewEchoResponse()
	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 4 || binary.LittleEndian.Uint16(wire[0:2]) != commands.EchoResponseStructureSize {
		t.Fatalf("wire = % x, want 4-byte StructureSize=4", wire)
	}
	var decoded commands.EchoResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

func TestEcho_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_ECHO); err != nil {
		t.Errorf("ECHO request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_ECHO); err != nil {
		t.Errorf("ECHO response dispatch: %v", err)
	}
}
