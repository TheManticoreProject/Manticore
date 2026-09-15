package client_test

import (
	"os"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// sentHeaderIdentifiers decodes the PID and MID of the captured request.
func sentHeaderIdentifiers(t *testing.T, raw []byte) (uint32, uint16) {
	t.Helper()
	msg := message.NewMessage()
	if err := msg.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode sent request: %v", err)
	}
	return uint32(msg.Header.GetPID()), uint16(msg.Header.MID)
}

// TestFileIORequestsCarryRealIdentifiers is the regression guard for the defect:
// every request went out with PID 0 and a MID equal to the server's advertised
// MaxMpxCount, so no request on a connection could be told from any other.
func TestFileIORequestsCarryRealIdentifiers(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewFlushResponse())}
	c := newSessionClient(tr)
	// The value the MID used to be taken from. Nothing may depend on it.
	c.Connection.MaxMpxCount = 50

	if err := c.Flush(1); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	firstPID, firstMID := sentHeaderIdentifiers(t, tr.sent)

	if firstPID == 0 {
		t.Error("request carries PID 0")
	}
	if want := uint32(os.Getpid()); firstPID != want && firstPID != want|1 {
		t.Errorf("request PID = %d, want this process (%d)", firstPID, want)
	}
	if firstMID == c.Connection.MaxMpxCount {
		t.Errorf("request MID = %d, which is the server's MaxMpxCount", firstMID)
	}

	if err := c.Flush(1); err != nil {
		t.Fatalf("second Flush: %v", err)
	}
	secondPID, secondMID := sentHeaderIdentifiers(t, tr.sent)

	if secondMID == firstMID {
		t.Errorf("two requests share MID %d", firstMID)
	}
	if secondPID != firstPID {
		t.Errorf("PID changed between requests: %d then %d", firstPID, secondPID)
	}
}

// TestLockRangeOwnerIsTheClientProcess checks the knock-on effect: a byte-range
// lock names its owner by the low half of the request PID, so a zero PID made
// every lock this client took claim the same owner.
func TestLockRangeOwnerIsTheClientProcess(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewLockingAndxResponse())}
	c := newSessionClient(tr)

	if err := c.LockFile(3, 0, 16, true); err != nil {
		t.Fatalf("LockFile: %v", err)
	}

	pid, _ := sentHeaderIdentifiers(t, tr.sent)
	req := sentLockingAndx(t, tr.sent)
	if len(req.Locks) != 1 {
		t.Fatalf("expected 1 lock range, got %d", len(req.Locks))
	}
	if req.Locks[0].PID == 0 {
		t.Error("lock range claims owner PID 0")
	}
	if uint16(req.Locks[0].PID) != uint16(pid) {
		t.Errorf("lock owner PID = %d, want the request PID low half %d", req.Locks[0].PID, uint16(pid))
	}
}
