package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

func TestDispatcher_ConnectionSetupCommands(t *testing.T) {
	cases := []codes.CommandCode{
		codes.SMB2_NEGOTIATE, codes.SMB2_SESSION_SETUP, codes.SMB2_LOGOFF,
		codes.SMB2_TREE_CONNECT, codes.SMB2_TREE_DISCONNECT,
	}
	for _, code := range cases {
		req, err := commands.CreateRequestCommand(code)
		if err != nil {
			t.Errorf("CreateRequestCommand(%v) error: %v", code, err)
		} else if req.GetCommandCode() != code {
			t.Errorf("request command code = %v, want %v", req.GetCommandCode(), code)
		}
		resp, err := commands.CreateResponseCommand(code)
		if err != nil {
			t.Errorf("CreateResponseCommand(%v) error: %v", code, err)
		} else if resp.GetCommandCode() != code {
			t.Errorf("response command code = %v, want %v", resp.GetCommandCode(), code)
		}
	}

	// A code without a structure yet (CREATE) is still unsupported.
	if _, err := commands.CreateRequestCommand(codes.SMB2_CREATE); err == nil {
		t.Errorf("expected SMB2_CREATE request to be unsupported")
	}
	if _, err := commands.CreateRequestCommand(codes.CommandCode(0x00FF)); err == nil {
		t.Errorf("expected unknown code to be unsupported")
	}
}

func TestNegotiateRequest_RoundTrip(t *testing.T) {
	req := commands.NewNegotiateRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.Capabilities = capabilities.SMB2_GLOBAL_CAP_DFS
	for i := range req.ClientGuid {
		req.ClientGuid[i] = byte(i)
	}
	req.AddDialect(dialects.SMB2_DIALECT_2_0_2)
	req.AddDialect(dialects.SMB2_DIALECT_2_1_0)

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// StructureSize is the fixed constant 36, and DialectCount reflects the slice.
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.NegotiateRequestStructureSize {
		t.Errorf("StructureSize = %d, want 36", got)
	}
	if got := binary.LittleEndian.Uint16(wire[2:4]); got != 2 {
		t.Errorf("DialectCount = %d, want 2", got)
	}
	if len(wire) != 36+2*2 {
		t.Errorf("wire length = %d, want %d", len(wire), 36+2*2)
	}

	var decoded commands.NegotiateRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.SecurityMode != req.SecurityMode || decoded.Capabilities != req.Capabilities {
		t.Errorf("scalar round-trip mismatch: %+v", decoded)
	}
	if decoded.ClientGuid != req.ClientGuid {
		t.Errorf("ClientGuid mismatch")
	}
	if len(decoded.Dialects) != 2 || decoded.Dialects[0] != dialects.SMB2_DIALECT_2_0_2 || decoded.Dialects[1] != dialects.SMB2_DIALECT_2_1_0 {
		t.Errorf("Dialects mismatch: %v", decoded.Dialects)
	}
}

func TestNegotiateResponse_RoundTrip(t *testing.T) {
	resp := commands.NewNegotiateResponse()
	resp.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_REQUIRED
	resp.DialectRevision = dialects.SMB2_DIALECT_2_0_2
	resp.MaxTransactSize = 0x10000
	resp.MaxReadSize = 0x10000
	resp.MaxWriteSize = 0x10000
	resp.SystemTime = 0x01D0000000000000
	resp.SecurityBuffer = []byte{0x60, 0x28, 0x06, 0x06} // pretend GSS token

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.NegotiateResponseStructureSize {
		t.Errorf("StructureSize = %d, want 65", got)
	}
	// SecurityBufferOffset is header-relative: 64 + 64 = 128.
	if got := binary.LittleEndian.Uint16(wire[56:58]); got != 128 {
		t.Errorf("SecurityBufferOffset = %d, want 128", got)
	}
	if got := binary.LittleEndian.Uint16(wire[58:60]); got != 4 {
		t.Errorf("SecurityBufferLength = %d, want 4", got)
	}

	var decoded commands.NegotiateResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.DialectRevision != resp.DialectRevision || decoded.MaxReadSize != resp.MaxReadSize || decoded.SystemTime != resp.SystemTime {
		t.Errorf("scalar round-trip mismatch: %+v", decoded)
	}
	if !bytes.Equal(decoded.SecurityBuffer, resp.SecurityBuffer) {
		t.Errorf("SecurityBuffer = % x, want % x", decoded.SecurityBuffer, resp.SecurityBuffer)
	}
}

