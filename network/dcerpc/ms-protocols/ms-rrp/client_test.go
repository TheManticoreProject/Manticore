package ms_rrp_test

import (
	"encoding/binary"
	"errors"
	"testing"

	ms_rrp "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-rrp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// fakeTransport is an in-memory dcerpc transport.Transport for driving RemoteRegistry
// without a network.
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

// fakeDialer hands the same fakeTransport to RemoteRegistry.Connect.
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

func le32(b []byte, v uint32) []byte { return binary.LittleEndian.AppendUint32(b, v) }

// connected returns a RemoteRegistry whose Connect has bound over ft (with a queued
// bind_ack), ready for one or more queued method responses.
func connected(t *testing.T, ft *fakeTransport) *ms_rrp.RemoteRegistry {
	t.Helper()
	ft.queue(bindAck(t))
	r := ms_rrp.New(&fakeDialer{ft: ft})
	if err := r.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if !r.IsConnected() {
		t.Fatal("IsConnected = false after Connect")
	}
	return r
}

// TestMethodsBeforeConnect verifies the not-connected guard fires instead of panicking on
// a nil association.
func TestMethodsBeforeConnect(t *testing.T) {
	r := ms_rrp.New(&fakeDialer{ft: &fakeTransport{}})
	if _, err := r.OpenLocalMachine(nil, 0); !errors.Is(err, ms_rrp.ErrNotConnected) {
		t.Errorf("OpenLocalMachine before Connect: err = %v, want ErrNotConnected", err)
	}
	if _, err := r.BaseRegGetVersion(ms_rrp.Handle{}); !errors.Is(err, ms_rrp.ErrNotConnected) {
		t.Errorf("BaseRegGetVersion before Connect: err = %v, want ErrNotConnected", err)
	}
	if r.IsConnected() {
		t.Error("IsConnected = true before Connect")
	}
}

// TestOpenLocalMachine drives the opnum-2 round-trip: a 20-byte context handle plus a
// success status.
func TestOpenLocalMachine(t *testing.T) {
	ft := &fakeTransport{}
	r := connected(t, ft)

	var stub []byte
	handle := make([]byte, 20)
	for i := range handle {
		handle[i] = byte(i + 1)
	}
	stub = append(stub, handle...) // PhKey
	stub = le32(stub, 0)           // status = ERROR_SUCCESS
	ft.queue(responsePDU(t, 2, stub))

	h, err := r.OpenLocalMachine(nil, 0x00020019)
	if err != nil {
		t.Fatalf("OpenLocalMachine: %v", err)
	}
	if h.IsZero() {
		t.Error("returned handle is zero")
	}
}

// TestBaseRegGetVersion drives the opnum-26 round-trip: a [out] DWORD then status.
func TestBaseRegGetVersion(t *testing.T) {
	ft := &fakeTransport{}
	r := connected(t, ft)

	var stub []byte
	stub = le32(stub, 6) // LpdwVersion
	stub = le32(stub, 0) // status
	ft.queue(responsePDU(t, 2, stub))

	ver, err := r.BaseRegGetVersion(ms_rrp.Handle{})
	if err != nil {
		t.Fatalf("BaseRegGetVersion: %v", err)
	}
	if uint32(ver) != 6 {
		t.Errorf("version = %d, want 6", uint32(ver))
	}
}

// TestStatusErrorSurfaced confirms a non-success Win32 status becomes a Go error.
func TestStatusErrorSurfaced(t *testing.T) {
	ft := &fakeTransport{}
	r := connected(t, ft)

	var stub []byte
	stub = append(stub, make([]byte, 20)...) // PhKey (zero)
	stub = le32(stub, 5)                     // status = ERROR_ACCESS_DENIED
	ft.queue(responsePDU(t, 2, stub))

	if _, err := r.OpenLocalMachine(nil, 0); err == nil {
		t.Fatal("OpenLocalMachine with ERROR_ACCESS_DENIED: err = nil, want non-nil")
	}
}

// TestCloseTearsDown verifies Close flips IsConnected and is safe to repeat.
func TestCloseTearsDown(t *testing.T) {
	ft := &fakeTransport{}
	r := connected(t, ft)
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if r.IsConnected() {
		t.Error("IsConnected = true after Close")
	}
	if err := r.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestRegistryValueCodecs(t *testing.T) {
	if got := ms_rrp.StringValue("Hello").String(); got != "Hello" {
		t.Errorf("StringValue/String round-trip = %q", got)
	}
	if got := ms_rrp.ExpandStringValue("%PATH%").String(); got != "%PATH%" {
		t.Errorf("ExpandStringValue/String round-trip = %q", got)
	}
	if v, ok := ms_rrp.DwordValue(0xDEADBEEF).Uint32(); !ok || v != 0xDEADBEEF {
		t.Errorf("DwordValue/Uint32 round-trip = %#x, ok=%v", v, ok)
	}
	if v, ok := ms_rrp.QwordValue(0x1122334455667788).Uint64(); !ok || v != 0x1122334455667788 {
		t.Errorf("QwordValue/Uint64 round-trip = %#x, ok=%v", v, ok)
	}
	multi := ms_rrp.MultiStringValue([]string{"a", "bb", "ccc"}).MultiString()
	if len(multi) != 3 || multi[0] != "a" || multi[1] != "bb" || multi[2] != "ccc" {
		t.Errorf("MultiStringValue/MultiString round-trip = %v", multi)
	}
	if v, ok := ms_rrp.DwordValue(0).Uint64(); ok {
		t.Errorf("Uint64 on a 4-byte REG_DWORD should report ok=false, got %#x", v)
	}
	raw := []byte{0xAA, 0xBB, 0xCC}
	if bv := ms_rrp.BinaryValue(raw); bv.Type != 3 || len(bv.Data) != 3 || bv.Data[0] != 0xAA {
		t.Errorf("BinaryValue round-trip = %+v", bv)
	}
	if got := ms_rrp.DwordValue(1).String(); got != "" {
		t.Errorf("String on a REG_DWORD should be empty, got %q", got)
	}
}
