package server

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
)

// TestPreNTFindLevelIsAdmitted asserts the pre-NT find levels are accepted rather
// than refused with STATUS_INVALID_INFO_CLASS.
//
// Admission is the gate that mattered: supportedFindLevel rejected these outright,
// so a client without CAP_NT_FIND could not enumerate a share at all. The entry
// encoding itself is asserted below, against the buffer assembly that had been
// written for the chained levels only.
func TestPreNTFindLevelIsAdmitted(t *testing.T) {
	for _, level := range []uint16{smbInfoStandard, smbInfoQueryEaSize, smbInfoQueryEasFromLst} {
		if !supportedFindLevel(level) {
			t.Errorf("find level 0x%04X is refused", level)
		}
	}
}

// TestPreNTFindEntryIsNotChained asserts a pre-NT entry is packed rather than
// linked, and that its dates survive.
//
// The NT levels begin each entry with the offset of the next, and the buffer
// assembly clears that field on the last one. A pre-NT entry begins with a
// creation date instead, so applying the same treatment would write zeroes over
// the first entry's creation and access dates — a listing that looks valid and
// carries wrong timestamps.
func TestPreNTFindEntryIsNotChained(t *testing.T) {
	if findLevelChains(smbInfoStandard) {
		t.Fatal("SMB_INFO_STANDARD is treated as a chained level")
	}
	if !findLevelChains(smbFindFileBothDirectoryInfo) {
		t.Fatal("SMB_FIND_FILE_BOTH_DIRECTORY_INFO is not treated as a chained level")
	}

	attr := FileAttr{
		Name:     "dated.txt",
		Size:     0x1234,
		Created:  mustTime(t, 2001, 2, 3, 4, 5, 6),
		Accessed: mustTime(t, 2002, 3, 4, 5, 6, 8),
		Modified: mustTime(t, 2003, 4, 5, 6, 7, 10),
	}

	entry, served := encodeLegacyFindEntry(smbInfoStandard, attr, false)
	if !served {
		t.Fatal("SMB_INFO_STANDARD is not served")
	}

	// The first field is a creation date, not a NextEntryOffset, so it must not
	// be zero for an entry with a creation time.
	if date := binary.LittleEndian.Uint16(entry[0:2]); date == 0 {
		t.Error("the entry begins with a zero, so its creation date was lost")
	}

	// The name follows the fixed part, with a one-byte length.
	fixed := informationlevels.SMB_INFO_STANDARD_SIZE
	if got := int(entry[fixed]); got != len(attr.Name) {
		t.Errorf("FileNameLength is %d, want %d", got, len(attr.Name))
	}
	if got := string(entry[fixed+1 : fixed+1+len(attr.Name)]); got != attr.Name {
		t.Errorf("the entry names %q, want %q", got, attr.Name)
	}
}

// TestPreNTQueryLevelsAreServed asserts the pre-NT query levels answer with the
// sizes the level can express.
func TestPreNTQueryLevelsAreServed(t *testing.T) {
	attr := FileAttr{
		Name:           "described.txt",
		Size:           0x2000,
		AllocationSize: 0x2000,
		Modified:       mustTime(t, 2003, 4, 5, 6, 7, 10),
	}

	standard, served := encodeLegacyFileInformation(smbInfoStandard, attr)
	if !served {
		t.Fatal("SMB_INFO_STANDARD is not served for a query")
	}
	if len(standard) != informationlevels.SMB_INFO_STANDARD_SIZE {
		t.Fatalf("SMB_INFO_STANDARD is %d bytes, want %d",
			len(standard), informationlevels.SMB_INFO_STANDARD_SIZE)
	}

	decoded := &informationlevels.SMB_INFO_STANDARD{}
	if _, err := decoded.Unmarshal(standard); err != nil {
		t.Fatalf("the answer did not parse: %v", err)
	}
	if int(decoded.Filedatasize) != int(attr.Size) {
		t.Errorf("FileDataSize is %d, want %d", decoded.Filedatasize, attr.Size)
	}

	// The EA-bearing level is the same structure plus a 4-byte EaSize of zero.
	withEa, served := encodeLegacyFileInformation(smbInfoQueryEaSize, attr)
	if !served {
		t.Fatal("SMB_INFO_QUERY_EA_SIZE is not served")
	}
	if len(withEa) != len(standard)+4 {
		t.Fatalf("SMB_INFO_QUERY_EA_SIZE is %d bytes, want %d", len(withEa), len(standard)+4)
	}
	if size := binary.LittleEndian.Uint32(withEa[len(standard):]); size != 0 {
		t.Errorf("EaSize is %d, want 0 — nothing here carries extended attributes", size)
	}
}

// TestIsNameValidAnswersWithNoData asserts the level answers success and nothing
// else.
//
// [MS-CIFS] section 2.2.8.3.5: "No parameters or data are returned on this
// InformationLevel request. An error is returned if the syntax of the name is
// incorrect." Reaching the encoder means the name already passed resolvePath, so
// success with an empty body is the complete answer.
func TestIsNameValidAnswersWithNoData(t *testing.T) {
	encoded, served := encodeLegacyFileInformation(smbInfoIsNameValid, FileAttr{Name: "valid.txt"})
	if !served {
		t.Fatal("SMB_INFO_IS_NAME_VALID is not served")
	}
	if len(encoded) != 0 {
		t.Errorf("the level returned %d bytes, want none", len(encoded))
	}
}

