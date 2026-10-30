package mspar

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// unionThenSentinel embeds the union in a struct, followed by the next field on the wire, so
// the codec applies union semantics and a wrong arm width shifts the sentinel.
type unionThenSentinel struct {
	Data     RPC_V2_NOTIFY_INFO_DATA_DATA
	Sentinel ndr.DWORD
}

// TestRPC_V2_NOTIFY_INFO_DATA_DATA_DwordWire pins the TABLE_DWORD arm to the IDL's
// "DWORD dwData[2]": the discriminant followed by two DWORDs and nothing else, so a field
// after the union is read from the right offset.
func TestRPC_V2_NOTIFY_INFO_DATA_DATA_DwordWire(t *testing.T) {
	in := unionThenSentinel{
		Data:     RPC_V2_NOTIFY_INFO_DATA_DATA{Tag: 1, DwData: [2]ndr.DWORD{0x11111111, 0x22222222}},
		Sentinel: 0xdeadbeef,
	}
	wire := []byte{
		0x01, 0x00, 0x00, 0x00, // discriminant TABLE_DWORD
		0x11, 0x11, 0x11, 0x11, 0x22, 0x22, 0x22, 0x22, // dwData[2]
		0xef, 0xbe, 0xad, 0xde, // Sentinel
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(raw, wire) {
		t.Errorf("Marshal = % x, want % x", raw, wire)
	}

	var out unionThenSentinel
	if err := ndr.Unmarshal(wire, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Data.Tag != in.Data.Tag || out.Data.DwData != in.Data.DwData || out.Sentinel != in.Sentinel {
		t.Errorf("Unmarshal = %+v, want %+v", out, in)
	}
}

// TestRPC_V2_NOTIFY_INFO_DATA_DATA_StringWire checks that only the selected TABLE_STRING arm
// is encoded: discriminant, STRING_CONTAINER (cbBuf + unique pointer + deferred conformant
// array), then the sentinel.
func TestRPC_V2_NOTIFY_INFO_DATA_DATA_StringWire(t *testing.T) {
	in := unionThenSentinel{
		Data: RPC_V2_NOTIFY_INFO_DATA_DATA{
			Tag:    2,
			String: STRING_CONTAINER{CbBuf: 4, PszString: []uint16{'h', 0}},
		},
		Sentinel: 0xdeadbeef,
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out unionThenSentinel
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Data.Tag != 2 || out.Data.String.CbBuf != 4 || len(out.Data.String.PszString) != 2 ||
		out.Data.String.PszString[0] != 'h' || out.Sentinel != in.Sentinel {
		t.Errorf("Unmarshal = %+v, want %+v", out, in)
	}
	// Non-selected arms must not be on the wire.
	if out.Data.DwData != [2]ndr.DWORD{} || out.Data.SystemTime.CbBuf != 0 {
		t.Errorf("non-selected arms decoded: %+v", out.Data)
	}
}
