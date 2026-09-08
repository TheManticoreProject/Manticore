package server

import (
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// sendNtRename sends an SMB_COM_NT_RENAME at the given information level.
func sendNtRename(
	t *testing.T,
	client *smb1client.Client,
	level uint16,
	oldName, newName string,
	searchAttributes uint16,
) uint32 {
	t.Helper()

	request := commands.NewNtRenameRequest()
	request.InformationLevel = types.USHORT(level)
	request.SearchAttributes.SetAttributes(searchAttributes)
	request.OldFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	if err := request.OldFileName.SetString(oldName); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}
	request.NewFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	if err := request.NewFileName.SetString(newName); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}

	status, _ := sendLegacy(t, client, codes.SMB_COM_NT_RENAME, request)
	return status
}

// renameServer serves one file and one directory.
func renameServer(t *testing.T, readOnly bool) (*MemoryFileSystem, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("before.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddDirectory("adirectory"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, readOnly)
	return fs, client
}

// TestNtRenameRenamesAtTheRenameLevel asserts the rename level moves the entry.
func TestNtRenameRenamesAtTheRenameLevel(t *testing.T) {
	fs, client := renameServer(t, false)

	if status := sendNtRename(t, client, smbNtRenameRenameFile, "before.txt", "after.txt", 0); status != 0 {
		t.Fatalf("the rename reported 0x%08X, want success", status)
	}

	if _, err := fs.Stat("after.txt"); err != nil {
		t.Errorf("the file is not at its new name: %v", err)
	}
	if _, err := fs.Stat("before.txt"); err == nil {
		t.Error("the file is still at its old name")
	}
}

// TestNtRenameLinksAtTheLinkLevel asserts the link level gives the file a second
// name and leaves the first in place.
//
// [MS-CIFS] section 3.3.5.51: "the original file MUST NOT be renamed. Instead, the
// server MUST attempt to create a hard link at the target". A server that renamed
// here would lose the original name and report success.
func TestNtRenameLinksAtTheLinkLevel(t *testing.T) {
	fs, client := renameServer(t, false)

	if status := sendNtRename(t, client, smbNtRenameSetLinkInfo, "before.txt", "alias.txt", 0); status != 0 {
		t.Fatalf("the link reported 0x%08X, want success", status)
	}

	// Both names resolve.
	if _, err := fs.Stat("before.txt"); err != nil {
		t.Errorf("the original name is gone, so this was a rename rather than a link: %v", err)
	}
	if _, err := fs.Stat("alias.txt"); err != nil {
		t.Errorf("the new name does not resolve: %v", err)
	}

	// And they are the same file, not two copies: a write through one is visible
	// through the other, which is the whole difference between a link and a copy.
	file, err := fs.Open("alias.txt", OpenFlags{Read: true, Write: true})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := file.WriteAt([]byte("CHANGED!"), 0); err != nil {
		t.Fatalf("WriteAt() error = %v", err)
	}
	file.Close()

	original, err := fs.Open("before.txt", OpenFlags{Read: true})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer original.Close()

	buffer := make([]byte, 8)
	if _, err := original.ReadAt(buffer, 0); err != nil {
		t.Fatalf("ReadAt() error = %v", err)
	}
	if string(buffer) != "CHANGED!" {
		t.Errorf("the original name holds %q, want the write made through the link", buffer)
	}
}

// TestNtRenameLinkRefusesATakenName asserts a link onto an existing name fails.
func TestNtRenameLinkRefusesATakenName(t *testing.T) {
	fs, client := renameServer(t, false)
	if err := fs.AddFile("taken.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	if status := sendNtRename(t, client, smbNtRenameSetLinkInfo, "before.txt", "taken.txt", 0); status == 0 {
		t.Error("a link onto an existing name succeeded")
	}
}

// TestNtRenameLinkOnABackendThatCannotLinkIsRefused asserts a share whose backend
// does not implement Linker says so rather than approximating.
//
// A copy would be a second file rather than a second name, and the two diverge as
// soon as either is written — so reporting success for one would be worse than
// refusing.
func TestNtRenameLinkOnABackendThatCannotLinkIsRefused(t *testing.T) {
	fs := &unlinkableFileSystem{MemoryFileSystem: NewMemoryFileSystem("FILES")}
	if err := fs.AddFile("before.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	status := sendNtRename(t, client, smbNtRenameSetLinkInfo, "before.txt", "alias.txt", 0)
	if status != uint32(nt_status.NT_STATUS_NOT_SUPPORTED) {
		t.Errorf("a backend that cannot link reported 0x%08X, want STATUS_NOT_SUPPORTED (0x%08X)",
			status, uint32(nt_status.NT_STATUS_NOT_SUPPORTED))
	}

	// A rename on the same share still works, so the refusal is about linking
	// rather than about the share.
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "before.txt", "after.txt", 0); status != 0 {
		t.Errorf("a rename on the same share reported 0x%08X", status)
	}
}

// unlinkableFileSystem is a MemoryFileSystem with Link hidden, standing in for a
// caller's backend that cannot create one.
//
// The embedded type's Link would otherwise satisfy Linker, so it is shadowed by a
// field rather than overridden — a method cannot be removed in Go, and this is the
// least surprising way to express "does not implement it".
type unlinkableFileSystem struct {
	*MemoryFileSystem

	// Link shadows the embedded method, so the type does not satisfy Linker.
	Link struct{}
}

// TestNtRenameRefusesAWildcardSource asserts the source name may not be a pattern.
//
// [MS-CIFS] section 3.3.5.51: "OldFileName MUST NOT contain wildcard characters;
// otherwise, the server MUST return an error response with a Status of
// STATUS_OBJECT_PATH_SYNTAX_BAD." Unlike SMB_COM_RENAME, this command acts on one
// named file, so a pattern is a malformed request rather than a match set.
func TestNtRenameRefusesAWildcardSource(t *testing.T) {
	_, client := renameServer(t, false)

	for _, pattern := range []string{"*.txt", "befor?.txt"} {
		status := sendNtRename(t, client, smbNtRenameRenameFile, pattern, "after.txt", 0)
		if status != uint32(nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD) {
			t.Errorf("a wildcard source %q reported 0x%08X, want STATUS_OBJECT_PATH_SYNTAX_BAD (0x%08X)",
				pattern, status, uint32(nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD))
		}
	}
}

// TestNtRenameRefusesAnUnknownLevel asserts a level that is neither of the two is
// refused with the status the specification names.
//
// [MS-CIFS] section 3.3.5.51 says STATUS_INVALID_SMB specifically, and a client
// distinguishing "unsupported" from "not a level" would be misled by
// STATUS_NOT_SUPPORTED or STATUS_INVALID_PARAMETER.
func TestNtRenameRefusesAnUnknownLevel(t *testing.T) {
	_, client := renameServer(t, false)

	status := sendNtRename(t, client, 0x0199, "before.txt", "after.txt", 0)
	if status != uint32(nt_status.NT_STATUS_INVALID_SMB) {
		t.Errorf("an unknown level reported 0x%08X, want STATUS_INVALID_SMB (0x%08X)",
			status, uint32(nt_status.NT_STATUS_INVALID_SMB))
	}
}

// TestNtRenameHonoursSearchAttributes asserts a directory is only matched when the
// request's search attributes admit one.
func TestNtRenameHonoursSearchAttributes(t *testing.T) {
	_, client := renameServer(t, false)

	// Without the directory bit, a directory is not a match.
	status := sendNtRename(t, client, smbNtRenameRenameFile, "adirectory", "renamed", 0)
	if status != uint32(nt_status.NT_STATUS_NO_SUCH_FILE) {
		t.Errorf("renaming a directory without the directory attribute reported 0x%08X, want STATUS_NO_SUCH_FILE",
			status)
	}

	// With it, the rename proceeds.
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "adirectory", "renamed",
		smbFileAttributeDirectory); status != 0 {
		t.Errorf("renaming a directory with the directory attribute reported 0x%08X", status)
	}
}