// TestStreamInformationDescribesTheDataStream asserts a file reports its unnamed
// data stream and a directory reports none.
func TestStreamInformationDescribesTheDataStream(t *testing.T) {
	file := FileAttr{Name: "streamed.txt", Size: 0x40, AllocationSize: 0x40}

	encoded, served := encodeLegacyFileInformation(smbQueryFileStreamInfo, file)
	if !served {
		t.Fatal("SMB_QUERY_FILE_STREAM_INFO is not served")
	}
	if len(encoded) == 0 {
		t.Fatal("a file reported no streams")
	}

	// NextEntryOffset(4) StreamNameLength(4) StreamSize(8) StreamAllocationSize(8).
	if next := binary.LittleEndian.Uint32(encoded[0:4]); next != 0 {
		t.Errorf("the only entry links to another at %d", next)
	}
	if size := binary.LittleEndian.Uint64(encoded[8:16]); size != uint64(file.Size) {
		t.Errorf("StreamSize is %d, want %d", size, file.Size)
	}

	// A directory has no streams, and an empty buffer is how that is reported.
	directory := FileAttr{Name: "adirectory", IsDir: true}
	empty, served := encodeLegacyFileInformation(smbQueryFileStreamInfo, directory)
	if !served {
		t.Fatal("SMB_QUERY_FILE_STREAM_INFO is not served for a directory")
	}
	if len(empty) != 0 {
		t.Errorf("a directory reported %d bytes of stream information, want none", len(empty))
	}
}

// TestCompressionInformationReportsNoCompression asserts the level says the file
// is not compressed rather than claiming a format.
//
// Claiming one would have a client compute sizes from a unit shift that does not
// describe the storage.
func TestCompressionInformationReportsNoCompression(t *testing.T) {
	attr := FileAttr{Name: "plain.txt", Size: 0x1000}

	encoded, served := encodeLegacyFileInformation(smbQueryFileCompressionInfo, attr)
	if !served {
		t.Fatal("SMB_QUERY_FILE_COMPRESSION_INFO is not served")
	}
	if size := binary.LittleEndian.Uint64(encoded[0:8]); size != uint64(attr.Size) {
		t.Errorf("CompressedFileSize is %d, want the file's %d", size, attr.Size)
	}
	if format := binary.LittleEndian.Uint16(encoded[8:10]); format != 0 {
		t.Errorf("CompressionFormat is %d, want 0 (COMPRESSION_FORMAT_NONE)", format)
	}
}

// TestPreNTVolumeLevelsAreServed asserts the two pre-NT volume levels answer, and
// that the allocation level's units agree with the volume's geometry.
func TestPreNTVolumeLevelsAreServed(t *testing.T) {
	volume := VolumeInfo{
		Label: "FILES", FileSystemName: "NTFS", SerialNumber: 0x11223344,
		TotalBytes: 1 << 30, FreeBytes: 1 << 29,
		SectorsPerAllocationUnit: 8, BytesPerSector: 512,
	}

	allocation, served := encodeLegacyVolumeInformation(smbInfoAllocation, volume, false)
	if !served {
		t.Fatal("SMB_INFO_ALLOCATION is not served")
	}
	// idFileSystem(4) cSectorUnit(4) cUnit(4) cUnitAvailable(4) cbSector(2).
	if len(allocation) != 18 {
		t.Fatalf("SMB_INFO_ALLOCATION is %d bytes, want 18", len(allocation))
	}
	unit := uint64(volume.SectorsPerAllocationUnit) * uint64(volume.BytesPerSector)
	if got := binary.LittleEndian.Uint32(allocation[8:12]); uint64(got) != uint64(volume.TotalBytes)/unit {
		t.Errorf("cUnit is %d, want %d", got, uint64(volume.TotalBytes)/unit)
	}
	if got := binary.LittleEndian.Uint16(allocation[16:18]); uint32(got) != volume.BytesPerSector {
		t.Errorf("cbSector is %d, want %d", got, volume.BytesPerSector)
	}

	// The volume level, in each encoding. cCharCount is the label's byte length,
	// which is what SMB_INFO_VOLUME.Marshal derives it as; under Unicode that is
	// twice the character count, and the two coincide in OEM.
	for _, unicode := range []bool{false, true} {
		info, served := encodeLegacyVolumeInformation(smbInfoVolume, volume, unicode)
		if !served {
			t.Fatalf("SMB_INFO_VOLUME is not served (unicode=%t)", unicode)
		}
		if serial := binary.LittleEndian.Uint32(info[0:4]); serial != volume.SerialNumber {
			t.Errorf("the serial number is 0x%08X, want 0x%08X (unicode=%t)",
				serial, volume.SerialNumber, unicode)
		}

		wantBytes := len(volume.Label)
		if unicode {
			wantBytes *= 2
		}
		if got := int(info[4]); got != wantBytes {
			t.Errorf("cCharCount is %d, want %d (unicode=%t)", got, wantBytes, unicode)
		}
		// And the count has to describe the label that follows it, or a client
		// reads the wrong extent.
		if len(info) != 5+wantBytes {
			t.Errorf("the level is %d bytes, want %d (unicode=%t)", len(info), 5+wantBytes, unicode)
		}
	}
}

// TestUnservedLevelIsStillRefused asserts adding these levels did not turn the
// range into a catch-all.
func TestUnservedLevelIsStillRefused(t *testing.T) {
	attr := FileAttr{Name: "described.txt"}

	if _, served := encodeLegacyFileInformation(0x0099, attr); served {
		t.Error("an unknown level was answered")
	}
	if supportedFindLevel(0x0099) {
		t.Error("an unknown find level is reported as supported")
	}
}

// mustTime builds a UTC time for a fixture.
func mustTime(t *testing.T, year, month, day, hour, minute, second int) time.Time {
	t.Helper()
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}
