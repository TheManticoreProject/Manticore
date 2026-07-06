package mswkst

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-WKST NDR
// structures in the absence of a live workstation service.
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

// TestScalarInfoStructs covers the plain fixed-layout WKSTA_INFO_* arms.
func TestScalarInfoStructs(t *testing.T) {
	roundTrip(t, "WKSTA_INFO_100", WKSTA_INFO_100{
		Wki100_platform_id:  500,
		Wki100_computername: wstr("HOST01"),
		Wki100_langroup:     wstr("WORKGROUP"),
		Wki100_ver_major:    10,
		Wki100_ver_minor:    0,
	})
	roundTrip(t, "WKSTA_INFO_102", WKSTA_INFO_102{
		Wki102_platform_id:     500,
		Wki102_computername:    wstr("HOST01"),
		Wki102_langroup:        wstr("CONTOSO"),
		Wki102_ver_major:       10,
		Wki102_ver_minor:       0,
		Wki102_lanroot:         nil,
		Wki102_logged_on_users: 3,
	})
	roundTrip(t, "WKSTA_INFO_1013", WKSTA_INFO_1013{Wki1013_keep_conn: 600})
	roundTrip(t, "WKSTA_INFO_502", WKSTA_INFO_502{
		Wki502_char_wait:                 3600,
		Wki502_keep_conn:                 600,
		Wki502_max_cmds:                  50,
		Wki502_use_opportunistic_locking: 1,
		Wki502_use_raw_read:              1,
	})
}

// TestStatWorkstation0 covers the large LARGE_INTEGER-heavy statistics structure.
func TestStatWorkstation0(t *testing.T) {
	roundTrip(t, "STAT_WORKSTATION_0", STAT_WORKSTATION_0{
		StatisticsStartTime: msdtyp.LARGE_INTEGER(0x01D7_0000_DEAD_BEEF),
		BytesReceived:       msdtyp.LARGE_INTEGER(1024),
		ReadOperations:      42,
		Sessions:            7,
	})
}

// TestWkstaInfoUnion exercises the WKSTA_INFO discriminated union across several arms.
func TestWkstaInfoUnion(t *testing.T) {
	roundTrip(t, "WKSTA_INFO/100", WKSTA_INFO{
		Tag:          100,
		WkstaInfo100: &WKSTA_INFO_100{Wki100_platform_id: 500, Wki100_computername: wstr("PC")},
	})
	roundTrip(t, "WKSTA_INFO/1013", WKSTA_INFO{
		Tag:           1013,
		WkstaInfo1013: &WKSTA_INFO_1013{Wki1013_keep_conn: 600},
	})
	roundTrip(t, "WKSTA_INFO/502", WKSTA_INFO{
		Tag:          502,
		WkstaInfo502: &WKSTA_INFO_502{Wki502_keep_conn: 600, Wki502_max_cmds: 50},
	})
}

// TestUseInfoContainers exercises the single-entry USE_INFO_*_CONTAINER (unique pointer to
// one struct) and the USE_ENUM_STRUCT union around them.
func TestUseInfoContainers(t *testing.T) {
	roundTrip(t, "USE_INFO_1_CONTAINER", USE_INFO_1_CONTAINER{
		EntriesRead: 1,
		Buffer: &USE_INFO_1{
			Ui1_local:    wstr("Z:"),
			Ui1_remote:   wstr(`\\srv\share`),
			Ui1_status:   0,
			Ui1_asg_type: 0,
		},
	})
	roundTrip(t, "USE_INFO_0_CONTAINER/nil", USE_INFO_0_CONTAINER{EntriesRead: 0, Buffer: nil})
	roundTrip(t, "USE_ENUM_STRUCT/1", USE_ENUM_STRUCT{
		Level: 1,
		UseInfo: USE_ENUM_UNION{
			Tag:    1,
			Level1: &USE_INFO_1_CONTAINER{EntriesRead: 1, Buffer: &USE_INFO_1{Ui1_local: wstr("Z:")}},
		},
	})
}

// TestUserEnumConformantArray exercises a WKSTA_USER_INFO_0_CONTAINER whose Buffer is a
// [unique] pointer to a conformant array of structs (the enum-buffer shape).
func TestUserEnumConformantArray(t *testing.T) {
	roundTrip(t, "WKSTA_USER_INFO_0_CONTAINER", WKSTA_USER_INFO_0_CONTAINER{
		EntriesRead: 2,
		Buffer: []WKSTA_USER_INFO_0{
			{Wkui0_username: wstr("alice")},
			{Wkui0_username: wstr("bob")},
		},
	})
	roundTrip(t, "WKSTA_TRANSPORT_INFO_0_CONTAINER", WKSTA_TRANSPORT_INFO_0_CONTAINER{
		EntriesRead: 1,
		Buffer: []WKSTA_TRANSPORT_INFO_0{
			{Wkti0_number_of_vcs: 4, Wkti0_transport_name: wstr(`\Device\NetBT_Tcpip`), Wkti0_wan_ish: 1},
		},
	})
	roundTrip(t, "WKSTA_USER_ENUM_STRUCT/0", WKSTA_USER_ENUM_STRUCT{
		Level: 0,
		WkstaUserInfo: WKSTA_USER_ENUM_UNION{
			Tag:    0,
			Level0: &WKSTA_USER_INFO_0_CONTAINER{EntriesRead: 1, Buffer: []WKSTA_USER_INFO_0{{Wkui0_username: wstr("carol")}}},
		},
	})
}

