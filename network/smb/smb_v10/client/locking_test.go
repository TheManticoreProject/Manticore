package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// sentLockingAndx decodes the captured request as a LockingAndxRequest.
func sentLockingAndx(t *testing.T, raw []byte) *commands.LockingAndxRequest {
	t.Helper()
	msg := message.NewMessage()
	if err := msg.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode sent request: %v", err)
	}
	req, ok := msg.Command.(*commands.LockingAndxRequest)
	if !ok {
		t.Fatalf("sent command is %T, want *LockingAndxRequest", msg.Command)
	}
	return req
}

func TestLockFileExclusiveRequestShape(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewLockingAndxResponse())}
	c := newSessionClient(tr)

	if err := c.LockFile(5, 0x1_0000_0020, 0x100, true); err != nil {
		t.Fatalf("LockFile: %v", err)
	}

	req := sentLockingAndx(t, tr.sent)
	// LARGE_FILES (0x10) set, SHARED_LOCK (0x01) clear for an exclusive lock.
	if req.TypeOfLock&0x10 == 0 {
		t.Errorf("expected LARGE_FILES bit set, TypeOfLock = 0x%02x", req.TypeOfLock)
	}
	if req.TypeOfLock&0x01 != 0 {
		t.Errorf("expected SHARED_LOCK bit clear for exclusive lock, TypeOfLock = 0x%02x", req.TypeOfLock)
	}
	if req.FID != 5 || req.NumberOfRequestedLocks != 1 || req.NumberOfRequestedUnlocks != 0 {
		t.Errorf("unexpected FID/counts: FID=%d locks=%d unlocks=%d", req.FID, req.NumberOfRequestedLocks, req.NumberOfRequestedUnlocks)
	}
	if len(req.Locks) != 1 {
		t.Fatalf("expected 1 lock range, got %d", len(req.Locks))
	}
	lock := req.Locks[0]
	gotOffset := uint64(lock.ByteOffsetHigh)<<32 | uint64(lock.ByteOffsetLow)
	gotLength := uint64(lock.LengthInBytesHigh)<<32 | uint64(lock.LengthInBytesLow)
	if gotOffset != 0x1_0000_0020 || gotLength != 0x100 {
		t.Errorf("range mismatch: offset=0x%x length=0x%x", gotOffset, gotLength)
	}
}

func TestLockFileSharedSetsSharedBit(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewLockingAndxResponse())}
	c := newSessionClient(tr)

	if err := c.LockFile(1, 0, 10, false); err != nil {
		t.Fatalf("LockFile (shared): %v", err)
	}
	req := sentLockingAndx(t, tr.sent)
	if req.TypeOfLock&0x01 == 0 {
		t.Errorf("expected SHARED_LOCK bit set for a shared lock, TypeOfLock = 0x%02x", req.TypeOfLock)
	}
}

func TestUnlockFileRequestShape(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewLockingAndxResponse())}
	c := newSessionClient(tr)

	if err := c.UnlockFile(9, 0x40, 0x80); err != nil {
		t.Fatalf("UnlockFile: %v", err)
	}
	req := sentLockingAndx(t, tr.sent)
	if req.NumberOfRequestedUnlocks != 1 || req.NumberOfRequestedLocks != 0 {
		t.Errorf("expected 1 unlock and 0 locks, got unlocks=%d locks=%d", req.NumberOfRequestedUnlocks, req.NumberOfRequestedLocks)
	}
	if len(req.Unlocks) != 1 {
		t.Fatalf("expected 1 unlock range, got %d", len(req.Unlocks))
	}
}

func TestLockFileWithoutSession(t *testing.T) {
	c := &client.Client{Connection: &client.Connection{Server: &client.Server{}}}
	if err := c.LockFile(1, 0, 1, true); err == nil {
		t.Error("expected error without a session")
	}
}
