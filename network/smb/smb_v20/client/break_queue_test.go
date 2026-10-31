package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// TestWaitOplockBreakReturnsBreakSkippedByRequest verifies that an OPLOCK_BREAK
// notification received while a request waits for its response is not lost:
// the request still completes with its own response, and the next
// WaitOplockBreak returns the break instead of reading the connection (where an
// unrelated ECHO response stands in for "no further break").
func TestWaitOplockBreakReturnsBreakSkippedByRequest(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{
		buildOplockBreakWire(t),
		cannedResponse(t, commands.NewCloseResponse(), 0, 0x5, 0x99),
		cannedResponse(t, commands.NewEchoResponse(), 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	if err := c.CloseFile(types.SMB2_FILEID{}); err != nil {
		t.Fatalf("CloseFile: %v", err)
	}
	brk, err := c.WaitOplockBreak()
	if err != nil {
		t.Fatalf("WaitOplockBreak: %v", err)
	}
	if brk.FileId != (types.SMB2_FILEID{Persistent: 0xAAAA, Volatile: 0xBBBB}) || brk.NewLevel != commands.SMB2_OPLOCK_LEVEL_II {
		t.Errorf("break = %+v, want FileId {0xAAAA 0xBBBB} at level II", brk)
	}
	if len(ft.responses) != 1 {
		t.Errorf("WaitOplockBreak read the connection although a break was queued (%d frames left, want 1)", len(ft.responses))
	}
}

// TestWaitBreakNotificationReturnsBreakSkippedByCompound verifies the same for a
// lease break that precedes a compounded response.
func TestWaitBreakNotificationReturnsBreakSkippedByCompound(t *testing.T) {
	key := [16]byte{0xDE, 0xAD, 0xBE, 0xEF}
	ft := &fakeTransport{responses: [][]byte{
		buildLeaseBreakWire(t, key, 0x07, 0x01, 3, true),
		createQueryCloseResponse(t, []byte{0x01}),
		cannedResponse(t, commands.NewEchoResponse(), 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	if _, err := c.CreateQueryInfoClose("file.txt", 0x00120089, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x05, 0); err != nil {
		t.Fatalf("CreateQueryInfoClose: %v", err)
	}
	bn, err := c.WaitBreakNotification()
	if err != nil {
		t.Fatalf("WaitBreakNotification: %v", err)
	}
	if bn.LeaseBreak == nil || bn.LeaseBreak.LeaseKey != key || bn.LeaseBreak.NewLeaseState != 0x01 {
		t.Fatalf("break = %+v, want the queued lease break", bn)
	}
	if len(ft.responses) != 1 {
		t.Errorf("WaitBreakNotification read the connection although a break was queued (%d frames left, want 1)", len(ft.responses))
	}
}

// TestBreakQueueIsBounded verifies that the queue of set-aside breaks keeps at
// most maxQueuedBreaks entries, dropping the oldest.
func TestBreakQueueIsBounded(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	for i := 0; i < maxQueuedBreaks+5; i++ {
		c.queueBreak([]byte{byte(i)}, false)
	}
	if len(c.pendingBreaks) != maxQueuedBreaks {
		t.Fatalf("queued %d breaks, want %d", len(c.pendingBreaks), maxQueuedBreaks)
	}
	if first := c.pendingBreaks[0].raw[0]; first != 5 {
		t.Errorf("oldest queued break = %d, want 5 (the 5 oldest dropped)", first)
	}
}
