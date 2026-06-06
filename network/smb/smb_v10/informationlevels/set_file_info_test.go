package informationlevels_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestSetFileBasicInfoRoundTrip verifies SMB_SET_FILE_BASIC_INFO marshals to the
// spec size (40 bytes) and round-trips every field.
func TestSetFileBasicInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_SET_FILE_BASIC_INFO{
		Creationtime:      types.FILETIME{DwLowDateTime: 0x11111111, DwHighDateTime: 0x22222222},
		Lastaccesstime:    types.FILETIME{DwLowDateTime: 0x33333333, DwHighDateTime: 0x44444444},
		Lastwritetime:     types.FILETIME{DwLowDateTime: 0x55555555, DwHighDateTime: 0x66666666},
		Changetime:        types.FILETIME{DwLowDateTime: 0x77777777, DwHighDateTime: 0x88888888},
		Extfileattributes: types.ATTR_HIDDEN | types.ATTR_ARCHIVE,
		Reserved:          0xDEADBEEF,
	}

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(raw) != 40 {
		t.Fatalf("expected 40 bytes, got %d", len(raw))
	}

	out := &informationlevels.SMB_SET_FILE_BASIC_INFO{}
	n, err := out.Unmarshal(raw)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != 40 {
		t.Errorf("expected 40 bytes consumed, got %d", n)
	}
	if *out != *in {
		t.Errorf("round-trip mismatch:\n in  %+v\n out %+v", in, out)
	}
}

// TestSetFileAllocationInfoRoundTrip verifies SMB_SET_FILE_ALLOCATION_INFO is an
// 8-byte LARGE_INTEGER that round-trips.
func TestSetFileAllocationInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_SET_FILE_ALLOCATION_INFO{}
	in.Allocationsize.QuadPart = 0x0123456789ABCDEF

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(raw) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(raw))
	}

	out := &informationlevels.SMB_SET_FILE_ALLOCATION_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if out.Allocationsize.QuadPart != in.Allocationsize.QuadPart {
		t.Errorf("AllocationSize = 0x%X, want 0x%X", out.Allocationsize.QuadPart, in.Allocationsize.QuadPart)
	}
}

// TestSetFileEndOfFileInfoRoundTrip verifies SMB_SET_FILE_END_OF_FILE_INFO is an
// 8-byte LARGE_INTEGER that round-trips.
func TestSetFileEndOfFileInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_SET_FILE_END_OF_FILE_INFO{}
	in.Endoffile.QuadPart = 0xFEDCBA9876543210

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(raw) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(raw))
	}

	out := &informationlevels.SMB_SET_FILE_END_OF_FILE_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if out.Endoffile.QuadPart != in.Endoffile.QuadPart {
		t.Errorf("EndOfFile = 0x%X, want 0x%X", out.Endoffile.QuadPart, in.Endoffile.QuadPart)
	}
}

// TestSetFileDispositionInfoRoundTrip verifies SMB_SET_FILE_DISPOSITION_INFO is a
// single DeletePending byte that round-trips.
func TestSetFileDispositionInfoRoundTrip(t *testing.T) {
	in := &informationlevels.SMB_SET_FILE_DISPOSITION_INFO{Deletepending: 0x01}

	raw, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(raw) != 1 || raw[0] != 0x01 {
		t.Fatalf("expected [0x01], got % x", raw)
	}

	out := &informationlevels.SMB_SET_FILE_DISPOSITION_INFO{}
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if out.Deletepending != in.Deletepending {
		t.Errorf("DeletePending = %d, want %d", out.Deletepending, in.Deletepending)
	}
}

// TestSetFileInfoUnmarshalBounds verifies truncated buffers return an error rather
// than panicking.
func TestSetFileInfoUnmarshalBounds(t *testing.T) {
	if _, err := (&informationlevels.SMB_SET_FILE_BASIC_INFO{}).Unmarshal(make([]byte, 39)); err == nil {
		t.Error("SMB_SET_FILE_BASIC_INFO: expected error on 39-byte buffer")
	}
	if _, err := (&informationlevels.SMB_SET_FILE_ALLOCATION_INFO{}).Unmarshal(make([]byte, 7)); err == nil {
		t.Error("SMB_SET_FILE_ALLOCATION_INFO: expected error on 7-byte buffer")
	}
	if _, err := (&informationlevels.SMB_SET_FILE_END_OF_FILE_INFO{}).Unmarshal(make([]byte, 7)); err == nil {
		t.Error("SMB_SET_FILE_END_OF_FILE_INFO: expected error on 7-byte buffer")
	}
	if _, err := (&informationlevels.SMB_SET_FILE_DISPOSITION_INFO{}).Unmarshal(nil); err == nil {
		t.Error("SMB_SET_FILE_DISPOSITION_INFO: expected error on empty buffer")
	}
}
