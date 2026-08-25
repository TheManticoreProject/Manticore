package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// testServer starts a server on an ephemeral loopback port with the given
// handlers registered, and returns it with the address to dial.
func testServer(t *testing.T, handlers ...Handler) (*Server, string) {
	t.Helper()

	srv, err := NewServer(Config{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	for _, handler := range handlers {
		srv.RegisterHandler(handler)
	}

	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	addr := listener.Addr().String()

	go func() {
		// Serve returns nil once the server is closed.
		_ = srv.Serve(listener)
	}()
	t.Cleanup(func() { srv.Close() })

	return srv, addr
}

// dialServer connects a client transport to the server.
func dialServer(t *testing.T, addr string) *tcp.TCPTransport {
	t.Helper()

	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort(%q) error = %v", addr, err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("Atoi(%q) error = %v", portText, err)
	}

	client := tcp.NewTCPTransport()
	client.SetTimeout(5 * time.Second)
	if err := client.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// newRequest builds a request message with the identifiers a test asserts are
// echoed back, and the Flags2 bits a modern client negotiates.
func newRequest(command codes.CommandCode) *message.Message {
	request := message.NewMessage()
	request.Header.Command = command
	request.Header.Flags = flags.Flags(flags.FLAGS_CASE_INSENSITIVE | flags.FLAGS_CANONICALIZED_PATHS)
	request.Header.Flags2 = flags2.Flags2(flags2.FLAGS2_UNICODE |
		flags2.FLAGS2_NT_STATUS_ERROR_CODES |
		flags2.FLAGS2_LONG_NAMES_ALLOWED)
	request.Header.SetPID(0x00001234)
	request.Header.TID = 0x0011
	request.Header.UID = 0x0022
	request.Header.MID = 0x0033
	return request
}

// receiveResponse reads one response frame and decodes it.
func receiveResponse(t *testing.T, client *tcp.TCPTransport) (*message.Message, []byte) {
	t.Helper()

	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode the response (% x): %v", raw, err)
	}
	return response, raw
}

// sendRequest marshals and sends a request.
func sendRequest(t *testing.T, client *tcp.TCPTransport, request *message.Message) {
	t.Helper()

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}
	if _, err := client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
}

// assertReplyHeader asserts a response carries the reply flag and echoes the
// request's identifiers, which is what lets a client correlate it.
func assertReplyHeader(t *testing.T, request, response *message.Message) {
	t.Helper()

	if !response.Header.Flags.IsReply() {
		t.Fatalf("response Flags = 0x%02X, SMB_FLAGS_REPLY is not set", uint8(response.Header.Flags))
	}
	if !response.Header.IsResponse() {
		t.Fatal("response does not report itself as a response")
	}
	if response.Header.Command != request.Header.Command {
		t.Fatalf("response command = 0x%02X, want 0x%02X", uint8(response.Header.Command), uint8(request.Header.Command))
	}
	if response.Header.GetPID() != request.Header.GetPID() {
		t.Fatalf("response PID = 0x%08X, want 0x%08X", response.Header.GetPID(), request.Header.GetPID())
	}
	if response.Header.TID != request.Header.TID {
		t.Fatalf("response TID = 0x%04X, want 0x%04X", response.Header.TID, request.Header.TID)
	}
	if response.Header.UID != request.Header.UID {
		t.Fatalf("response UID = 0x%04X, want 0x%04X", response.Header.UID, request.Header.UID)
	}
	if response.Header.MID != request.Header.MID {
		t.Fatalf("response MID = 0x%04X, want 0x%04X", response.Header.MID, request.Header.MID)
	}
	if !bytes.Equal(response.Header.Protocol[:], smbMagic) {
		t.Fatalf("response protocol identifier = % x, want % x", response.Header.Protocol[:], smbMagic)
	}
}