func TestSessionSetupRequest_RoundTrip(t *testing.T) {
	req := commands.NewSessionSetupRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.PreviousSessionId = 0
	req.SecurityBuffer = []byte{0x4E, 0x54, 0x4C, 0x4D} // "NTLM"

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.SessionSetupRequestStructureSize {
		t.Errorf("StructureSize = %d, want 25", got)
	}
	// SecurityBufferOffset is header-relative: 64 + 24 = 88.
	if got := binary.LittleEndian.Uint16(wire[12:14]); got != 88 {
		t.Errorf("SecurityBufferOffset = %d, want 88", got)
	}

	var decoded commands.SessionSetupRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.SecurityMode != req.SecurityMode || !bytes.Equal(decoded.SecurityBuffer, req.SecurityBuffer) {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestSessionSetupResponse_RoundTrip(t *testing.T) {
	resp := commands.NewSessionSetupResponse()
	resp.SessionFlags = commands.SMB2_SESSION_FLAG_IS_GUEST
	resp.SecurityBuffer = []byte{0xA1, 0x82}

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.SessionSetupResponseStructureSize {
		t.Errorf("StructureSize = %d, want 9", got)
	}
	// SecurityBufferOffset is header-relative: 64 + 8 = 72.
	if got := binary.LittleEndian.Uint16(wire[4:6]); got != 72 {
		t.Errorf("SecurityBufferOffset = %d, want 72", got)
	}

	var decoded commands.SessionSetupResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.SessionFlags != resp.SessionFlags || !bytes.Equal(decoded.SecurityBuffer, resp.SecurityBuffer) {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestTreeConnectRequest_RoundTrip(t *testing.T) {
	req := commands.NewTreeConnectRequest()
	req.Path = `\\server\share`

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.TreeConnectRequestStructureSize {
		t.Errorf("StructureSize = %d, want 9", got)
	}
	// PathOffset is header-relative: 64 + 8 = 72; PathLength is the UTF-16 byte count.
	if got := binary.LittleEndian.Uint16(wire[4:6]); got != 72 {
		t.Errorf("PathOffset = %d, want 72", got)
	}
	if got := binary.LittleEndian.Uint16(wire[6:8]); int(got) != len(req.Path)*2 {
		t.Errorf("PathLength = %d, want %d", got, len(req.Path)*2)
	}

	var decoded commands.TreeConnectRequest
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Path != req.Path {
		t.Errorf("Path = %q, want %q", decoded.Path, req.Path)
	}
}

func TestTreeConnectResponse_RoundTrip(t *testing.T) {
	resp := commands.NewTreeConnectResponse()
	resp.ShareType = commands.SMB2_SHARE_TYPE_DISK
	resp.ShareFlags = 0x00000030
	resp.MaximalAccess = 0x001F01FF

	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 16 {
		t.Fatalf("wire length = %d, want 16", len(wire))
	}

	var decoded commands.TreeConnectResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.ShareType != resp.ShareType || decoded.ShareFlags != resp.ShareFlags || decoded.MaximalAccess != resp.MaximalAccess {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestLogoffAndTreeDisconnect_RoundTrip(t *testing.T) {
	for _, c := range []interface {
		Marshal() ([]byte, error)
	}{
		commands.NewLogoffRequest(), commands.NewLogoffResponse(),
		commands.NewTreeDisconnectRequest(), commands.NewTreeDisconnectResponse(),
	} {
		wire, err := c.Marshal()
		if err != nil {
			t.Fatalf("Marshal %T: %v", c, err)
		}
		if len(wire) != 4 || binary.LittleEndian.Uint16(wire[0:2]) != 4 {
			t.Errorf("%T: wire = % x, want 4-byte StructureSize=4", c, wire)
		}
	}
}
