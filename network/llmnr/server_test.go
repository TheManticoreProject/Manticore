package llmnr_test

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr"
)

func TestNewIPv4Server(t *testing.T) {
	server, err := llmnr.NewIPv4Server()
	if err != nil {
		t.Fatalf("Failed to create new IPv4 server: %v", err)
	}
	if server == nil {
		t.Fatal("NewIPv4Server returned nil")
	}

	if server.Network != "udp4" {
		t.Errorf("Expected network to be 'udp4', got %s", server.Network)
	}

	listenAddr := fmt.Sprintf("%s:%d", llmnr.IPv4MulticastAddr, llmnr.LLMNRPort)
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
	emptyHandler := func(server *llmnr.Server, remoteAddr net.Addr, writer llmnr.ResponseWriter, message *llmnr.Message) bool {
		return true
	}

	server, err := llmnr.NewIPv4ServerWithHandlers([]llmnr.Handler{llmnr.HandlerFunc(emptyHandler)})
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

	if server.Conn == nil {
		t.Error("Expected server connection to be initialized, got nil")
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

	server, err := llmnr.NewIPv6Server()
	if err != nil {
		t.Fatalf("Failed to create new IPv6 server: %v", err)
	}
	if server == nil {
		t.Fatal("NewIPv6Server returned nil")
	}

	if server.Network != "udp6" {
		t.Errorf("Expected network to be 'udp6', got %s", server.Network)
	}

	listenAddr := fmt.Sprintf("[%s]:%d", llmnr.IPv6MulticastAddr, llmnr.LLMNRPort)
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

	emptyHandler := func(server *llmnr.Server, remoteAddr net.Addr, writer llmnr.ResponseWriter, message *llmnr.Message) bool {
		return true
	}

	server, err := llmnr.NewIPv6ServerWithHandlers([]llmnr.Handler{llmnr.HandlerFunc(emptyHandler)})
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

	if server.Conn == nil {
		t.Error("Expected server connection to be initialized, got nil")
	}

	server.Close()

	select {
	case <-server.Closed:
		// Server closed successfully
	case <-time.After(1 * time.Second):
		t.Error("Expected server to close within 1 second, but it did not")
	}
}
