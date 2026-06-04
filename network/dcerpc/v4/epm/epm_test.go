package epm

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// sampleIface is an arbitrary interface UUID used in the request tests.
var sampleIface = guid.GUID{A: 0x11223344, B: 0x5566, C: 0x7788, D: 0x99aa, E: 0xbbccddeeff00}

func TestMarshalEptMapRequestStructure(t *testing.T) {
	tower := BuildMapTower(sampleIface, 1, 0)
	towerBytes := tower.Marshal()
	stub := marshalEptMapRequest(nil, tower, DefaultMaxTowers)

	off := 0
	u32 := func(name string) uint32 {
		if off+4 > len(stub) {
			t.Fatalf("stub too short reading %s", name)
		}
		v := binary.LittleEndian.Uint32(stub[off:])
		off += 4
		return v
	}

	if v := u32("object referent"); v != 0 {
		t.Errorf("object referent = 0x%08x, want 0 (null pointer)", v)
	}
	if v := u32("map_tower referent"); v != refMapTower {
		t.Errorf("map_tower referent = 0x%08x, want 0x%08x", v, refMapTower)
	}
	if v := u32("max_count"); v != uint32(len(towerBytes)) {
		t.Errorf("conformant max_count = %d, want %d", v, len(towerBytes))
	}
	if v := u32("tower_length"); v != uint32(len(towerBytes)) {
		t.Errorf("tower_length = %d, want %d", v, len(towerBytes))
	}
	// The tower octet string follows.
	if off+len(towerBytes) > len(stub) {
		t.Fatal("stub too short for tower bytes")
	}
	off += len(towerBytes)
	for off%4 != 0 { // skip alignment padding
		off++
	}
	// entry_handle: 20 zero octets.
	for i := 0; i < contextHandleSize; i++ {
		if stub[off+i] != 0 {
			t.Fatalf("entry_handle octet %d = 0x%02x, want 0", i, stub[off+i])
		}
	}
	off += contextHandleSize
	if v := u32("max_towers"); v != DefaultMaxTowers {
		t.Errorf("max_towers = %d, want %d", v, DefaultMaxTowers)
	}
	if off != len(stub) {
		t.Errorf("trailing %d bytes after max_towers", len(stub)-off)
	}
}

func TestMarshalEptMapRequestWithObject(t *testing.T) {
	obj := guid.GUID{A: 0xcafef00d}
	stub := marshalEptMapRequest(&obj, BuildMapTower(sampleIface, 1, 0), 1)
	if binary.LittleEndian.Uint32(stub[:4]) != refObject {
		t.Fatalf("object referent = 0x%08x, want 0x%08x", binary.LittleEndian.Uint32(stub[:4]), refObject)
	}
	// The 16-byte object UUID must follow the referent id.
	gotUUID := stub[4:20]
	if string(gotUUID) != string(obj.ToBytes()) {
		t.Fatalf("object UUID bytes = % x, want % x", gotUUID, obj.ToBytes())
	}
}

// buildEptMapResponse builds a spec-shaped ept_map response stub carrying the given
// towers and status, so the parser can be exercised against the documented layout.
func buildEptMapResponse(towers []Tower, status uint32) []byte {
	w := &ndrWriter{}
	w.bytes(make([]byte, contextHandleSize)) // entry_handle
	w.u32(uint32(len(towers)))               // num_towers
	w.u32(uint32(len(towers)))               // array maximum_count
	w.u32(0)                                 // array offset
	w.u32(uint32(len(towers)))               // array actual_count
	for i := range towers {
		w.u32(uint32(0x00030000 + i)) // non-null referent id per element
	}
	for _, tw := range towers {
		tb := tw.Marshal()
		w.u32(uint32(len(tb))) // hoisted conformant maximum_count
		w.u32(uint32(len(tb))) // tower_length
		w.bytes(tb)
		w.align(4)
	}
	w.u32(status)
	return w.buf
}

func fullTower(port uint16, ip net.IP) Tower {
	return Tower{Floors: []Floor{
		InterfaceFloor(sampleIface, 1, 0),
		TransferSyntaxFloor(),
		protocolFloor(FloorProtoNCADG, 0),
		UDPFloor(port),
		IPFloor(ip),
	}}
}

