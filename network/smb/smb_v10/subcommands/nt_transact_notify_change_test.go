package subcommands

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestNtTransactNotifyChangeSetupGolden(t *testing.T) {
	s := NtTransactNotifyChangeSetup{
		CompletionFilter: FILE_NOTIFY_CHANGE_NAME, // 0x03
		FID:              0x4002,
		WatchTree:        true,
	}
	got, err := s.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x03, 0x00, 0x00, 0x00, // CompletionFilter
		0x02, 0x40, // FID
		0x01, // WatchTree = true
		0x00, // Reserved
	}
	if !bytes.Equal(got, want) {
		t.Errorf("NT_TRANSACT_NOTIFY_CHANGE setup:\n got % x\nwant % x", got, want)
	}
	var out NtTransactNotifyChangeSetup
	if _, err := out.Unmarshal(got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != s {
		t.Errorf("round trip: got %+v want %+v", out, s)
	}
}

func TestFileNotifyInformationGolden(t *testing.T) {
	f := FileNotifyInformation{Action: FILE_ACTION_ADDED, FileName: "a.txt"}
	got, err := f.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// FileNameLength = 5 chars * 2 = 10 octets.
	if got := binary.LittleEndian.Uint32(got[8:12]); got != 10 {
		t.Errorf("FileNameLength: got %d want 10", got)
	}
	if binary.LittleEndian.Uint32(got[4:8]) != FILE_ACTION_ADDED {
		t.Errorf("Action mismatch")
	}
	if len(got) != 12+10 {
		t.Fatalf("length: got %d want %d", len(got), 12+10)
	}

	var out FileNotifyInformation
	n, err := out.Unmarshal(got)
	if err != nil || n != len(got) {
		t.Fatalf("Unmarshal: n=%d err=%v", n, err)
	}
	if out.Action != f.Action || out.FileName != f.FileName {
		t.Errorf("round trip: got %+v want %+v", out, f)
	}
}

func TestFileNotifyInformationListRoundTrip(t *testing.T) {
	items := []FileNotifyInformation{
		{Action: FILE_ACTION_ADDED, FileName: "new.txt"},   // 7 chars -> 14 octets, record 26 -> padded 28
		{Action: FILE_ACTION_REMOVED, FileName: "old.txt"}, // last entry, NextEntryOffset 0
	}
	raw, err := MarshalFileNotifyInformationList(items)
	if err != nil {
		t.Fatalf("Marshal list: %v", err)
	}
	// First record NextEntryOffset = align4(12 + 14) = 28.
	if got := binary.LittleEndian.Uint32(raw[0:4]); got != 28 {
		t.Errorf("first NextEntryOffset: got %d want 28", got)
	}

	out, err := ParseFileNotifyInformationList(raw)
	if err != nil {
		t.Fatalf("Parse list: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("entries: got %d want 2", len(out))
	}
	if out[0].Action != FILE_ACTION_ADDED || out[0].FileName != "new.txt" {
		t.Errorf("entry 0: %+v", out[0])
	}
	if out[1].Action != FILE_ACTION_REMOVED || out[1].FileName != "old.txt" || out[1].NextEntryOffset != 0 {
		t.Errorf("entry 1: %+v", out[1])
	}
}