// TestEchoRoundTrip drives SMB_COM_ECHO end to end and asserts the response
// carries the reply header, the echoed data and the sequence number.
func TestEchoRoundTrip(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	payload := []byte("manticore echo payload")
	request := newRequest(codes.SMB_COM_ECHO)
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = payload
	request.AddCommand(echo)

	sendRequest(t, client, request)
	response, _ := receiveResponse(t, client)

	assertReplyHeader(t, request, response)
	if response.Header.Status != 0 {
		t.Fatalf("response Status = 0x%08X, want success", response.Header.Status)
	}

	echoResponse, ok := response.Command.(*commands.EchoResponse)
	if !ok {
		t.Fatalf("response command is %T, want *commands.EchoResponse", response.Command)
	}
	if echoResponse.SequenceNumber != 1 {
		t.Fatalf("SequenceNumber = %d, want 1", echoResponse.SequenceNumber)
	}
	if !bytes.Equal(echoResponse.Data, payload) {
		t.Fatalf("echoed data = %q, want %q", echoResponse.Data, payload)
	}
}

// TestEchoCountProducesOneResponsePerCount asserts the server sends EchoCount
// responses, numbered from 1, as the command is defined to do.
func TestEchoCountProducesOneResponsePerCount(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	const count = 3
	request := newRequest(codes.SMB_COM_ECHO)
	echo := commands.NewEchoRequest()
	echo.EchoCount = count
	echo.Data = []byte("x")
	request.AddCommand(echo)

	sendRequest(t, client, request)

	for expected := 1; expected <= count; expected++ {
		response, _ := receiveResponse(t, client)
		echoResponse, ok := response.Command.(*commands.EchoResponse)
		if !ok {
			t.Fatalf("response %d command is %T, want *commands.EchoResponse", expected, response.Command)
		}
		if int(echoResponse.SequenceNumber) != expected {
			t.Fatalf("response %d has SequenceNumber %d", expected, echoResponse.SequenceNumber)
		}
	}
}

// TestEchoCountZeroProducesNoResponse asserts an EchoCount of zero is answered
// with silence, per the command definition, and that the connection stays usable
// afterwards.
func TestEchoCountZeroProducesNoResponse(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	silent := newRequest(codes.SMB_COM_ECHO)
	echo := commands.NewEchoRequest()
	echo.EchoCount = 0
	echo.Data = []byte("ignored")
	silent.AddCommand(echo)
	sendRequest(t, client, silent)

	// Nothing should arrive for that request, so a short read must time out.
	client.SetTimeout(250 * time.Millisecond)
	if raw, err := client.Receive(); err == nil {
		t.Fatalf("an EchoCount of 0 produced a response (% x)", raw)
	}

	// The connection is still good: a subsequent request is answered.
	client.SetTimeout(5 * time.Second)
	request := newRequest(codes.SMB_COM_ECHO)
	followUp := commands.NewEchoRequest()
	followUp.EchoCount = 1
	followUp.Data = []byte("still here")
	request.AddCommand(followUp)
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	assertReplyHeader(t, request, response)
}

// TestEchoCountIsCapped asserts a client cannot turn one request into an
// unbounded number of responses.
func TestEchoCountIsCapped(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	request := newRequest(codes.SMB_COM_ECHO)
	echo := commands.NewEchoRequest()
	echo.EchoCount = types.USHORT(MaxEchoCount + 100)
	echo.Data = []byte("y")
	request.AddCommand(echo)
	sendRequest(t, client, request)

	for i := 1; i <= MaxEchoCount; i++ {
		receiveResponse(t, client)
	}

	// The cap is the last response; nothing follows it.
	client.SetTimeout(250 * time.Millisecond)
	if raw, err := client.Receive(); err == nil {
		t.Fatalf("server sent more than the %d-response cap (% x)", MaxEchoCount, raw)
	}
}

// TestUnimplementedCommandIsRefused asserts a recognized command with no handler
// is answered STATUS_NOT_IMPLEMENTED, in a payload-less error response.
func TestUnimplementedCommandIsRefused(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	// SMB_COM_TREE_CONNECT_ANDX is a recognized command that no handler serves yet.
	request := newRequest(codes.SMB_COM_TREE_CONNECT_ANDX)
	request.AddCommand(commands.NewTreeConnectAndxRequest())
	sendRequest(t, client, request)

	response, raw := receiveResponse(t, client)
	assertReplyHeader(t, request, response)

	if response.Header.Status != uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_NOT_IMPLEMENTED (0x%08X)",
			response.Header.Status, uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED))
	}

	// An error response is the header plus WordCount 0 and ByteCount 0.
	wantLength := header.SMB_HEADER_SIZE + 3
	if len(raw) != wantLength {
		t.Fatalf("error response is %d bytes, want %d (header + WordCount 0 + ByteCount 0)", len(raw), wantLength)
	}
	if raw[header.SMB_HEADER_SIZE] != 0x00 {
		t.Fatalf("error response WordCount = 0x%02X, want 0x00", raw[header.SMB_HEADER_SIZE])
	}
	if byteCount := binary.LittleEndian.Uint16(raw[header.SMB_HEADER_SIZE+1:]); byteCount != 0 {
		t.Fatalf("error response ByteCount = %d, want 0", byteCount)
	}
}

