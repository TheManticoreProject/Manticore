package dsrepl_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestDS_REPL_VALUE_META_DATA_BLOB_RoundTrip(t *testing.T) {
	g, _ := guid.FromString("fedcba98-7654-3210-fedc-ba9876543210")

	original := &dsrepl.DS_REPL_VALUE_META_DATA_BLOB{
		AttributeName:                  strptr("member"),
		ObjectDn:                       strptr("CN=Group,DC=example,DC=com"),
		Data:                           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		Deleted:                        data_structures.FILETIME{DwLowDateTime: 0, DwHighDateTime: 0},
		Created:                        data_structures.FILETIME{DwLowDateTime: 0xAAAA, DwHighDateTime: 0x01D00001},
		Version:                        2,
		LastOriginatingChange:          data_structures.FILETIME{DwLowDateTime: 0xBBBB, DwHighDateTime: 0x01D00002},
		LastOriginatingDsaInvocationID: *g,
		OriginatingChange:              111,
		LocalChange:                    222,
		LastOriginatingDsaDN:           strptr("CN=NTDS Settings,CN=DC06"),
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_VALUE_META_DATA_BLOB()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !sameStr(parsed.AttributeName, original.AttributeName) {
		t.Errorf("AttributeName = %v, want %v", parsed.AttributeName, original.AttributeName)
	}
	if !sameStr(parsed.ObjectDn, original.ObjectDn) {
		t.Errorf("ObjectDn = %v, want %v", parsed.ObjectDn, original.ObjectDn)
	}
	if !bytes.Equal(parsed.Data, original.Data) {
		t.Errorf("Data = %v, want %v", parsed.Data, original.Data)
	}
	if parsed.Created != original.Created {
		t.Errorf("Created = %+v, want %+v", parsed.Created, original.Created)
	}
	if parsed.Version != original.Version {
		t.Errorf("Version = %d, want %d", parsed.Version, original.Version)
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
	if !sameStr(parsed.LastOriginatingDsaDN, original.LastOriginatingDsaDN) {
		t.Errorf("LastOriginatingDsaDN = %v, want %v", parsed.LastOriginatingDsaDN, original.LastOriginatingDsaDN)
	}
}

func TestDS_REPL_VALUE_META_DATA_BLOB_NoData(t *testing.T) {
	original := &dsrepl.DS_REPL_VALUE_META_DATA_BLOB{
		AttributeName: strptr("member"),
		ObjectDn:      strptr("CN=Group,DC=example,DC=com"),
		Data:          nil,
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := dsrepl.NewDS_REPL_VALUE_META_DATA_BLOB()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(parsed.Data) != 0 {
		t.Errorf("Data = %v, want empty", parsed.Data)
	}
}

func TestDS_REPL_VALUE_META_DATA_BLOB_ShortData(t *testing.T) {
	parsed := dsrepl.NewDS_REPL_VALUE_META_DATA_BLOB()
	if _, err := parsed.Unmarshal(make([]byte, 40)); err == nil {
		t.Error("expected error for short data, got nil")
	}
}
