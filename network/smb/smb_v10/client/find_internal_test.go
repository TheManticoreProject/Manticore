package client

import (
	"encoding/binary"
	"testing"
	"time"
)

// makeBothDirInfoEntry builds one SMB_FIND_FILE_BOTH_DIRECTORY_INFO record
// (MS-CIFS 2.2.8.1.7): a 94-byte fixed portion followed by the OEM FileName.
func makeBothDirInfoEntry(name string, short string, size uint64, attrs uint32, created, accessed, modified, changed uint64, next uint32) []byte {
	fixed := make([]byte, bothDirInfoFixedSize)
	binary.LittleEndian.PutUint32(fixed[0:4], next)
	binary.LittleEndian.PutUint64(fixed[8:16], created)
	binary.LittleEndian.PutUint64(fixed[16:24], accessed)
	binary.LittleEndian.PutUint64(fixed[24:32], modified)
	binary.LittleEndian.PutUint64(fixed[32:40], changed)
	binary.LittleEndian.PutUint64(fixed[40:48], size)
	binary.LittleEndian.PutUint32(fixed[56:60], attrs)
	binary.LittleEndian.PutUint32(fixed[60:64], uint32(len(name)))
	// ShortName: UTF-16LE in the 24-byte field at offset 70, length byte at 68.
	su16 := []byte{}
	for _, r := range short {
		su16 = binary.LittleEndian.AppendUint16(su16, uint16(r))
	}
	fixed[68] = byte(len(su16))
	copy(fixed[70:94], su16)
	return append(fixed, []byte(name)...)
}

func TestParseBothDirInfo(t *testing.T) {
	const tick = uint64(116444736000000000) // FILETIME for the Unix epoch
	dir := makeBothDirInfoEntry("subdir", "SUBDIR", 0, fileAttributeDirectory, tick, tick, tick, tick, uint32(bothDirInfoFixedSize+len("subdir")))
	file := makeBothDirInfoEntry("report.txt", "REPORT~1.TXT", 4096, 0x20, tick+10_000_000, tick, tick, tick, 0)

	entries := parseBothDirInfo(append(dir, file...))

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	if entries[0].LongName != "subdir" {
		t.Errorf("entry0 LongName = %q, want %q", entries[0].LongName, "subdir")
	}
	if entries[0].ShortName != "SUBDIR" {
		t.Errorf("entry0 ShortName = %q, want %q", entries[0].ShortName, "SUBDIR")
	}
	if !entries[0].IsDirectory {
		t.Errorf("entry0 should be a directory (attrs=0x%x)", entries[0].Attributes)
	}

	if entries[1].LongName != "report.txt" {
		t.Errorf("entry1 LongName = %q, want %q", entries[1].LongName, "report.txt")
	}
	if entries[1].Size != 4096 {
		t.Errorf("entry1 Size = %d, want 4096", entries[1].Size)
	}
	if entries[1].IsDirectory {
		t.Errorf("entry1 should not be a directory")
	}
	if !entries[1].CreatedAt.Equal(time.Unix(1, 0).UTC()) {
		t.Errorf("entry1 CreatedAt = %v, want 1970-01-01T00:00:01Z", entries[1].CreatedAt)
	}
}

func TestParseBothDirInfoEmpty(t *testing.T) {
	if got := parseBothDirInfo(nil); len(got) != 0 {
		t.Errorf("parseBothDirInfo(nil) = %d entries, want 0", len(got))
	}
	// A truncated buffer (shorter than the fixed portion) yields no entries and no panic.
	if got := parseBothDirInfo(make([]byte, bothDirInfoFixedSize-1)); len(got) != 0 {
		t.Errorf("parseBothDirInfo(truncated) = %d entries, want 0", len(got))
	}
}

func TestParseTrans2Response(t *testing.T) {
	// Build a minimal standard Transaction2 response: 32-byte header placeholder,
	// WordCount=10, 10 parameter words, then parameter and data buffers located by
	// ParameterOffset/DataOffset (measured from the header start).
	const hdr = 32
	wordCount := 10
	params := []byte{0xAA, 0xBB}
	data := []byte{0x01, 0x02, 0x03}

	wordsStart := hdr + 1
	paramOffset := wordsStart + 2*wordCount + 2 // after words + ByteCount
	dataOffset := paramOffset + len(params)

	raw := make([]byte, dataOffset+len(data))
	raw[hdr] = byte(wordCount)
	w := raw[wordsStart:]
	binary.LittleEndian.PutUint16(w[6:8], uint16(len(params)))  // ParameterCount
	binary.LittleEndian.PutUint16(w[8:10], uint16(paramOffset)) // ParameterOffset
	binary.LittleEndian.PutUint16(w[12:14], uint16(len(data)))  // DataCount
	binary.LittleEndian.PutUint16(w[14:16], uint16(dataOffset)) // DataOffset
	copy(raw[paramOffset:], params)
	copy(raw[dataOffset:], data)

	gotParams, gotData, err := parseTrans2Response(raw)
	if err != nil {
		t.Fatalf("parseTrans2Response error: %v", err)
	}
	if string(gotParams) != string(params) {
		t.Errorf("params = % x, want % x", gotParams, params)
	}
	if string(gotData) != string(data) {
		t.Errorf("data = % x, want % x", gotData, data)
	}
}

