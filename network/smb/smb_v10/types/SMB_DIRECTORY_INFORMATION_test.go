package types_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestSMB_DIRECTORY_INFORMATION_Size asserts the entry is the fixed 43 bytes
// [MS-CIFS] section 2.2.8.1.4 specifies, whatever it carries.
//
// The size is the property that matters most about this structure. The entries
// are packed with nothing linking them, so a client walks the array by this
// stride: an entry of any other length desynchronises every entry after it, and
// the listing is read as garbage rather than reported as malformed.
//
// This test previously compared against golden hex that had been captured from
// the implementation rather than from the specification, so it agreed with a
// 53-byte entry — a 24-byte resume key carrying a variable-block string header,
// an 8-byte FILETIME in a 2-byte field, and a name field with a buffer-format
// byte in front of it. The assertions here are structural for that reason:
// derived from the field table rather than from what the code happened to emit.
func TestSMB_DIRECTORY_INFORMATION_Size(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		fileName string
	}{
		{name: "short name", fileName: "A.TXT"},
		{name: "full 8.3 name", fileName: "ABCDEFGH.TXT"},
		{name: "empty name", fileName: ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			entry := types.NewSMB_DIRECTORY_INFORMATION()
			entry.FileAttributes = types.UCHAR(types.ATTR_ARCHIVE)
			entry.LastWriteTime = *types.NewSMB_TIME_DOSFromTime(4, 5, 6)
			entry.LastWriteDate = *types.NewSMB_DATEFromDate(2021, 12, 3)
			entry.FileSize = types.ULONG(1024)
			entry.FileName = *types.NewOEM_STRINGFromString(testCase.fileName)

			marshalled, err := entry.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if len(marshalled) != types.SMB_DIRECTORY_INFORMATION_SIZE {
				t.Fatalf("the entry is %d bytes, want %d",
					len(marshalled), types.SMB_DIRECTORY_INFORMATION_SIZE)
			}
		})
	}
}

