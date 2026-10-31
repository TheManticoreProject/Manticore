package server

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// renameShareAll is the share access every handle in these tests opens with, so
// that nothing but the rename itself is under test.
const renameShareAll = 7

// TestRenameByHandleMovesTheHandle renames a file through the handle that holds
// it for delete-on-close. A later set through the handle, and the close, must act
// on the renamed file, not on a file created under the old name in the meantime.
func TestRenameByHandleMovesTheHandle(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("victim.txt", []byte("old")); err != nil {
		t.Fatal(err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("victim.txt", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE|fileflags.DELETE,
		renameShareAll, fileflags.FILE_OPEN, fileflags.FILE_DELETE_ON_CLOSE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	data, err := (&filesystem.FileRenameInformation{FileName: "moved.txt"}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SetFileInformation(fid, smbInfoPassthrough+uint16(infoclass.FileRenameInformation), data); err != nil {
		t.Fatalf("rename by handle: %v", err)
	}

	// Another client creates an unrelated file under the old name.
	if err := fs.AddFile("victim.txt", []byte("unrelated")); err != nil {
		t.Fatal(err)
	}

	// A set through the handle acts on the file the handle holds.
	eof, err := (&filesystem.FileEndOfFileInformation{EndOfFile: 1}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SetFileInformation(fid, smbInfoPassthrough+uint16(infoclass.FileEndOfFileInformation), eof); err != nil {
		t.Fatalf("end-of-file set by handle: %v", err)
	}
	if attr, err := fs.Stat("moved.txt"); err != nil || attr.Size != 1 {
		t.Errorf("Stat(moved.txt) = size %d, %v, want the handle's set to make it 1 byte", attr.Size, err)
	}
	if attr, err := fs.Stat("victim.txt"); err != nil || attr.Size != int64(len("unrelated")) {
		t.Errorf("Stat(victim.txt) = size %d, %v, want the unrelated file untouched", attr.Size, err)
	}
	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}

	if _, err := fs.Stat("victim.txt"); err != nil {
		t.Errorf("the unrelated victim.txt was deleted by the renamed handle's delete-on-close: %v", err)
	}
	if _, err := fs.Stat("moved.txt"); err == nil {
		t.Error("moved.txt survived the delete-on-close of the handle that renamed it")
	}
}

// TestRenameMovesHandlesOnAnotherConnection renames a file by path, through
// each rename command, while a handle on another connection holds it for
// delete-on-close.
func TestRenameMovesHandlesOnAnotherConnection(t *testing.T) {
	renames := map[string]func(client *smb1client.Client, from, to string) error{
		"SMB_COM_RENAME":    func(c *smb1client.Client, from, to string) error { return c.RenameFile(from, to) },
		"SMB_COM_NT_RENAME": func(c *smb1client.Client, from, to string) error { return c.NtRename(from, to) },
	}
	for name, rename := range renames {
		t.Run(name, func(t *testing.T) {
			fs := NewMemoryFileSystem("FILES")
			if err := fs.AddFile("victim.txt", []byte("old")); err != nil {
				t.Fatal(err)
			}
			srv, holder := fileServer(t, fs, false)
			renamer := attachClient(t, srv)

			fid, err := holder.OpenFile("victim.txt", fileflags.GENERIC_READ|fileflags.DELETE,
				renameShareAll, fileflags.FILE_OPEN, fileflags.FILE_DELETE_ON_CLOSE)
			if err != nil {
				t.Fatalf("OpenFile() error = %v", err)
			}
			if err := rename(renamer, "victim.txt", "moved.txt"); err != nil {
				t.Fatalf("rename: %v", err)
			}

			if err := fs.AddFile("victim.txt", []byte("unrelated")); err != nil {
				t.Fatal(err)
			}
			if err := holder.CloseFile(fid); err != nil {
				t.Fatalf("CloseFile() error = %v", err)
			}
			if _, err := fs.Stat("victim.txt"); err != nil {
				t.Errorf("the unrelated victim.txt was deleted by the other handle's delete-on-close: %v", err)
			}
			if _, err := fs.Stat("moved.txt"); err == nil {
				t.Error("moved.txt survived the delete-on-close of a handle on it")
			}
		})
	}
}

// TestRenameDirectoryMovesHandlesBeneathIt renames a directory while a handle on
// another connection holds a file inside it for delete-on-close.
func TestRenameDirectoryMovesHandlesBeneathIt(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.Mkdir("dir1"); err != nil {
		t.Fatal(err)
	}
	if err := fs.AddFile("dir1/child.txt", []byte("child")); err != nil {
		t.Fatal(err)
	}
	srv, holder := fileServer(t, fs, false)
	renamer := attachClient(t, srv)

	fid, err := holder.OpenFile("dir1\\child.txt", fileflags.GENERIC_READ|fileflags.DELETE,
		renameShareAll, fileflags.FILE_OPEN, fileflags.FILE_DELETE_ON_CLOSE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if err := renamer.RenameFile("dir1", "dir2"); err != nil {
		t.Fatalf("RenameFile() error = %v", err)
	}

	// The old names are reused by unrelated entries.
	if err := fs.Mkdir("dir1"); err != nil {
		t.Fatal(err)
	}
	if err := fs.AddFile("dir1/child.txt", []byte("unrelated")); err != nil {
		t.Fatal(err)
	}
	if err := holder.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}
	if _, err := fs.Stat("dir1/child.txt"); err != nil {
		t.Errorf("the unrelated dir1/child.txt was deleted by the moved handle's delete-on-close: %v", err)
	}
	if _, err := fs.Stat("dir2/child.txt"); err == nil {
		t.Error("dir2/child.txt survived the delete-on-close of a handle on it")
	}
}

// TestRenameMovesByteRangeLocks checks that a lock held through a handle on a
// renamed file still excludes a writer that reaches the file by its new name.
func TestRenameMovesByteRangeLocks(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("locked.bin", make([]byte, 64)); err != nil {
		t.Fatal(err)
	}
	srv, holder := fileServer(t, fs, false)
	other := attachClient(t, srv)

	fid, err := holder.OpenFile("locked.bin", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		renameShareAll, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if err := holder.LockFile(fid, 0, 16, true); err != nil {
		t.Fatalf("LockFile() error = %v", err)
	}
	if err := other.RenameFile("locked.bin", "moved.bin"); err != nil {
		t.Fatalf("RenameFile() error = %v", err)
	}

	writer, err := other.OpenFile("moved.bin", fileflags.GENERIC_WRITE, renameShareAll, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("OpenFile(moved.bin) error = %v", err)
	}
	if _, err := other.WriteFile(writer, 0, []byte("over the lock")); err == nil {
		t.Error("a write into a range locked through another handle succeeded after the rename")
	}

	// Closing the holder releases the lock under the name it is now filed under.
	if err := holder.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}
	share := srv.shares["FILES"]
	if share == nil {
		t.Fatal("the share is not registered")
	}
	if held := share.locks.HeldOn("moved.bin") + share.locks.HeldOn("locked.bin"); held != 0 {
		t.Errorf("%d locks survive the close of the handle that held them", held)
	}
}

// TestRenameReplaceRefusedWhileTargetIsOpen replaces a file another connection
// holds open. [MS-FSA] section 2.1.5.15.11 refuses that with
// STATUS_ACCESS_DENIED; accepting it left the holder's handle naming the
// renamed file, so the holder's delete-on-close deleted it.
func TestRenameReplaceRefusedWhileTargetIsOpen(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("src.txt", []byte("source")); err != nil {
		t.Fatal(err)
	}
	if err := fs.AddFile("dst.txt", []byte("target")); err != nil {
		t.Fatal(err)
	}
	srv, holder := fileServer(t, fs, false)
	renamer := attachClient(t, srv)

	held, err := holder.OpenFile("dst.txt", fileflags.GENERIC_READ|fileflags.DELETE,
		renameShareAll, fileflags.FILE_OPEN, fileflags.FILE_DELETE_ON_CLOSE)
	if err != nil {
		t.Fatalf("OpenFile(dst.txt) error = %v", err)
	}
	replace, err := (&filesystem.FileRenameInformation{ReplaceIfExists: true, FileName: "dst.txt"}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	level := smbInfoPassthrough + uint16(infoclass.FileRenameInformation)

	// By handle and by path, both refused while dst.txt is held.
	source, err := renamer.OpenFile("src.txt", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE|fileflags.DELETE,
		renameShareAll, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("OpenFile(src.txt) error = %v", err)
	}
	if err := renamer.SetFileInformation(source, level, replace); err == nil || !strings.Contains(err.Error(), "c0000022") {
		t.Errorf("replacing an open target by handle = %v, want STATUS_ACCESS_DENIED", err)
	}
	if err := renamer.SetPathInformation("src.txt", level, replace); err == nil || !strings.Contains(err.Error(), "c0000022") {
		t.Errorf("replacing an open target by path = %v, want STATUS_ACCESS_DENIED", err)
	}
	if attr, err := fs.Stat("src.txt"); err != nil || attr.Size != int64(len("source")) {
		t.Errorf("Stat(src.txt) = size %d, %v after the refused renames, want it untouched", attr.Size, err)
	}

	// The holder's delete-on-close takes its own file, and only that.
	if err := holder.CloseFile(held); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}
	if _, err := fs.Stat("dst.txt"); err == nil {
		t.Error("dst.txt survived the delete-on-close of the handle that held it")
	}
	if _, err := fs.Stat("src.txt"); err != nil {
		t.Errorf("src.txt was deleted by another file's delete-on-close: %v", err)
	}

	// With nothing open on the target any more, the replace goes through.
	if err := fs.AddFile("dst.txt", []byte("target again")); err != nil {
		t.Fatal(err)
	}
	if err := renamer.SetFileInformation(source, level, replace); err != nil {
		t.Fatalf("replacing a target nobody holds: %v", err)
	}
	if attr, err := fs.Stat("dst.txt"); err != nil || attr.Size != int64(len("source")) {
		t.Errorf("Stat(dst.txt) = size %d, %v after the replace, want the renamed file", attr.Size, err)
	}
	if _, err := fs.Stat("src.txt"); err == nil {
		t.Error("src.txt still exists after it replaced dst.txt")
	}
	if err := renamer.CloseFile(source); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}
}

// TestRenameUnderConcurrentConnections renames files and their directory back and
// forth from several connections at once, while each also opens, locks, writes
// and closes. Whatever interleaving happens, once every handle is closed the
// share must hold no handle and no lock, under any name. Run it with -race.
func TestRenameUnderConcurrentConnections(t *testing.T) {
	const workers, rounds = 6, 40

	fs := NewMemoryFileSystem("FILES")
	if err := fs.Mkdir("d"); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < workers; n++ {
		if err := fs.AddFile(fmt.Sprintf("d/f%d", n), make([]byte, 32)); err != nil {
			t.Fatal(err)
		}
	}
	srv, first := fileServer(t, fs, false)
	clients := []*smb1client.Client{first}
	for len(clients) < workers {
		clients = append(clients, attachClient(t, srv))
	}

	var wg sync.WaitGroup
	for n, client := range clients {
		wg.Add(1)
		go func(n int, client *smb1client.Client) {
			defer wg.Done()
			for round := 0; round < rounds; round++ {
				// Half the workers move the directory; every rename may fail
				// because another worker moved it first, which is fine.
				if n%2 == 0 {
					if round%2 == 0 {
						_ = client.RenameFile("d", "e")
					} else {
						_ = client.RenameFile("e", "d")
					}
				}
				for _, dir := range []string{"d", "e"} {
					name, other := fmt.Sprintf("%s\\f%d", dir, n), fmt.Sprintf("g%d", n)
					if round%2 == 1 {
						name, other = fmt.Sprintf("%s\\g%d", dir, n), fmt.Sprintf("f%d", n)
					}
					fid, err := client.OpenFile(name, fileflags.GENERIC_READ|fileflags.GENERIC_WRITE|fileflags.DELETE,
						renameShareAll, fileflags.FILE_OPEN, 0)
					if err != nil {
						continue
					}
					_ = client.LockFile(fid, 0, 8, true)
					_, _ = client.WriteFile(fid, 16, []byte("data"))
					if data, err := (&filesystem.FileRenameInformation{FileName: dir + "\\" + other}).Marshal(); err == nil {
						_ = client.SetFileInformation(fid, smbInfoPassthrough+uint16(infoclass.FileRenameInformation), data)
					}
					_, _ = client.WriteFile(fid, 8, []byte("more"))
					_ = client.CloseFile(fid)
				}
			}
		}(n, client)
	}
	wg.Wait()

	share := srv.shares["FILES"]
	if share == nil {
		t.Fatal("the share is not registered")
	}
	share.opens.mutex.Lock()
	open := len(share.opens.opens)
	share.opens.mutex.Unlock()
	if open != 0 {
		t.Errorf("%d handles remain registered after every handle closed", open)
	}
	share.locks.mutex.Lock()
	locked := len(share.locks.granted)
	share.locks.mutex.Unlock()
	if locked != 0 {
		t.Errorf("locks remain on %d paths after every handle closed", locked)
	}
}

// TestRenamedPath covers the path arithmetic a rename applies to each handle.
func TestRenamedPath(t *testing.T) {
	for _, tc := range []struct {
		path, from, to string
		want           string
		moved          bool
	}{
		{"a.txt", "a.txt", "b.txt", "b.txt", true},
		{"dir/a.txt", "dir", "new", "new/a.txt", true},
		{"dir/sub/a.txt", "dir", "x/y", "x/y/sub/a.txt", true},
		{"dirx/a.txt", "dir", "new", "", false},
		{"other.txt", "a.txt", "b.txt", "", false},
		{"dir", "dir/a.txt", "b.txt", "", false},
	} {
		got, moved := renamedPath(tc.path, tc.from, tc.to)
		if got != tc.want || moved != tc.moved {
			t.Errorf("renamedPath(%q, %q, %q) = %q, %v, want %q, %v", tc.path, tc.from, tc.to, got, moved, tc.want, tc.moved)
		}
	}
}
