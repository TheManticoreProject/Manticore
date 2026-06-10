package filesystem

import (
	"testing"
)

func TestFileNetworkOpenInformationRoundTrip(t *testing.T) {
	in := &FileNetworkOpenInformation{
		CreationTime:   0x01D0000000000001,
		LastAccessTime: 0x01D0000000000002,
		LastWriteTime:  0x01D0000000000003,
		ChangeTime:     0x01D0000000000004,
		AllocationSize: 4096,
		EndOfFile:      1234,
		FileAttributes: 0x80,
		Reserved:       0,
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(b) != FileNetworkOpenInformationSize {
		t.Fatalf("marshalled size = %d, want %d", len(b), FileNetworkOpenInformationSize)
	}
	var out FileNetworkOpenInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
	if err := (&FileNetworkOpenInformation{}).Unmarshal(b[:10]); err == nil {
		t.Error("Unmarshal should reject a short buffer")
	}
}

func TestFileAllInformationRoundTrip(t *testing.T) {
	in := &FileAllInformation{
		Basic:                FileBasicInformation{CreationTime: 11, LastAccessTime: 22, LastWriteTime: 33, ChangeTime: 44, FileAttributes: 0x20},
		Standard:             FileStandardInformation{AllocationSize: 8192, EndOfFile: 4096, NumberOfLinks: 1, DeletePending: false, Directory: false},
		IndexNumber:          0x1122334455667788,
		EaSize:               7,
		AccessFlags:          0x001F01FF,
		CurrentByteOffset:    256,
		Mode:                 0x10,
		AlignmentRequirement: 1,
		FileName:             "report.txt",
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out FileAllInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Basic != in.Basic || out.Standard != in.Standard || out.IndexNumber != in.IndexNumber ||
		out.EaSize != in.EaSize || out.AccessFlags != in.AccessFlags || out.CurrentByteOffset != in.CurrentByteOffset ||
		out.Mode != in.Mode || out.AlignmentRequirement != in.AlignmentRequirement || out.FileName != in.FileName {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestFileBothDirectoryInformationRoundTrip(t *testing.T) {
	in := &FileBothDirectoryInformation{
		FileIndex:      0,
		CreationTime:   100,
		LastAccessTime: 200,
		LastWriteTime:  300,
		ChangeTime:     400,
		EndOfFile:      512,
		AllocationSize: 1024,
		FileAttributes: 0x20,
		EaSize:         0,
		ShortName:      "REPORT~1.TXT",
		FileName:       "report.txt",
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out FileBothDirectoryInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.FileName != in.FileName || out.ShortName != in.ShortName || out.EndOfFile != in.EndOfFile ||
		out.AllocationSize != in.AllocationSize || out.FileAttributes != in.FileAttributes || out.CreationTime != in.CreationTime {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestFileIdBothDirectoryInformationRoundTrip(t *testing.T) {
	in := &FileIdBothDirectoryInformation{
		CreationTime:   1,
		LastAccessTime: 2,
		LastWriteTime:  3,
		ChangeTime:     4,
		EndOfFile:      99,
		AllocationSize: 100,
		FileAttributes: 0x10,
		ShortName:      "DIR~1",
		FileId:         0xAABBCCDDEEFF0011,
		FileName:       "directory",
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out FileIdBothDirectoryInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.FileName != in.FileName || out.ShortName != in.ShortName || out.FileId != in.FileId || out.FileAttributes != in.FileAttributes {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestFileFullDirectoryInformationRoundTrip(t *testing.T) {
	in := &FileFullDirectoryInformation{
		CreationTime:   5,
		EndOfFile:      7,
		AllocationSize: 8,
		FileAttributes: 0x20,
		EaSize:         42,
		FileName:       "a.bin",
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out FileFullDirectoryInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.FileName != in.FileName || out.EaSize != in.EaSize || out.EndOfFile != in.EndOfFile || out.FileAttributes != in.FileAttributes {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestDirectoryInformationListChaining(t *testing.T) {
	// Build two chained FILE_BOTH_DIR_INFORMATION entries and parse them back.
	e1 := &FileBothDirectoryInformation{FileName: "alpha", FileAttributes: 0x20, EndOfFile: 1}
	e2 := &FileBothDirectoryInformation{FileName: "beta", FileAttributes: 0x20, EndOfFile: 2}
	b1, _ := e1.Marshal()
	b2, _ := e2.Marshal()
	// First entry's NextEntryOffset points to the second; build the buffer.
	e1.NextEntryOffset = uint32(len(b1))
	b1, _ = e1.Marshal()
	buf := append(append([]byte{}, b1...), b2...)

	got := ParseFileBothDirectoryInformationList(buf)
	if len(got) != 2 {
		t.Fatalf("parsed %d entries, want 2", len(got))
	}
	if got[0].FileName != "alpha" || got[1].FileName != "beta" {
		t.Errorf("names = %q,%q want alpha,beta", got[0].FileName, got[1].FileName)
	}
}

func TestFileFsVolumeInformationRoundTrip(t *testing.T) {
	in := &FileFsVolumeInformation{
		VolumeCreationTime: 0x01D0000000000009,
		VolumeSerialNumber: 0xDEADBEEF,
		SupportsObjects:    true,
		VolumeLabel:        "SYSTEM",
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out FileFsVolumeInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestFileFsFullSizeInformationRoundTrip(t *testing.T) {
	in := &FileFsFullSizeInformation{
		TotalAllocationUnits:           1000,
		CallerAvailableAllocationUnits: 400,
		ActualAvailableAllocationUnits: 500,
		SectorsPerAllocationUnit:       8,
		BytesPerSector:                 512,
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(b) != FileFsFullSizeInformationSize {
		t.Fatalf("size = %d, want %d", len(b), FileFsFullSizeInformationSize)
	}
	var out FileFsFullSizeInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestFileFsAttributeInformationRoundTrip(t *testing.T) {
	in := &FileFsAttributeInformation{
		FileSystemAttributes:       0x00000007,
		MaximumComponentNameLength: 255,
		FileSystemName:             "NTFS",
	}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out FileFsAttributeInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, *in)
	}
}

func TestFileFsDeviceInformationRoundTrip(t *testing.T) {
	in := &FileFsDeviceInformation{DeviceType: 7, Characteristics: 0x20}
	b, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(b) != FileFsDeviceInformationSize {
		t.Fatalf("size = %d, want %d", len(b), FileFsDeviceInformationSize)
	}
	var out FileFsDeviceInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip mismatch: got %+v want %+v", out, *in)
	}
}

func TestUTF16Helpers(t *testing.T) {
	for _, s := range []string{"", "a", "report.txt", "café", "日本語"} {
		if got := decodeUTF16LE(encodeUTF16LE(s)); got != s {
			t.Errorf("UTF-16 round trip for %q = %q", s, got)
		}
	}
	// A trailing NUL unit must be trimmed on decode.
	withNul := append(encodeUTF16LE("x"), 0x00, 0x00)
	if got := decodeUTF16LE(withNul); got != "x" {
		t.Errorf("decodeUTF16LE trailing NUL = %q, want x", got)
	}
}
