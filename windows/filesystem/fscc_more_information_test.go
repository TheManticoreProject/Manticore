package filesystem

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

// le builds a little-endian buffer from a list of fixed-width values, so each
// test spells out the [MS-FSCC] field layout independently of Marshal.
func le(fields ...any) []byte {
	var b bytes.Buffer
	for _, f := range fields {
		if err := binary.Write(&b, binary.LittleEndian, f); err != nil {
			panic(err)
		}
	}
	return b.Bytes()
}

func utf16le(s string) []byte { return encodeUTF16LE(s) }

func TestFileDirectoryInformationWireLayout(t *testing.T) {
	name := utf16le("report.txt")
	want := le(uint32(0), uint32(7), uint64(11), uint64(22), uint64(33), uint64(44),
		int64(1234), int64(4096), uint32(0x20), uint32(len(name)))
	want = append(want, name...)
	if len(want)-len(name) != 64 {
		t.Fatalf("fixed part is %d bytes, want 64", len(want)-len(name))
	}

	in := FileDirectoryInformation{FileIndex: 7, CreationTime: 11, LastAccessTime: 22, LastWriteTime: 33,
		ChangeTime: 44, EndOfFile: 1234, AllocationSize: 4096, FileAttributes: 0x20, FileName: "report.txt"}
	got, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Marshal:\n got %x\nwant %x", got, want)
	}
	var out FileDirectoryInformation
	if err := out.Unmarshal(want); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("Unmarshal = %+v, want %+v", out, in)
	}
	if err := out.Unmarshal(want[:63]); err == nil {
		t.Error("Unmarshal accepted a 63-byte buffer")
	}
}

func TestFileIdExtdDirectoryInformationWireLayout(t *testing.T) {
	name := utf16le("a")
	fileID := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	want := le(uint32(0), uint32(0), uint64(1), uint64(2), uint64(3), uint64(4),
		int64(5), int64(6), uint32(0x400), uint32(len(name)), uint32(9), uint32(0xA000000C), fileID)
	want = append(want, name...)
	if len(want)-len(name) != 88 {
		t.Fatalf("fixed part is %d bytes, want 88", len(want)-len(name))
	}

	in := FileIdExtdDirectoryInformation{CreationTime: 1, LastAccessTime: 2, LastWriteTime: 3, ChangeTime: 4,
		EndOfFile: 5, AllocationSize: 6, FileAttributes: 0x400, EaSize: 9, ReparsePointTag: 0xA000000C,
		FileId: fileID, FileName: "a"}
	got, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Marshal:\n got %x\nwant %x", got, want)
	}
	var out FileIdExtdDirectoryInformation
	if err := out.Unmarshal(want); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("Unmarshal = %+v, want %+v", out, in)
	}
	if err := out.Unmarshal(want[:87]); err == nil {
		t.Error("Unmarshal accepted an 87-byte buffer")
	}
}

func TestFileStreamInformationListWireLayout(t *testing.T) {
	// "::$DATA" is 14 bytes, so the first entry is 38 bytes and the second
	// starts at the next 8-byte boundary, 40.
	first, second := utf16le("::$DATA"), utf16le(":ads:$DATA")
	want := le(uint32(40), uint32(len(first)), int64(10), int64(4096))
	want = append(want, first...)
	want = append(want, 0, 0)
	want = append(want, le(uint32(0), uint32(len(second)), int64(3), int64(8))...)
	want = append(want, second...)

	streams := []FileStreamInformation{
		{NextEntryOffset: 999, StreamSize: 10, StreamAllocationSize: 4096, StreamName: "::$DATA"},
		{StreamSize: 3, StreamAllocationSize: 8, StreamName: ":ads:$DATA"},
	}
	got, err := MarshalFileStreamInformationList(streams)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("MarshalFileStreamInformationList:\n got %x\nwant %x", got, want)
	}

	parsed := ParseFileStreamInformationList(want)
	if len(parsed) != 2 {
		t.Fatalf("parsed %d entries, want 2", len(parsed))
	}
	streams[0].NextEntryOffset = 40
	for i := range parsed {
		if parsed[i] != streams[i] {
			t.Errorf("entry %d = %+v, want %+v", i, parsed[i], streams[i])
		}
	}

	if got := ParseFileStreamInformationList(nil); len(got) != 0 {
		t.Errorf("empty buffer parsed to %d entries", len(got))
	}
	// An offset past the buffer ends the chain after the first entry.
	bad := append([]byte(nil), want...)
	binary.LittleEndian.PutUint32(bad[0:4], 0x7FFFFFFF)
	if got := ParseFileStreamInformationList(bad); len(got) != 1 {
		t.Errorf("out-of-range NextEntryOffset parsed to %d entries, want 1", len(got))
	}
}

