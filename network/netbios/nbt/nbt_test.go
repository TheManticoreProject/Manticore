package nbt_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbns"
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
	// These helpers exercise the raw framing path, so skip the session
	// establishment handshake that Connect performs by default.
	tr.SetHandshakeEnabled(false)
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

// readSessionRequest reads a full SESSION REQUEST (4-byte header + body) from
// the raw server-side connection and returns the header and body bytes.
func readSessionRequest(t *testing.T, srv net.Conn) (header, body []byte) {
	t.Helper()
	header = make([]byte, 4)
	if _, err := io.ReadFull(srv, header); err != nil {
		t.Fatalf("failed to read SESSION REQUEST header: %v", err)
	}
	length := (int(header[1]&0x01) << 16) | (int(header[2]) << 8) | int(header[3])
	body = make([]byte, length)
	if _, err := io.ReadFull(srv, body); err != nil {
		t.Fatalf("failed to read SESSION REQUEST body: %v", err)
	}
	return header, body
}

// listenerReplying starts a TCP listener that, on the first accepted
// connection, reads the SESSION REQUEST and writes reply. It returns the
// listener address and a cleanup function.
func listenerReplying(t *testing.T, reply []byte, capture func(header, body []byte)) (*net.TCPAddr, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		header, body := readSessionRequest(t, c)
		if capture != nil {
			capture(header, body)
		}
		c.Write(reply)
		// Keep the connection open briefly so the client can read the reply.
		time.Sleep(200 * time.Millisecond)
	}()
	return ln.Addr().(*net.TCPAddr), func() { ln.Close() }
}

// TestNBTTransport_SessionRequestWireBytes asserts the exact SESSION REQUEST a
// default transport emits for the "*SMBSERVER" called name: type 0x81, length
// 68, and the 34-byte second-level-encoded called and calling names.
func TestNBTTransport_SessionRequestWireBytes(t *testing.T) {
	var gotHeader, gotBody []byte
	addr, cleanup := listenerReplying(t, []byte{0x82, 0x00, 0x00, 0x00}, func(h, b []byte) {
		gotHeader, gotBody = h, b
	})
	defer cleanup()

	tr := nbt.NewNBTTransport()
	tr.SetCallingName("CLIENT")
	tr.SetTimeout(5 * time.Second)
	if err := tr.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer tr.Close()

	// Header: TYPE 0x81, FLAGS 0x00, LENGTH 0x0044 (68).
	wantHeader := []byte{0x81, 0x00, 0x00, 0x44}
	if !bytes.Equal(gotHeader, wantHeader) {
		t.Fatalf("SESSION REQUEST header = % x, want % x", gotHeader, wantHeader)
	}
	if len(gotBody) != 68 {
		t.Fatalf("SESSION REQUEST body length = %d, want 68", len(gotBody))
	}

	// The well-known second-level encoding of "*SMBSERVER"<20>: 0x20 length
	// prefix, 32 encoded chars, 0x00 terminator.
	wantCalled := append(append([]byte{0x20}, []byte("CKFDENECFDEFFCFGEFFCCACACACACACA")...), 0x00)
	if !bytes.Equal(gotBody[:34], wantCalled) {
		t.Fatalf("called name = % x, want % x", gotBody[:34], wantCalled)
	}
	// Calling name "CLIENT"<00>: verify the framing bytes (length prefix and
	// terminator) and that it decodes back to the workstation name.
	calling := gotBody[34:68]
	if calling[0] != 0x20 || calling[33] != 0x00 {
		t.Fatalf("calling name framing = % x, want 0x20 ... 0x00", calling)
	}
}

// TestNBTTransport_HandshakePositive verifies a POSITIVE SESSION RESPONSE
// completes Connect without error.
func TestNBTTransport_HandshakePositive(t *testing.T) {
	addr, cleanup := listenerReplying(t, []byte{0x82, 0x00, 0x00, 0x00}, nil)
	defer cleanup()

	tr := nbt.NewNBTTransport()
	tr.SetTimeout(5 * time.Second)
	if err := tr.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("Connect() with POSITIVE response error = %v", err)
	}
	tr.Close()
}

