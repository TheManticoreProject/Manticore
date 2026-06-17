package dsrepl_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestDS_REPL_OPW_BLOB_RoundTrip(t *testing.T) {
	g1, _ := guid.FromString("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	g2, _ := guid.FromString("ffffffff-0000-1111-2222-333333333333")

	original := &dsrepl.DS_REPL_OPW_BLOB{
		Enqueued:             data_structures.FILETIME{DwLowDateTime: 0x12345678, DwHighDateTime: 0x01D0ABCD},
		SerialNumber:         42,
		Priority:             100,
		OpType:               dsrepl.DS_REPL_OP_TYPE_SYNC,
		Options:              0x00000010,
		NamingContext:        strptr("DC=example,DC=com"),
		DsaDN:                strptr("CN=NTDS Settings,CN=DC03"),
		DsaAddress:           strptr("dc03.example.com"),
		NamingContextObjGuid: *g1,
		DsaObjGuid:           *g2,
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_OPW_BLOB()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.Enqueued != original.Enqueued {
		t.Errorf("Enqueued = %+v, want %+v", parsed.Enqueued, original.Enqueued)
	}
	if parsed.SerialNumber != original.SerialNumber {
		t.Errorf("SerialNumber = %d, want %d", parsed.SerialNumber, original.SerialNumber)
	}
	if parsed.Priority != original.Priority {
		t.Errorf("Priority = %d, want %d", parsed.Priority, original.Priority)
	}
	if parsed.OpType != original.OpType {
		t.Errorf("OpType = %d, want %d", parsed.OpType, original.OpType)
	}
	if parsed.Options != original.Options {
		t.Errorf("Options = 0x%08x, want 0x%08x", parsed.Options, original.Options)
	}
	if !sameStr(parsed.NamingContext, original.NamingContext) {
		t.Errorf("NamingContext = %v, want %v", parsed.NamingContext, original.NamingContext)
	}
	if !sameStr(parsed.DsaDN, original.DsaDN) {
		t.Errorf("DsaDN = %v, want %v", parsed.DsaDN, original.DsaDN)
	}
	if !sameStr(parsed.DsaAddress, original.DsaAddress) {
		t.Errorf("DsaAddress = %v, want %v", parsed.DsaAddress, original.DsaAddress)
	}
	if !parsed.NamingContextObjGuid.Equal(&original.NamingContextObjGuid) {
		t.Errorf("NamingContextObjGuid mismatch")
	}
	if !parsed.DsaObjGuid.Equal(&original.DsaObjGuid) {
		t.Errorf("DsaObjGuid mismatch")
	}
}

func TestDS_REPL_OPW_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_OPW_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 20)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}