func TestParseTrans2ResponseInterim(t *testing.T) {
	// WordCount < 10 (interim/short response) yields empty params/data, no error.
	raw := make([]byte, 32+1)
	raw[32] = 0
	params, data, err := parseTrans2Response(raw)
	if err != nil || len(params) != 0 || len(data) != 0 {
		t.Fatalf("interim parse = (%v, %v, %v), want empty/no error", params, data, err)
	}
}

func TestParseTrans2ResponseOutOfBounds(t *testing.T) {
	const hdr = 32
	wordCount := 10
	wordsStart := hdr + 1
	raw := make([]byte, wordsStart+2*wordCount)
	raw[hdr] = byte(wordCount)
	w := raw[wordsStart:]
	binary.LittleEndian.PutUint16(w[6:8], 100)   // ParameterCount far beyond buffer
	binary.LittleEndian.PutUint16(w[8:10], 9999) // ParameterOffset out of bounds
	if _, _, err := parseTrans2Response(raw); err == nil {
		t.Fatal("expected out-of-bounds error, got nil")
	}
}

func TestFiletimeToTime(t *testing.T) {
	if !filetimeToTime(0).IsZero() {
		t.Error("filetimeToTime(0) should be the zero time")
	}
	const epoch = uint64(116444736000000000)
	if got := filetimeToTime(epoch); !got.Equal(time.Unix(0, 0).UTC()) {
		t.Errorf("filetimeToTime(epoch) = %v, want 1970-01-01T00:00:00Z", got)
	}
	if got := filetimeToTime(epoch + 10_000_000); !got.Equal(time.Unix(1, 0).UTC()) {
		t.Errorf("filetimeToTime(epoch+1s) = %v, want 1970-01-01T00:00:01Z", got)
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	in := []byte{0x41, 0x00, 0x42, 0x00, 0x00, 0x00} // "AB" + NUL
	if got := decodeUTF16LE(in); got != "AB" {
		t.Errorf("decodeUTF16LE = %q, want %q", got, "AB")
	}
	if got := decodeUTF16LE(nil); got != "" {
		t.Errorf("decodeUTF16LE(nil) = %q, want empty", got)
	}
}

func TestBuildFindFirst2Params(t *testing.T) {
	b := buildFindFirst2Params("\\*")
	// SearchAttributes, SearchCount, Flags, InformationLevel, SearchStorageType(4), pattern, NUL
	if len(b) != 2+2+2+2+4+len("\\*")+1 {
		t.Fatalf("unexpected length %d", len(b))
	}
	if binary.LittleEndian.Uint16(b[6:8]) != smbFindFileBothDirectoryInfo {
		t.Errorf("InformationLevel = 0x%04x, want 0x%04x", binary.LittleEndian.Uint16(b[6:8]), smbFindFileBothDirectoryInfo)
	}
	if b[len(b)-1] != 0x00 {
		t.Error("pattern must be NUL-terminated")
	}
	if string(b[12:len(b)-1]) != "\\*" {
		t.Errorf("pattern = %q, want %q", string(b[12:len(b)-1]), "\\*")
	}
}

func TestBuildFindNext2Params(t *testing.T) {
	b := buildFindNext2Params(0x1234)
	if binary.LittleEndian.Uint16(b[0:2]) != 0x1234 {
		t.Errorf("SID = 0x%04x, want 0x1234", binary.LittleEndian.Uint16(b[0:2]))
	}
	if binary.LittleEndian.Uint16(b[4:6]) != smbFindFileBothDirectoryInfo {
		t.Errorf("InformationLevel = 0x%04x, want 0x%04x", binary.LittleEndian.Uint16(b[4:6]), smbFindFileBothDirectoryInfo)
	}
	// Flags must request SMB_FIND_CONTINUE_FROM_LAST (0x0008).
	if binary.LittleEndian.Uint16(b[10:12]) != 0x0008 {
		t.Errorf("Flags = 0x%04x, want 0x0008", binary.LittleEndian.Uint16(b[10:12]))
	}
}
