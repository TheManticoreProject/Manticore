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
		// Send a Direct TCP header claiming a 0x100000 (1 MiB) payload, above the
		// cap the test installs below; Receive should reject before allocating or
		// reading.
		_, _ = c.Write([]byte{0x00, 0x10, 0x00, 0x00})
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
	// A frame of 1 MiB is legal by default now, so bound this transport explicitly.
	tr.SetMaxPayloadSize(64 * 1024)

	_, err = tr.Receive()
	if err == nil {
		t.Fatal("TCPTransport.Receive() should return error for oversized length, got nil")
	}
}

// TestTCPTransport_ReceiveAcceptsLargeNegotiatedFrame guards the defect: a frame
// larger than the old 1 MiB cap but within what the 24-bit length field describes
// must be accepted, because a Windows server advertises an 8 MiB MaxReadSize and is
// entitled to answer a read in one frame of that size.
func TestTCPTransport_ReceiveAcceptsLargeNegotiatedFrame(t *testing.T) {
	const payloadLen = 2 * 1024 * 1024 // above the old cap, within the field

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
		header := []byte{0x00, byte((payloadLen >> 16) & 0xFF), byte((payloadLen >> 8) & 0xFF), byte(payloadLen & 0xFF)}
		if _, err := c.Write(header); err != nil {
			return
		}
		_, _ = c.Write(make([]byte, payloadLen))
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
	tr.SetTimeout(10 * time.Second)
	if err := tr.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("TCPTransport.Connect() error = %v", err)
	}
	defer tr.Close()

	got, err := tr.Receive()
	if err != nil {
		t.Fatalf("TCPTransport.Receive() rejected a %d-byte frame: %v", payloadLen, err)
	}
	if len(got) != payloadLen {
		t.Errorf("Receive() returned %d bytes, want %d", len(got), payloadLen)
	}
}

// TestTCPTransport_SetMaxPayloadSizeBounds checks the listener-side lever: an
// explicit cap is honoured, and 0 or an unrepresentable value restores the default.
func TestTCPTransport_SetMaxPayloadSizeBounds(t *testing.T) {
	// Drive each case through Receive against a peer that announces a frame of a
	// known size, since the cap is only observable there.
	for _, tc := range []struct {
		name      string
		cap       uint32
		announce  int
		wantError bool
	}{
		{"within explicit cap", 64 * 1024, 1024, false},
		{"above explicit cap", 1024, 64 * 1024, true},
		{"zero restores default", 0, 2 * 1024 * 1024, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("listen: %v", err)
			}
			defer ln.Close()

			go func() {
				c, err := ln.Accept()
				if err != nil {
					return
				}
				defer c.Close()
				n := tc.announce
				_, _ = c.Write([]byte{0x00, byte(n >> 16), byte(n >> 8), byte(n)})
				_, _ = c.Write(make([]byte, n))
			}()

			host, portStr, _ := net.SplitHostPort(ln.Addr().String())
			port, _ := strconv.Atoi(portStr)

			conn := tcp.NewTCPTransport()
			conn.SetTimeout(10 * time.Second)
			if err := conn.Connect(net.ParseIP(host), port); err != nil {
				t.Fatalf("connect: %v", err)
			}
			defer conn.Close()
			conn.SetMaxPayloadSize(tc.cap)

			_, err = conn.Receive()
			if tc.wantError && err == nil {
				t.Errorf("Receive() accepted %d bytes under a %d cap", tc.announce, tc.cap)
			}
			if !tc.wantError && err != nil {
				t.Errorf("Receive() rejected %d bytes under a %d cap: %v", tc.announce, tc.cap, err)
			}
		})
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
