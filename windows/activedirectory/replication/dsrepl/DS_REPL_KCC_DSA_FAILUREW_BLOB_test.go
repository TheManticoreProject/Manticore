package dsrepl_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestDS_REPL_KCC_DSA_FAILUREW_BLOB_RoundTrip(t *testing.T) {
	g, _ := guid.FromString("deadbeef-1234-5678-9abc-def012345678")

	original := &dsrepl.DS_REPL_KCC_DSA_FAILUREW_BLOB{
		DsaDN:        "CN=NTDS Settings,CN=DC02",
		DsaObjGuid:   *g,
		FirstFailure: data_structures.FILETIME{DwLowDateTime: 0xAABBCCDD, DwHighDateTime: 0x01D12345},
		NumFailures:  7,
		LastResult:   1722,
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_KCC_DSA_FAILUREW_BLOB()
	n, err := parsed.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(marshalled) {
		t.Errorf("Unmarshal consumed %d bytes, expected %d", n, len(marshalled))
	}

	if parsed.DsaDN != original.DsaDN {
		t.Errorf("DsaDN = %q, want %q", parsed.DsaDN, original.DsaDN)
	}
	if !parsed.DsaObjGuid.Equal(&original.DsaObjGuid) {
		t.Errorf("DsaObjGuid mismatch")
	}
	if parsed.FirstFailure != original.FirstFailure {
		t.Errorf("FirstFailure = %+v, want %+v", parsed.FirstFailure, original.FirstFailure)
	}
	if parsed.NumFailures != original.NumFailures {
		t.Errorf("NumFailures = %d, want %d", parsed.NumFailures, original.NumFailures)
	}
	if parsed.LastResult != original.LastResult {
		t.Errorf("LastResult = %d, want %d", parsed.LastResult, original.LastResult)
	}
}

func TestDS_REPL_KCC_DSA_FAILUREW_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_KCC_DSA_FAILUREW_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 10)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}
