package dsrepl_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

func TestDS_REPL_QUEUE_STATISTICSW_BLOB_RoundTrip(t *testing.T) {
	original := &dsrepl.DS_REPL_QUEUE_STATISTICSW_BLOB{
		CurrentOpStarted: msdtyp.FILETIME{DwLowDateTime: 1, DwHighDateTime: 0x01D00001},
		NumPendingOps:    5,
		OldestSync:       msdtyp.FILETIME{DwLowDateTime: 2, DwHighDateTime: 0x01D00002},
		OldestAdd:        msdtyp.FILETIME{DwLowDateTime: 3, DwHighDateTime: 0x01D00003},
		OldestMod:        msdtyp.FILETIME{DwLowDateTime: 4, DwHighDateTime: 0x01D00004},
		OldestDel:        msdtyp.FILETIME{DwLowDateTime: 5, DwHighDateTime: 0x01D00005},
		OldestUpdRefs:    msdtyp.FILETIME{DwLowDateTime: 6, DwHighDateTime: 0x01D00006},
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(marshalled) != 52 {
		t.Errorf("marshalled length = %d, want 52", len(marshalled))
	}

	parsed := dsrepl.NewDS_REPL_QUEUE_STATISTICSW_BLOB()
	n, err := parsed.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != 52 {
		t.Errorf("Unmarshal consumed %d bytes, want 52", n)
	}

	if *parsed != *original {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", *parsed, *original)
	}
}

func TestDS_REPL_QUEUE_STATISTICSW_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_QUEUE_STATISTICSW_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 51)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}
