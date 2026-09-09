package tcp_test

import (
	"bytes"
	"io"
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

// dialLoopback starts a listener, hands the accepted connection back on a channel,
// and returns a transport already connected to it.
func dialLoopback(t *testing.T) (*tcp.TCPTransport, <-chan net.Conn) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			close(accepted)
			return
		}
		accepted <- c
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
	tr.SetTimeout(2 * time.Second)
	if err := tr.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("TCPTransport.Connect() error = %v", err)
	}
	t.Cleanup(func() { tr.Close() })

	return tr, accepted
}

// TestTCPTransport_SendWritesDirectTCPHeader pins the framing Send emits: a zero
// byte followed by the 3-byte big-endian payload length ([MS-SMB2] 2.1).
func TestTCPTransport_SendWritesDirectTCPHeader(t *testing.T) {
	tr, accepted := dialLoopback(t)

	payload := []byte{0xFE, 'S', 'M', 'B'}
	if _, err := tr.Send(payload); err != nil {
		t.Fatalf("TCPTransport.Send() error = %v", err)
	}

	conn := <-accepted
	defer conn.Close()

	got := make([]byte, 4+len(payload))
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("failed to read frame: %v", err)
	}

	want := []byte{0x00, 0x00, 0x00, 0x04, 0xFE, 'S', 'M', 'B'}
	if !bytes.Equal(got, want) {
		t.Errorf("Send() wrote % x, want % x", got, want)
	}
}

// TestTCPTransport_SendRejectsOversizedPayload guards the length-truncation bug: a
// payload above the 24-bit field must be refused, not encoded modulo 2^24. At
// exactly MaxDirectTCPFrameLength+1 the truncated length would be 0, so the peer
// would frame the remainder of the stream at the wrong offset.
func TestTCPTransport_SendRejectsOversizedPayload(t *testing.T) {
	tr, accepted := dialLoopback(t)

	// Drain the peer for the duration of the test. Without this the unguarded code
	// path blocks forever on TCP backpressure partway through the 16 MiB write, so a
	// regression would hang the suite instead of failing it.
	conn := <-accepted
	defer conn.Close()
	go io.Copy(io.Discard, conn)

	n, err := tr.Send(make([]byte, tcp.MaxDirectTCPFrameLength+1))
	if err == nil {
		t.Fatalf("TCPTransport.Send() accepted a payload larger than the 24-bit length field, wrote %d bytes", n)
	}
	if n != 0 {
		t.Errorf("TCPTransport.Send() wrote %d byte(s) for a rejected message, want 0", n)
	}
}
