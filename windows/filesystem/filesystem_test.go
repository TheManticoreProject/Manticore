package filesystem

import (
	"encoding/binary"
	"testing"
)

func TestFileBasicInformation_RoundTrip(t *testing.T) {
	in := &FileBasicInformation{
		CreationTime:   0x01D7A0000000000A,
		LastAccessTime: 0x01D7A0000000000B,
		LastWriteTime:  0x01D7A0000000000C,
		ChangeTime:     0x01D7A0000000000D,
		FileAttributes: 0x00000020,
	}
	b, _ := in.Marshal()
	if len(b) != FileBasicInformationSize {
		t.Fatalf("size = %d, want %d", len(b), FileBasicInformationSize)
	}
	var out FileBasicInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatal(err)
	}
	if out != *in {
		t.Errorf("round-trip mismatch: %+v vs %+v", out, *in)
	}
}

func TestFileStandardInformation_RoundTrip(t *testing.T) {
	in := &FileStandardInformation{AllocationSize: 4096, EndOfFile: 1234, NumberOfLinks: 1, DeletePending: true, Directory: false}
	b, _ := in.Marshal()
	if len(b) != FileStandardInformationSize {
		t.Fatalf("size = %d, want %d", len(b), FileStandardInformationSize)
	}
	var out FileStandardInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatal(err)
	}
	if out != *in {
		t.Errorf("round-trip mismatch: %+v vs %+v", out, *in)
	}
}

func TestFileEndOfFileInformation_RoundTrip(t *testing.T) {
	in := &FileEndOfFileInformation{EndOfFile: 0x1122334455}
	b, _ := in.Marshal()
	var out FileEndOfFileInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatal(err)
	}
	if out.EndOfFile != in.EndOfFile {
		t.Errorf("EndOfFile = %d, want %d", out.EndOfFile, in.EndOfFile)
	}
}

func TestFileDispositionInformation(t *testing.T) {
	b, _ := (&FileDispositionInformation{DeletePending: true}).Marshal()
	if len(b) != 1 || b[0] != 1 {
		t.Fatalf("DeletePending=true marshalled to % x", b)
	}
}

// TestFileRenameInformation_SMB2Layout verifies the SMB2 wire layout:
// ReplaceIfExists(1) Reserved(7) RootDirectory(8) FileNameLength(4) FileName(UTF-16LE).
func TestFileRenameInformation_SMB2Layout(t *testing.T) {
	in := &FileRenameInformation{ReplaceIfExists: true, FileName: `dir\new.txt`}
	b, _ := in.Marshal()

	if b[0] != 1 {
		t.Errorf("ReplaceIfExists byte = %d, want 1", b[0])
	}
	for i := 1; i < 16; i++ { // Reserved(7) + RootDirectory(8) must be zero
		if b[i] != 0 {
			t.Errorf("byte %d = %d, want 0 (Reserved/RootDirectory)", i, b[i])
		}
	}
	nameLen := binary.LittleEndian.Uint32(b[16:20])
	if int(nameLen) != len(in.FileName)*2 {
		t.Errorf("FileNameLength = %d, want %d", nameLen, len(in.FileName)*2)
	}
	if 20+int(nameLen) != len(b) {
		t.Errorf("total %d != 20 + name %d", len(b), nameLen)
	}

	var out FileRenameInformation
	if err := out.Unmarshal(b); err != nil {
		t.Fatal(err)
	}
	if out.FileName != in.FileName || out.ReplaceIfExists != in.ReplaceIfExists {
		t.Errorf("round-trip mismatch: %+v vs %+v", out, *in)
	}
}

func TestFileFsSizeInformation(t *testing.T) {
	b := make([]byte, FileFsSizeInformationSize)
	binary.LittleEndian.PutUint64(b[0:8], 1000)  // total units
	binary.LittleEndian.PutUint64(b[8:16], 400)  // available units
	binary.LittleEndian.PutUint32(b[16:20], 8)   // sectors/unit
	binary.LittleEndian.PutUint32(b[20:24], 512) // bytes/sector
	var fi FileFsSizeInformation
	if err := fi.Unmarshal(b); err != nil {
		t.Fatal(err)
	}
	if fi.TotalBytes() != 1000*8*512 || fi.AvailableBytes() != 400*8*512 {
		t.Errorf("byte totals wrong: total=%d avail=%d", fi.TotalBytes(), fi.AvailableBytes())
	}
}

func TestUnmarshal_ShortBuffers(t *testing.T) {
	if (&FileBasicInformation{}).Unmarshal(make([]byte, 10)) == nil {
		t.Error("expected error on short FileBasicInformation buffer")
	}
	if (&FileFsSizeInformation{}).Unmarshal([]byte{}) == nil {
		t.Error("expected error on empty FileFsSizeInformation buffer")
	}
	// A valid empty-name rename still needs the 20-byte fixed head.
	if (&FileRenameInformation{}).Unmarshal(make([]byte, 4)) == nil {
		t.Error("expected error on short FileRenameInformation buffer")
	}
}
