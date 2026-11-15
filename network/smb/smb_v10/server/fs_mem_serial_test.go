package server

import "testing"

// TestMemoryFileSystemSerialIsRandom guards against the former hardcoded
// 0x4D454D30 ("0MEM") volume serial: a MemoryFileSystem must report a
// non-constant serial that differs between instances.
func TestMemoryFileSystemSerialIsRandom(t *testing.T) {
	const oldConstant = 0x4D454D30

	a, _ := NewMemoryFileSystem("DATA").VolumeInfo()
	b, _ := NewMemoryFileSystem("DATA").VolumeInfo()

	if a.SerialNumber == oldConstant || b.SerialNumber == oldConstant {
		t.Fatalf("serial is the hardcoded 0x4D454D30 constant")
	}
	if a.SerialNumber == b.SerialNumber {
		t.Fatalf("two MemoryFileSystem instances share serial %#08x, want distinct random serials", a.SerialNumber)
	}
}
