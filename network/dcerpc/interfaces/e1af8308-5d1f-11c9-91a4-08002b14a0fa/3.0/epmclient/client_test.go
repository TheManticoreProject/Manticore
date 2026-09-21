package epmclient_test

import (
	"errors"
	"net"
	"testing"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/epmclient"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msrpce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpce"
)

// --- test helpers ---

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

type fakeDialer struct{ ft *fakeTransport }

func (d *fakeDialer) RPCTransport(string) (dcerpctransport.Transport, error) { return d.ft, nil }

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

func le32(b []byte, v uint32) []byte {
	return append(b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func pad4(b []byte) []byte {
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}

func sampleIface() guid.GUID {
	return guid.GUID{A: 0xc681d488, B: 0xd850, C: 0x11d0, D: 0x8c52, E: 0x00c04fd90f7e}
}

func connectedEPM(t *testing.T, ft *fakeTransport) *epmclient.EndpointMapper {
	t.Helper()
	ft.queue(bindAck(t))
	e := epmclient.New(&fakeDialer{ft: ft})
	if err := e.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if !e.IsConnected() {
		t.Fatal("IsConnected = false after Connect")
	}
	return e
}

func eptLookupStub(t *testing.T, handle msrpce.ContextHandle, status uint32, entries []msrpce.EptEntry) []byte {
	t.Helper()
	var b []byte
	b = append(b, handle[:]...)
	b = le32(b, uint32(len(entries)))
	b = le32(b, functions.DefaultMaxEnts)
	b = le32(b, 0)
	b = le32(b, uint32(len(entries)))

	refid := uint32(0x00020000)
	for _, e := range entries {
		b = append(b, e.Object.Octets[:]...)
		if e.Tower != nil {
			b = le32(b, refid)
			refid += 4
		} else {
			b = le32(b, 0)
		}
		ann := string(e.Annotation)
		b = le32(b, 0)
		b = le32(b, uint32(len(ann)+1))
		b = append(b, []byte(ann)...)
		b = append(b, 0)
		b = pad4(b)
	}

	for _, e := range entries {
		if e.Tower == nil {
			continue
		}
		tw := e.Tower.TowerOctetString
		b = le32(b, uint32(len(tw)))
		b = le32(b, uint32(len(tw)))
		b = append(b, tw...)
		b = pad4(b)
	}

	b = le32(b, status)
	return b
}

func tcpEntry(port uint16, annotation string) msrpce.EptEntry {
	tw := msrpce.NewTwr(msrpce.Tower{Floors: []msrpce.Floor{
		msrpce.InterfaceFloor(sampleIface(), 1, 0),
		msrpce.TransferSyntaxFloor(),
		{LHS: []byte{msrpce.FloorProtoNCACN}, RHS: []byte{0, 0}},
		msrpce.TCPFloor(port),
		msrpce.IPFloor(net.IPv4(10, 0, 0, 1)),
	}})
	return msrpce.EptEntry{
		Object:     msrpce.NewEptUUID(sampleIface()),
		Tower:      &tw,
		Annotation: msrpce.Annotation(annotation),
	}
}

func eptMapResponseStub(tower msrpce.Tower, maxTowers uint32) []byte {
	towerBytes := tower.Marshal()
	var b []byte
	b = append(b, make([]byte, msrpce.ContextHandleSize)...)
	b = le32(b, 1)
	b = le32(b, maxTowers)
	b = le32(b, 0)
	b = le32(b, 1)
	b = le32(b, 0x00020004)
	b = le32(b, uint32(len(towerBytes)))
	b = le32(b, uint32(len(towerBytes)))
	b = append(b, towerBytes...)
	b = pad4(b)
	b = le32(b, epm.EptStatusSuccess)
	return b
}

// --- tests ---

func TestMethodsBeforeConnect(t *testing.T) {
	e := epmclient.New(&fakeDialer{ft: &fakeTransport{}})
	if _, err := e.Lookup(); !errors.Is(err, epmclient.ErrNotConnected) {
		t.Errorf("Lookup before Connect: err = %v, want ErrNotConnected", err)
	}
	if _, err := e.Map(sampleIface(), 1, 0); !errors.Is(err, epmclient.ErrNotConnected) {
		t.Errorf("Map before Connect: err = %v, want ErrNotConnected", err)
	}
	if e.IsConnected() {
		t.Error("IsConnected = true before Connect")
	}
}

func TestConnectIdempotent(t *testing.T) {
	ft := &fakeTransport{}
	e := connectedEPM(t, ft)
	if err := e.Connect(); err != nil {
		t.Errorf("second Connect: %v", err)
	}
}

func TestLookup(t *testing.T) {
	ft := &fakeTransport{}
	e := connectedEPM(t, ft)

	entries := []msrpce.EptEntry{
		tcpEntry(49664, "Service A"),
		tcpEntry(135, "Service B"),
	}
	ft.queue(responsePDU(t, 2, eptLookupStub(t, msrpce.ContextHandle{}, epm.EptStatusSuccess, entries)))

	got, err := e.Lookup()
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	if got[0].Annotation != "Service A" {
		t.Errorf("entry 0 annotation = %q, want %q", got[0].Annotation, "Service A")
	}
}

func TestMap(t *testing.T) {
	ft := &fakeTransport{}
	e := connectedEPM(t, ft)

	resolved := msrpce.Tower{Floors: []msrpce.Floor{
		msrpce.InterfaceFloor(sampleIface(), 1, 0),
		msrpce.TransferSyntaxFloor(),
		{LHS: []byte{msrpce.FloorProtoNCACN}, RHS: []byte{0, 0}},
		msrpce.TCPFloor(49664),
		msrpce.IPFloor(net.IPv4(10, 0, 0, 30)),
	}}
	ft.queue(responsePDU(t, 2, eptMapResponseStub(resolved, functions.DefaultMaxTowers)))

	eps, err := e.Map(sampleIface(), 1, 0)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(eps) != 1 || eps[0].Port != 49664 {
		t.Fatalf("Map = %v, want one endpoint on port 49664", eps)
	}
}

func TestCloseTearsDown(t *testing.T) {
	ft := &fakeTransport{}
	e := connectedEPM(t, ft)
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if e.IsConnected() {
		t.Error("IsConnected = true after Close")
	}
	if err := e.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestCloseBeforeConnect(t *testing.T) {
	e := epmclient.New(&fakeDialer{ft: &fakeTransport{}})
	if err := e.Close(); err != nil {
		t.Errorf("Close before Connect: %v", err)
	}
}

func TestInterface(t *testing.T) {
	e := epmclient.New(&fakeDialer{ft: &fakeTransport{}})
	if got := e.Interface(); got != epm.SyntaxID() {
		t.Errorf("Interface = %v, want %v", got, epm.SyntaxID())
	}
}
