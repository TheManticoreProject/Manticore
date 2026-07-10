package server_test

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/server"
)

func TestNewIPv4Server(t *testing.T) {
	server, err := server.NewIPv4Server()
	if err != nil {
		t.Fatalf("Failed to create new IPv4 server: %v", err)
	}
	if server == nil {
		t.Fatal("NewIPv4Server returned nil")
	}

	if server.Network != "udp4" {
		t.Errorf("Expected network to be 'udp4', got %s", server.Network)
	}

	listenAddr := fmt.Sprintf("%s:%d", constants.IPv4MulticastAddr, constants.ListenPort)
	if !strings.EqualFold(server.Address.String(), listenAddr) {
		t.Errorf("Expected address to be %s, got %s", listenAddr, server.Address.String())
	}

	if server.Conn != nil {
		t.Errorf("Expected connection to be nil, got %v", server.Conn)
	}

	if server.Closed == nil {
		t.Error("Expected Closed channel to be initialized, got nil")
	}

	if server.Debug {
		t.Error("Expected Debug to be false by default, got true")
	}
}

func TestIPv4ServerStartAndStop(t *testing.T) {
	emptyHandler := func(server *server.Server, remoteAddr net.Addr, writer server.ResponseWriter, message *message.Message) bool {
		return true
	}

	server, err := server.NewIPv4ServerWithHandlers(
		[]server.Handler{
			server.HandlerFunc(emptyHandler),
		},
	)
	if err != nil {
		t.Fatalf("Failed to create new IPv4 server: %v", err)
	}
	if server == nil {
		t.Fatal("NewIPv4Server returned nil")
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(250 * time.Millisecond)

	if !server.Listening() {
		t.Error("Expected server to be listening after startup, got not listening")
	}

	server.Close()

	select {
	case <-server.Closed:
		// Server closed successfully
	case <-time.After(1 * time.Second):
		t.Error("Expected server to close within 1 second, but it did not")
	}
}

func TestIPv6NewServer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
		// server_test.go:125: Failed to start server: failed to listen: listen udp6 [ff02::1:3]:5355: setsockopt: not supported by windows
		// https://github.com/golang/go/issues/63529
	}

	server, err := server.NewIPv6Server()
	if err != nil {
		t.Fatalf("Failed to create new IPv6 server: %v", err)
	}
	if server == nil {
		t.Fatal("NewIPv6Server returned nil")
	}

	if server.Network != "udp6" {
		t.Errorf("Expected network to be 'udp6', got %s", server.Network)
	}

	listenAddr := fmt.Sprintf("[%s]:%d", constants.IPv6MulticastAddr, constants.ListenPort)
	if !strings.EqualFold(server.Address.String(), listenAddr) {
		t.Errorf("Expected address to be %s, got %s", listenAddr, server.Address.String())
	}

	if server.Conn != nil {
		t.Errorf("Expected connection to be nil, got %v", server.Conn)
	}

	if server.Closed == nil {
		t.Error("Expected Closed channel to be initialized, got nil")
	}

	if server.Debug {
		t.Error("Expected Debug to be false by default, got true")
	}
}

func TestIPv6ServerStartAndStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
		// server_test.go:125: Failed to start server: failed to listen: listen udp6 [ff02::1:3]:5355: setsockopt: not supported by windows
		// https://github.com/golang/go/issues/63529
	}

	emptyHandler := func(server *server.Server, remoteAddr net.Addr, writer server.ResponseWriter, message *message.Message) bool {
		return true
	}

	server, err := server.NewIPv6ServerWithHandlers(
		[]server.Handler{
			server.HandlerFunc(emptyHandler),
		},
	)
	if err != nil {
		t.Fatalf("Failed to create new IPv6 server: %v", err)
	}
	if server == nil {
		t.Fatal("NewIPv6Server returned nil")
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(250 * time.Millisecond)

	if !server.Listening() {
		t.Error("Expected server to be listening after startup, got not listening")
	}

	server.Close()

	select {
	case <-server.Closed:
		// Server closed successfully
	case <-time.After(1 * time.Second):
		t.Error("Expected server to close within 1 second, but it did not")
	}
}

