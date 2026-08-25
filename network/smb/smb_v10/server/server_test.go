package server

import (
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
)

// TestNewServerRejectsInvalidConfig asserts a configuration that cannot be
// honoured is refused at construction rather than misbehaving later.
func TestNewServerRejectsInvalidConfig(t *testing.T) {
	if _, err := NewServer(Config{MaxConnections: -1}); err == nil {
		t.Fatal("NewServer() should reject a negative MaxConnections")
	}
	if _, err := NewServer(Config{Timeout: -time.Second}); err == nil {
		t.Fatal("NewServer() should reject a negative Timeout")
	}
	if _, err := NewServer(Config{}); err != nil {
		t.Fatalf("NewServer() with the zero config error = %v", err)
	}
}

// TestServerListeningAndAddr asserts the server reports its listening state and
// bound address, which is how a caller discovers the port it asked to have
// chosen for it.
func TestServerListeningAndAddr(t *testing.T) {
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if srv.Listening() {
		t.Fatal("a server that has not been given a listener reports itself listening")
	}
	if srv.Addr() != nil {
		t.Fatalf("Addr() = %v before serving, want nil", srv.Addr())
	}

	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(listener) }()

	waitFor(t, func() bool { return srv.Listening() }, "server did not report itself listening")

	addr, ok := srv.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("Addr() = %T, want *net.TCPAddr", srv.Addr())
	}
	if addr.Port == 0 {
		t.Fatal("Addr() reports port 0, so the bound port was not resolved")
	}

	if err := srv.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := <-served; err != nil {
		t.Fatalf("Serve() returned %v after Close(), want nil", err)
	}
	if srv.Listening() {
		t.Fatal("server still reports itself listening after Close()")
	}
}

// TestServerCloseIsIdempotent asserts Close can be called more than once, which
// matters because a caller commonly defers it and also calls it on a shutdown
// path.
func TestServerCloseIsIdempotent(t *testing.T) {
	srv, addr := testServer(t)
	// Establish a connection so Close has something to tear down.
	dialServer(t, addr)
	waitFor(t, func() bool { return srv.Connections() == 1 }, "server did not register the connection")

	if err := srv.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := srv.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if got := srv.Connections(); got != 0 {
		t.Fatalf("Connections() = %d after Close(), want 0", got)
	}
}

// TestServerClosesLiveConnections asserts Close tears down connections that are
// parked in a read, rather than waiting for clients to disconnect. Without that,
// Close would block for as long as a client cared to stay idle.
func TestServerClosesLiveConnections(t *testing.T) {
	// No read timeout, so the connection goroutine is parked indefinitely and
	// only Close can free it.
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	go func() { _ = srv.Serve(listener) }()

	client := dialServer(t, listener.Addr().String())
	waitFor(t, func() bool { return srv.Connections() == 1 }, "server did not register the connection")

	closed := make(chan error, 1)
	go func() { closed <- srv.Close() }()

	select {
	case err := <-closed:
		if err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close() blocked on an idle connection")
	}

	// The client's side of the connection is now gone.
	client.SetTimeout(2 * time.Second)
	if _, err := client.Receive(); err == nil {
		t.Fatal("client connection survived the server's Close()")
	}
}

// TestServerRejectsServeAfterClose asserts a listener handed to a closed server
// is refused and closed, rather than being accepted into a server that will
// never serve it.
func TestServerRejectsServeAfterClose(t *testing.T) {
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if err := srv.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	if err := srv.Serve(listener); err == nil {
		t.Fatal("Serve() on a closed server should fail")
	}
	// The listener was closed on the way out, so binding it again is not needed
	// and accepting from it must fail.
	if _, _, err := listener.Accept(); err == nil {
		t.Fatal("Serve() on a closed server left the listener open")
	}

	if err := srv.Serve(nil); err == nil {
		t.Fatal("Serve(nil) should fail")
	}
}

