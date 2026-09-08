package server

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/tcp"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// deferringHandler is a Handler that defers every ECHO it sees and answers it
// from another goroutine when told to.
//
// ECHO is used because it needs no session, no share and no handle, so the test
// exercises the deferral machinery and nothing else.
type deferringHandler struct {
	// deferred receives what the test needs about each intercepted request, so
	// the test drives the answer rather than racing it.
	deferred chan deferredRequest
}

// deferredRequest is what the handler saw, published so the test can build a
// cancel that names the same request rather than guessing the client's
// identifiers.
type deferredRequest struct {
	responder  AsyncResponder
	connection *Connection
	pid        uint32
	mid        uint16
}

func (h *deferringHandler) Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
	if req.Header.Command != codes.SMB_COM_ECHO {
		return false
	}
	h.deferred <- deferredRequest{
		responder:  w.Defer(),
		connection: conn,
		pid:        req.Header.GetPID(),
		mid:        uint16(req.Header.MID),
	}
	// Handled: the responder owns the answer now, and writing one here as well
	// would put two replies on the connection for one request.
	return true
}

// deferringServer starts a server with a deferring handler registered.
func deferringServer(t *testing.T) (*Server, *smb1client.Client, *deferringHandler) {
	t.Helper()

	// The client is fully authenticated first: SMB_COM_ECHO needs no session to
	// be dispatched, but the client refuses to send one without having
	// established the connection, and a failure there would look like a handler
	// that never fired.
	srv, client := pipedClient(t, conformanceConfig(SigningDisabled), true)

	// Registered after the exchange, so nothing in it is intercepted.
	handler := &deferringHandler{deferred: make(chan deferredRequest, 4)}
	srv.RegisterHandler(handler)

	return srv, client, handler
}

// TestDeferredAnswerReachesTheClient asserts a response sent after the handler
// returned arrives, and carries the identifiers of the request it answers.
//
// Correlation is the whole point: a client matches a reply to a request by MID, so
// a deferred answer that lost it would be read against the wrong request.
func TestDeferredAnswerReachesTheClient(t *testing.T) {
	_, client, handler := deferringServer(t)

	payload := []byte("answered later")

	// Sent from a goroutine, because the client's Echo blocks until the answer
	// arrives and the answer is this test's job to produce.
	type echoResult struct {
		data []byte
		err  error
	}
	results := make(chan echoResult, 1)
	go func() {
		data, err := client.Echo(payload)
		results <- echoResult{data: data, err: err}
	}()

	var deferred deferredRequest
	select {
	case deferred = <-handler.deferred:
	case <-time.After(5 * time.Second):
		t.Fatal("the handler never deferred the request")
	}
	responder := deferred.responder

	response := commands.NewEchoResponse()
	response.SequenceNumber = types.USHORT(1)
	response.Data = []types.UCHAR(payload)
	if err := responder.Respond(response); err != nil {
		t.Fatalf("Respond() error = %v", err)
	}

	select {
	case result := <-results:
		if result.err != nil {
			t.Fatalf("the client did not receive the deferred answer: %v", result.err)
		}
		if !bytes.Equal(result.data, payload) {
			t.Errorf("the deferred answer carried %q, want %q", result.data, payload)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the deferred answer never reached the client")
	}
}

// TestDeferredAnswerIsTrackedUntilItFires asserts the connection knows what it
// still owes, and forgets it once answered.
//
// A responder that is never released would keep the connection remembering it for
// as long as the connection lives, so the accounting is what makes a leak visible.
func TestDeferredAnswerIsTrackedUntilItFires(t *testing.T) {
	_, client, handler := deferringServer(t)

	go client.Echo([]byte("tracked"))

	deferred := <-handler.deferred
	responder, conn := deferred.responder, deferred.connection
	waitFor(t, func() bool { return conn.OutstandingResponses() == 1 },
		"the deferred request was not recorded as outstanding")

	response := commands.NewEchoResponse()
	response.SequenceNumber = types.USHORT(1)
	response.Data = []types.UCHAR("tracked")
	if err := responder.Respond(response); err != nil {
		t.Fatalf("Respond() error = %v", err)
	}

	waitFor(t, func() bool { return conn.OutstandingResponses() == 0 },
		"the deferred request was not released after answering")
}

// TestDeferredAnswerRefusesASecondSend asserts a responder answers once.
//
// A request has one response, and a second would leave a reply on the connection
// that no request accounts for — after which the client reads every later
// response against the wrong request.
func TestDeferredAnswerRefusesASecondSend(t *testing.T) {
	_, client, handler := deferringServer(t)

	go client.Echo([]byte("once"))
	responder := (<-handler.deferred).responder

	response := commands.NewEchoResponse()
	response.SequenceNumber = types.USHORT(1)
	response.Data = []types.UCHAR("once")
	if err := responder.Respond(response); err != nil {
		t.Fatalf("the first Respond() failed: %v", err)
	}

	second := commands.NewEchoResponse()
	second.SequenceNumber = types.USHORT(1)
	second.Data = []types.UCHAR("twice")
	if err := responder.Respond(second); err == nil {
		t.Error("a second Respond() succeeded, so two replies went out for one request")
	}

	// Done after answering is harmless, so it is safe in a defer.
	responder.Done()
}

// TestNtCancelCancelsAnOutstandingRequest asserts SMB_COM_NT_CANCEL reaches the
// deferred request it names.
//
// The command was previously accepted and ignored, which was only correct because
// nothing could be outstanding. Now that something can be, it has to arrive.
func TestNtCancelCancelsAnOutstandingRequest(t *testing.T) {
	_, client, handler := deferringServer(t)

	go client.Echo([]byte("cancel me"))
	deferred := <-handler.deferred
	responder := deferred.responder

	// The cancel has to name the same PID and MID the deferred request carried,
	// which the handler published rather than the test guessing at the client's
	// numbering.
	request := newRequest(codes.SMB_COM_NT_CANCEL)
	// The cancel needs the session's UID: SMB_COM_NT_CANCEL is not one of the
	// commands that may arrive without a session, so the dispatcher refuses it
	// before the handler sees it otherwise.
	request.Header.UID = client.Session.SessionUID
	request.Header.SetPID(deferred.pid)
	request.Header.MID = types.USHORT(deferred.mid)
	request.AddCommand(commands.NewNtCancelRequest())

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the cancel: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	select {
	case <-responder.Cancelled():
	case <-time.After(5 * time.Second):
		t.Fatal("the cancel never reached the deferred request")
	}

	// The request still has to be released, and answering a cancelled request is
	// the responder's choice: here it reports the cancellation.
	if err := responder.RespondWithError(nt_status.NT_STATUS_CANCELLED); err != nil {
		t.Fatalf("RespondWithError() error = %v", err)
	}
}

// TestClosingTheConnectionCancelsWhatItOwes asserts a deferred answer is told to
// stop when the connection goes away, rather than discovering it on a failed write.
func TestClosingTheConnectionCancelsWhatItOwes(t *testing.T) {
	srv, client, handler := deferringServer(t)

	go client.Echo([]byte("abandoned"))
	responder := (<-handler.deferred).responder

	if err := srv.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case <-responder.Cancelled():
	case <-time.After(5 * time.Second):
		t.Fatal("closing the connection did not cancel the deferred request")
	}
}

