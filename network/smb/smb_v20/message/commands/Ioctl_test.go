package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestIoctlRequest_RoundTrip(t *testing.T) {
	req := commands.NewIoctlRequest()
	req.CtlCode = 0x0011C017 // FSCTL_PIPE_TRANSCEIVE
	req.Flags = commands.SMB2_0_IOCTL_IS_FSCTL
	req.MaxOutputResponse = 0x10000
	req.FileId = types.SMB2_FILEID{Persistent: 0x1, Volatile: 0x2}
	req.Input = []byte{0x05, 0x00, 0x0B, 0x03} // pretend DCERPC bind fragment

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.IoctlRequestStructureSize {
		t.Errorf("StructureSize = %d, want 57", got)
	}
	// InputOffset is header-relative: 64 + 56 = 120.
	if got := binary.LittleEndian.Uint32(wire[24:28]); got != 120 {
		t.Errorf("InputOffset = %d, want 120", got)
	}

	decoded := commands.NewIoctlRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.CtlCode != req.CtlCode || decoded.Flags != req.Flags || !bytes.Equal(decoded.Input, req.Input) {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestIoctlResponse_RoundTrip(t *testing.T) {
	resp := commands.NewIoctlResponse()
	resp.CtlCode = 0x0011C017
	resp.FileId = types.SMB2_FILEID{Persistent: 0x3, Volatile: 0x4}
	resp.Output = []byte{0x05, 0x00, 0x0C, 0x03, 0x10, 0x00}

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.IoctlResponseStructureSize {
		t.Errorf("StructureSize = %d, want 49", got)
	}
	// With no input, output starts at header-relative 64 + 48 = 112 (already 8-aligned).
	if got := binary.LittleEndian.Uint32(wire[32:36]); got != 112 {
		t.Errorf("OutputOffset = %d, want 112", got)
	}

	decoded := commands.NewIoctlResponse()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.CtlCode != resp.CtlCode || !bytes.Equal(decoded.Output, resp.Output) {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestIoctl_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_IOCTL); err != nil {
		t.Errorf("IOCTL request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_IOCTL); err != nil {
		t.Errorf("IOCTL response dispatch: %v", err)
	}
}
