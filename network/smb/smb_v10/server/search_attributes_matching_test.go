package server

import (
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// sendDelete sends an SMB_COM_DELETE with the given SearchAttributes.
func sendDelete(t *testing.T, client *smb1client.Client, fileName string, searchAttributes uint16) uint32 {
	t.Helper()

	request := commands.NewDeleteRequest()
	request.SearchAttributes.SetAttributes(searchAttributes)
	request.FileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	if err := request.FileName.SetString(fileName); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}

	status, _ := sendLegacy(t, client, codes.SMB_COM_DELETE, request)
	return status
}

// sendRename sends an SMB_COM_RENAME with the given SearchAttributes.
func sendRename(t *testing.T, client *smb1client.Client, oldName, newName string, searchAttributes uint16) uint32 {
	t.Helper()

	request := commands.NewRenameRequest()
	request.SearchAttributes.SetAttributes(searchAttributes)
	request.OldFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	if err := request.OldFileName.SetString(oldName); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}
	request.NewFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	if err := request.NewFileName.SetString(newName); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}

	status, _ := sendLegacy(t, client, codes.SMB_COM_RENAME, request)
	return status
}

// attributeServer serves a normal file, a hidden file and a system file, so the
// SearchAttributes matching rule can be exercised against each.
func attributeServer(t *testing.T) (*MemoryFileSystem, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("normal.txt", []byte("n")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddFile("secret.txt", []byte("s")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.SetAttr("secret.txt", FileAttr{Hidden: true}, AttrMask{Hidden: true}); err != nil {
		t.Fatalf("SetAttr(hidden) error = %v", err)
	}
	if err := fs.AddFile("driver.sys", []byte("d")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.SetAttr("driver.sys", FileAttr{System: true}, AttrMask{System: true}); err != nil {
		t.Fatalf("SetAttr(system) error = %v", err)
	}
	_, client := fileServer(t, fs, false)
	return fs, client
}

// TestNtRenameHonoursHiddenAndSystemAttributes asserts SMB_COM_NT_RENAME applies
// the inclusive SearchAttributes rule to the hidden and system bits, not only the
// directory bit: a special file matches only when its own bit is set
// ([MS-CIFS] section 2.2.4.66.1).
func TestNtRenameHonoursHiddenAndSystemAttributes(t *testing.T) {
	fs, client := attributeServer(t)

	// A hidden file is not matched when the hidden bit is clear.
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "secret.txt", "x.txt", 0); status != uint32(nt_status.NT_STATUS_NO_SUCH_FILE) {
		t.Errorf("renaming a hidden file without the hidden attribute reported 0x%08X, want STATUS_NO_SUCH_FILE", status)
	}
	// With the hidden bit it renames.
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "secret.txt", "x.txt", smbFileAttributeHidden); status != 0 {
		t.Errorf("renaming a hidden file with the hidden attribute reported 0x%08X", status)
	}
	if _, err := fs.Stat("x.txt"); err != nil {
		t.Errorf("the hidden file is not at its new name: %v", err)
	}

	// A system file needs the system bit; the hidden bit alone does not admit it.
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "driver.sys", "y.sys", smbFileAttributeHidden); status != uint32(nt_status.NT_STATUS_NO_SUCH_FILE) {
		t.Errorf("renaming a system file with only the hidden attribute reported 0x%08X, want STATUS_NO_SUCH_FILE", status)
	}
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "driver.sys", "y.sys", smbFileAttributeSystem); status != 0 {
		t.Errorf("renaming a system file with the system attribute reported 0x%08X", status)
	}

	// A normal file always matches, whatever SearchAttributes holds.
	if status := sendNtRename(t, client, smbNtRenameRenameFile, "normal.txt", "z.txt", 0); status != 0 {
		t.Errorf("renaming a normal file reported 0x%08X", status)
	}
}

// TestDeleteHonoursSearchAttributes asserts SMB_COM_DELETE applies the same rule
// to a single named file ([MS-CIFS] section 3.3.5.9).
func TestDeleteHonoursSearchAttributes(t *testing.T) {
	fs, client := attributeServer(t)

	// A hidden file is not deleted when the hidden bit is clear, and survives.
	if status := sendDelete(t, client, "secret.txt", 0); status != uint32(nt_status.NT_STATUS_NO_SUCH_FILE) {
		t.Errorf("deleting a hidden file without the hidden attribute reported 0x%08X, want STATUS_NO_SUCH_FILE", status)
	}
	if _, err := fs.Stat("secret.txt"); err != nil {
		t.Errorf("the hidden file was deleted despite the attribute mismatch: %v", err)
	}
	// With the hidden bit it is deleted.
	if status := sendDelete(t, client, "secret.txt", smbFileAttributeHidden); status != 0 {
		t.Errorf("deleting a hidden file with the hidden attribute reported 0x%08X", status)
	}
	if _, err := fs.Stat("secret.txt"); err == nil {
		t.Error("the hidden file was not deleted")
	}

	// A normal file is always deleted.
	if status := sendDelete(t, client, "normal.txt", 0); status != 0 {
		t.Errorf("deleting a normal file reported 0x%08X", status)
	}
}

// TestDeleteWildcardSkipsHiddenWithoutTheBit asserts the wildcard path of
// SMB_COM_DELETE skips a hidden file whose bit is clear while still deleting the
// normal files the pattern matches.
func TestDeleteWildcardSkipsHiddenWithoutTheBit(t *testing.T) {
	fs, client := attributeServer(t)

	if status := sendDelete(t, client, "*.txt", 0); status != 0 {
		t.Fatalf("a wildcard delete that matched a normal file reported 0x%08X, want success", status)
	}
	if _, err := fs.Stat("normal.txt"); err == nil {
		t.Error("the normal file was not deleted by the wildcard")
	}
	if _, err := fs.Stat("secret.txt"); err != nil {
		t.Errorf("the hidden file was deleted by a wildcard without the hidden attribute: %v", err)
	}
}

// TestRenameHonoursSearchAttributes asserts SMB_COM_RENAME applies the rule to its
// source ([MS-CIFS] section 2.2.4.8.1).
func TestRenameHonoursSearchAttributes(t *testing.T) {
	fs, client := attributeServer(t)

	// A hidden source is not matched when the hidden bit is clear.
	if status := sendRename(t, client, "secret.txt", "x.txt", 0); status != uint32(nt_status.NT_STATUS_NO_SUCH_FILE) {
		t.Errorf("renaming a hidden file without the hidden attribute reported 0x%08X, want STATUS_NO_SUCH_FILE", status)
	}
	// With the hidden bit it renames.
	if status := sendRename(t, client, "secret.txt", "x.txt", smbFileAttributeHidden); status != 0 {
		t.Errorf("renaming a hidden file with the hidden attribute reported 0x%08X", status)
	}
	if _, err := fs.Stat("x.txt"); err != nil {
		t.Errorf("the hidden file is not at its new name: %v", err)
	}

	// A normal file always renames.
	if status := sendRename(t, client, "normal.txt", "z.txt", 0); status != 0 {
		t.Errorf("renaming a normal file reported 0x%08X", status)
	}
}