// TestServerTCPRoundTrip exercises the server's TCP responder end to end: it
// starts ListenAndServeTCP on an ephemeral loopback port, dials it, sends a
// length-prefixed query (RFC 1035 §4.2.2), and confirms the handler chain runs
// and a correctly framed, well-formed response comes back.
func TestServerTCPRoundTrip(t *testing.T) {
	// Handler that answers every A query for the queried name with 10.7.0.10.
	answerHandler := func(_ *server.Server, _ net.Addr, w server.ResponseWriter, msg *message.Message) bool {
		if len(msg.Questions) == 0 {
			return true
		}
		resp := message.NewMessage()
		resp.Header.Identifier = msg.Header.Identifier
		resp.SetResponse()
		if err := resp.AddAnswerClassINTypeA(string(msg.Questions[0].Name), "10.7.0.10"); err != nil {
			return true
		}
		_ = w.WriteMessage(resp)
		return false
	}

	srv, err := server.NewIPv4ServerWithHandlers([]server.Handler{server.HandlerFunc(answerHandler)})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	// Bind an ephemeral loopback port so the test needs no privileges and cannot
	// collide with a real responder on 5355.
	srv.TCPListenAddr = "127.0.0.1:0"

	go func() {
		if err := srv.ListenAndServeTCP(); err != nil {
			t.Errorf("ListenAndServeTCP() error = %v", err)
		}
	}()
	defer srv.Close()

	// Wait for the TCP responder to come up using the synchronized accessor.
	deadline := time.Now().Add(2 * time.Second)
	for !srv.ListeningTCP() {
		if time.Now().After(deadline) {
			t.Fatal("TCP responder did not come up within 2s")
		}
		time.Sleep(5 * time.Millisecond)
	}

	addr := srv.TCPAddr()
	if addr == nil {
		t.Fatal("TCPAddr() = nil after the responder came up")
	}

	conn, err := net.DialTimeout("tcp", addr.String(), 2*time.Second)
	if err != nil {
		t.Fatalf("DialTimeout() error = %v", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	// Build and send a length-prefixed query for "host.local".
	query := message.NewMessage()
	query.SetQuery()
	if err := query.AddQuestion("host.local", llmnr_type.TypeA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion() error = %v", err)
	}
	encoded, err := query.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := message.WriteTCPMessage(conn, encoded); err != nil {
		t.Fatalf("WriteTCPMessage() error = %v", err)
	}

	// Read and decode the length-prefixed response.
	payload, err := message.ReadTCPMessage(conn)
	if err != nil {
		t.Fatalf("ReadTCPMessage() error = %v", err)
	}
	resp := &message.Message{}
	if _, err := resp.Unmarshal(payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !resp.IsResponse() {
		t.Error("TCP response does not have the QR (response) bit set")
	}
	if resp.Header.Identifier != query.Header.Identifier {
		t.Errorf("TCP response ID = %d, want %d", resp.Header.Identifier, query.Header.Identifier)
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("TCP response answers = %d, want 1", len(resp.Answers))
	}
	if got := net.IP(resp.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("TCP response answer RDATA = %q, want %q", got, "10.7.0.10")
	}
}

func TestIsIPv4AndIsIPv6AreMutuallyExclusive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	ipv4Server, err := server.NewIPv4Server()
	if err != nil {
		t.Fatalf("Failed to create IPv4 server: %v", err)
	}
	if !ipv4Server.IsIPv4() {
		t.Errorf("Expected IPv4 server IsIPv4()=true, got false")
	}
	if ipv4Server.IsIPv6() {
		t.Errorf("Expected IPv4 server IsIPv6()=false, got true (IPv4 address reports as IPv6)")
	}

	ipv6Server, err := server.NewIPv6Server()
	if err != nil {
		t.Fatalf("Failed to create IPv6 server: %v", err)
	}
	if !ipv6Server.IsIPv6() {
		t.Errorf("Expected IPv6 server IsIPv6()=true, got false")
	}
	if ipv6Server.IsIPv4() {
		t.Errorf("Expected IPv6 server IsIPv4()=false, got true")
	}
}
