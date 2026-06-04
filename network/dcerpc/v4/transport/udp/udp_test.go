package udp_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport/udp"
)

// newServer opens a loopback UDP socket and returns it with its port. The caller
// closes the socket via t.Cleanup.
func newServer(t *testing.T) (*net.UDPConn, int) {
	t.Helper()
	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Skipf("loopback UDP unavailable in this environment: %v", err)
	}
	t.Cleanup(func() { pc.Close() })
	return pc, pc.LocalAddr().(*net.UDPAddr).Port
}

// startEcho runs an echo loop on the server socket until it is closed.
func startEcho(srv *net.UDPConn) {
	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := srv.ReadFromUDP(buf)
			if err != nil {
				return
			}
			_, _ = srv.WriteToUDP(buf[:n], addr)
		}
	}()
}

func dialed(t *testing.T, port int, opts ...udp.Option) *udp.Transport {
	t.Helper()
	tr := udp.New("127.0.0.1", port, opts...)
	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { tr.Close() })
	return tr
}

func TestSendRecvRoundTrip(t *testing.T) {
	srv, port := newServer(t)
	startEcho(srv)
	tr := dialed(t, port)

	msg := []byte("connectionless datagram")
	n, err := tr.Send(msg)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if n != len(msg) {
		t.Fatalf("Send wrote %d bytes, want %d", n, len(msg))
	}
	got, err := tr.Recv()
	if err != nil {
		t.Fatalf("Recv: %v", err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatalf("Recv = %q, want %q", got, msg)
	}
}

// TestMessageBoundariesPreserved verifies the datagram contract: two Sends produce
// two distinct Recvs, never a coalesced stream.
func TestMessageBoundariesPreserved(t *testing.T) {
	srv, port := newServer(t)
	startEcho(srv)
	tr := dialed(t, port)

	first := []byte("AAA")
	second := []byte("BBBBB")
	if _, err := tr.Send(first); err != nil {
		t.Fatalf("Send first: %v", err)
	}
	if _, err := tr.Send(second); err != nil {
		t.Fatalf("Send second: %v", err)
	}
	got1, err := tr.Recv()
	if err != nil {
		t.Fatalf("Recv first: %v", err)
	}
	got2, err := tr.Recv()
	if err != nil {
		t.Fatalf("Recv second: %v", err)
	}
	if !bytes.Equal(got1, first) || !bytes.Equal(got2, second) {
		t.Fatalf("boundaries not preserved: got %q then %q, want %q then %q", got1, got2, first, second)
	}
}

func TestSendRejectsOversizeDatagram(t *testing.T) {
	srv, port := newServer(t)
	startEcho(srv)
	tr := dialed(t, port, udp.WithMaxPDUSize(8))

	if tr.MaxPDUSize() != 8 {
		t.Fatalf("MaxPDUSize = %d, want 8", tr.MaxPDUSize())
	}
	if _, err := tr.Send(make([]byte, 9)); err == nil {
		t.Fatal("expected Send to reject a datagram larger than MaxPDUSize, got nil")
	}
	if _, err := tr.Send(make([]byte, 8)); err != nil {
		t.Fatalf("Send of exactly MaxPDUSize bytes failed: %v", err)
	}
}

func TestRecvTimeout(t *testing.T) {
	// Server that never replies, so Recv must time out rather than block forever.
	_, port := newServer(t)
	tr := dialed(t, port, udp.WithTimeout(150*time.Millisecond))

	if _, err := tr.Send([]byte("hello")); err != nil {
		t.Fatalf("Send: %v", err)
	}
	start := time.Now()
	if _, err := tr.Recv(); err == nil {
		t.Fatal("expected Recv to time out, got nil error")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Recv took %v, expected to time out near 150ms", elapsed)
	}
}

func TestOperationsOnClosedTransport(t *testing.T) {
	srv, port := newServer(t)
	startEcho(srv)
	tr := udp.New("127.0.0.1", port)

	// Before Connect.
	if tr.IsConnected() {
		t.Fatal("IsConnected true before Connect")
	}
	if _, err := tr.Send([]byte("x")); err == nil {
		t.Fatal("expected Send before Connect to error")
	}
	if _, err := tr.Recv(); err == nil {
		t.Fatal("expected Recv before Connect to error")
	}

	if err := tr.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if !tr.IsConnected() {
		t.Fatal("IsConnected false after Connect")
	}
	if tr.RemoteAddr() == nil {
		t.Fatal("RemoteAddr nil after Connect")
	}

	// After Close.
	if err := tr.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if tr.IsConnected() {
		t.Fatal("IsConnected true after Close")
	}
	if err := tr.Close(); err != nil {
		t.Fatalf("second Close should be a no-op, got: %v", err)
	}
	if _, err := tr.Send([]byte("x")); err == nil {
		t.Fatal("expected Send after Close to error")
	}
}

func TestConnectIdempotent(t *testing.T) {
	srv, port := newServer(t)
	startEcho(srv)
	tr := dialed(t, port)
	if err := tr.Connect(); err != nil {
		t.Fatalf("second Connect should be a no-op, got: %v", err)
	}
}

func TestDefaultMaxPDUSize(t *testing.T) {
	tr := udp.New("127.0.0.1", 135)
	if tr.MaxPDUSize() != transport.MaxPDUSizeDefault {
		t.Fatalf("default MaxPDUSize = %d, want %d", tr.MaxPDUSize(), transport.MaxPDUSizeDefault)
	}
}