// TestJoinPasswordFixedArrays covers the fixed-size join password buffers, whose arrays are
// NDR fixed arrays (no conformance / referent) rather than conformant slices.
func TestJoinPasswordFixedArrays(t *testing.T) {
	var up JOINPR_USER_PASSWORD
	copy(up.Obfuscator[:], []byte{1, 2, 3, 4, 5, 6, 7, 8})
	up.Buffer[0] = 'P'
	up.Buffer[1] = 'w'
	up.Length = 4
	roundTrip(t, "JOINPR_USER_PASSWORD", up)

	var enc JOINPR_ENCRYPTED_USER_PASSWORD
	for i := range enc.Buffer {
		enc.Buffer[i] = uint8(i)
	}
	roundTrip(t, "JOINPR_ENCRYPTED_USER_PASSWORD", enc)

	roundTrip(t, "JOINPR_ENCRYPTED_USER_PASSWORD_AES", JOINPR_ENCRYPTED_USER_PASSWORD_AES{
		CbCipher: 3,
		Cipher:   []uint8{0xAA, 0xBB, 0xCC},
	})
}

// TestNetComputerNameArray covers the [unique] pointer to a conformant array of
// RPC_UNICODE_STRING returned by NetrEnumerateComputerNames.
func TestNetComputerNameArray(t *testing.T) {
	roundTrip(t, "NET_COMPUTER_NAME_ARRAY", NET_COMPUTER_NAME_ARRAY{
		EntryCount: 2,
		ComputerNames: []msdtyp.RPC_UNICODE_STRING{
			msdtyp.NewUnicodeString("host.contoso.com"),
			msdtyp.NewUnicodeString("alias.contoso.com"),
		},
	})
	roundTrip(t, "NET_COMPUTER_NAME_ARRAY/empty", NET_COMPUTER_NAME_ARRAY{EntryCount: 0, ComputerNames: nil})
}

// TestUnionArmReferentEmitted pins the wire shape of a pointer-arm union as it is actually
// used: embedded as a field (the codec gates a union to its selected arm only when it is a
// struct field, not when marshaled as a bare top-level value). An inline-value arm would
// emit [discriminant DWORD][body] for the union, whereas the correct [unique]-pointer arm
// emits [discriminant DWORD][arm referent id DWORD][body]. This length guard catches the
// arm-referent regression that a symmetric marshal/unmarshal round trip cannot (both
// directions would share the same wrong model).
func TestUnionArmReferentEmitted(t *testing.T) {
	// Mirror the response-struct usage: the union is a field, preceded by nothing else.
	type wrap struct {
		Info WKSTA_INFO
	}
	in := wrap{Info: WKSTA_INFO{Tag: 1013, WkstaInfo1013: &WKSTA_INFO_1013{Wki1013_keep_conn: 600}}}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Gated union: discriminant (1013) + arm referent id + deferred one-DWORD body (600).
	if len(raw) != 12 {
		t.Fatalf("embedded WKSTA_INFO(1013) marshaled to %d bytes (% x), want 12 (discriminant + arm referent id + one-DWORD body); an inline-value arm drops the referent id", len(raw), raw)
	}
	disc := uint32(raw[0]) | uint32(raw[1])<<8 | uint32(raw[2])<<16 | uint32(raw[3])<<24
	if disc != 1013 {
		t.Errorf("discriminant = %d, want 1013; raw = % x", disc, raw)
	}
	ref := uint32(raw[4]) | uint32(raw[5])<<8 | uint32(raw[6])<<16 | uint32(raw[7])<<24
	if ref == 0 {
		t.Errorf("arm referent id is zero (% x); the [unique] arm pointer was not emitted", raw)
	}
	body := uint32(raw[8]) | uint32(raw[9])<<8 | uint32(raw[10])<<16 | uint32(raw[11])<<24
	if body != 600 {
		t.Errorf("arm body = %d, want 600 (decode desynced?); raw = % x", body, raw)
	}

	// And it must round-trip when embedded.
	var out wrap
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Info.WkstaInfo1013 == nil || out.Info.WkstaInfo1013.Wki1013_keep_conn != 600 || out.Info.Tag != 1013 {
		t.Errorf("embedded union round trip mismatch: %+v", out.Info)
	}
}

// TestEnumWidths guards that the NDR enums are 16-bit named types (not accidentally 4-byte).
func TestEnumWidths(t *testing.T) {
	cases := map[string]int{
		"NETSETUP_JOIN_STATUS":   int(reflect.TypeOf(NETSETUP_JOIN_STATUS(0)).Size()),
		"NETSETUP_NAME_TYPE":     int(reflect.TypeOf(NETSETUP_NAME_TYPE(0)).Size()),
		"NET_COMPUTER_NAME_TYPE": int(reflect.TypeOf(NET_COMPUTER_NAME_TYPE(0)).Size()),
	}
	for name, sz := range cases {
		if sz != 2 {
			t.Errorf("%s: Go size %d bytes, expected 2 (NDR enums are 16-bit)", name, sz)
		}
	}
}
