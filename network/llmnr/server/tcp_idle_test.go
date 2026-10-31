package server_test

import (
	"errors"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/server"
)

// startTCPResponder starts ListenAndServeTCP on an ephemeral loopback port with
// a no-op handler and returns the server once its TCP socket is bound.
func startTCPResponder(t *testing.T, configure func(*server.Server)) *server.Server {
	t.Helper()

	nop := func(_ *server.Server, _ net.Addr, _ server.ResponseWriter, _ *message.Message) bool { return true }
	srv, err := server.NewIPv4ServerWithHandlers([]server.Handler{server.HandlerFunc(nop)})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	srv.TCPListenAddr = "127.0.0.1:0"
	if configure != nil {
		configure(srv)
	}

	go func() {
		if err := srv.ListenAndServeTCP(); err != nil {
			t.Errorf("ListenAndServeTCP() error = %v", err)
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for !srv.ListeningTCP() {
		if time.Now().After(deadline) {
			t.Fatal("TCP responder did not come up within 2s")
		}
		time.Sleep(5 * time.Millisecond)
	}
	return srv
}

// expectServerClosed reads from conn until the server side closes it, failing
// the test if conn is still open after wait.
func expectServerClosed(t *testing.T, conn net.Conn, wait time.Duration) {
	t.Helper()

	_ = conn.SetReadDeadline(time.Now().Add(wait))
	_, err := conn.Read(make([]byte, 1))
	if errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("connection still open after %s", wait)
	}
	if err != io.EOF {
		// A reset is also an acceptable sign the server dropped the connection.
		var opErr *net.OpError
		if !errors.As(err, &opErr) {
			t.Fatalf("Read() error = %v, want EOF or connection reset", err)
		}
	}
}

// TestServerTCPCloseClosesIdleConnections verifies that Close tears down TCP
// connections that were accepted but never sent a query, so their handler
// goroutines do not outlive the server.
func TestServerTCPCloseClosesIdleConnections(t *testing.T) {
	srv := startTCPResponder(t, nil)

	var conns []net.Conn
	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", srv.TCPAddr().String(), 2*time.Second)
		if err != nil {
			t.Fatalf("DialTimeout() error = %v", err)
		}
		defer conn.Close()
		conns = append(conns, conn)
	}
	// Let the server accept every connection before closing it.
	time.Sleep(100 * time.Millisecond)

	srv.Close()

	for _, conn := range conns {
		expectServerClosed(t, conn, 2*time.Second)
	}
}

// TestServerTCPIdleConnectionTimesOut verifies that a client that connects but
// never sends a query is disconnected once TCPReadTimeout elapses, while the
// server keeps running.
func TestServerTCPIdleConnectionTimesOut(t *testing.T) {
	srv := startTCPResponder(t, func(s *server.Server) {
		s.TCPReadTimeout = 200 * time.Millisecond
	})
	defer srv.Close()

	conn, err := net.DialTimeout("tcp", srv.TCPAddr().String(), 2*time.Second)
	if err != nil {
		t.Fatalf("DialTimeout() error = %v", err)
	}
	defer conn.Close()

	expectServerClosed(t, conn, 2*time.Second)
}