// TestServerMaxConnections asserts the connection limit is enforced: a client
// arriving while the server is full is closed immediately, and a slot freed by a
// disconnect is reused.
func TestServerMaxConnections(t *testing.T) {
	srv, err := NewServer(Config{MaxConnections: 1})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	addr := listener.Addr().String()
	first := dialServer(t, addr)
	waitFor(t, func() bool { return srv.Connections() == 1 }, "server did not register the first connection")

	// The second connection is accepted by the kernel and then closed by the
	// server, so the request goes out but no answer comes back.
	second := dialServer(t, addr)
	sendRequest(t, second, echoRequest("over the limit"))
	second.SetTimeout(2 * time.Second)
	if raw, err := second.Receive(); err == nil {
		t.Fatalf("a connection over the limit was served (% x)", raw)
	}

	// The first connection is unaffected.
	sendRequest(t, first, echoRequest("within the limit"))
	response, _ := receiveResponse(t, first)
	if response.Header.Status != 0 {
		t.Fatalf("Status = 0x%08X on the accepted connection, want success", response.Header.Status)
	}

	// Freeing the slot lets a new client in.
	first.Close()
	waitFor(t, func() bool { return srv.Connections() == 0 }, "server did not release the connection")

	third := dialServer(t, addr)
	sendRequest(t, third, echoRequest("after the slot was freed"))
	if response, _ := receiveResponse(t, third); response.Header.Status != 0 {
		t.Fatalf("Status = 0x%08X on the reused slot, want success", response.Header.Status)
	}
}

// TestServerTimeoutClosesIdleConnection asserts the configured timeout bounds an
// idle connection, so a client that connects and says nothing does not hold a
// goroutine indefinitely.
func TestServerTimeoutClosesIdleConnection(t *testing.T) {
	srv, err := NewServer(Config{Timeout: 250 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	dialServer(t, listener.Addr().String())
	waitFor(t, func() bool { return srv.Connections() == 1 }, "server did not register the connection")
	waitFor(t, func() bool { return srv.Connections() == 0 }, "server did not close the idle connection")
}

// TestServerOverNBT asserts the server serves a connection that arrived over the
// NetBIOS session service, so both listeners are interchangeable above the
// transport.
func TestServerOverNBT(t *testing.T) {
	srv, err := NewServer(Config{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener, err := transport.ListenNBT("127.0.0.1:0", []string{"MANTICORE"})
	if err != nil {
		t.Fatalf("ListenNBT() error = %v", err)
	}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	addr := listener.Addr().(*net.TCPAddr)
	client := nbt.NewNBTTransport()
	client.SetCalledName("MANTICORE")
	client.SetCallingName("TESTCLIENT")
	client.SetTimeout(5 * time.Second)
	if err := client.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	defer client.Close()

	request := echoRequest("over netbios")
	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}
	if _, err := client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode the response: %v", err)
	}
	assertReplyHeader(t, request, response)

	echoResponse, ok := response.Command.(*commands.EchoResponse)
	if !ok {
		t.Fatalf("response command is %T, want *commands.EchoResponse", response.Command)
	}
	if string(echoResponse.Data) != "over netbios" {
		t.Fatalf("echoed data = %q", echoResponse.Data)
	}
}

// TestRegisterHandlerIgnoresNil asserts a nil handler is not appended, so the
// dispatch loop cannot be made to call one.
func TestRegisterHandlerIgnoresNil(t *testing.T) {
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	srv.RegisterHandler(nil)
	if got := len(srv.snapshotHandlers()); got != 0 {
		t.Fatalf("snapshotHandlers() has %d entries after registering nil, want 0", got)
	}
}

// TestConfigIsReported asserts the server reports back the configuration it was
// built with.
func TestConfigIsReported(t *testing.T) {
	want := Config{Timeout: 3 * time.Second, MaxConnections: 7}
	srv, err := NewServer(want)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if got := srv.Config(); got != want {
		t.Fatalf("Config() = %+v, want %+v", got, want)
	}
}

// echoRequest builds a one-response SMB_COM_ECHO request carrying payload.
func echoRequest(payload string) *message.Message {
	request := newRequest(codes.SMB_COM_ECHO)
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte(payload)
	request.AddCommand(echo)
	return request
}

// waitFor polls a condition until it holds, failing the test if it does not
// within a couple of seconds. Server state changes on another goroutine, so a
// test cannot read it synchronously.
func waitFor(t *testing.T, condition func() bool, message string) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal(message)
}
