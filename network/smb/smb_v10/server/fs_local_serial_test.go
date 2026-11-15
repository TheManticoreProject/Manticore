package server

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLocalFileSystemSerialFromHostVolume checks the volume serial is derived
// from the host volume, not the share path. Two shares on sibling directories of
// the same volume must report the same serial when the OS exposes one (the old
// path-hash gave two different serials). When no host serial is available the
// stable path-derived fallback is used, so the serial is always non-zero.
func TestLocalFileSystemSerialFromHostVolume(t *testing.T) {
	base := t.TempDir()
	dirA := filepath.Join(base, "a")
	dirB := filepath.Join(base, "b")
	for _, d := range []string{dirA, dirB} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	fsA, err := NewLocalFileSystem(dirA, "A")
	if err != nil {
		t.Fatalf("NewLocalFileSystem(a): %v", err)
	}
	defer fsA.Close()
	fsB, err := NewLocalFileSystem(dirB, "B")
	if err != nil {
		t.Fatalf("NewLocalFileSystem(b): %v", err)
	}
	defer fsB.Close()

	a, _ := fsA.VolumeInfo()
	b, _ := fsB.VolumeInfo()

	if a.SerialNumber == 0 || b.SerialNumber == 0 {
		t.Fatalf("serials must be non-zero, got a=%#x b=%#x", a.SerialNumber, b.SerialNumber)
	}

	// Both directories live on the same volume, so when the OS exposes a host
	// serial both shares report it and the two agree. If no host serial is
	// available both fell back to the path hash, which differs per path — that is
	// the acceptable fallback, so only assert equality when a host serial exists.
	if serialA, okA := hostVolumeSerial(dirA); okA {
		if serialB, okB := hostVolumeSerial(dirB); okB {
			if serialA != serialB {
				t.Fatalf("host volume serial differs for sibling dirs on one volume: %#x vs %#x", serialA, serialB)
			}
			if a.SerialNumber != serialA || b.SerialNumber != serialB {
				t.Fatalf("VolumeInfo serial did not use the host volume serial: a=%#x b=%#x host=%#x", a.SerialNumber, b.SerialNumber, serialA)
			}
			if a.SerialNumber != b.SerialNumber {
				t.Fatalf("two shares on one volume report different serials: %#x vs %#x", a.SerialNumber, b.SerialNumber)
			}
		}
	}
}
