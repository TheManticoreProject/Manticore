package nbt_test

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
)

// connectedPair returns a connected NBTTransport together with the raw
// server-side connection it is talking to, so tests can inspect or produce
// exact wire bytes. The returned cleanup closes both ends and the listener.
func connectedPair(t *testing.T) (*nbt.NBTTransport, net.Conn, func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}

	type accepted struct {
		conn net.Conn
		err  error
	}
	ch := make(chan accepted, 1)
	go func() {
		c, err := ln.Accept()
		ch <- accepted{c, err}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	tr := nbt.NewNBTTransport()
	tr.SetTimeout(5 * time.Second)
	if err := tr.Connect(addr.IP, addr.Port); err != nil {
		ln.Close()
		t.Fatalf("NBTTransport.Connect() error = %v", err)
	}

	a := <-ch
	if a.err != nil {
		tr.Close()
		ln.Close()
		t.Fatalf("failed to accept connection: %v", a.err)
	}

	cleanup := func() {
		tr.Close()
		if a.conn != nil {
			a.conn.Close()
		}
		ln.Close()
	}
	return tr, a.conn, cleanup
}

// TestNBTTransport_SendFramesSub64K asserts the exact wire bytes emitted for a
// payload below 65536 bytes: the FLAGS byte (header[1]) is 0x00 and LENGTH is
// carried in the low 16 bits.
func TestNBTTransport_SendFramesSub64K(t *testing.T) {
	tr, srv, cleanup := connectedPair(t)
	defer cleanup()

	payload := []byte("test")
	if _, err := tr.Send(payload); err != nil {
		t.Fatalf("NBTTransport.Send() error = %v", err)
	}

	got := make([]byte, 4+len(payload))
	if _, err := io.ReadFull(srv, got); err != nil {
		t.Fatalf("failed to read framed message: %v", err)
	}

	want := []byte{0x00, 0x00, 0x00, 0x04, 't', 'e', 's', 't'}
	if !bytes.Equal(got, want) {
		t.Fatalf("framed bytes = % x, want % x", got, want)
	}
}

// TestNBTTransport_SendFramesOver64K asserts that a payload larger than 65535
// bytes sets the length-extension bit (header[1] & 0x01) and encodes the 17th
// length bit correctly alongside the low 16 bits.
func TestNBTTransport_SendFramesOver64K(t *testing.T) {
	tr, srv, cleanup := connectedPair(t)
	defer cleanup()

	// 0x1ABCD == 109517 bytes: high bit set, low bytes 0xAB and 0xCD.
	const n = 0x1ABCD
	payload := make([]byte, n)

	errCh := make(chan error, 1)
	go func() {
		_, err := tr.Send(payload)
		errCh <- err
	}()

	header := make([]byte, 4)
	if _, err := io.ReadFull(srv, header); err != nil {
		t.Fatalf("failed to read framed header: %v", err)
	}
	want := []byte{0x00, 0x01, 0xAB, 0xCD}
	if !bytes.Equal(header, want) {
		t.Fatalf("framed header = % x, want % x", header, want)
	}
	if _, err := io.CopyN(io.Discard, srv, n); err != nil {
		t.Fatalf("failed to drain payload: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("NBTTransport.Send() error = %v", err)
	}
}

// TestNBTTransport_SendRejectsTooLarge verifies that a payload exceeding the
// 17-bit maximum (131071 bytes) is rejected instead of being mis-framed.
func TestNBTTransport_SendRejectsTooLarge(t *testing.T) {
	tr, srv, cleanup := connectedPair(t)
	defer cleanup()

	// Keep draining so a (mistaken) write cannot block the test.
	go io.Copy(io.Discard, srv)

	payload := make([]byte, 0x1FFFF+1)
	if _, err := tr.Send(payload); err == nil {
		t.Fatal("NBTTransport.Send() should reject payloads larger than 131071 bytes")
	}
}

// TestNBTTransport_ReceiveParsesExtensionBit feeds fixed wire vectors into
// Receive and checks that the extension bit in header[1] is combined as the
// 17th length bit, for both the sub-64K and over-64K cases.
func TestNBTTransport_ReceiveParsesExtensionBit(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
		length int
	}{
		{name: "sub-64K", header: []byte{0x00, 0x00, 0x00, 0x04}, length: 4},
		{name: "over-64K", header: []byte{0x00, 0x01, 0x00, 0x02}, length: (1 << 16) | 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, srv, cleanup := connectedPair(t)
			defer cleanup()

			go func() {
				srv.Write(tt.header)
				srv.Write(make([]byte, tt.length))
			}()

			got, err := tr.Receive()
			if err != nil {
				t.Fatalf("NBTTransport.Receive() error = %v", err)
			}
			if len(got) != tt.length {
				t.Fatalf("Receive() returned %d bytes, want %d", len(got), tt.length)
			}
		})
	}
}

// TestNBTTransport_SendReceiveRoundTrip proves the framing round-trips: the
// bytes produced by Send are echoed back and parsed by Receive into the exact
// original payload, for both the sub-64K and over-64K length paths.
func TestNBTTransport_SendReceiveRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "sub-64K", length: 512},
		{name: "over-64K", length: 65538},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, srv, cleanup := connectedPair(t)
			defer cleanup()

			payload := make([]byte, tt.length)
			for i := range payload {
				payload[i] = byte(i)
			}

			// Echo whatever bytes Send produces straight back to the transport.
			go io.Copy(srv, srv)

			if _, err := tr.Send(payload); err != nil {
				t.Fatalf("NBTTransport.Send() error = %v", err)
			}
			got, err := tr.Receive()
			if err != nil {
				t.Fatalf("NBTTransport.Receive() error = %v", err)
			}
			if !bytes.Equal(got, payload) {
				t.Fatalf("round-trip payload mismatch: got %d bytes, want %d", len(got), len(payload))
			}
		})
	}
}

func TestNBTTransport_Connect(t *testing.T) {
	tests := []struct {
		name    string
		ip      net.IP
		port    int
		wantErr bool
	}{
		{
			name:    "Valid IPv4 connection with default port",
			ip:      net.ParseIP("127.0.0.1"),
			port:    0,
			wantErr: true, // Will fail since no server is running
		},
		{
			name:    "Valid IPv4 connection with custom port",
			ip:      net.ParseIP("127.0.0.1"),
			port:    139,
			wantErr: true,
		},
		{
			name:    "Valid IPv6 connection",
			ip:      net.ParseIP("::1"),
			port:    139,
			wantErr: true,
		},
		{
			name:    "Invalid IP",
			ip:      nil,
			port:    139,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := nbt.NewNBTTransport()
			err := tr.Connect(tt.ip, tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NBTTransport.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			tr.Close()
		})
	}
}

func TestNBTTransport_Send(t *testing.T) {
	tr := nbt.NewNBTTransport()

	// Test sending without connection
	_, err := tr.Send([]byte("test"))
	if err == nil {
		t.Error("NBTTransport.Send() should error when not connected")
	}

	tr.Close()
}

func TestNBTTransport_ReceiveTimesOutOnSilentServer(t *testing.T) {
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

	addr := ln.Addr().(*net.TCPAddr)

	tr := nbt.NewNBTTransport()
	tr.SetTimeout(100 * time.Millisecond)
	if err := tr.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("NBTTransport.Connect() error = %v", err)
	}
	defer tr.Close()

	start := time.Now()
	_, err = tr.Receive()
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("NBTTransport.Receive() should time out on a silent server, got nil error")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("NBTTransport.Receive() took %v to fail, want a bounded timeout", elapsed)
	}
}