// TestNBTTransport_HandshakeNegative verifies each NEGATIVE SESSION RESPONSE
// error code is surfaced as a mapped error and fails Connect.
func TestNBTTransport_HandshakeNegative(t *testing.T) {
	codes := []byte{0x80, 0x81, 0x82, 0x83, 0x8F}
	for _, code := range codes {
		code := code
		t.Run(fmt.Sprintf("0x%02X", code), func(t *testing.T) {
			addr, cleanup := listenerReplying(t, []byte{0x83, 0x00, 0x00, 0x01, code}, nil)
			defer cleanup()

			tr := nbt.NewNBTTransport()
			// Use a non-wildcard called name so the NEGATIVE response is
			// surfaced directly without the "*SMBSERVER" NODE STATUS fallback.
			tr.SetCalledName("TESTSERVER")
			tr.SetTimeout(5 * time.Second)
			err := tr.Connect(addr.IP, addr.Port)
			if err == nil {
				tr.Close()
				t.Fatalf("Connect() with NEGATIVE 0x%02X should fail", code)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("0x%02X", code)) {
				t.Fatalf("error %q does not mention code 0x%02X", err.Error(), code)
			}
		})
	}
}

// TestNBTTransport_HandshakeRetarget verifies a RETARGET SESSION RESPONSE
// causes the transport to re-dial the advertised endpoint and retry the
// SESSION REQUEST, succeeding when the second endpoint answers POSITIVE.
func TestNBTTransport_HandshakeRetarget(t *testing.T) {
	// Second endpoint answers POSITIVE.
	addr2, cleanup2 := listenerReplying(t, []byte{0x82, 0x00, 0x00, 0x00}, nil)
	defer cleanup2()

	// First endpoint answers RETARGET pointing at addr2.
	ip4 := addr2.IP.To4()
	retarget := []byte{0x84, 0x00, 0x00, 0x06, ip4[0], ip4[1], ip4[2], ip4[3], byte(addr2.Port >> 8), byte(addr2.Port & 0xFF)}
	addr1, cleanup1 := listenerReplying(t, retarget, nil)
	defer cleanup1()

	tr := nbt.NewNBTTransport()
	tr.SetTimeout(5 * time.Second)
	if err := tr.Connect(addr1.IP, addr1.Port); err != nil {
		t.Fatalf("Connect() through RETARGET error = %v", err)
	}
	if !tr.IsConnected() {
		t.Fatal("transport should be connected after following RETARGET")
	}
	tr.Close()
}

// TestNBTTransport_HandshakeSkipsKeepAlive verifies a SESSION KEEP ALIVE frame
// preceding the POSITIVE response is transparently ignored.
func TestNBTTransport_HandshakeSkipsKeepAlive(t *testing.T) {
	// 0x85 keep-alive (empty body) followed by a POSITIVE response.
	reply := []byte{0x85, 0x00, 0x00, 0x00, 0x82, 0x00, 0x00, 0x00}
	addr, cleanup := listenerReplying(t, reply, nil)
	defer cleanup()

	tr := nbt.NewNBTTransport()
	tr.SetTimeout(5 * time.Second)
	if err := tr.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("Connect() past keep-alive error = %v", err)
	}
	tr.Close()
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
	tr.SetHandshakeEnabled(false)
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

// acceptSessionResult carries the outcome of a server-side AcceptSession run
// back to the test goroutine.
type acceptSessionResult struct {
	transport *nbt.NBTTransport
	called    string
	calling   string
	err       error
}

// serveAcceptSession listens on an ephemeral port and runs AcceptSession on the
// first connection it accepts, delivering the outcome on the returned channel.
// It is the server-side counterpart of the connectedPair helper above.
func serveAcceptSession(t *testing.T, acceptedNames []string) (*net.TCPAddr, <-chan acceptSessionResult, func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}

	ch := make(chan acceptSessionResult, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			ch <- acceptSessionResult{err: err}
			return
		}
		tr := nbt.NewNBTTransportFromConn(conn)
		tr.SetTimeout(5 * time.Second)
		called, calling, err := tr.AcceptSession(acceptedNames)
		ch <- acceptSessionResult{transport: tr, called: called, calling: calling, err: err}
	}()

	return ln.Addr().(*net.TCPAddr), ch, func() { ln.Close() }
}

