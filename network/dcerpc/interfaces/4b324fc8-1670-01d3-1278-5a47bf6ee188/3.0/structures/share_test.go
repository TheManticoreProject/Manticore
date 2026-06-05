package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestShareInfo1ContainerRoundTrip marshals a SHARE_INFO_1_CONTAINER holding one entry
// and verifies it survives an NDR round trip unchanged.
func TestShareInfo1ContainerRoundTrip(t *testing.T) {
	in := SHARE_INFO_1_CONTAINER{
		EntriesRead: 1,
		Buffer: []SHARE_INFO_1{
			{
				Shi1Netname: "IPC$",
				Shi1Type:    3,
				Shi1Remark:  "Remote IPC",
			},
		},
	}

	data, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SHARE_INFO_1_CONTAINER
	if err := ndr.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

// TestShareEnumStructRoundTrip marshals a SHARE_ENUM_STRUCT at Level 1 whose union
// discriminant is 1 (Level1 container with one element) and verifies it survives an NDR
// round trip unchanged.
func TestShareEnumStructRoundTrip(t *testing.T) {
	in := SHARE_ENUM_STRUCT{
		Level: 1,
		ShareInfo: SHARE_ENUM_UNION{
			Tag: 1,
			Level1: &SHARE_INFO_1_CONTAINER{
				EntriesRead: 1,
				Buffer: []SHARE_INFO_1{
					{
						Shi1Netname: "C$",
						Shi1Type:    0,
						Shi1Remark:  "Default share",
					},
				},
			},
		},
	}

	data, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SHARE_ENUM_STRUCT
	if err := ndr.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}