// TestNtRenameOnAReadOnlyShareIsRefused asserts the share's refusal applies.
func TestNtRenameOnAReadOnlyShareIsRefused(t *testing.T) {
	_, client := renameServer(t, true)

	if status := sendNtRename(t, client, smbNtRenameRenameFile, "before.txt", "after.txt", 0); status == 0 {
		t.Error("a rename succeeded on a read-only share")
	}
}

// TestObsoleteNameCommandsStayRefused asserts SMB_COM_COPY and SMB_COM_MOVE are
// still answered STATUS_NOT_IMPLEMENTED.
//
// They were rendered obsolete in the NT LAN Manager dialect — the only dialect this
// server speaks — and [MS-CIFS] sections 2.2.4.37 and 2.2.4.38 say servers
// receiving them "SHOULD return STATUS_NOT_IMPLEMENTED". Serving them would be
// implementing something the specification tells servers to refuse, so this test
// pins the refusal rather than leaving it to look like an oversight.
func TestObsoleteNameCommandsStayRefused(t *testing.T) {
	_, client := renameServer(t, false)

	for _, testCase := range []struct {
		name    string
		command codes.CommandCode
	}{
		{name: "SMB_COM_COPY", command: codes.SMB_COM_COPY},
		{name: "SMB_COM_MOVE", command: codes.SMB_COM_MOVE},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var status uint32
			switch testCase.command {
			case codes.SMB_COM_COPY:
				status, _ = sendLegacy(t, client, codes.SMB_COM_COPY, commands.NewCopyRequest())
			case codes.SMB_COM_MOVE:
				status, _ = sendLegacy(t, client, codes.SMB_COM_MOVE, commands.NewMoveRequest())
			}
			if status != uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED) {
				t.Errorf("%s reported 0x%08X, want STATUS_NOT_IMPLEMENTED (0x%08X)",
					testCase.name, status, uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED))
			}
		})
	}
}
