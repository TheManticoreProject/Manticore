package server

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// Handler observes or answers an inbound request before the built-in command
// dispatch sees it.
//
// Run reports whether it handled the request: true stops the chain and the
// request is not dispatched, so a handler that returns true is responsible for
// having written a response. false passes the request to the next handler and
// ultimately to the dispatch table, which makes a handler that only observes
// (logging, capture, packet description) a one-liner.
//
// Handlers are shared across connections and run on the goroutine that owns the
// connection, so an implementation that keeps state must guard it.
type Handler interface {
	Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool

// Run calls f.
func (f HandlerFunc) Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
	return f(srv, conn, w, req)
}

// ResponseWriter sends a response correlated to the request being handled. It
// copies the request's TID, UID, PID and MID into the reply, sets
// SMB_FLAGS_REPLY, mirrors the negotiated Flags2 bits, and frames the result on
// the connection's transport, so a caller only has to build the command body.
type ResponseWriter interface {
	// RemoteAddr returns the address of the client being answered.
	RemoteAddr() net.Addr

	// WriteResponse sends a successful response carrying cmd. It may be called
	// more than once for a request whose command is defined to produce several
	// responses, such as SMB_COM_ECHO.
	WriteResponse(cmd command_interface.CommandInterface) error

	// WriteResponseWithStatus sends a response carrying cmd together with a
	// status other than success. An interim response needs this: a status such
	// as STATUS_MORE_PROCESSING_REQUIRED travels alongside a payload rather than
	// instead of one, which is how a multi-leg authentication reports that it is
	// unfinished.
	WriteResponseWithStatus(cmd command_interface.CommandInterface, status nt_status.NT_STATUS) error

	// WriteError sends an error response: a header carrying status, with no
	// command payload (WordCount 0, ByteCount 0). The status is encoded in
	// whichever form the request selected.
	WriteError(status nt_status.NT_STATUS) error

	// SetResponseUID overrides the user identifier echoed in responses to this
	// request. A request arrives with the UID the client knows, which is zero
	// until the server assigns one, so the leg that assigns it has to say so.
	SetResponseUID(uid uint16)

	// SignResponse signs responses to this request with the given key and
	// sequence number.
	//
	// The dispatch loop arms this for a connection that is already signing. The
	// exchange that establishes signing has to arm it itself, because the key
	// does not exist until that exchange has been verified.
	SignResponse(macKey []byte, sequenceNumber uint32)
}

// responseWriter is the ResponseWriter bound to one request on one connection.
type responseWriter struct {
	conn    *Connection
	request *message.Message

	// uid, when set, replaces the request's UID in responses.
	uid    uint16
	uidSet bool

	// signKey and signSequence, when set, sign responses to this request.
	signKey      []byte
	signSequence uint32
}

// RemoteAddr returns the address of the client being answered.
func (w *responseWriter) RemoteAddr() net.Addr {
	return w.conn.Remote
}

// WriteResponse sends a successful response carrying cmd.
func (w *responseWriter) WriteResponse(cmd command_interface.CommandInterface) error {
	return w.write(cmd, nt_status.NT_STATUS_SUCCESS)
}

// WriteResponseWithStatus sends a response carrying cmd and a non-success status.
func (w *responseWriter) WriteResponseWithStatus(cmd command_interface.CommandInterface, status nt_status.NT_STATUS) error {
	return w.write(cmd, status)
}

// SetResponseUID overrides the UID echoed in responses to this request.
func (w *responseWriter) SetResponseUID(uid uint16) {
	w.uid = uid
	w.uidSet = true
}

// SignResponse signs responses to this request with the given key and sequence
// number.
func (w *responseWriter) SignResponse(macKey []byte, sequenceNumber uint32) {
	w.signKey = macKey
	w.signSequence = sequenceNumber
}

// WriteError sends a payload-less error response carrying status.
func (w *responseWriter) WriteError(status nt_status.NT_STATUS) error {
	return w.write(newErrorResponse(w.request.Header.Command), status)
}

// write builds the reply, marshals it and frames it on the transport.
func (w *responseWriter) write(cmd command_interface.CommandInterface, status nt_status.NT_STATUS) error {
	if cmd == nil {
		return fmt.Errorf("cannot write a response with no command")
	}

	reply := message.NewMessage()
	reply.Header = replyHeader(w.request.Header, status, len(w.signKey) > 0)
	if w.uidSet {
		reply.Header.UID = types.USHORT(w.uid)
	}

	// A command marshals its string fields according to the message's Unicode
	// setting, so the response must use the same encoding the request did.
	cmd.SetUnicode(w.request.Header.Flags2.IsUnicode())
	reply.AddCommand(cmd)

	// AddCommand takes the header's command code from the command it is given.
	// The reply must echo the request's code regardless, so that a handler
	// returning a mismatched command cannot desynchronize the client.
	reply.Header.Command = w.request.Header.Command

	marshalled, err := reply.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal the response: %v", err)
	}

	// Signing happens after marshalling and over the whole message, because the
	// signature occupies a field inside the header it covers.
	if len(w.signKey) > 0 {
		signing.Sign(w.signKey, marshalled, w.signSequence)
	}

	if _, err := w.conn.Transport.Send(marshalled); err != nil {
		return fmt.Errorf("failed to send the response: %v", err)
	}
	return nil
}

// Compile-time assurance that responseWriter satisfies the contract.
var _ ResponseWriter = (*responseWriter)(nil)
