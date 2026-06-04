package interfaces_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/interfaces"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

var (
	testActivity = guid.GUID{A: 0x01020304, B: 0x0506, C: 0x0708, D: 0x090a, E: 0x0b0c0d0e0f10}
	testIface    = guid.GUID{A: 0x11111111, B: 0x2222, C: 0x3333, D: 0x4444, E: 0x555566667777}
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

type mockConn struct {
	sent   [][]byte
	events [][]byte
}

var _ transport.Transport = (*mockConn)(nil)

func (m *mockConn) Connect() error { return nil }
func (m *mockConn) Send(b []byte) (int, error) {
	m.sent = append(m.sent, append([]byte(nil), b...))
	return len(b), nil
}
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

func responsePDU(t *testing.T, body []byte) []byte {
	t.Helper()
	h := pdu.NewHeader(pdu.PacketTypeResponse)
	h.ActivityID = testActivity
	h.SequenceNumber = 0
	h.ServerBoot = 0x12345678
	raw, err := (&pdu.PDU{Header: h, Body: body}).Marshal()
	if err != nil {
		t.Fatalf("build response PDU: %v", err)
	}
	return raw
}

func TestBindingInvokeRoutesToInterface(t *testing.T) {
	conn := &mockConn{events: [][]byte{responsePDU(t, []byte("out"))}}
	rpc := client.New(conn, client.WithActivityID(testActivity))
	b := interfaces.NewBinding(rpc, testIface, 2, 1)

	in := []byte{0xaa, 0xbb, 0xcc}
	out, err := b.Invoke(7, in)
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if string(out) != "out" {
		t.Fatalf("response = %q, want out", out)
	}
	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 sent PDU, got %d", len(conn.sent))
	}

	var req pdu.PDU
	if _, err := req.Unmarshal(conn.sent[0]); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if req.Header.PacketType != pdu.PacketTypeRequest {
		t.Errorf("packet type = %s, want request", req.Header.PacketType)
	}
	if !req.Header.InterfaceID.Equal(&testIface) {
		t.Errorf("interface = %s, want %s", req.Header.InterfaceID.ToFormatD(), testIface.ToFormatD())
	}
	if req.Header.InterfaceVersion != 2 {
		t.Errorf("interface version = %d, want 2", req.Header.InterfaceVersion)
	}
	if req.Header.OpNum != 7 {
		t.Errorf("opnum = %d, want 7", req.Header.OpNum)
	}
	if !req.Header.Flags1.Has(pdu.Flags1Idempotent) {
		t.Errorf("request should be idempotent, flags = %s", req.Header.Flags1)
	}
	if !bytes.Equal(req.Body, in) {
		t.Errorf("request stub = % x, want % x", req.Body, in)
	}
}

func TestBindingAccessors(t *testing.T) {
	b := interfaces.NewBinding(client.New(&mockConn{}, client.WithActivityID(testActivity)), testIface, 3, 1)
	iface := b.Interface()
	if !iface.Equal(&testIface) {
		t.Errorf("Interface() = %s, want %s", iface.ToFormatD(), testIface.ToFormatD())
	}
	if maj, min := b.Version(); maj != 3 || min != 1 {
		t.Errorf("Version() = %d.%d, want 3.1", maj, min)
	}
}
