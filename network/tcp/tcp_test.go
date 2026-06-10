package tcp_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/tcp"
)

func TestTCPTransport_Connect(t *testing.T) {
	t.Run("Connect succeeds to running IPv4 server", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to start IPv4 test server: %v", err)
		}
		defer ln.Close()

		// Accept a single connection in background
		go func() {
			c, err := ln.Accept()
			if err == nil {
				c.Close()
			}
		}()

		host, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatalf("failed to parse listener address: %v", err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatalf("failed to parse port: %v", err)
		}

		tr := tcp.NewTCPTransport()
		if err := tr.Connect(net.ParseIP(host), port); err != nil {
			t.Fatalf("TCPTransport.Connect() error = %v, want no error", err)
		}
		_ = tr.Close()
	})

	t.Run("Connect succeeds to running IPv6 server (if available)", func(t *testing.T) {
		ln, err := net.Listen("tcp", "[::1]:0")
		if err != nil {
			t.Skipf("IPv6 loopback not available: %v", err)
		}
		defer ln.Close()

		go func() {
			c, err := ln.Accept()
			if err == nil {
				c.Close()
			}
		}()

		host, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatalf("failed to parse listener address: %v", err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatalf("failed to parse port: %v", err)
		}

		tr := tcp.NewTCPTransport()
		if err := tr.Connect(net.ParseIP(host), port); err != nil {
			t.Fatalf("TCPTransport.Connect() error = %v, want no error", err)
		}
		_ = tr.Close()
	})

	t.Run("Invalid IP returns error", func(t *testing.T) {
		tr := tcp.NewTCPTransport()
		if err := tr.Connect(nil, 445); err == nil {
			t.Error("TCPTransport.Connect() should return error when IP is nil")
		}
		_ = tr.Close()
	})
}

func TestTCPTransport_Send(t *testing.T) {
	tr := tcp.NewTCPTransport()

	// Test sending without connection
	_, err := tr.Send([]byte("test"))
	if err == nil {
		t.Error("TCPTransport.Send() should return error when not connected")
	}
}

func TestTCPTransport_Close(t *testing.T) {
	tr := tcp.NewTCPTransport()

	// Test closing without connection
	err := tr.Close()
	if err != nil {
		t.Error("TCPTransport.Close() should not return error when not connected")
	}
}

func TestTCPTransport_ReceiveRejectsOversizedLength(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		// Send a Direct TCP header claiming a 0xFFFFFF (≈16 MiB) payload,
		// well above MaxDirectTCPPayloadSize; Receive should reject before
		// allocating or reading.
		_, _ = c.Write([]byte{0x00, 0xFF, 0xFF, 0xFF})
	}()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse listener address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	tr := tcp.NewTCPTransport()
	if err := tr.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("TCPTransport.Connect() error = %v", err)
	}
	defer tr.Close()

	_, err = tr.Receive()
	if err == nil {
		t.Fatal("TCPTransport.Receive() should return error for oversized length, got nil")
	}
}

func TestTCPTransport_ReceiveTimesOutOnSilentServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	// Accept the connection but never write anything.
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		select {}
	}()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse listener address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	tr := tcp.NewTCPTransport()
	tr.SetTimeout(100 * time.Millisecond)
	if err := tr.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("TCPTransport.Connect() error = %v", err)
	}
	defer tr.Close()

	start := time.Now()
	_, err = tr.Receive()
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("TCPTransport.Receive() should time out on a silent server, got nil error")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("TCPTransport.Receive() took %v to fail, want a bounded timeout", elapsed)
	}
}
