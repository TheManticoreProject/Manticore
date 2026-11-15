package server

import (
	"testing"
)

// TestBackendsReportExFATName checks both shipped backends name the volume exFAT,
// not NTFS. The server reports only FILE_CASE_PRESERVED_NAMES | FILE_UNICODE_ON_DISK
// (0x6) — the capability set of an exFAT volume, not NTFS, which always reports
// persistent ACLs and case-sensitive search. Naming it NTFS beside a 0x6 attribute
// mask that no real NTFS volume reports was the fingerprint.
func TestBackendsReportExFATName(t *testing.T) {
	mem := NewMemoryFileSystem("FILES")
	if v, _ := mem.VolumeInfo(); v.FileSystemName != "exFAT" {
		t.Errorf("MemoryFileSystem FileSystemName = %q, want exFAT", v.FileSystemName)
	}

	root := t.TempDir()
	local, err := NewLocalFileSystem(root, "FILES")
	if err != nil {
		t.Fatalf("NewLocalFileSystem: %v", err)
	}
	defer local.Close()
	if v, _ := local.VolumeInfo(); v.FileSystemName != "exFAT" {
		t.Errorf("LocalFileSystem FileSystemName = %q, want exFAT", v.FileSystemName)
	}

	// The SMB1 TREE_CONNECT native file system name agrees.
	if nativeFileSystemName != "exFAT" {
		t.Errorf("nativeFileSystemName = %q, want exFAT", nativeFileSystemName)
	}

	// The attribute set is unchanged and consistent with exFAT.
	if serverFileSystemAttributes != fileCasePreservedNames|fileUnicodeOnDisk {
		t.Errorf("serverFileSystemAttributes = %#x, want 0x6 (case-preserved | unicode)", serverFileSystemAttributes)
	}
}
