package server

import (
	"encoding/binary"
	"testing"
	"time"
)

// TestVolumeCreationTimeEncoded guards against the former always-zero
// VolumeCreationTime: the SMB1 SMB_QUERY_FS_VOLUME_INFO encoder must emit the
// volume's creation time as a FILETIME, and zero only when none is set.
func TestVolumeCreationTimeEncoded(t *testing.T) {
	created := time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC)
	volume := VolumeInfo{
		Label: "FILES", FileSystemName: "NTFS", SerialNumber: 0x11223344,
		CreationTime:             created,
		SectorsPerAllocationUnit: 8, BytesPerSector: 512,
	}

	info, served := encodeVolumeInformation(smbQueryFsVolumeInfo, volume, true)
	if !served {
		t.Fatal("SMB_QUERY_FS_VOLUME_INFO is not served")
	}
	if got := binary.LittleEndian.Uint64(info[0:8]); got != filetimeOf(created) {
		t.Fatalf("VolumeCreationTime = %#x, want %#x", got, filetimeOf(created))
	}
	if got := binary.LittleEndian.Uint64(info[0:8]); got == 0 {
		t.Fatal("VolumeCreationTime is zero despite a creation time being set")
	}

	// With no creation time the field stays zero.
	volume.CreationTime = time.Time{}
	info, _ = encodeVolumeInformation(smbQueryFsVolumeInfo, volume, true)
	if got := binary.LittleEndian.Uint64(info[0:8]); got != 0 {
		t.Fatalf("VolumeCreationTime = %#x with no time set, want 0", got)
	}
}

// TestBackendsReportCreationTime confirms the shipped backends populate a
// non-zero volume creation time.
func TestBackendsReportCreationTime(t *testing.T) {
	mem, _ := NewMemoryFileSystem("DATA").VolumeInfo()
	if mem.CreationTime.IsZero() {
		t.Error("MemoryFileSystem reports a zero volume creation time")
	}

	fs, err := NewLocalFileSystem(t.TempDir(), "DATA")
	if err != nil {
		t.Fatalf("NewLocalFileSystem: %v", err)
	}
	local, _ := fs.VolumeInfo()
	if local.CreationTime.IsZero() {
		t.Error("LocalFileSystem reports a zero volume creation time")
	}
}
