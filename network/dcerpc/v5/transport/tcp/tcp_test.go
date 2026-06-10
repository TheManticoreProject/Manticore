package tcp

import (
	"bytes"
	"net"
	"testing"
	"time"

	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// pipeTransport returns a TCP transport whose dialer hands back one end of an in-memory
// net.Pipe, together with the other (peer) end the test drives, and a counter of dial
// calls. net.Pipe is synchronous, so reads/writes on either end are exercised from a
// goroutine.
func pipeTransport(t *testing.T) (tr *TCPTransport, peer net.Conn, dials *int) {
	t.Helper()
	clientEnd, peerEnd := net.Pipe()
	n := 0
	tr = New("198.51.100.7", 49152)
	tr.dial = func(address string, timeout time.Duration) (socket, error) {
		n++
		return clientEnd, nil
	}
	return tr, peerEnd, &n
}

func TestTCPTransport_ImplementsInterface(t *testing.T) {
	var _ dcerpctransport.Transport = New("127.0.0.1", 135)
}

func TestTCPTransport_Defaults(t *testing.T) {
	tr := New("127.0.0.1", 135)
	if tr.MaxXmitFrag() != DefaultMaxXmitFrag {
		t.Errorf("MaxXmitFrag = %d, want %d", tr.MaxXmitFrag(), DefaultMaxXmitFrag)
	}
	if tr.MaxRecvFrag() != DefaultMaxRecvFrag {
		t.Errorf("MaxRecvFrag = %d, want %d", tr.MaxRecvFrag(), DefaultMaxRecvFrag)
	}
	if tr.address != "127.0.0.1:135" {
		t.Errorf("address = %q, want 127.0.0.1:135", tr.address)
	}
}

func TestTCPTransport_AddressIPv6(t *testing.T) {
	tr := New("fe80::1", 135)
	if tr.address != "[fe80::1]:135" {
		t.Errorf("address = %q, want [fe80::1]:135", tr.address)
	}
}

func TestTCPTransport_SetMaxFrag(t *testing.T) {
	tr := New("127.0.0.1", 135)
	tr.SetMaxFrag(1024, 2048)
	if tr.MaxXmitFrag() != 1024 || tr.MaxRecvFrag() != 2048 {
		t.Errorf("frag sizes = %d/%d, want 1024/2048", tr.MaxXmitFrag(), tr.MaxRecvFrag())
	}
}

func TestTCPTransport_ConnectIdempotent(t *testing.T) {
	tr, peer, dials := pipeTransport(t)
	defer peer.Close()
	defer tr.Close()

	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := tr.Connect(); err != nil {
		t.Fatalf("second Connect() error = %v", err)
	}
	if *dials != 1 {
		t.Errorf("dialed %d times, want 1 (Connect must be idempotent)", *dials)
	}
}

func TestTCPTransport_SendWritesWholePDU(t *testing.T) {
	tr, peer, _ := pipeTransport(t)
	defer peer.Close()
	defer tr.Close()
	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	pdu := []byte{0x05, 0x00, 0x0b, 0x03, 0xde, 0xad, 0xbe, 0xef}
	got := make(chan []byte, 1)
	go func() {
		b := make([]byte, len(pdu))
		_, _ = peer.Read(b)
		got <- b
	}()

	if err := tr.Send(pdu); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if received := <-got; !bytes.Equal(received, pdu) {
		t.Errorf("peer received %x, want %x", received, pdu)
	}
}

func TestTCPTransport_RecvReturnsPeerBytes(t *testing.T) {
	tr, peer, _ := pipeTransport(t)
	defer peer.Close()
	defer tr.Close()
	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	payload := []byte{0x05, 0x00, 0x02, 0x03, 0x10, 0x00, 0x00, 0x00}
	go func() { _, _ = peer.Write(payload) }()

	got, err := tr.Recv()
	if err != nil {
		t.Fatalf("Recv() error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("Recv() = %x, want %x", got, payload)
	}
}

func TestTCPTransport_RecvErrorsOnClose(t *testing.T) {
	tr, peer, _ := pipeTransport(t)
	defer tr.Close()
	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	peer.Close()
	if _, err := tr.Recv(); err == nil {
		t.Fatal("Recv() after peer close: error = nil, want non-nil")
	}
}

func TestTCPTransport_SendBeforeConnect(t *testing.T) {
	tr := New("127.0.0.1", 135)
	if err := tr.Send([]byte{1, 2, 3}); err == nil {
		t.Fatal("Send() before Connect: error = nil, want non-nil")
	}
}

func TestTCPTransport_SendEmptyPDU(t *testing.T) {
	tr, peer, _ := pipeTransport(t)
	defer peer.Close()
	defer tr.Close()
	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := tr.Send(nil); err == nil {
		t.Fatal("Send(nil): error = nil, want non-nil")
	}
}

func TestTCPTransport_RecvBeforeConnect(t *testing.T) {
	tr := New("127.0.0.1", 135)
	if _, err := tr.Recv(); err == nil {
		t.Fatal("Recv() before Connect: error = nil, want non-nil")
	}
}

func TestTCPTransport_CloseIdempotent(t *testing.T) {
	tr, peer, _ := pipeTransport(t)
	defer peer.Close()
	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := tr.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := tr.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if tr.conn != nil {
		t.Error("conn not cleared after Close")
	}
}