// TestUnimplementedCommandUsesLegacyStatusEncoding asserts a client that did not
// negotiate NT status codes receives the legacy SMBSTATUS class/code encoding
// instead of the NTSTATUS value.
func TestUnimplementedCommandUsesLegacyStatusEncoding(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	request := newRequest(codes.SMB_COM_TREE_CONNECT_ANDX)
	// Clear the NT-status bit, leaving an old-style client.
	request.Header.Flags2 &= ^flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES)
	request.AddCommand(commands.NewTreeConnectAndxRequest())
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	// STATUS_NOT_IMPLEMENTED is ERRDOS/ERRbadfunc: class 0x01, code 0x0001.
	if want := uint32(0x00010001); response.Header.Status != want {
		t.Fatalf("Status = 0x%08X, want the legacy encoding 0x%08X", response.Header.Status, want)
	}
	if response.Header.Flags2&flags2.FLAGS2_NT_STATUS_ERROR_CODES != 0 {
		t.Fatal("response set SMB_FLAGS2_NT_STATUS_ERROR_CODES although the request did not")
	}
}

// TestUnknownCommandCodeIsRefused asserts a command code the message layer does
// not recognize is answered ERRSRV/ERRbadcmd rather than being confused with a
// recognized command the server has not implemented.
func TestUnknownCommandCodeIsRefused(t *testing.T) {
	unknown, ok := firstUnknownCommandCode()
	if !ok {
		t.Skip("every command code in 0x00..0xFF is recognized, so there is no unknown code to test")
	}

	_, addr := testServer(t)
	client := dialServer(t, addr)

	// Hand-build the frame: the message layer cannot marshal a command it does
	// not know.
	frame := rawFrame(unknown, flags.Flags(0), flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES), []byte{0x00, 0x00, 0x00})
	if _, err := client.Send(frame); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if len(raw) < header.SMB_HEADER_SIZE {
		t.Fatalf("response is only %d bytes", len(raw))
	}
	status := binary.LittleEndian.Uint32(raw[5:9])
	if status != uint32(nt_status.NT_STATUS_SMB_BAD_COMMAND) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_SMB_BAD_COMMAND (0x%08X)",
			status, uint32(nt_status.NT_STATUS_SMB_BAD_COMMAND))
	}
}

// TestMalformedBodyKeepsConnection asserts a request whose header is intact but
// whose body will not decode is answered with an error and the connection
// survives, so one bad request does not cost a client its session.
func TestMalformedBodyKeepsConnection(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	// A WordCount claiming more parameter words than the frame carries.
	frame := rawFrame(codes.SMB_COM_ECHO, flags.Flags(0), flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES), []byte{0x20, 0x00, 0x00})
	if _, err := client.Send(frame); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != uint32(nt_status.NT_STATUS_INVALID_SMB) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_INVALID_SMB (0x%08X)",
			status, uint32(nt_status.NT_STATUS_INVALID_SMB))
	}

	// The connection is still serving.
	request := newRequest(codes.SMB_COM_ECHO)
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("after the bad frame")
	request.AddCommand(echo)
	sendRequest(t, client, request)
	response, _ := receiveResponse(t, client)
	assertReplyHeader(t, request, response)
}