// TestNBTTransport_AcceptSessionPositive verifies the accept side completes the
// handshake against the client side of the same implementation, and that the
// resulting transports carry SESSION MESSAGEs in both directions.
func TestNBTTransport_AcceptSessionPositive(t *testing.T) {
	addr, results, cleanup := serveAcceptSession(t, []string{"FILESERVER"})
	defer cleanup()

	client := nbt.NewNBTTransport()
	client.SetCalledName("FILESERVER")
	client.SetCallingName("TESTCLIENT")
	client.SetTimeout(5 * time.Second)
	if err := client.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	defer client.Close()

	got := <-results
	if got.err != nil {
		t.Fatalf("AcceptSession() error = %v", got.err)
	}
	defer got.transport.Close()

	if got.called != "FILESERVER" {
		t.Fatalf("called name = %q, want %q", got.called, "FILESERVER")
	}
	if got.calling != "TESTCLIENT" {
		t.Fatalf("calling name = %q, want %q", got.calling, "TESTCLIENT")
	}

	// The session is established, so both ends must now frame SESSION MESSAGEs
	// for each other.
	payload := []byte{0xFF, 'S', 'M', 'B', 0x72}
	if _, err := client.Send(payload); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	received, err := got.transport.Receive()
	if err != nil {
		t.Fatalf("server Receive() error = %v", err)
	}
	if !bytes.Equal(received, payload) {
		t.Fatalf("server received % x, want % x", received, payload)
	}

	reply := []byte{0xFF, 'S', 'M', 'B', 0x72, 0x00}
	if _, err := got.transport.Send(reply); err != nil {
		t.Fatalf("server Send() error = %v", err)
	}
	back, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if !bytes.Equal(back, reply) {
		t.Fatalf("client received % x, want % x", back, reply)
	}
}

// TestNBTTransport_AcceptSessionWildcard verifies the "*SMBSERVER" convention is
// answered even when it is not in the accepted-name list, and that an empty list
// answers to any name.
func TestNBTTransport_AcceptSessionWildcard(t *testing.T) {
	cases := []struct {
		name          string
		acceptedNames []string
		calledName    string
	}{
		{"wildcard against a named endpoint", []string{"FILESERVER"}, "*SMBSERVER"},
		{"any name against an empty list", nil, "SOMEOTHERHOST"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			addr, results, cleanup := serveAcceptSession(t, tc.acceptedNames)
			defer cleanup()

			client := nbt.NewNBTTransport()
			client.SetCalledName(tc.calledName)
			client.SetTimeout(5 * time.Second)
			if err := client.Connect(addr.IP, addr.Port); err != nil {
				t.Fatalf("client Connect() to %q error = %v", tc.calledName, err)
			}
			defer client.Close()

			got := <-results
			if got.err != nil {
				t.Fatalf("AcceptSession() error = %v", got.err)
			}
			got.transport.Close()
			if got.called != tc.calledName {
				t.Fatalf("called name = %q, want %q", got.called, tc.calledName)
			}
		})
	}
}

// TestNBTTransport_AcceptSessionRefusesUnknownName verifies a CALLED name the
// endpoint does not serve is refused with a NEGATIVE SESSION RESPONSE carrying
// NEGATIVE_SESSION_NOT_LISTENING_ON_CALLED_NAME (0x80), which is what the client
// side surfaces as an error.
func TestNBTTransport_AcceptSessionRefusesUnknownName(t *testing.T) {
	addr, results, cleanup := serveAcceptSession(t, []string{"FILESERVER"})
	defer cleanup()

	// Drive the client at the raw byte level so the response can be inspected
	// exactly, rather than through Connect's fallback logic.
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		t.Fatalf("dial error = %v", err)
	}
	defer conn.Close()

	request := sessionRequestBytes(t, "OTHERHOST", 0x20, "TESTCLIENT", 0x00)
	if _, err := conn.Write(request); err != nil {
		t.Fatalf("write SESSION REQUEST error = %v", err)
	}

	response := make([]byte, 5)
	if _, err := io.ReadFull(conn, response); err != nil {
		t.Fatalf("read SESSION RESPONSE error = %v", err)
	}
	want := []byte{0x83, 0x00, 0x00, 0x01, 0x80}
	if !bytes.Equal(response, want) {
		t.Fatalf("response = % x, want % x", response, want)
	}

	got := <-results
	if got.err == nil {
		t.Fatal("AcceptSession() should reject a CALLED name the endpoint does not serve")
	}
	if !strings.Contains(got.err.Error(), "OTHERHOST") {
		t.Fatalf("error %q does not name the refused CALLED name", got.err.Error())
	}
}

