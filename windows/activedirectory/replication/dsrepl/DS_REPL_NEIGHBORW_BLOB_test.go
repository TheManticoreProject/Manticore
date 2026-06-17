package dsrepl_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestDS_REPL_NEIGHBORW_BLOB_RoundTrip(t *testing.T) {
	g1, _ := guid.FromString("11111111-1111-1111-1111-111111111111")
	g2, _ := guid.FromString("22222222-2222-2222-2222-222222222222")
	g3, _ := guid.FromString("33333333-3333-3333-3333-333333333333")
	g4, _ := guid.FromString("44444444-4444-4444-4444-444444444444")

	original := &dsrepl.DS_REPL_NEIGHBORW_BLOB{
		NamingContext:                  strptr("DC=example,DC=com"),
		SourceDsaDN:                    strptr("CN=NTDS Settings,CN=DC01"),
		SourceDsaAddress:               strptr("abcd.example.com"),
		AsyncIntersiteTransportDN:      strptr("CN=IP,CN=Inter-Site Transports"),
		ReplicaFlags:                   0x00000050,
		Reserved:                       0,
		NamingContextObjGuid:           *g1,
		SourceDsaObjGuid:               *g2,
		SourceDsaInvocationID:          *g3,
		AsyncIntersiteTransportObjGuid: *g4,
		LastObjChangeSynced:            123456,
		AttributeFilter:                123000,
		LastSyncSuccess:                data_structures.FILETIME{DwLowDateTime: 0x11111111, DwHighDateTime: 0x01D00000},
		LastSyncAttempt:                data_structures.FILETIME{DwLowDateTime: 0x22222222, DwHighDateTime: 0x01D00001},
		LastSyncResult:                 0,
		NumConsecutiveSyncFailures:     3,
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_NEIGHBORW_BLOB()
	n, err := parsed.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(marshalled) {
		t.Errorf("Unmarshal consumed %d bytes, expected %d", n, len(marshalled))
	}

	if !sameStr(parsed.NamingContext, original.NamingContext) {
		t.Errorf("NamingContext = %v, want %v", parsed.NamingContext, original.NamingContext)
	}
	if !sameStr(parsed.SourceDsaDN, original.SourceDsaDN) {
		t.Errorf("SourceDsaDN = %v, want %v", parsed.SourceDsaDN, original.SourceDsaDN)
	}
	if !sameStr(parsed.SourceDsaAddress, original.SourceDsaAddress) {
		t.Errorf("SourceDsaAddress = %v, want %v", parsed.SourceDsaAddress, original.SourceDsaAddress)
	}
	if !sameStr(parsed.AsyncIntersiteTransportDN, original.AsyncIntersiteTransportDN) {
		t.Errorf("AsyncIntersiteTransportDN = %v, want %v", parsed.AsyncIntersiteTransportDN, original.AsyncIntersiteTransportDN)
	}
	if parsed.ReplicaFlags != original.ReplicaFlags {
		t.Errorf("ReplicaFlags = 0x%08x, want 0x%08x", parsed.ReplicaFlags, original.ReplicaFlags)
	}
	if !parsed.NamingContextObjGuid.Equal(&original.NamingContextObjGuid) {
		t.Errorf("NamingContextObjGuid mismatch")
	}
	if !parsed.SourceDsaInvocationID.Equal(&original.SourceDsaInvocationID) {
		t.Errorf("SourceDsaInvocationID mismatch")
	}
	if parsed.LastObjChangeSynced != original.LastObjChangeSynced {
		t.Errorf("LastObjChangeSynced = %d, want %d", parsed.LastObjChangeSynced, original.LastObjChangeSynced)
	}
	if parsed.LastSyncSuccess != original.LastSyncSuccess {
		t.Errorf("LastSyncSuccess = %+v, want %+v", parsed.LastSyncSuccess, original.LastSyncSuccess)
	}
	if parsed.NumConsecutiveSyncFailures != original.NumConsecutiveSyncFailures {
		t.Errorf("NumConsecutiveSyncFailures = %d, want %d", parsed.NumConsecutiveSyncFailures, original.NumConsecutiveSyncFailures)
	}

	// Re-marshalling the parsed structure must yield identical bytes.
	remarshalled, err := parsed.Marshal()
	if err != nil {
		t.Fatalf("re-Marshal failed: %v", err)
	}
	if string(remarshalled) != string(marshalled) {
		t.Errorf("re-marshalled bytes differ from original")
	}
}

func TestDS_REPL_NEIGHBORW_BLOB_NullVsEmptyStrings(t *testing.T) {
	// AsyncIntersiteTransportDN is NULL for RPC/IP replication: a nil pointer must
	// round-trip as an offset of 0 (absent), while a present-but-empty string must
	// round-trip as a non-nil pointer to "" — the distinction *string preserves.
	original := &dsrepl.DS_REPL_NEIGHBORW_BLOB{
		NamingContext:             strptr("DC=example,DC=com"),
		SourceDsaDN:               strptr(""), // present but empty
		AsyncIntersiteTransportDN: nil,        // absent (NULL)
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_NEIGHBORW_BLOB()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if parsed.AsyncIntersiteTransportDN != nil {
		t.Errorf("AsyncIntersiteTransportDN = %v, want nil (NULL)", parsed.AsyncIntersiteTransportDN)
	}
	if parsed.SourceDsaDN == nil || *parsed.SourceDsaDN != "" {
		t.Errorf("SourceDsaDN = %v, want non-nil pointer to empty string", parsed.SourceDsaDN)
	}
	if !sameStr(parsed.NamingContext, original.NamingContext) {
		t.Errorf("NamingContext = %v, want %v", parsed.NamingContext, original.NamingContext)
	}
}

func TestDS_REPL_NEIGHBORW_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_NEIGHBORW_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 16)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}
