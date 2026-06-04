package epm

import (
	"bytes"
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

func TestFloorMarshalRoundTrip(t *testing.T) {
	f := Floor{LHS: []byte{FloorProtoUDP}, RHS: []byte{0x00, 0x87}}
	raw := f.Marshal()
	got, n, err := unmarshalFloor(raw)
	if err != nil {
		t.Fatalf("unmarshalFloor: %v", err)
	}
	if n != len(raw) {
		t.Fatalf("consumed %d, want %d", n, len(raw))
	}
	if !bytes.Equal(got.LHS, f.LHS) || !bytes.Equal(got.RHS, f.RHS) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, f)
	}
}

func TestTowerGoldenBytes(t *testing.T) {
	// One UDP floor for port 135 (0x0087, big-endian in the RHS).
	tw := Tower{Floors: []Floor{UDPFloor(135)}}
	want := []byte{
		0x01, 0x00, // floor count = 1
		0x01, 0x00, // LHS length = 1
		0x08,       // LHS = UDP protocol id
		0x02, 0x00, // RHS length = 2
		0x00, 0x87, // RHS = port 135, big-endian
	}
	if got := tw.Marshal(); !bytes.Equal(got, want) {
		t.Fatalf("tower bytes = % x, want % x", got, want)
	}
}

func TestTowerMarshalRoundTrip(t *testing.T) {
	iface := guid.GUID{A: 0x12345678, B: 0x9abc, C: 0xdef0, D: 0x1122, E: 0x334455667788}
	in := BuildMapTower(iface, 2, 1)
	out, err := UnmarshalTower(in.Marshal())
	if err != nil {
		t.Fatalf("UnmarshalTower: %v", err)
	}
	if len(out.Floors) != len(in.Floors) {
		t.Fatalf("got %d floors, want %d", len(out.Floors), len(in.Floors))
	}
	for i := range in.Floors {
		if !bytes.Equal(out.Floors[i].LHS, in.Floors[i].LHS) || !bytes.Equal(out.Floors[i].RHS, in.Floors[i].RHS) {
			t.Fatalf("floor %d mismatch", i)
		}
	}
}

func TestBuildMapTowerStructure(t *testing.T) {
	iface := guid.GUID{A: 0xdeadbeef}
	tw := BuildMapTower(iface, 1, 0)
	if len(tw.Floors) != 5 {
		t.Fatalf("ncadg_ip_udp map tower must have 5 floors, got %d", len(tw.Floors))
	}
	wantProto := []byte{FloorProtoUUID, FloorProtoUUID, FloorProtoNCADG, FloorProtoUDP, FloorProtoIP}
	for i, p := range wantProto {
		if got := tw.Floors[i].Protocol(); got != p {
			t.Errorf("floor %d protocol = 0x%02x, want 0x%02x", i, got, p)
		}
	}
	// Interface floor LHS = id byte + 16-byte UUID + 2-byte major; RHS = 2-byte minor.
	if l := len(tw.Floors[0].LHS); l != 19 {
		t.Errorf("interface floor LHS = %d bytes, want 19", l)
	}
	if l := len(tw.Floors[0].RHS); l != 2 {
		t.Errorf("interface floor RHS = %d bytes, want 2", l)
	}
	// Map tower requests a wildcard endpoint: zero port and zero address.
	if ep, ok := tw.Endpoint(); !ok || ep.Port != 0 || !ep.IP.Equal(net.IPv4zero) {
		t.Errorf("map tower endpoint = %v ok=%v, want 0.0.0.0:0", ep, ok)
	}
}

func TestTowerEndpointExtraction(t *testing.T) {
	tw := Tower{Floors: []Floor{
		InterfaceFloor(guid.GUID{A: 1}, 1, 0),
		TransferSyntaxFloor(),
		protocolFloor(FloorProtoNCADG, 0),
		UDPFloor(49152),
		IPFloor(net.IPv4(10, 0, 0, 5)),
	}}
	ep, ok := tw.Endpoint()
	if !ok {
		t.Fatal("expected endpoint extraction to succeed")
	}
	if ep.Port != 49152 {
		t.Errorf("port = %d, want 49152", ep.Port)
	}
	if !ep.IP.Equal(net.IPv4(10, 0, 0, 5)) {
		t.Errorf("ip = %s, want 10.0.0.5", ep.IP)
	}
}

func TestTowerEndpointMissingFloors(t *testing.T) {
	// Only a UDP floor, no IP floor -> not a complete endpoint.
	tw := Tower{Floors: []Floor{UDPFloor(135)}}
	if _, ok := tw.Endpoint(); ok {
		t.Fatal("expected ok=false when the IP floor is absent")
	}
}

func TestUnmarshalTowerTruncated(t *testing.T) {
	// Declares 2 floors but provides none.
	if _, err := UnmarshalTower([]byte{0x02, 0x00}); err == nil {
		t.Fatal("expected error for tower with missing floors")
	}
}
