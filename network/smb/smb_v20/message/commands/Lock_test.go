package commands_test

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestLockRequest_RoundTrip(t *testing.T) {
	req := commands.NewLockRequest()
	req.FileId = types.SMB2_FILEID{Persistent: 0x1, Volatile: 0x2}
	req.Locks = []commands.LockElement{
		{Offset: 0, Length: 0x100, Flags: commands.SMB2_LOCKFLAG_EXCLUSIVE_LOCK | commands.SMB2_LOCKFLAG_FAIL_IMMEDIATELY},
		{Offset: 0x200, Length: 0x50, Flags: commands.SMB2_LOCKFLAG_UNLOCK},
	}

	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// StructureSize is always 48; LockCount reflects the array; body grows per element.
	if got := binary.LittleEndian.Uint16(wire[0:2]); got != commands.LockRequestStructureSize {
		t.Errorf("StructureSize = %d, want 48", got)
	}
	if got := binary.LittleEndian.Uint16(wire[2:4]); got != 2 {
		t.Errorf("LockCount = %d, want 2", got)
	}
	if len(wire) != 24+2*commands.LockElementSize {
		t.Errorf("wire length = %d, want %d", len(wire), 24+2*commands.LockElementSize)
	}

	decoded := commands.NewLockRequest()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.FileId != req.FileId || !reflect.DeepEqual(decoded.Locks, req.Locks) {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestLockResponse_RoundTrip(t *testing.T) {
	resp := commands.NewLockResponse()
	wire, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 4 || binary.LittleEndian.Uint16(wire[0:2]) != commands.LockResponseStructureSize {
		t.Fatalf("wire = % x, want 4-byte StructureSize=4", wire)
	}
	var decoded commands.LockResponse
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

func TestLock_Dispatcher(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_LOCK); err != nil {
		t.Errorf("LOCK request dispatch: %v", err)
	}
	if _, err := commands.CreateResponseCommand(codes.SMB2_LOCK); err != nil {
		t.Errorf("LOCK response dispatch: %v", err)
	}
}
