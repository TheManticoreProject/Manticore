package data_structures_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestGUID_Compatibility(t *testing.T) {
	// Test that GUID is properly aliased to guid.GUID
	var msGUID data_structures.GUID
	var windowsGUID guid.GUID

	// Both types should be the same
	var _ = msGUID == windowsGUID // This will fail to compile if types are incompatible

	// Create a GUID and verify it works with the underlying implementation
	testGUID := data_structures.GUID{
		A: 0x12345678,
		B: 0x1234,
		C: 0x5678,
		D: 0x9abc,
		E: 0xdef012345678,
	}

	// Test string representation
	expected := "12345678-1234-5678-9abc-def012345678"
	if result := testGUID.ToFormatD(); result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}

	// Test byte conversion
	expectedBytes := []byte{
		0x78, 0x56, 0x34, 0x12,
		0x34, 0x12,
		0x78, 0x56,
		0x9a, 0xbc,
		0xde, 0xf0, 0x12, 0x34, 0x56, 0x78,
	}
	resultBytes := testGUID.ToBytes()
	for i, b := range resultBytes {
		if b != expectedBytes[i] {
			t.Errorf("Expected byte %x at position %d, but got %x", expectedBytes[i], i, b)
		}
	}

	// Test involution (conversion to bytes and back)
	var reconstructedGUID data_structures.GUID
	reconstructedGUID.FromRawBytes(resultBytes)
	if !reconstructedGUID.Equal(&testGUID) {
		t.Errorf("GUIDs are not equal after involution. Before: %s, After: %s",
			testGUID.ToFormatD(), reconstructedGUID.ToFormatD())
	}
}
