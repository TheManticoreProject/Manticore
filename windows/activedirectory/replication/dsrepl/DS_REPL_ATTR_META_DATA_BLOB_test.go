package dsrepl_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestDS_REPL_ATTR_META_DATA_BLOB_RoundTrip(t *testing.T) {
	g, _ := guid.FromString("01234567-89ab-cdef-0123-456789abcdef")

	original := &dsrepl.DS_REPL_ATTR_META_DATA_BLOB{
		AttributeName:                  "objectClass",
		Version:                        4,
		LastOriginatingChange:          data_structures.FILETIME{DwLowDateTime: 0x11223344, DwHighDateTime: 0x01D05566},
		LastOriginatingDsaInvocationID: *g,
		OriginatingChange:              555000,
		LocalChange:                    556000,
		LastOriginatingDsaDN:           "CN=NTDS Settings,CN=DC05",
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_ATTR_META_DATA_BLOB()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.AttributeName != original.AttributeName {
		t.Errorf("AttributeName = %q, want %q", parsed.AttributeName, original.AttributeName)
	}
	if parsed.Version != original.Version {
		t.Errorf("Version = %d, want %d", parsed.Version, original.Version)
	}
	if parsed.LastOriginatingChange != original.LastOriginatingChange {
		t.Errorf("LastOriginatingChange = %+v, want %+v", parsed.LastOriginatingChange, original.LastOriginatingChange)
	}
	if !parsed.LastOriginatingDsaInvocationID.Equal(&original.LastOriginatingDsaInvocationID) {
		t.Errorf("LastOriginatingDsaInvocationID mismatch")
	}
	if parsed.OriginatingChange != original.OriginatingChange {
		t.Errorf("OriginatingChange = %d, want %d", parsed.OriginatingChange, original.OriginatingChange)
	}
	if parsed.LocalChange != original.LocalChange {
		t.Errorf("LocalChange = %d, want %d", parsed.LocalChange, original.LocalChange)
	}
	if parsed.LastOriginatingDsaDN != original.LastOriginatingDsaDN {
		t.Errorf("LastOriginatingDsaDN = %q, want %q", parsed.LastOriginatingDsaDN, original.LastOriginatingDsaDN)
	}
}

func TestDS_REPL_ATTR_META_DATA_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_ATTR_META_DATA_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 16)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}
