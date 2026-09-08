package server

import (
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// passthroughLevel is the SMB information level that carries a native
// information class ([MS-SMB] section 2.2.2.3.5).
func passthroughLevel(class infoclass.FileInformationClass) uint16 {
	return smbInfoPassthrough + uint16(class)
}

// passthroughServer serves one file of known contents and returns a connected
// client with a handle on it.
func passthroughServer(t *testing.T, name, contents string) (*smb1client.Client, smb1client.FID) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile(name, []byte(contents)); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile(name,
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening %q failed: %v", name, err)
	}
	return client, fid
}

// TestPassthroughQueryClassesAreServed asserts each served native query class
// comes back and parses with the repository's own [MS-FSCC] structures.
//
// Parsing the answer back with windows/filesystem rather than with assertions on
// raw offsets is deliberate: those structures are what the SMB2 client and the
// rest of the repository already agree on, so a layout this server got wrong
// fails here rather than agreeing with itself.
func TestPassthroughQueryClassesAreServed(t *testing.T) {
	const contents = "0123456789abcdef"
	client, fid := passthroughServer(t, "described.txt", contents)

	t.Run("basic information", func(t *testing.T) {
		raw, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileBasicInformation))
		if err != nil {
			t.Fatalf("querying FileBasicInformation failed: %v", err)
		}
		basic := filesystem.FileBasicInformation{}
		if err := basic.Unmarshal(raw); err != nil {
			t.Fatalf("the answer did not parse as FILE_BASIC_INFORMATION: %v", err)
		}
		if basic.LastWriteTime == 0 {
			t.Error("the reported last-write time is zero")
		}
		if basic.FileAttributes == 0 {
			t.Error("the reported attributes are zero, so the file describes nothing")
		}
	})

	t.Run("standard information", func(t *testing.T) {
		raw, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileStandardInformation))
		if err != nil {
			t.Fatalf("querying FileStandardInformation failed: %v", err)
		}
		standard := filesystem.FileStandardInformation{}
		if err := standard.Unmarshal(raw); err != nil {
			t.Fatalf("the answer did not parse as FILE_STANDARD_INFORMATION: %v", err)
		}
		if standard.EndOfFile != int64(len(contents)) {
			t.Errorf("EndOfFile is %d, want %d", standard.EndOfFile, len(contents))
		}
		if standard.Directory {
			t.Error("a file is reported as a directory")
		}
		if standard.NumberOfLinks != 1 {
			t.Errorf("NumberOfLinks is %d, want 1", standard.NumberOfLinks)
		}
	})

	t.Run("network open information", func(t *testing.T) {
		raw, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileNetworkOpenInformation))
		if err != nil {
			t.Fatalf("querying FileNetworkOpenInformation failed: %v", err)
		}
		open := filesystem.FileNetworkOpenInformation{}
		if err := open.Unmarshal(raw); err != nil {
			t.Fatalf("the answer did not parse as FILE_NETWORK_OPEN_INFORMATION: %v", err)
		}
		if open.EndOfFile != int64(len(contents)) {
			t.Errorf("EndOfFile is %d, want %d", open.EndOfFile, len(contents))
		}
	})

	t.Run("all information", func(t *testing.T) {
		raw, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileAllInformation))
		if err != nil {
			t.Fatalf("querying FileAllInformation failed: %v", err)
		}
		all := filesystem.FileAllInformation{}
		if err := all.Unmarshal(raw); err != nil {
			t.Fatalf("the answer did not parse as FILE_ALL_INFORMATION: %v", err)
		}
		if all.Standard.EndOfFile != int64(len(contents)) {
			t.Errorf("EndOfFile is %d, want %d", all.Standard.EndOfFile, len(contents))
		}
		if !strings.EqualFold(all.FileName, `\described.txt`) {
			t.Errorf("FileName is %q, want %q", all.FileName, `\described.txt`)
		}
	})

	t.Run("name information carries UTF-16 whatever the message declared", func(t *testing.T) {
		raw, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileNameInformation))
		if err != nil {
			t.Fatalf("querying FileNameInformation failed: %v", err)
		}
		if len(raw) < 4 {
			t.Fatalf("the answer is %d bytes, too short for FileNameLength", len(raw))
		}
		// FileNameLength(4) FileName(variable). A native structure's strings are
		// UTF-16LE regardless of SMB_FLAGS2_UNICODE, so decoding it that way is
		// the assertion.
		name := utf16.DecodeUTF16LE([]types.UCHAR(raw[4:]))
		if !strings.EqualFold(name, `\described.txt`) {
			t.Errorf("the name decoded as %q, want %q — a native name is UTF-16LE", name, `\described.txt`)
		}
	})

	t.Run("classes with nothing behind them are still answered", func(t *testing.T) {
		for _, class := range []infoclass.FileInformationClass{
			infoclass.FileInternalInformation,
			infoclass.FileEaInformation,
			infoclass.FileAccessInformation,
			infoclass.FilePositionInformation,
			infoclass.FileAlternateNameInformation,
		} {
			if _, err := client.QueryFileInformation(fid, passthroughLevel(class)); err != nil {
				t.Errorf("querying class %d failed: %v", class, err)
			}
		}
	})
}

