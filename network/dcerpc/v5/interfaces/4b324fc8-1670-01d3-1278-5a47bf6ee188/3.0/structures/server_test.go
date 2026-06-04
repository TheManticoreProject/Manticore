package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

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
