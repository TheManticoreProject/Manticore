package structures

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

func sampleIface() guid.GUID {
	// efsrpc: c681d488-d850-11d0-8c52-00c04fd90f7e v1.0, an arbitrary lookup target.
	return guid.GUID{A: 0xc681d488, B: 0xd850, C: 0x11d0, D: 0x8c52, E: 0x00c04fd90f7e}
}

func TestTower_RoundTrip(t *testing.T) {
	in := BuildMapTowerTCP(sampleIface(), 1, 0)
	raw := in.Marshal()
	out, err := UnmarshalTower(raw)
	if err != nil {
		t.Fatalf("UnmarshalTower() error = %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("tower round-trip mismatch:\n in  %+v\n out %+v", in, out)
	}
	if len(out.Floors) != 5 {
		t.Fatalf("floor count = %d, want 5", len(out.Floors))
	}
	// Floor protocol identifiers, in order: interface, transfer syntax, ncacn, TCP, IP.
	want := []byte{FloorProtoUUID, FloorProtoUUID, FloorProtoNCACN, FloorProtoTCP, FloorProtoIP}
	for i, p := range want {
		if got := out.Floors[i].Protocol(); got != p {
			t.Errorf("floor %d protocol = 0x%02x, want 0x%02x", i, got, p)
		}
	}
}

func TestTower_EndpointTCP(t *testing.T) {
	tower := Tower{Floors: []Floor{
		InterfaceFloor(sampleIface(), 1, 0),
		TransferSyntaxFloor(),
		Floor{LHS: []byte{FloorProtoNCACN}, RHS: []byte{0, 0}},
		TCPFloor(49664),
		IPFloor(net.IPv4(10, 0, 0, 30)),
	}}
	ep, ok := tower.Endpoint()
	if !ok {
		t.Fatal("Endpoint() ok = false, want true")
	}
	if ep.Port != 49664 {
		t.Errorf("port = %d, want 49664", ep.Port)
	}
	if !ep.IP.Equal(net.IPv4(10, 0, 0, 30)) {
		t.Errorf("ip = %s, want 10.0.0.30", ep.IP)
	}
}

func TestTower_EndpointNoTCPFloor(t *testing.T) {
	tower := Tower{Floors: []Floor{InterfaceFloor(sampleIface(), 1, 0), TransferSyntaxFloor()}}
	if _, ok := tower.Endpoint(); ok {
		t.Error("Endpoint() ok = true for a tower with no TCP floor, want false")
	}
}

func TestUnmarshalTower_Truncated(t *testing.T) {
	if _, err := UnmarshalTower([]byte{0x05}); err == nil {
		t.Error("UnmarshalTower(truncated) error = nil, want non-nil")
	}
	// A floor count of 2 but bytes for none.
	if _, err := UnmarshalTower([]byte{0x02, 0x00}); err == nil {
		t.Error("UnmarshalTower(missing floors) error = nil, want non-nil")
	}
}

func TestTwr_RoundTrip(t *testing.T) {
	tower := BuildMapTowerTCP(sampleIface(), 1, 0)
	in := NewTwr(tower)

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	// Conformant struct: hoisted maximum_count, then tower_length, then the octet string.
	if len(raw) < 8 {
		t.Fatalf("marshalled twr_t too short: %d bytes", len(raw))
	}
	towerBytes := tower.Marshal()
	wantHead := make([]byte, 8)
	// maximum_count == tower_length == len(towerBytes), both little-endian.
	for i, n := 0, uint32(len(towerBytes)); i < 4; i++ {
		wantHead[i] = byte(n >> (8 * uint(i)))
		wantHead[i+4] = byte(n >> (8 * uint(i)))
	}
	if !bytes.Equal(raw[:8], wantHead) {
		t.Errorf("twr_t header:\n got %x\nwant %x", raw[:8], wantHead)
	}

	var out Twr
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !bytes.Equal(out.TowerOctetString, in.TowerOctetString) {
		t.Errorf("octet string round-trip mismatch:\n got %x\nwant %x", out.TowerOctetString, in.TowerOctetString)
	}
	got, err := out.DecodeTower()
	if err != nil {
		t.Fatalf("DecodeTower() error = %v", err)
	}
	if !reflect.DeepEqual(got, tower) {
		t.Errorf("decoded tower mismatch")
	}
}

func TestContextHandle_RoundTrip(t *testing.T) {
	var in ContextHandle
	for i := range in {
		in[i] = byte(i + 1)
	}
	raw, err := ndr.Marshal(&struct{ H ContextHandle }{in})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if len(raw) != ContextHandleSize {
		t.Fatalf("marshalled size = %d, want %d", len(raw), ContextHandleSize)
	}
	if !bytes.Equal(raw, in[:]) {
		t.Errorf("context handle bytes:\n got %x\nwant %x", raw, in[:])
	}
	var out struct{ H ContextHandle }
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if out.H != in {
		t.Errorf("context handle round-trip = %x, want %x", out.H, in)
	}
}

func TestEptUUID_RoundTrip(t *testing.T) {
	g := sampleIface()
	in := NewEptUUID(g)
	raw, err := ndr.Marshal(&struct{ U EptUUID }{in})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("marshalled size = %d, want 16", len(raw))
	}
	if !bytes.Equal(raw, g.ToBytes()) {
		t.Errorf("uuid bytes:\n got %x\nwant %x", raw, g.ToBytes())
	}
	var out struct{ U EptUUID }
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := out.U.GUID(); !got.Equal(&g) {
		t.Errorf("uuid round-trip mismatch: got %s", got.ToFormatD())
	}
}