// TestSMB_DIRECTORY_INFORMATION_FieldOffsets asserts each field lands where the
// field table puts it: ResumeKey(21) FileAttributes(1) LastWriteTime(2)
// LastWriteDate(2) FileSize(4) FileName(13).
func TestSMB_DIRECTORY_INFORMATION_FieldOffsets(t *testing.T) {
	entry := types.NewSMB_DIRECTORY_INFORMATION()
	entry.ResumeKey = types.SMB_RESUME_KEY{
		Reserved:    0x11,
		ServerState: [16]types.UCHAR{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		ClientState: [4]types.UCHAR{0xAA, 0xBB, 0xCC, 0xDD},
	}
	entry.FileAttributes = types.UCHAR(types.ATTR_DIRECTORY)
	entry.LastWriteTime = *types.NewSMB_TIME_DOSFromTime(4, 5, 6)
	entry.LastWriteDate = *types.NewSMB_DATEFromDate(2021, 12, 3)
	entry.FileSize = types.ULONG(0x11223344)
	entry.FileName = *types.NewOEM_STRINGFromString("FOLDER")

	marshalled, err := entry.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// The resume key is bare: no buffer-format byte and no length ahead of it,
	// which is the difference from the request's data block.
	if marshalled[0] != 0x11 {
		t.Errorf("the entry starts with 0x%02X, want the resume key's Reserved byte 0x11 — "+
			"a buffer-format byte is being emitted", marshalled[0])
	}
	if !bytes.Equal(marshalled[17:21], []byte{0xAA, 0xBB, 0xCC, 0xDD}) {
		t.Errorf("ClientState is at the wrong offset: got % x", marshalled[17:21])
	}

	if got := marshalled[21]; got != byte(types.ATTR_DIRECTORY) {
		t.Errorf("FileAttributes is 0x%02X at offset 21, want 0x%02X", got, byte(types.ATTR_DIRECTORY))
	}
	if got := binary.LittleEndian.Uint32(marshalled[26:30]); got != 0x11223344 {
		t.Errorf("FileSize is 0x%08X at offset 26, want 0x11223344", got)
	}

	// FileName is 12 bytes of space-padded name and a terminating null.
	nameField := marshalled[30:43]
	if got := string(nameField[:6]); got != "FOLDER" {
		t.Errorf("the name field holds %q, want %q", got, "FOLDER")
	}
	if got := string(nameField[6:12]); got != "      " {
		t.Errorf("the name is padded with %q, want spaces", got)
	}
	if nameField[12] != 0x00 {
		t.Errorf("the name field ends with 0x%02X, want a terminating null", nameField[12])
	}
}

// TestSMB_DIRECTORY_INFORMATION_RoundTrip asserts the entry decodes what it
// encodes, which it could not while its marshalled length disagreed with the
// length its decoder consumed.
func TestSMB_DIRECTORY_INFORMATION_RoundTrip(t *testing.T) {
	entry := types.NewSMB_DIRECTORY_INFORMATION()
	entry.ResumeKey = types.SMB_RESUME_KEY{
		Reserved:    0x11,
		ServerState: [16]types.UCHAR{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		ClientState: [4]types.UCHAR{0xAA, 0xBB, 0xCC, 0xDD},
	}
	entry.FileAttributes = types.UCHAR(types.ATTR_ARCHIVE)
	entry.LastWriteTime = *types.NewSMB_TIME_DOSFromTime(4, 5, 6)
	entry.LastWriteDate = *types.NewSMB_DATEFromDate(2021, 12, 3)
	entry.FileSize = types.ULONG(1024)
	entry.FileName = *types.NewOEM_STRINGFromString("TEST.TXT")

	marshalled, err := entry.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	decoded := types.NewSMB_DIRECTORY_INFORMATION()
	read, err := decoded.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if read != types.SMB_DIRECTORY_INFORMATION_SIZE {
		t.Fatalf("Unmarshal consumed %d bytes, want %d", read, types.SMB_DIRECTORY_INFORMATION_SIZE)
	}

	if decoded.ResumeKey != entry.ResumeKey {
		t.Errorf("the resume key round-tripped to %+v, want %+v", decoded.ResumeKey, entry.ResumeKey)
	}
	if decoded.FileAttributes != entry.FileAttributes {
		t.Errorf("FileAttributes round-tripped to 0x%02X, want 0x%02X",
			decoded.FileAttributes, entry.FileAttributes)
	}
	if decoded.FileSize != entry.FileSize {
		t.Errorf("FileSize round-tripped to %d, want %d", decoded.FileSize, entry.FileSize)
	}
	if decoded.LastWriteTime != entry.LastWriteTime {
		t.Errorf("LastWriteTime round-tripped to %+v, want %+v", decoded.LastWriteTime, entry.LastWriteTime)
	}
	// The padding and terminator are stripped on the way back.
	if got := decoded.FileName.GetString(); got != "TEST.TXT" {
		t.Errorf("the name round-tripped to %q, want %q", got, "TEST.TXT")
	}
}

// TestSMB_DIRECTORY_INFORMATION_ArrayIsWalkableByStride asserts an array of
// entries can be walked by the fixed stride, which is the only way a client can
// read one.
func TestSMB_DIRECTORY_INFORMATION_ArrayIsWalkableByStride(t *testing.T) {
	names := []string{"ONE.TXT", "TWO.TXT", "THREE.TXT"}

	array := []byte{}
	for index, name := range names {
		entry := types.NewSMB_DIRECTORY_INFORMATION()
		entry.FileSize = types.ULONG(index + 1)
		entry.FileName = *types.NewOEM_STRINGFromString(name)

		marshalled, err := entry.Marshal()
		if err != nil {
			t.Fatalf("Marshal(%q) error = %v", name, err)
		}
		array = append(array, marshalled...)
	}

	if len(array) != len(names)*types.SMB_DIRECTORY_INFORMATION_SIZE {
		t.Fatalf("the array is %d bytes, want %d", len(array),
			len(names)*types.SMB_DIRECTORY_INFORMATION_SIZE)
	}

	for index, want := range names {
		at := index * types.SMB_DIRECTORY_INFORMATION_SIZE
		decoded := types.NewSMB_DIRECTORY_INFORMATION()
		if _, err := decoded.Unmarshal(array[at:]); err != nil {
			t.Fatalf("entry %d did not decode: %v", index, err)
		}
		if got := decoded.FileName.GetString(); got != want {
			t.Errorf("entry %d at offset %d names %q, want %q", index, at, got, want)
		}
		if got := int(decoded.FileSize); got != index+1 {
			t.Errorf("entry %d carries size %d, want %d", index, got, index+1)
		}
	}
}

// TestSMB_DIRECTORY_INFORMATION_NameTooLongIsRefused asserts a name that does not
// fit the 8.3 field is refused rather than truncated. A truncated name is a
// different file.
func TestSMB_DIRECTORY_INFORMATION_NameTooLongIsRefused(t *testing.T) {
	entry := types.NewSMB_DIRECTORY_INFORMATION()
	entry.FileName = *types.NewOEM_STRINGFromString("ABCDEFGHIJKLM")

	if _, err := entry.Marshal(); err == nil {
		t.Fatal("a 13-character name marshalled into the 12-byte field, want an error")
	}
}

func TestNewSMB_DIRECTORY_INFORMATION(t *testing.T) {
	dirInfo := types.NewSMB_DIRECTORY_INFORMATION()

	if dirInfo == nil {
		t.Fatal("NewSMB_DIRECTORY_INFORMATION returned nil")
	}
	if dirInfo.FileAttributes != 0 {
		t.Errorf("Expected FileAttributes to be 0, got %d", dirInfo.FileAttributes)
	}
	if dirInfo.FileSize != 0 {
		t.Errorf("Expected FileSize to be 0, got %d", dirInfo.FileSize)
	}
}
