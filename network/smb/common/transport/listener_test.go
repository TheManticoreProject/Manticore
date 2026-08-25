package transport_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/tcp"
)

// acceptResult carries the outcome of a Listener.Accept run back to the test
// goroutine.
type acceptResult struct {
	transport transport.Transport
	remote    net.Addr
	err       error
}

// acceptOnce runs Accept in the background and delivers its outcome, so a test
// can drive the client side on the main goroutine.
func acceptOnce(l transport.Listener) <-chan acceptResult {
	ch := make(chan acceptResult, 1)
	go func() {
		tr, remote, err := l.Accept()
		ch <- acceptResult{tr, remote, err}
	}()
	return ch
}

// exchange asserts a bidirectional SMB-message round trip over an established
// client/server transport pair, which is the contract a Listener promises: the
// transport it returns is ready to carry SMB messages in both directions.
func exchange(t *testing.T, client, server transport.Transport) {
	t.Helper()

	request := []byte{0xFF, 'S', 'M', 'B', 0x72, 0x00, 0x00, 0x00}
	if _, err := client.Send(request); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	got, err := server.Receive()
	if err != nil {
		t.Fatalf("server Receive() error = %v", err)
	}
	if !bytes.Equal(got, request) {
		t.Fatalf("server received % x, want % x", got, request)
	}

	response := []byte{0xFF, 'S', 'M', 'B', 0x72, 0x01}
	if _, err := server.Send(response); err != nil {
		t.Fatalf("server Send() error = %v", err)
	}
	back, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if !bytes.Equal(back, response) {
		t.Fatalf("client received % x, want % x", back, response)
	}
}

// TestListenTCPRoundTrip verifies a Direct TCP listener hands back a connected
// transport that round-trips SMB messages with the client-side transport.
func TestListenTCPRoundTrip(t *testing.T) {
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	defer listener.Close()

	results := acceptOnce(listener)

	addr := listener.Addr().(*net.TCPAddr)
	client := tcp.NewTCPTransport()
	client.SetTimeout(5 * time.Second)
	if err := client.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	defer client.Close()

	got := <-results
	if got.err != nil {
		t.Fatalf("Accept() error = %v", got.err)
	}
	defer got.transport.Close()

	if !got.transport.IsConnected() {
		t.Fatal("Accept() returned a transport that reports itself disconnected")
	}
	if got.remote == nil {
		t.Fatal("Accept() returned a nil remote address")
	}
	got.transport.SetTimeout(5 * time.Second)

	exchange(t, client, got.transport)
}

// TestListenNBTRoundTrip verifies an NBT listener completes the RFC 1002 session
// handshake before returning, so the transport it hands back is immediately
// usable for SMB messages.
func TestListenNBTRoundTrip(t *testing.T) {
	listener, err := transport.ListenNBT("127.0.0.1:0", []string{"FILESERVER"})
	if err != nil {
		t.Fatalf("ListenNBT() error = %v", err)
	}
	defer listener.Close()

	results := acceptOnce(listener)

	addr := listener.Addr().(*net.TCPAddr)
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
		t.Fatalf("Accept() error = %v", got.err)
	}
	defer got.transport.Close()
	got.transport.SetTimeout(5 * time.Second)

	exchange(t, client, got.transport)
}

// TestListenNBTSkipsFailedHandshake verifies a connection that fails the NetBIOS
// handshake is discarded rather than surfaced from Accept, so one bad client
// cannot take the listener down: the next well-formed client is still served.
func TestListenNBTSkipsFailedHandshake(t *testing.T) {
	listener, err := transport.ListenNBT("127.0.0.1:0", []string{"FILESERVER"})
	if err != nil {
		t.Fatalf("ListenNBT() error = %v", err)
	}
	defer listener.Close()

	results := acceptOnce(listener)
	addr := listener.Addr().(*net.TCPAddr)

	// A client asking for a CALLED name this endpoint does not serve is refused
	// during the handshake. Connect fails, and Accept must still be waiting.
	refused := nbt.NewNBTTransport()
	refused.SetCalledName("OTHERHOST")
	refused.SetTimeout(5 * time.Second)
	if err := refused.Connect(addr.IP, addr.Port); err == nil {
		refused.Close()
		t.Fatal("Connect() with an unserved CALLED name should fail")
	}

	select {
	case got := <-results:
		if got.transport != nil {
			got.transport.Close()
		}
		t.Fatalf("Accept() returned after a failed handshake (err = %v), it should keep waiting", got.err)
	case <-time.After(150 * time.Millisecond):
	}

	// A well-formed client is still served on the same listener.
	client := nbt.NewNBTTransport()
	client.SetCalledName("FILESERVER")
	client.SetTimeout(5 * time.Second)
	if err := client.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("client Connect() after a failed handshake error = %v", err)
	}
	defer client.Close()

	got := <-results
	if got.err != nil {
		t.Fatalf("Accept() error = %v", got.err)
	}
	defer got.transport.Close()
	got.transport.SetTimeout(5 * time.Second)

	exchange(t, client, got.transport)
}

// TestListenerCloseUnblocksAccept verifies Close makes a blocked Accept return an
// error, which is how a server's accept loop terminates.
func TestListenerCloseUnblocksAccept(t *testing.T) {
	listeners := map[string]func() (transport.Listener, error){
		"tcp": func() (transport.Listener, error) { return transport.ListenTCP("127.0.0.1:0") },
		"nbt": func() (transport.Listener, error) { return transport.ListenNBT("127.0.0.1:0", nil) },
	}
	for name, newListener := range listeners {
		newListener := newListener
		t.Run(name, func(t *testing.T) {
			listener, err := newListener()
			if err != nil {
				t.Fatalf("listen error = %v", err)
			}

			results := acceptOnce(listener)
			if err := listener.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			select {
			case got := <-results:
				if got.err == nil {
					got.transport.Close()
					t.Fatal("Accept() should fail once the listener is closed")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Accept() did not return after Close()")
			}
		})
	}
}

// TestListenDefaultPorts verifies an address with no port is bound on the
// dialect's default SMB port. Binding 445 and 139 needs privileges the test
// environment may not have, so an inability to bind is skipped rather than
// failed; what matters is that a successful bind lands on the right port.
func TestListenDefaultPorts(t *testing.T) {
	cases := []struct {
		name string
		open func(string) (transport.Listener, error)
		port string
	}{
		{"tcp", transport.ListenTCP, transport.DefaultDirectTCPPort},
		{"nbt", func(addr string) (transport.Listener, error) { return transport.ListenNBT(addr, nil) }, transport.DefaultNBTPort},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			listener, err := tc.open("127.0.0.1")
			if err != nil {
				t.Skipf("cannot bind the default port %s: %v", tc.port, err)
			}
			defer listener.Close()

			_, port, err := net.SplitHostPort(listener.Addr().String())
			if err != nil {
				t.Fatalf("SplitHostPort(%q) error = %v", listener.Addr().String(), err)
			}
			if port != tc.port {
				t.Fatalf("bound port = %s, want the default %s", port, tc.port)
			}
		})
	}
}
