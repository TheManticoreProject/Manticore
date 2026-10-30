package dsrepl_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/activedirectory/replication/dsrepl"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

func TestDS_REPL_VALUE_META_DATA_BLOB_RoundTrip(t *testing.T) {
	g, _ := guid.FromString("fedcba98-7654-3210-fedc-ba9876543210")

	original := &dsrepl.DS_REPL_VALUE_META_DATA_BLOB{
		AttributeName:                  strptr("member"),
		ObjectDn:                       strptr("CN=Group,DC=example,DC=com"),
		Data:                           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		Deleted:                        msdtyp.FILETIME{DwLowDateTime: 0, DwHighDateTime: 0},
		Created:                        msdtyp.FILETIME{DwLowDateTime: 0xAAAA, DwHighDateTime: 0x01D00001},
		Version:                        2,
		LastOriginatingChange:          msdtyp.FILETIME{DwLowDateTime: 0xBBBB, DwHighDateTime: 0x01D00002},
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

// TestDS_REPL_VALUE_META_DATA_BLOB_HugeOffsets checks that string offsets and
// pbData/cbData values >= 2^31 are rejected with an error. On 32-bit builds the
// int conversions of these values were negative (or their sum wrapped), so the
// range checks passed and the subsequent slice or make panicked.
func TestDS_REPL_VALUE_META_DATA_BLOB_HugeOffsets(t *testing.T) {
	put32 := func(b []byte, off int, v uint32) {
		b[off], b[off+1], b[off+2], b[off+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}

	cases := []struct {
		name   string
		oszAtt uint32
		cbData uint32
		pbData uint32
	}{
		{name: "oszAttributeName=0x80000000", oszAtt: 0x80000000},
		{name: "oszAttributeName=0xFFFFFFF0", oszAtt: 0xFFFFFFF0},
		{name: "cbData=0xFFFFFFF0", cbData: 0xFFFFFFF0, pbData: 0x10},
		{name: "pbData+cbData wraps", cbData: 0x80000010, pbData: 0x80000000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Unmarshal panicked: %v", r)
				}
			}()
			data := make([]byte, 96)
			put32(data, 0, tc.oszAtt)
			put32(data, 8, tc.cbData)
			put32(data, 12, tc.pbData)
			parsed := dsrepl.NewDS_REPL_VALUE_META_DATA_BLOB()
			if _, err := parsed.Unmarshal(data); err == nil {
				t.Fatal("expected error for out-of-range offset, got nil")
			}
		})
	}
}
