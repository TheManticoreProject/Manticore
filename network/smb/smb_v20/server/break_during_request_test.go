package server_test

import (
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestServerOplockBreakDuringRequest checks that an oplock break the server sends
// while the holder is waiting for the response to another request is still
// delivered by WaitBreakNotification, on signed and encrypted sessions.
func TestServerOplockBreakDuringRequest(t *testing.T) {
	for _, config := range []server.Config{{}, {RequireEncryption: true}} {
		_, port := startServer(t, config)
		holder, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		id, level, err := holder.CreateFileWithOplock("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
			fileflags.FILE_OPEN, 0, commands.SMB2_OPLOCK_LEVEL_BATCH)
		if err != nil || level != commands.SMB2_OPLOCK_LEVEL_II {
			t.Fatalf("open with oplock = level %d, %v; want level II", level, err)
		}

		writer, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.CreateFile("hello.txt", fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, 0); err != nil {
			t.Fatalf("open for writing: %v", err)
		}
		// The server sends the break from its own goroutine; let it reach the
		// holder's socket ahead of the ECHO response.
		time.Sleep(100 * time.Millisecond)
		if err := holder.Echo(); err != nil {
			t.Fatalf("Echo (encryption %v): %v", config.RequireEncryption, err)
		}

		type result struct {
			brk *client.BreakNotification
			err error
		}
		done := make(chan result, 1)
		go func() {
			brk, err := holder.WaitBreakNotification()
			done <- result{brk, err}
		}()
		select {
		case r := <-done:
			if r.err != nil {
				t.Fatalf("WaitBreakNotification (encryption %v): %v", config.RequireEncryption, r.err)
			}
			if b := r.brk.OplockBreak; b == nil || b.FileId != id || b.NewLevel != commands.SMB2_OPLOCK_LEVEL_NONE {
				t.Fatalf("break = %+v; want FileId %+v broken to level none", b, id)
			}
		case <-time.After(2 * time.Second):
			holder.Disconnect()
			<-done
			t.Fatalf("WaitBreakNotification (encryption %v): the break received during Echo was lost", config.RequireEncryption)
		}
	}
}
