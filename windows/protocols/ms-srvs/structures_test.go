package mssrvs

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTripConnFileSession marshals in, unmarshals into a fresh value of the same
// type, and asserts the result is deeply equal to in. It is named distinctly to
// avoid colliding with helpers defined by other test files in this package.
func roundTripConnFileSession[T any](t *testing.T, name string, in T) {
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

// TestSESSION_INFO_1_CONTAINER_RoundTrip exercises a [unique] pointer to a
// conformant array of structs that each carry [unique] string pointers.
func TestSESSION_INFO_1_CONTAINER_RoundTrip(t *testing.T) {
	in := SESSION_INFO_1_CONTAINER{
		EntriesRead: 2,
		Buffer: []SESSION_INFO_1{
			{
				Sesi1Cname:     "\\\\CLIENT1",
				Sesi1Username:  "alice",
				Sesi1NumOpens:  3,
				Sesi1Time:      120,
				Sesi1IdleTime:  10,
				Sesi1UserFlags: 1,
			},
			{
				Sesi1Cname:     "\\\\CLIENT2",
				Sesi1Username:  "",
				Sesi1NumOpens:  0,
				Sesi1Time:      0,
				Sesi1IdleTime:  0,
				Sesi1UserFlags: 0,
			},
		},
	}
	roundTripConnFileSession(t, "SESSION_INFO_1_CONTAINER", in)
}

// TestSESSION_ENUM_STRUCT_Level1_RoundTrip exercises the enum struct with a level-1
// union arm pointing at a one-element container.
func TestSESSION_ENUM_STRUCT_Level1_RoundTrip(t *testing.T) {
	in := SESSION_ENUM_STRUCT{
		Level: 1,
		SessionInfo: SESSION_ENUM_UNION{
			Tag: 1,
			Level1: &SESSION_INFO_1_CONTAINER{
				EntriesRead: 1,
				Buffer: []SESSION_INFO_1{
					{
						Sesi1Cname:     "\\\\HOST",
						Sesi1Username:  "bob",
						Sesi1NumOpens:  5,
						Sesi1Time:      300,
						Sesi1IdleTime:  42,
						Sesi1UserFlags: 0,
					},
				},
			},
		},
	}
	roundTripConnFileSession(t, "SESSION_ENUM_STRUCT_Level1", in)
}

// TestServerInfo101RoundTrip exercises a full-detail SERVER_INFO level with
// both embedded [unique,string] fields populated.
func TestServerInfo101RoundTrip(t *testing.T) {
	in := SERVER_INFO_101{
		Sv101PlatformId:   500,
		Sv101Name:         "SERVER01",
		Sv101VersionMajor: 6,
		Sv101VersionMinor: 1,
		Sv101Type:         0x00000003,
		Sv101Comment:      "a comment",
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SERVER_INFO_101
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

// TestServerInfoUnion101 exercises the SERVER_INFO union with Tag=101 and the
// matching ServerInfo101 arm selected.
func TestServerInfoUnion101(t *testing.T) {
	in := SERVER_INFO{
		Tag: 101,
		ServerInfo101: &SERVER_INFO_101{
			Sv101PlatformId:   500,
			Sv101Name:         "SERVER01",
			Sv101VersionMajor: 10,
			Sv101VersionMinor: 0,
			Sv101Type:         0x00001000,
			Sv101Comment:      "round trip",
		},
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SERVER_INFO
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if out.Tag != in.Tag {
		t.Errorf("Tag: got %d want %d", out.Tag, in.Tag)
	}
	if out.ServerInfo101 == nil {
		t.Fatalf("ServerInfo101 arm is nil after round trip")
	}
	if !reflect.DeepEqual(*in.ServerInfo101, *out.ServerInfo101) {
		t.Errorf("arm mismatch:\n got %+v\nwant %+v", *out.ServerInfo101, *in.ServerInfo101)
	}
}

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

// roundTripTransportMisc marshals in, unmarshals into a fresh value of the same
// type, and asserts the result is deeply equal to in. It is named distinctly to
// avoid colliding with helpers defined by other test files in this package.
func roundTripTransportMisc[T any](t *testing.T, name string, in T) {
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

// TestSERVER_TRANSPORT_INFO_2_RoundTrip exercises a struct that carries a
// [unique] pointer to a conformant byte array (size_is) alongside several
// [unique] string pointers.
func TestSERVER_TRANSPORT_INFO_2_RoundTrip(t *testing.T) {
	in := SERVER_TRANSPORT_INFO_2{
		Svti2Numberofvcs:            7,
		Svti2Transportname:          "\\Device\\NetbiosSmb",
		Svti2Transportaddress:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		Svti2Transportaddresslength: 5,
		Svti2Networkaddress:         "HOST",
		Svti2Domain:                 "WORKGROUP",
		Svti2Flags:                  0x00000001,
	}
	roundTripTransportMisc(t, "SERVER_TRANSPORT_INFO_2", in)
}

// TestTIME_OF_DAY_INFO_RoundTrip exercises a flat all-scalar struct including
// the signed TodTimezone field.
func TestTIME_OF_DAY_INFO_RoundTrip(t *testing.T) {
	in := TIME_OF_DAY_INFO{
		TodElapsedt:  1700000000,
		TodMsecs:     123,
		TodHours:     14,
		TodMins:      30,
		TodSecs:      45,
		TodHunds:     50,
		TodTimezone:  -120,
		TodTinterval: 310,
		TodDay:       4,
		TodMonth:     6,
		TodYear:      2026,
		TodWeekday:   4,
	}
	roundTripTransportMisc(t, "TIME_OF_DAY_INFO", in)
}
