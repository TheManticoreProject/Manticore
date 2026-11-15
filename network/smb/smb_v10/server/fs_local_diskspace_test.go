package server

import "testing"

// TestLocalFileSystemVolumeInfoReportsCapacity guards against the former
// zero-capacity report: a LocalFileSystem share must advertise the real size and
// free space of the host volume, so clients do not see a disk of 0 bytes.
func TestLocalFileSystemVolumeInfoReportsCapacity(t *testing.T) {
	fs, err := NewLocalFileSystem(t.TempDir(), "TEST")
	if err != nil {
		t.Fatalf("NewLocalFileSystem: %v", err)
	}
	defer fs.Close()

	vi, err := fs.VolumeInfo()
	if err != nil {
		t.Fatalf("VolumeInfo: %v", err)
	}
	if vi.TotalBytes <= 0 {
		t.Fatalf("TotalBytes = %d, want the host volume capacity (> 0)", vi.TotalBytes)
	}
	if vi.FreeBytes < 0 || vi.FreeBytes > vi.TotalBytes {
		t.Fatalf("FreeBytes = %d, want 0..TotalBytes (%d)", vi.FreeBytes, vi.TotalBytes)
	}
}