// TestNBTTransport_AcceptSessionRejectsWrongSuffix verifies a CALLED name
// addressed to a service other than the server service (0x20) is refused with
// NEGATIVE_SESSION_CALLED_NAME_NOT_PRESENT (0x82).
func TestNBTTransport_AcceptSessionRejectsWrongSuffix(t *testing.T) {
	addr, results, cleanup := serveAcceptSession(t, nil)
	defer cleanup()

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		t.Fatalf("dial error = %v", err)
	}
	defer conn.Close()

	// Suffix 0x00 is the workstation service, not the server service.
	request := sessionRequestBytes(t, "FILESERVER", 0x00, "TESTCLIENT", 0x00)
	if _, err := conn.Write(request); err != nil {
		t.Fatalf("write SESSION REQUEST error = %v", err)
	}

	response := make([]byte, 5)
	if _, err := io.ReadFull(conn, response); err != nil {
		t.Fatalf("read SESSION RESPONSE error = %v", err)
	}
	want := []byte{0x83, 0x00, 0x00, 0x01, 0x82}
	if !bytes.Equal(response, want) {
		t.Fatalf("response = % x, want % x", response, want)
	}

	if got := <-results; got.err == nil {
		t.Fatal("AcceptSession() should reject a non-server service suffix")
	}
}

// TestNBTTransport_AcceptSessionRejectsMalformed verifies the accept side refuses
// a frame that is not a SESSION REQUEST, and refuses a SESSION REQUEST whose
// names are malformed, without hanging or panicking.
func TestNBTTransport_AcceptSessionRejectsMalformed(t *testing.T) {
	cases := []struct {
		name  string
		frame []byte
	}{
		// A SESSION MESSAGE (0x00) where a SESSION REQUEST is required.
		{"wrong message type", []byte{0x00, 0x00, 0x00, 0x02, 0xFF, 0x53}},
		// A SESSION REQUEST with an empty body.
		{"empty body", []byte{0x81, 0x00, 0x00, 0x00}},
		// A SESSION REQUEST whose CALLED name has a bad length byte.
		{"bad called name", append([]byte{0x81, 0x00, 0x00, 0x22, 0x10}, bytes.Repeat([]byte{'A'}, 33)...)},
		// A well-formed CALLED name with no CALLING name following it.
		{"missing calling name", func() []byte {
			called := sessionServiceNameBytes(0x20)
			frame := []byte{0x81, 0x00, 0x00, byte(len(called))}
			return append(frame, called...)
		}()},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			addr, results, cleanup := serveAcceptSession(t, nil)
			defer cleanup()

			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				t.Fatalf("dial error = %v", err)
			}
			defer conn.Close()

			if _, err := conn.Write(tc.frame); err != nil {
				t.Fatalf("write error = %v", err)
			}

			got := <-results
			if got.err == nil {
				if got.transport != nil {
					got.transport.Close()
				}
				t.Fatalf("AcceptSession() should reject %s", tc.name)
			}
		})
	}
}

// sessionRequestBytes builds a SESSION REQUEST frame carrying the given CALLED
// and CALLING names with explicit service suffixes, so a test can address a
// service other than the one the client-side helpers assume.
func sessionRequestBytes(t *testing.T, calledName string, calledSuffix byte, callingName string, callingSuffix byte) []byte {
	t.Helper()

	called, err := nbns.EncodeSessionServiceName(calledName, calledSuffix)
	if err != nil {
		t.Fatalf("EncodeSessionServiceName(%q) error = %v", calledName, err)
	}
	calling, err := nbns.EncodeSessionServiceName(callingName, callingSuffix)
	if err != nil {
		t.Fatalf("EncodeSessionServiceName(%q) error = %v", callingName, err)
	}

	length := len(called) + len(calling)
	frame := []byte{0x81, 0x00, byte((length >> 8) & 0xFF), byte(length & 0xFF)}
	frame = append(frame, called...)
	return append(frame, calling...)
}

// sessionServiceNameBytes returns one encoded session-service name for use in
// hand-built frames.
func sessionServiceNameBytes(suffix byte) []byte {
	encoded, err := nbns.EncodeSessionServiceName("FILESERVER", suffix)
	if err != nil {
		panic(err)
	}
	return encoded
}