// TestUnsolicitedSendGoesUnsignedOnASigningConnection asserts an oplock break is
// sent, and sent without a signature, on a connection that is signing.
//
// [MS-CIFS] section 3.3.4.1 requires exactly that: a message the server sends is
// signed with the number in ServerSendSequenceNumber[PID,MID], and then "OpLock
// Break Notification messages are exempt from signing". The exemption is what
// makes the message possible — that table is keyed by a request's PID and MID,
// and a message with no request behind it has no entry — and sending it unsigned
// consumes no number, so the two sides' numbering stays in step.
func TestUnsolicitedSendGoesUnsignedOnASigningConnection(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	t.Cleanup(func() {
		serverSide.Close()
		clientSide.Close()
	})

	conn := &Connection{
		Transport: tcp.NewTCPTransportFromConn(serverSide),
		Remote:    &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1},
		// Signing active, with a key that would produce a visible signature if
		// one were applied.
		SigningActive: true,
		SigningKey:    bytes.Repeat([]byte{0xAB}, 16),
	}

	clientTransport := tcp.NewTCPTransportFromConn(clientSide)
	clientTransport.SetTimeout(5 * time.Second)

	received := make(chan []byte, 1)
	go func() {
		frame, err := clientTransport.Receive()
		if err != nil {
			received <- nil
			return
		}
		received <- frame
	}()

	notification := commands.NewLockingAndxRequest()
	notification.FID = types.USHORT(0x1234)
	notification.TypeOfLock = types.UCHAR(commands.LockingAndxOplockRelease)
	if err := conn.SendUnsolicited(notification, 1, 1, conn.nextUnsolicitedMID()); err != nil {
		t.Fatalf("the break was refused on a signing connection: %v", err)
	}

	var frame []byte
	select {
	case frame = <-received:
	case <-time.After(5 * time.Second):
		t.Fatal("the break was never sent")
	}
	if frame == nil {
		t.Fatal("reading the break failed")
	}

	// The SecuritySignature field is bytes 14 to 22 of the SMB header. An exempt
	// message leaves it zero.
	if len(frame) < header.SMB_HEADER_SIZE {
		t.Fatalf("the frame is %d bytes, shorter than an SMB header", len(frame))
	}
	signature := frame[14:22]
	if !bytes.Equal(signature, make([]byte, 8)) {
		t.Errorf("the break carries the signature % x, want it left zero as an exempt message", signature)
	}
}