// TestUnparseableFrameDropsConnection asserts a frame that is not an SMB message
// at all ends the connection rather than being answered: there is nothing to
// correlate a response to, and the frame stream can no longer be trusted.
func TestUnparseableFrameDropsConnection(t *testing.T) {
	cases := []struct {
		name  string
		frame []byte
	}{
		{"short frame", []byte{0xFF, 'S', 'M', 'B'}},
		{"wrong protocol identifier", append([]byte{0xFE, 'S', 'M', 'B'}, make([]byte, 31)...)},
		{"empty frame", []byte{}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, addr := testServer(t)
			client := dialServer(t, addr)

			if _, err := client.Send(tc.frame); err != nil {
				t.Fatalf("client Send() error = %v", err)
			}
			client.SetTimeout(2 * time.Second)
			if raw, err := client.Receive(); err == nil {
				t.Fatalf("server answered an unparseable frame with % x", raw)
			}
		})
	}
}

// TestRequestWithReplyFlagIsRejected asserts an inbound frame claiming to be a
// response is refused. It matters beyond tidiness: the message layer selects
// request or response decoders from that bit, so honouring it on input would let
// a client choose which decoder runs against its bytes.
func TestRequestWithReplyFlagIsRejected(t *testing.T) {
	_, addr := testServer(t)
	client := dialServer(t, addr)

	frame := rawFrame(codes.SMB_COM_ECHO, flags.Flags(flags.FLAGS_REPLY), flags2.Flags2(0), []byte{0x01, 0x00, 0x00, 0x00, 0x00})
	if _, err := client.Send(frame); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	client.SetTimeout(2 * time.Second)
	if raw, err := client.Receive(); err == nil {
		t.Fatalf("server answered a frame carrying SMB_FLAGS_REPLY with % x", raw)
	}
}

// TestHandlerChainInterceptsBeforeDispatch asserts a registered handler sees the
// request first, that returning true suppresses the built-in dispatch, and that
// returning false falls through to it.
func TestHandlerChainInterceptsBeforeDispatch(t *testing.T) {
	t.Run("intercepting handler answers instead of dispatch", func(t *testing.T) {
		intercepted := make(chan codes.CommandCode, 1)
		handler := HandlerFunc(func(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
			intercepted <- req.Header.Command
			// Answer with a status the built-in dispatch would never send for
			// this command, so the test can tell which path replied.
			if err := w.WriteError(nt_status.NT_STATUS_ACCESS_DENIED); err != nil {
				t.Errorf("handler WriteError() error = %v", err)
			}
			return true
		})

		_, addr := testServer(t, handler)
		client := dialServer(t, addr)

		request := newRequest(codes.SMB_COM_ECHO)
		echo := commands.NewEchoRequest()
		echo.EchoCount = 1
		echo.Data = []byte("intercept me")
		request.AddCommand(echo)
		sendRequest(t, client, request)

		response, _ := receiveResponse(t, client)
		if response.Header.Status != uint32(nt_status.NT_STATUS_ACCESS_DENIED) {
			t.Fatalf("Status = 0x%08X, want the handler's ACCESS_DENIED", response.Header.Status)
		}
		select {
		case got := <-intercepted:
			if got != codes.SMB_COM_ECHO {
				t.Fatalf("handler saw command 0x%02X", uint8(got))
			}
		default:
			t.Fatal("handler was not called")
		}
	})

	t.Run("observing handler falls through to dispatch", func(t *testing.T) {
		observed := make(chan struct{}, 1)
		handler := HandlerFunc(func(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
			select {
			case observed <- struct{}{}:
			default:
			}
			return false
		})

		_, addr := testServer(t, handler)
		client := dialServer(t, addr)

		request := newRequest(codes.SMB_COM_ECHO)
		echo := commands.NewEchoRequest()
		echo.EchoCount = 1
		echo.Data = []byte("observe me")
		request.AddCommand(echo)
		sendRequest(t, client, request)

		response, _ := receiveResponse(t, client)
		if response.Header.Status != 0 {
			t.Fatalf("Status = 0x%08X, want the echo handler's success", response.Header.Status)
		}
		if _, ok := response.Command.(*commands.EchoResponse); !ok {
			t.Fatalf("response command is %T, want the dispatched echo response", response.Command)
		}
		select {
		case <-observed:
		default:
			t.Fatal("observing handler was not called")
		}
	})
}