// TestPassthroughUnservedClassIsRefused asserts a class in the range that the
// server does not serve is refused rather than answered with something else.
func TestPassthroughUnservedClassIsRefused(t *testing.T) {
	client, fid := passthroughServer(t, "described.txt", "x")

	// 12 is FileNamesInformation, which describes a directory enumeration rather
	// than one file and is not served here.
	if _, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileNamesInformation)); err == nil {
		t.Error("an unserved pass-through class was answered")
	}
}

// TestPassthroughSetEndOfFileResizes asserts the pass-through set classes act,
// rather than being accepted and dropped.
func TestPassthroughSetEndOfFileResizes(t *testing.T) {
	client, fid := passthroughServer(t, "resized.txt", "0123456789abcdef")

	endOfFile := filesystem.FileEndOfFileInformation{EndOfFile: 4}
	encoded, err := endOfFile.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if err := client.SetFileInformation(fid, passthroughLevel(infoclass.FileEndOfFileInformation), encoded); err != nil {
		t.Fatalf("setting FileEndOfFileInformation failed: %v", err)
	}

	raw, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileStandardInformation))
	if err != nil {
		t.Fatalf("querying the size back failed: %v", err)
	}
	standard := filesystem.FileStandardInformation{}
	if err := standard.Unmarshal(raw); err != nil {
		t.Fatalf("the answer did not parse: %v", err)
	}
	if standard.EndOfFile != 4 {
		t.Errorf("the file is %d bytes after being truncated to 4", standard.EndOfFile)
	}
}

// TestPassthroughSetBasicInformationAppliesTimes asserts a native basic set is
// applied, and that a zero timestamp leaves that one alone.
func TestPassthroughSetBasicInformationAppliesTimes(t *testing.T) {
	client, fid := passthroughServer(t, "stamped.txt", "x")

	before, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileBasicInformation))
	if err != nil {
		t.Fatalf("querying the times first failed: %v", err)
	}
	original := filesystem.FileBasicInformation{}
	if err := original.Unmarshal(before); err != nil {
		t.Fatalf("the answer did not parse: %v", err)
	}

	// Only the write time is supplied; the rest are zero and so must not move.
	wanted := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	basic := filesystem.FileBasicInformation{LastWriteTime: filetimeOf(wanted)}
	encoded, err := basic.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := client.SetFileInformation(fid, passthroughLevel(infoclass.FileBasicInformation), encoded); err != nil {
		t.Fatalf("setting FileBasicInformation failed: %v", err)
	}

	after, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileBasicInformation))
	if err != nil {
		t.Fatalf("querying the times back failed: %v", err)
	}
	applied := filesystem.FileBasicInformation{}
	if err := applied.Unmarshal(after); err != nil {
		t.Fatalf("the answer did not parse: %v", err)
	}

	if applied.LastWriteTime != filetimeOf(wanted) {
		t.Errorf("the write time is %d, want %d", applied.LastWriteTime, filetimeOf(wanted))
	}
	if applied.CreationTime != original.CreationTime {
		t.Errorf("the creation time moved to %d from %d, but none was supplied",
			applied.CreationTime, original.CreationTime)
	}
}

// TestPassthroughRenameMovesTheEntry asserts a rename through the pass-through
// class works, which is how a Windows client renames a file.
func TestPassthroughRenameMovesTheEntry(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("before.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	rename := filesystem.FileRenameInformation{FileName: `after.txt`}
	encoded, err := rename.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if err := client.SetPathInformation("before.txt",
		passthroughLevel(infoclass.FileRenameInformation), encoded); err != nil {
		t.Fatalf("renaming through the pass-through class failed: %v", err)
	}

	if _, err := fs.Stat("after.txt"); err != nil {
		t.Errorf("the file is not at its new name: %v", err)
	}
	if _, err := fs.Stat("before.txt"); err == nil {
		t.Error("the file is still at its old name")
	}
}

// TestPassthroughRenameRefusesARelativeRoot asserts a rename whose new name is
// relative to a directory handle is refused rather than misapplied.
//
// RootDirectory names an open directory the new name is relative to. Nothing here
// can resolve one, and treating the name as share-relative instead would move the
// file somewhere the client did not ask for and report success.
func TestPassthroughRenameRefusesARelativeRoot(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("before.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	rename := filesystem.FileRenameInformation{RootDirectory: 0x1234, FileName: `after.txt`}
	encoded, err := rename.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if err := client.SetPathInformation("before.txt",
		passthroughLevel(infoclass.FileRenameInformation), encoded); err == nil {
		t.Fatal("a rename relative to a directory handle was accepted")
	}
	if _, err := fs.Stat("before.txt"); err != nil {
		t.Errorf("the refused rename moved the file anyway: %v", err)
	}
}

// TestPassthroughRenameRefusesEscapingTheShare asserts the new name goes through
// the same path resolution every other name does.
func TestPassthroughRenameRefusesEscapingTheShare(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("before.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	rename := filesystem.FileRenameInformation{FileName: `..\escaped.txt`}
	encoded, err := rename.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if err := client.SetPathInformation("before.txt",
		passthroughLevel(infoclass.FileRenameInformation), encoded); err == nil {
		t.Fatal("a rename out of the share was accepted")
	}
}

// TestPassthroughCapabilityIsAdvertised asserts the server says it serves the
// range it serves.
func TestPassthroughCapabilityIsAdvertised(t *testing.T) {
	if serverCapabilities&capabilities.CAP_INFOLEVEL_PASSTHROUGH == 0 {
		t.Error("CAP_INFOLEVEL_PASSTHROUGH is not advertised, but the pass-through range is served")
	}
}
