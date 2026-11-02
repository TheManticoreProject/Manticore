package client

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

// TestCancelTargetsASynchronousRequest checks a request with no interim
// response yet is cancelled in the synchronous form: no ASYNC flag, and the
// request's MessageId, TreeId and SessionId ([MS-SMB2] 3.2.4.24). Before, Cancel
// refused because no AsyncId was known.
func TestCancelTargetsASynchronousRequest(t *testing.T) {
	ft := &fakeTransport{}
	c := withConnectedTree(ft)
	c.setOutstanding(17, 0x5, 0x99)

	if err := c.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if len(ft.sent) != 1 || len(ft.sent[0]) < 64 {
		t.Fatalf("sent %d frames", len(ft.sent))
	}
	sent := ft.sent[0]
	if flags.Flags(binary.LittleEndian.Uint32(sent[16:20])).IsAsync() {
		t.Error("a synchronous CANCEL set SMB2_FLAGS_ASYNC_COMMAND")
	}
	if got := binary.LittleEndian.Uint64(sent[24:32]); got != 17 {
		t.Errorf("MessageId = %d, want 17", got)
	}
	if got := binary.LittleEndian.Uint32(sent[36:40]); got != 0x5 {
		t.Errorf("TreeId = %#x, want 0x5", got)
	}
	if got := binary.LittleEndian.Uint64(sent[40:48]); got != 0x99 {
		t.Errorf("SessionId = %#x, want 0x99", got)
	}
}

// TestCancelPrefersTheAsyncForm checks an interim response's AsyncId takes
// precedence over the synchronous record.
func TestCancelPrefersTheAsyncForm(t *testing.T) {
	ft := &fakeTransport{}
	c := withConnectedTree(ft)
	c.setOutstanding(17, 0x5, 0x99)
	c.setPendingAsync(17, 0xABCD)
	if err := c.Cancel(); err != nil {
		t.Fatal(err)
	}
	if !flags.Flags(binary.LittleEndian.Uint32(ft.sent[0][16:20])).IsAsync() {
		t.Error("an async request was cancelled in the synchronous form")
	}
}
