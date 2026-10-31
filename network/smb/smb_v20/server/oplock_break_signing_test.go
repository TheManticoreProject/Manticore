package server_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestServerOplockBreakOnSignedSession checks that the client accepts an
// OPLOCK_BREAK notification on a signed session. The notification carries the
// reserved MessageId 0xFFFFFFFFFFFFFFFF and is sent unsigned (MS-SMB2 3.3.4.6),
// which MS-SMB2 3.2.5.1.3 exempts from signature verification.
func TestServerOplockBreakOnSignedSession(t *testing.T) {
	for _, dialect := range []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0, dialects.SMB2_DIALECT_3_1_1} {
		t.Run(dialect.String(), func(t *testing.T) {
			_, port := startServer(t, server.Config{RequireSigning: true})
			holder, err := dial(t, port, dialect, "Passw0rd!")
			if err != nil {
				t.Fatal(err)
			}
			if !holder.Session.SigningActive {
				t.Fatal("signing is not active on the client")
			}
			id, level, err := holder.CreateFileWithOplock("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
				fileflags.FILE_OPEN, 0, commands.SMB2_OPLOCK_LEVEL_BATCH)
			if err != nil || level != commands.SMB2_OPLOCK_LEVEL_II {
				t.Fatalf("open with oplock = level %d, %v; want level II", level, err)
			}

			writer, err := dial(t, port, dialect, "Passw0rd!")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.CreateFile("hello.txt", fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, 0); err != nil {
				t.Fatalf("open for writing: %v", err)
			}

			brk, err := holder.WaitOplockBreak()
			if err != nil {
				t.Fatalf("WaitOplockBreak: %v", err)
			}
			if brk.FileId != id || brk.NewLevel != commands.SMB2_OPLOCK_LEVEL_NONE {
				t.Fatalf("break = %+v; want FileId %+v broken to level none", brk, id)
			}
		})
	}
}