// TestReplyHeaderMirrorsNegotiatedFlags asserts the response mirrors exactly the
// Flags2 bits that record what the two sides agreed to speak, and no others.
func TestReplyHeaderMirrorsNegotiatedFlags(t *testing.T) {
	request := header.NewHeader()
	request.Command = codes.SMB_COM_ECHO
	request.Flags = flags.Flags(flags.FLAGS_CASE_INSENSITIVE | flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_OPLOCK)
	request.Flags2 = flags2.Flags2(flags2.FLAGS2_UNICODE |
		flags2.FLAGS2_NT_STATUS_ERROR_CODES |
		flags2.FLAGS2_LONG_NAMES_ALLOWED |
		flags2.FLAGS2_EXTENDED_SECURITY |
		flags2.FLAGS2_DFS)

	reply := replyHeader(request, nt_status.NT_STATUS_SUCCESS)

	if !reply.Flags.IsReply() {
		t.Fatal("reply does not set SMB_FLAGS_REPLY")
	}
	// FLAGS_OPLOCK is a request-only bit and must not come back.
	if reply.Flags&flags.FLAGS_OPLOCK != 0 {
		t.Fatal("reply mirrored the request-only SMB_FLAGS_OPLOCK bit")
	}
	if reply.Flags&echoedRequestFlags != request.Flags&echoedRequestFlags {
		t.Fatalf("reply Flags = 0x%02X, want the echoed bits from 0x%02X", uint8(reply.Flags), uint8(request.Flags))
	}
	// FLAGS2_DFS is not part of the mirrored set.
	if reply.Flags2&flags2.FLAGS2_DFS != 0 {
		t.Fatal("reply mirrored SMB_FLAGS2_DFS, which is not in the echoed set")
	}
	if reply.Flags2 != request.Flags2&echoedRequestFlags2 {
		t.Fatalf("reply Flags2 = 0x%04X, want 0x%04X", uint16(reply.Flags2), uint16(request.Flags2&echoedRequestFlags2))
	}
	if !reply.Flags2.IsUnicode() {
		t.Fatal("reply dropped the Unicode bit the request set")
	}
}

// TestErrorResponseMarshalsEmptyBody asserts the error-response body is exactly
// WordCount 0 followed by ByteCount 0.
func TestErrorResponseMarshalsEmptyBody(t *testing.T) {
	marshalled, err := newErrorResponse(codes.SMB_COM_ECHO).Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if want := []byte{0x00, 0x00, 0x00}; !bytes.Equal(marshalled, want) {
		t.Fatalf("error body = % x, want % x", marshalled, want)
	}
}

// TestIsKnownCommand asserts the check that separates an unrecognized command
// code from a recognized one the server has not implemented.
func TestIsKnownCommand(t *testing.T) {
	if !isKnownCommand(codes.SMB_COM_ECHO) {
		t.Fatal("SMB_COM_ECHO should be a known command")
	}
	if unknown, ok := firstUnknownCommandCode(); ok && isKnownCommand(unknown) {
		t.Fatalf("0x%02X should not be a known command", uint8(unknown))
	}
}

// firstUnknownCommandCode returns the lowest command code the message layer does
// not recognize, so a test does not have to hardcode one that a later commit
// might implement.
func firstUnknownCommandCode() (codes.CommandCode, bool) {
	for value := 0; value <= 0xFF; value++ {
		command := codes.CommandCode(value)
		if !isKnownCommand(command) {
			return command, true
		}
	}
	return 0, false
}

// rawFrame builds an SMB frame by hand, for the cases the message layer cannot
// produce: an unrecognized command code, a deliberately malformed body, or a
// request carrying the reply flag.
func rawFrame(command codes.CommandCode, flagBits flags.Flags, flags2Bits flags2.Flags2, body []byte) []byte {
	frame := make([]byte, header.SMB_HEADER_SIZE)
	copy(frame, smbMagic)
	frame[4] = byte(command)
	// Status stays zero.
	frame[9] = byte(flagBits)
	binary.LittleEndian.PutUint16(frame[10:12], uint16(flags2Bits))
	// PIDHigh, SecurityFeatures and Reserved stay zero.
	binary.LittleEndian.PutUint16(frame[24:26], 0x0011) // TID
	binary.LittleEndian.PutUint16(frame[26:28], 0x1234) // PIDLow
	binary.LittleEndian.PutUint16(frame[28:30], 0x0022) // UID
	binary.LittleEndian.PutUint16(frame[30:32], 0x0033) // MID
	return append(frame, body...)
}
