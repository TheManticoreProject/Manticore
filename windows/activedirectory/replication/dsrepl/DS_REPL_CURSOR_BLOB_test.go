package dsrepl_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestDS_REPL_CURSOR_BLOB_RoundTrip(t *testing.T) {
	g, _ := guid.FromString("0a0b0c0d-0e0f-1011-1213-141516171819")

	original := &dsrepl.DS_REPL_CURSOR_BLOB{
		SourceDsaInvocationID: *g,
		AttributeFilter:       987654321,
		LastSyncSuccess:       data_structures.FILETIME{DwLowDateTime: 0xCAFEBABE, DwHighDateTime: 0x01D0FACE},
		SourceDsaDN:           strptr("CN=NTDS Settings,CN=DC04"),
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_CURSOR_BLOB()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !parsed.SourceDsaInvocationID.Equal(&original.SourceDsaInvocationID) {
		t.Errorf("SourceDsaInvocationID mismatch")
	}
	if parsed.AttributeFilter != original.AttributeFilter {
		t.Errorf("AttributeFilter = %d, want %d", parsed.AttributeFilter, original.AttributeFilter)
	}
	if parsed.LastSyncSuccess != original.LastSyncSuccess {
		t.Errorf("LastSyncSuccess = %+v, want %+v", parsed.LastSyncSuccess, original.LastSyncSuccess)
	}
	if !sameStr(parsed.SourceDsaDN, original.SourceDsaDN) {
		t.Errorf("SourceDsaDN = %v, want %v", parsed.SourceDsaDN, original.SourceDsaDN)
	}
}

func TestDS_REPL_CURSOR_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_CURSOR_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 8)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}

func TestDS_REPL_CURSOR_BLOB_FixedBytes(t *testing.T) {
	// Hand-built blob to validate the on-the-wire field layout independently of
	// Marshal: 16-byte GUID, 8-byte USN, 8-byte FILETIME, 4-byte string offset,
	// then a packed UTF-16LE string.
	data := []byte{
		// uuidSourceDsaInvocationID (16 bytes)
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		// usnAttributeFilter = 5 (little-endian)
		0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// fTimeLastSyncSuccess: low = 10, high = 20
		0x0a, 0x00, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00,
		// oszSourceDsaDN = 36 (points right after the 36-byte header)
		0x24, 0x00, 0x00, 0x00,
		// data: "DC" in UTF-16LE plus null terminator
		0x44, 0x00, 0x43, 0x00, 0x00, 0x00,
	}

	parsed := dsrepl.NewDS_REPL_CURSOR_BLOB()
	if _, err := parsed.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.AttributeFilter != 5 {
		t.Errorf("AttributeFilter = %d, want 5", parsed.AttributeFilter)
	}
	if parsed.LastSyncSuccess.DwLowDateTime != 10 || parsed.LastSyncSuccess.DwHighDateTime != 20 {
		t.Errorf("LastSyncSuccess = %+v, want {Low:10 High:20}", parsed.LastSyncSuccess)
	}
	if !sameStr(parsed.SourceDsaDN, strptr("DC")) {
		t.Errorf("SourceDsaDN = %v, want %q", parsed.SourceDsaDN, "DC")
	}
}

func TestDS_REPL_CURSOR_BLOB_BadOffset(t *testing.T) {
	// A 36-byte header whose oszSourceDsaDN points past the end of the buffer must
	// produce an out-of-range error rather than panicking.
	data := make([]byte, 36)
	data[32] = 0xFF // oszSourceDsaDN = 0x000000FF, out of range
	parsed := dsrepl.NewDS_REPL_CURSOR_BLOB()
	if _, err := parsed.Unmarshal(data); err == nil {
		t.Error("expected out-of-range error for bad offset, got nil")
	}
}