func TestFileFsSectorSizeInformationWireLayout(t *testing.T) {
	want := le(uint32(512), uint32(4096), uint32(4096), uint32(4096),
		uint32(SSINFO_FLAGS_ALIGNED_DEVICE|SSINFO_FLAGS_NO_SEEK_PENALTY), uint32(0), uint32(SSINFO_OFFSET_UNKNOWN))
	in := FileFsSectorSizeInformation{
		LogicalBytesPerSector: 512, PhysicalBytesPerSectorForAtomicity: 4096,
		PhysicalBytesPerSectorForPerformance: 4096, FileSystemEffectivePhysicalBytesPerSectorForAtomicity: 4096,
		Flags: 0x5, ByteOffsetForPartitionAlignment: 0xFFFFFFFF,
	}
	got, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) || len(got) != FileFsSectorSizeInformationSize {
		t.Fatalf("Marshal:\n got %x\nwant %x", got, want)
	}
	var out FileFsSectorSizeInformation
	if err := out.Unmarshal(want); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("Unmarshal = %+v, want %+v", out, in)
	}
	if err := out.Unmarshal(want[:27]); err == nil {
		t.Error("Unmarshal accepted a 27-byte buffer")
	}
}

func TestFileFsControlInformationWireLayout(t *testing.T) {
	want := le(int64(1), int64(2), int64(3), uint64(QuotaNoLimit), uint64(1<<30),
		uint32(FILE_VC_QUOTA_TRACK|FILE_VC_LOG_QUOTA_LIMIT), uint32(0))
	in := FileFsControlInformation{FreeSpaceStartFiltering: 1, FreeSpaceThreshold: 2, FreeSpaceStopFiltering: 3,
		DefaultQuotaThreshold: QuotaNoLimit, DefaultQuotaLimit: 1 << 30, FileSystemControlFlags: 0x21}
	got, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) || len(got) != FileFsControlInformationSize {
		t.Fatalf("Marshal:\n got %x\nwant %x", got, want)
	}
	var out FileFsControlInformation
	if err := out.Unmarshal(want); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("Unmarshal = %+v, want %+v", out, in)
	}
	// The padding is optional on receipt; anything shorter is not.
	if err := out.Unmarshal(want[:44]); err != nil || out != in {
		t.Errorf("Unmarshal without padding = %+v, %v", out, err)
	}
	if err := out.Unmarshal(want[:43]); err == nil {
		t.Error("Unmarshal accepted a 43-byte buffer")
	}
}

// TestFileStreamInformationFromWindows decodes the FileStreamInformation answer
// Windows Server 2025 gave over SMB2 for a 16-byte file with a 3-byte ":mcads"
// alternate stream, and re-encodes it to the same bytes.
func TestFileStreamInformationFromWindows(t *testing.T) {
	raw, err := hex.DecodeString("280000000e000000100000000000000010000000000000003a003a002400440041005400410000000000000018000000030000000000000008000000000000003a006d0063006100640073003a0024004400410054004100")
	if err != nil {
		t.Fatal(err)
	}
	streams := ParseFileStreamInformationList(raw)
	want := []FileStreamInformation{
		{NextEntryOffset: 40, StreamSize: 16, StreamAllocationSize: 16, StreamName: "::$DATA"},
		{StreamSize: 3, StreamAllocationSize: 8, StreamName: ":mcads:$DATA"},
	}
	if len(streams) != len(want) {
		t.Fatalf("parsed %d streams, want %d", len(streams), len(want))
	}
	for i := range want {
		if streams[i] != want[i] {
			t.Errorf("stream %d = %+v, want %+v", i, streams[i], want[i])
		}
	}
	again, err := MarshalFileStreamInformationList(streams)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, raw) {
		t.Errorf("re-encoded:\n got %x\nwant %x", again, raw)
	}
}
