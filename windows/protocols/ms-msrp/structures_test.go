package msmsrp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in. This is the wire-shape acceptance gate for the
// MS-MSRP NDR structures in the absence of a live Messenger service.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

// TestMSG_INFO_0 covers the single [unique,string] wchar_t* name, present and null.
func TestMSG_INFO_0(t *testing.T) {
	roundTrip(t, "MSG_INFO_0/name", MSG_INFO_0{Msgi0_name: wstr("ALICE")})
	roundTrip(t, "MSG_INFO_0/null", MSG_INFO_0{})
}

// TestMSG_INFO_1 covers two [unique,string] arms plus the forward flag.
func TestMSG_INFO_1(t *testing.T) {
	roundTrip(t, "MSG_INFO_1/full", MSG_INFO_1{
		Msgi1_name:         wstr("ALICE"),
		Msgi1_forward_flag: 1,
		Msgi1_forward:      wstr("BOB"),
	})
	roundTrip(t, "MSG_INFO_1/no-forward", MSG_INFO_1{Msgi1_name: wstr("ALICE")})
	roundTrip(t, "MSG_INFO_1/null", MSG_INFO_1{})
}

// TestMSG_INFO_0_CONTAINER covers the [unique] pointer to a conformant array of
// pointer-bearing structs, plus the empty and null-buffer shapes.
func TestMSG_INFO_0_CONTAINER(t *testing.T) {
	roundTrip(t, "MSG_INFO_0_CONTAINER/two", MSG_INFO_0_CONTAINER{
		EntriesRead: 2,
		Buffer:      []MSG_INFO_0{{Msgi0_name: wstr("ALICE")}, {Msgi0_name: wstr("BOB")}},
	})
	roundTrip(t, "MSG_INFO_0_CONTAINER/empty", MSG_INFO_0_CONTAINER{EntriesRead: 0, Buffer: []MSG_INFO_0{}})
}

// TestMSG_INFO_1_CONTAINER mirrors the Level-1 container.
func TestMSG_INFO_1_CONTAINER(t *testing.T) {
	roundTrip(t, "MSG_INFO_1_CONTAINER/one", MSG_INFO_1_CONTAINER{
		EntriesRead: 1,
		Buffer: []MSG_INFO_1{{
			Msgi1_name:         wstr("ALICE"),
			Msgi1_forward_flag: 0,
			Msgi1_forward:      nil,
		}},
	})
	roundTrip(t, "MSG_INFO_1_CONTAINER/empty", MSG_INFO_1_CONTAINER{EntriesRead: 0, Buffer: []MSG_INFO_1{}})
}

// TestMSG_ENUM_STRUCT exercises the [switch_is(Level)] union across both arms; the
// inline union Tag is set to match Level, as a caller must do before marshalling.
func TestMSG_ENUM_STRUCT(t *testing.T) {
	roundTrip(t, "MSG_ENUM_STRUCT/level0", MSG_ENUM_STRUCT{
		Level: 0,
		MsgInfo: MSG_ENUM_UNION{
			Tag: 0,
			Level0: &MSG_INFO_0_CONTAINER{
				EntriesRead: 1,
				Buffer:      []MSG_INFO_0{{Msgi0_name: wstr("ALICE")}},
			},
		},
	})
	roundTrip(t, "MSG_ENUM_STRUCT/level1", MSG_ENUM_STRUCT{
		Level: 1,
		MsgInfo: MSG_ENUM_UNION{
			Tag: 1,
			Level1: &MSG_INFO_1_CONTAINER{
				EntriesRead: 1,
				Buffer:      []MSG_INFO_1{{Msgi1_name: wstr("BOB"), Msgi1_forward_flag: 1, Msgi1_forward: wstr("CAROL")}},
			},
		},
	})
}

// TestMSG_INFO exercises the [switch_type(DWORD)] union returned by
// NetrMessageNameGetInfo across both arms.
func TestMSG_INFO(t *testing.T) {
	roundTrip(t, "MSG_INFO/level0", MSG_INFO{
		Tag:      0,
		MsgInfo0: &MSG_INFO_0{Msgi0_name: wstr("ALICE")},
	})
	roundTrip(t, "MSG_INFO/level1", MSG_INFO{
		Tag:      1,
		MsgInfo1: &MSG_INFO_1{Msgi1_name: wstr("ALICE"), Msgi1_forward_flag: 1, Msgi1_forward: wstr("BOB")},
	})
}
