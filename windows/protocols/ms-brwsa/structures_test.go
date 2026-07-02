package msbrwsa

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestServerInfo100RoundTrip exercises the element type: a scalar platform id plus a
// [unique] wide-string name.
func TestServerInfo100RoundTrip(t *testing.T) {
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		in := SERVER_INFO_100{Sv100PlatformId: 500, Sv100Name: ndr.WSTR("CONTOSO")}
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s marshal: %v", s, err)
		}
		var out SERVER_INFO_100
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s unmarshal: %v", s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s: SERVER_INFO_100 round-trip got %+v want %+v", s, out, in)
		}
	}
}

// TestServerInfo100ContainerRoundTrip exercises the [unique] pointer to a
// [size_is(EntriesRead)] conformant array of SERVER_INFO_100 (elements carrying their own
// [unique] string pointers).
func TestServerInfo100ContainerRoundTrip(t *testing.T) {
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		in := SERVER_INFO_100_CONTAINER{
			EntriesRead: 2,
			Buffer: []SERVER_INFO_100{
				{Sv100PlatformId: 500, Sv100Name: ndr.WSTR("CONTOSO")},
				{Sv100PlatformId: 500, Sv100Name: ndr.WSTR("FABRIKAM")},
			},
		}
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s marshal: %v", s, err)
		}
		var out SERVER_INFO_100_CONTAINER
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s unmarshal: %v", s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s: SERVER_INFO_100_CONTAINER round-trip got %+v want %+v", s, out, in)
		}
	}
}

// TestServerEnumStructCase100RoundTrip exercises the switch_is(Level) union with the
// case-100 arm populated: a [unique] pointer to a container holding one entry.
func TestServerEnumStructCase100RoundTrip(t *testing.T) {
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		in := SERVER_ENUM_STRUCT{
			Level: 100,
			ServerInfo: SERVER_ENUM_UNION{
				Tag: 100,
				Level100: &SERVER_INFO_100_CONTAINER{
					EntriesRead: 1,
					Buffer:      []SERVER_INFO_100{{Sv100PlatformId: 500, Sv100Name: ndr.WSTR("CORP")}},
				},
			},
		}
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s marshal: %v", s, err)
		}
		var out SERVER_ENUM_STRUCT
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s unmarshal: %v", s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s: SERVER_ENUM_STRUCT(case=100) round-trip got %+v want %+v", s, out, in)
		}
	}
}

// TestServerEnumStructCase100NilBuffer exercises the case-100 arm with a NULL container
// pointer (the request shape a client sends: Level=100, no data yet).
func TestServerEnumStructCase100NilBuffer(t *testing.T) {
	in := SERVER_ENUM_STRUCT{Level: 100, ServerInfo: SERVER_ENUM_UNION{Tag: 100, Level100: nil}}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out SERVER_ENUM_STRUCT
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("SERVER_ENUM_STRUCT(nil buffer) round-trip got %+v want %+v", out, in)
	}
}
