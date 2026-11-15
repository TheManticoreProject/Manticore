package server

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestLockAndReadThenWriteAndUnlock drives SMB_COM_LOCK_AND_READ and
// SMB_COM_WRITE_AND_UNLOCK end to end against the server and proves both the data
// transfer and the lock side effects: LOCK_AND_READ returns the bytes and leaves
// the range locked against another handle, and WRITE_AND_UNLOCK writes and then
// releases the range. These commands back the advertised CAP_LOCK_AND_READ.
func TestLockAndReadThenWriteAndUnlock(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	// LOCK_AND_READ [0,8) on the first handle.
	req := newRequest(codes.SMB_COM_LOCK_AND_READ)
	req.Header.UID = client.Session.SessionUID
	req.Header.TID = client.Session.TreeID
	lar := commands.NewLockAndReadRequest()
	lar.FID = types.USHORT(first)
	lar.CountOfBytesToRead = 8
	lar.ReadOffsetInBytes = 0
	req.AddCommand(lar)
	marshalled, err := req.Marshal()
	if err != nil {
		t.Fatalf("marshal LOCK_AND_READ: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("send LOCK_AND_READ: %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("receive LOCK_AND_READ: %v", err)
	}
	if st := replyStatus(raw); st != uint32(nt_status.NT_STATUS_SUCCESS) {
		t.Fatalf("LOCK_AND_READ status = 0x%08X, want success", st)
	}

	resp := message.NewMessage()
	if err := resp.Unmarshal(raw); err != nil {
		t.Fatalf("decode LOCK_AND_READ response: %v", err)
	}
	larResp, ok := resp.Command.(*commands.LockAndReadResponse)
	if !ok {
		t.Fatalf("response is %T, want *LockAndReadResponse", resp.Command)
	}
	if got := string([]byte(larResp.BytesRead.Buffer)); got != "01234567" {
		t.Fatalf("LOCK_AND_READ returned %q, want %q", got, "01234567")
	}

	// The range is now locked: another handle cannot take it.
	if err := client.LockFile(second, 0, 8, true); err == nil {
		t.Fatal("LOCK_AND_READ did not leave the range locked: a second handle locked it")
	}

	// WRITE_AND_UNLOCK [0,8) on the first handle, writing new bytes and releasing
	// the lock LOCK_AND_READ took.
	wreq := newRequest(codes.SMB_COM_WRITE_AND_UNLOCK)
	wreq.Header.UID = client.Session.SessionUID
	wreq.Header.TID = client.Session.TreeID
	wau := commands.NewWriteAndUnlockRequest()
	wau.FID = types.USHORT(first)
	wau.CountOfBytesToWrite = 8
	wau.WriteOffsetInBytes = 0
	wau.Data.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	wau.Data.Buffer = []types.UCHAR("ABCDEFGH")
	wreq.AddCommand(wau)
	wmarshalled, err := wreq.Marshal()
	if err != nil {
		t.Fatalf("marshal WRITE_AND_UNLOCK: %v", err)
	}
	if _, err := client.Transport.Send(wmarshalled); err != nil {
		t.Fatalf("send WRITE_AND_UNLOCK: %v", err)
	}
	wraw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("receive WRITE_AND_UNLOCK: %v", err)
	}
	if st := replyStatus(wraw); st != uint32(nt_status.NT_STATUS_SUCCESS) {
		t.Fatalf("WRITE_AND_UNLOCK status = 0x%08X, want success", st)
	}
	wresp := message.NewMessage()
	if err := wresp.Unmarshal(wraw); err != nil {
		t.Fatalf("decode WRITE_AND_UNLOCK response: %v", err)
	}
	wauResp, ok := wresp.Command.(*commands.WriteAndUnlockResponse)
	if !ok {
		t.Fatalf("response is %T, want *WriteAndUnlockResponse", wresp.Command)
	}
	if wauResp.CountOfBytesWritten != 8 {
		t.Fatalf("WRITE_AND_UNLOCK wrote %d bytes, want 8", wauResp.CountOfBytesWritten)
	}

	// The range is unlocked again: the other handle can now take it.
	if err := client.LockFile(second, 0, 8, true); err != nil {
		t.Fatalf("WRITE_AND_UNLOCK did not release the range: %v", err)
	}
	_ = client.UnlockFile(second, 0, 8)

	// The write took effect.
	data, err := client.ReadFile(first, 0, 16)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got := string(data); got != "ABCDEFGH89abcdef" {
		t.Fatalf("file after WRITE_AND_UNLOCK = %q, want %q", got, "ABCDEFGH89abcdef")
	}
}

// TestCapLockAndReadAdvertised checks the server advertises CAP_LOCK_AND_READ now
// that the backing commands are implemented, keeping the advertisement and the
// behaviour in step.
func TestCapLockAndReadAdvertised(t *testing.T) {
	if serverCapabilities&capabilities.CAP_LOCK_AND_READ == 0 {
		t.Fatal("CAP_LOCK_AND_READ is implemented but not advertised in serverCapabilities")
	}
}
