package filesystem

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// notifyEntry builds one FILE_NOTIFY_INFORMATION record. next is written to
// NextEntryOffset verbatim.
func notifyEntry(action FileAction, name string, next uint32) []byte {
	u16 := utf16.Encode([]rune(name))
	nameBytes := make([]byte, len(u16)*2)
	for i, u := range u16 {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], u)
	}
	e := make([]byte, 12+len(nameBytes))
	binary.LittleEndian.PutUint32(e[0:4], next)
	binary.LittleEndian.PutUint32(e[4:8], uint32(action))
	binary.LittleEndian.PutUint32(e[8:12], uint32(len(nameBytes)))
	copy(e[12:], nameBytes)
	return e
}

func TestParseFileNotifyInformation(t *testing.T) {
	first := notifyEntry(FileActionAdded, "a.txt", 0)
	first[0] = byte(len(first)) // NextEntryOffset -> second entry
	second := notifyEntry(FileActionModified, `sub\b.txt`, 0)

	buf := append(append([]byte{}, first...), second...)
	got := ParseFileNotifyInformation(buf)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(got), got)
	}
	if got[0].Action != FileActionAdded || got[0].FileName != "a.txt" {
		t.Errorf("entry 0 = %+v", got[0])
	}
	if got[1].Action != FileActionModified || got[1].FileName != `sub\b.txt` {
		t.Errorf("entry 1 = %+v", got[1])
	}
}

func TestParseFileNotifyInformation_Empty(t *testing.T) {
	if got := ParseFileNotifyInformation(nil); got != nil {
		t.Errorf("expected nil for empty buffer, got %+v", got)
	}
}

func TestFileActionString(t *testing.T) {
	if FileActionAdded.String() != "ADDED" || FileActionRenamedNewName.String() != "RENAMED_NEW_NAME" {
		t.Error("FileAction.String mismatch")
	}
}
