package functions_test

import (
	"encoding/binary"
	"errors"
	"net"
	"testing"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// fakeTransport is an in-memory transport.Transport for driving the client without a
// network.
type fakeTransport struct {
	sent      [][]byte
	recvQueue [][]byte
}

func (f *fakeTransport) Connect() error { return nil }
func (f *fakeTransport) Send(p []byte) error {
	f.sent = append(f.sent, append([]byte(nil), p...))
	return nil
}
func (f *fakeTransport) Recv() ([]byte, error) {
	if len(f.recvQueue) == 0 {
		return nil, errors.New("recv queue empty")
	}
	c := f.recvQueue[0]
	f.recvQueue = f.recvQueue[1:]
	return c, nil
}
func (f *fakeTransport) Close() error        { return nil }
func (f *fakeTransport) MaxXmitFrag() uint16 { return 5840 }
func (f *fakeTransport) MaxRecvFrag() uint16 { return 5840 }
func (f *fakeTransport) queue(b []byte)      { f.recvQueue = append(f.recvQueue, b) }

func bindAck(t *testing.T) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag: 5840,
		MaxRecvFrag: 5840,
		Results:     []pdu.PresentationResult{{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()}},
	}
	b, err := ack.Marshal()
	if err != nil {
		t.Fatalf("bind_ack marshal: %v", err)
	}
	return b
}

func responsePDU(t *testing.T, callID uint32, stub []byte) []byte {
	t.Helper()
	resp := &pdu.Response{Stub: stub}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID)
	b, err := resp.Marshal()
	if err != nil {
		t.Fatalf("response marshal: %v", err)
	}
	return b
}

func boundClient(t *testing.T, ft *fakeTransport) *client.Client {
	t.Helper()
	ft.queue(bindAck(t))
	c := client.NewClient(ft)
	if err := c.Bind(epm.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return c
}

func sampleIface() guid.GUID {
	return guid.GUID{A: 0xc681d488, B: 0xd850, C: 0x11d0, D: 0x8c52, E: 0x00c04fd90f7e}
}

// le32 appends v as a little-endian uint32.
func le32(b []byte, v uint32) []byte { return binary.LittleEndian.AppendUint32(b, v) }

// pad4 appends zero octets until len(b) is a multiple of 4.
func pad4(b []byte) []byte {
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}

// eptMapResponseStub assembles the [out] NDR stub of an ept_map call that returns a
// single tower resolving to the given TCP endpoint. The layout is: 20-octet context
// handle, num_towers, the ITowers full pointer, the conformant-varying array header,
// one element referent, the twr_t body, then the status.
func eptMapResponseStub(tower structures.Tower, maxTowers uint32) []byte {
	towerBytes := tower.Marshal()

	var b []byte
	b = append(b, make([]byte, structures.ContextHandleSize)...) // entry_handle
	b = le32(b, 1)                                               // num_towers
	b = le32(b, 0x00020000)                                      // ITowers referent id (non-null)
	b = le32(b, maxTowers)                                       // array maximum_count (size_is)
	b = le32(b, 0)                                               // array offset
	b = le32(b, 1)                                               // array actual_count (length_is)
	b = le32(b, 0x00020004)                                      // element[0] referent id (non-null)
	// twr_t body: hoisted maximum_count, tower_length, octet string, pad to 4.
	b = le32(b, uint32(len(towerBytes)))
	b = le32(b, uint32(len(towerBytes)))
	b = append(b, towerBytes...)
	b = pad4(b)
	b = le32(b, epm.StatusSuccess) // status
	return b
}

func TestEptMap_RoundTrip(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	resolved := structures.Tower{Floors: []structures.Floor{
		structures.InterfaceFloor(sampleIface(), 1, 0),
		structures.TransferSyntaxFloor(),
		{LHS: []byte{structures.FloorProtoNCACN}, RHS: []byte{0, 0}},
		structures.TCPFloor(49664),
		structures.IPFloor(net.IPv4(10, 0, 0, 30)),
	}}
	ft.queue(responsePDU(t, 2, eptMapResponseStub(resolved, functions.DefaultMaxTowers)))

	towers, err := functions.EptMap(c, nil, structures.BuildMapTowerTCP(sampleIface(), 1, 0), functions.DefaultMaxTowers)
	if err != nil {
		t.Fatalf("EptMap() error = %v", err)
	}
	if len(towers) != 1 {
		t.Fatalf("got %d towers, want 1", len(towers))
	}
	ep, ok := towers[0].Endpoint()
	if !ok {
		t.Fatal("returned tower has no endpoint")
	}
	if ep.Port != 49664 {
		t.Errorf("port = %d, want 49664", ep.Port)
	}

	// Verify the request: opnum 3, null object pointer, then the map tower.
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != epm.OpnumEptMap {
		t.Errorf("opnum = %d, want %d", req.Opnum, epm.OpnumEptMap)
	}
	if got := binary.LittleEndian.Uint32(req.Stub[:4]); got != 0 {
		t.Errorf("object pointer = 0x%08x, want 0 (null)", got)
	}
}

func TestMap_ReturnsEndpoints(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	resolved := structures.Tower{Floors: []structures.Floor{
		structures.InterfaceFloor(sampleIface(), 1, 0),
		structures.TransferSyntaxFloor(),
		{LHS: []byte{structures.FloorProtoNCACN}, RHS: []byte{0, 0}},
		structures.TCPFloor(135),
		structures.IPFloor(net.IPv4(192, 0, 2, 1)),
	}}
	ft.queue(responsePDU(t, 2, eptMapResponseStub(resolved, functions.DefaultMaxTowers)))

	eps, err := functions.Map(c, sampleIface(), 1, 0)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if len(eps) != 1 || eps[0].Port != 135 {
		t.Fatalf("Map() = %v, want one endpoint on port 135", eps)
	}
}

func TestEptMap_StatusError(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	// num_towers = 0, null ITowers pointer, non-zero status.
	var stub []byte
	stub = append(stub, make([]byte, structures.ContextHandleSize)...)
	stub = le32(stub, 0)                          // num_towers
	stub = le32(stub, 0)                          // ITowers null pointer
	stub = le32(stub, epm.EptStatusNotRegistered) // status
	ft.queue(responsePDU(t, 2, stub))

	if _, err := functions.EptMap(c, nil, structures.BuildMapTowerTCP(sampleIface(), 1, 0), functions.DefaultMaxTowers); err == nil {
		t.Fatal("EptMap() with ept_s_not_registered: error = nil, want non-nil")
	}
}
