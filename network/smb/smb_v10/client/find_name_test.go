package client

import (
	"encoding/binary"
	"testing"
)

// bothDirInfoEntry builds one SMB_FIND_FILE_BOTH_DIRECTORY_INFO entry whose
// FileName is a null-terminated SMB_STRING, as Windows sends it: FileNameLength
// counts the terminator ([MS-SMB] 2.2.8.1.2).
func bothDirInfoEntry(name string, last bool, attrs uint32) []byte {
	nameBytes := append([]byte(name), 0x00)
	size := bothDirInfoFixedSize + len(nameBytes)
	// Entries are padded to a 4-byte boundary by NextEntryOffset.
	for size%4 != 0 {
		size++
	}
	buf := make([]byte, size)

	next := uint32(size)
	if last {
		next = 0
	}
	binary.LittleEndian.PutUint32(buf[0:4], next)
	binary.LittleEndian.PutUint32(buf[56:60], attrs)
	binary.LittleEndian.PutUint32(buf[60:64], uint32(len(nameBytes)))
	buf[68] = 0 // ShortNameLength
	copy(buf[bothDirInfoFixedSize:], nameBytes)
	return buf
}

// TestParseBothDirInfoStripsNameTerminator guards the defect: FileNameLength counts
// the SMB_STRING terminator, and copying all declared bytes left every name with a
// trailing NUL — so a listing could not be compared against the files it named.
//
// The lengths here are the ones a live Windows Server 2012 R2 sent for these exact
// names ("ADFS" declared as 5 bytes, "AppCompat" as 10).
func TestParseBothDirInfoStripsNameTerminator(t *testing.T) {
	data := append(bothDirInfoEntry("ADFS", false, fileAttributeDirectory), bothDirInfoEntry("AppCompat", true, 0)...)

	entries := parseBothDirInfo(data, false)
	if len(entries) != 2 {
		t.Fatalf("parsed %d entries, want 2", len(entries))
	}

	for i, want := range []string{"ADFS", "AppCompat"} {
		if got := entries[i].LongName; got != want {
			t.Errorf("entry %d LongName = %q, want %q", i, got, want)
		}
	}
	if !entries[0].IsDirectory {
		t.Error("entry 0 should be a directory")
	}
}

// TestParseBothDirInfoStripsUTF16Terminator covers a Unicode name, whose SMB_STRING
// terminator is two bytes.
func TestParseBothDirInfoStripsUTF16Terminator(t *testing.T) {
	// "AB" in UTF-16LE plus a two-byte terminator.
	nameBytes := []byte{'A', 0x00, 'B', 0x00, 0x00, 0x00}
	size := bothDirInfoFixedSize + len(nameBytes)
	buf := make([]byte, size)
	binary.LittleEndian.PutUint32(buf[0:4], 0)
	binary.LittleEndian.PutUint32(buf[60:64], uint32(len(nameBytes)))
	copy(buf[bothDirInfoFixedSize:], nameBytes)

	entries := parseBothDirInfo(buf, false)
	if len(entries) != 1 {
		t.Fatalf("parsed %d entries, want 1", len(entries))
	}
	// The raw bytes are kept (the client issues OEM requests), but neither
	// terminator byte may survive.
	if got := entries[0].LongName; got != "A\x00B" {
		t.Errorf("LongName = %q, want %q with both terminator bytes removed", got, "A\x00B")
	}
}