func TestParseEptMapResponseRoundTrip(t *testing.T) {
	towers := []Tower{
		fullTower(49152, net.IPv4(10, 0, 0, 5)),
		fullTower(1025, net.IPv4(192, 168, 1, 1)),
	}
	got, status, err := parseEptMapResponse(buildEptMapResponse(towers, 0))
	if err != nil {
		t.Fatalf("parseEptMapResponse: %v", err)
	}
	if status != 0 {
		t.Fatalf("status = 0x%08x, want 0", status)
	}
	if len(got) != 2 {
		t.Fatalf("got %d towers, want 2", len(got))
	}
	ep0, ok := got[0].Endpoint()
	if !ok || ep0.Port != 49152 || !ep0.IP.Equal(net.IPv4(10, 0, 0, 5)) {
		t.Errorf("tower 0 endpoint = %v ok=%v", ep0, ok)
	}
	ep1, _ := got[1].Endpoint()
	if ep1.Port != 1025 || !ep1.IP.Equal(net.IPv4(192, 168, 1, 1)) {
		t.Errorf("tower 1 endpoint = %v", ep1)
	}
}

func TestParseEptMapResponseRejectsHugeCount(t *testing.T) {
	w := &ndrWriter{}
	w.bytes(make([]byte, contextHandleSize))
	w.u32(0)          // num_towers
	w.u32(0xffffffff) // maximum_count
	w.u32(0)          // offset
	w.u32(0xffffffff) // actual_count -> implausible
	if _, _, err := parseEptMapResponse(w.buf); err == nil {
		t.Fatal("expected error on implausible tower count")
	}
}

func TestParseEptMapResponseTruncated(t *testing.T) {
	if _, _, err := parseEptMapResponse([]byte{0x00, 0x01, 0x02}); err == nil {
		t.Fatal("expected error on truncated response")
	}
}

// --- end-to-end Map() over a scripted mock transport ---

var testActivity = guid.GUID{A: 0xa1b2c3d4, B: 0x1111, C: 0x2222, D: 0x3333, E: 0x444455556666}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

type mockConn struct {
	events [][]byte
}

var _ transport.Transport = (*mockConn)(nil)

func (m *mockConn) Connect() error           { return nil }
func (m *mockConn) Send([]byte) (int, error) { return 0, nil }
func (m *mockConn) Recv() ([]byte, error) {
	if len(m.events) == 0 {
		return nil, timeoutError{}
	}
	e := m.events[0]
	m.events = m.events[1:]
	return e, nil
}
func (m *mockConn) SetDeadline(time.Time) error { return nil }
func (m *mockConn) MaxPDUSize() int             { return transport.MaxPDUSizeDefault }
func (m *mockConn) RemoteAddr() net.Addr        { return nil }
func (m *mockConn) IsConnected() bool           { return true }
func (m *mockConn) Close() error                { return nil }

// responsePDU wraps an ept_map response stub in a connectionless response PDU for the
// test conversation.
func responsePDU(t *testing.T, stub []byte) []byte {
	t.Helper()
	h := pdu.NewHeader(pdu.PacketTypeResponse)
	h.ActivityID = testActivity
	h.SequenceNumber = 0
	h.ServerBoot = 0x11112222
	raw, err := (&pdu.PDU{Header: h, Body: stub}).Marshal()
	if err != nil {
		t.Fatalf("build response PDU: %v", err)
	}
	return raw
}

func TestMapEndToEnd(t *testing.T) {
	stub := buildEptMapResponse([]Tower{fullTower(49152, net.IPv4(10, 0, 0, 5))}, 0)
	conn := &mockConn{events: [][]byte{responsePDU(t, stub)}}
	rpc := client.New(conn, client.WithActivityID(testActivity))

	eps, err := New(rpc).Map(sampleIface, 1, 0)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(eps) != 1 {
		t.Fatalf("got %d endpoints, want 1", len(eps))
	}
	if eps[0].Port != 49152 || !eps[0].IP.Equal(net.IPv4(10, 0, 0, 5)) {
		t.Fatalf("endpoint = %s, want 10.0.0.5:49152", eps[0])
	}
}

func TestMapReturnsErrorOnStatus(t *testing.T) {
	stub := buildEptMapResponse(nil, 0x16c9a0d6) // EPT_S_NOT_REGISTERED-style status
	conn := &mockConn{events: [][]byte{responsePDU(t, stub)}}
	rpc := client.New(conn, client.WithActivityID(testActivity))

	if _, err := New(rpc).Map(sampleIface, 1, 0); err == nil {
		t.Fatal("expected an error when ept_map returns a non-zero status")
	}
}
