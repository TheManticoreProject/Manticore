package server

import (
	"bytes"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// smbMagic is the four-byte protocol identifier every SMB message starts with.
// The message layer's header decoder copies these bytes without checking them,
// so the server validates them itself: a frame that does not carry them is not
// an SMB message and there is nothing to answer.
var smbMagic = []byte{0xFF, 'S', 'M', 'B'}

// Connection is the server-side state of one client connection. It is created by
// the accept loop and owned by the single goroutine that serves the connection,
// so its fields need no locking.
//
// It holds only what the current phase populates. The session and tree tables and
// the signing state arrive with the phases that establish them, rather than being
// declared here in advance.
type Connection struct {
	// Server is the server this connection was accepted by.
	Server *Server

	// Transport carries SMB messages to and from the client, already framed and
	// with any transport-level handshake complete.
	Transport transport.Transport

	// Remote is the client's address, used for logging and for a handler that
	// wants to record who it is talking to.
	Remote net.Addr

	// Dialect is the dialect string selected during negotiation, empty before it.
	Dialect string

	// Negotiated records that a dialect has been agreed. A second NEGOTIATE on
	// one connection is a protocol violation, and every command other than
	// NEGOTIATE and ECHO requires one to have happened.
	Negotiated bool

	// ClientMaxBufferSize and ClientCapabilities are what the client advertised
	// in its session setup, bounding what the server may send back.
	ClientMaxBufferSize uint32
	ClientCapabilities  capabilities.Capabilities

	// UseUnicode and UseNTStatus record what the client negotiated, so a handler
	// does not have to re-derive them from each request header.
	UseUnicode  bool
	UseNTStatus bool

	// ExtendedSecurity records that the client negotiated extended security, and
	// so expects a GSS security blob rather than a challenge in the clear.
	ExtendedSecurity bool

	// SigningActive reports that every request must carry a valid signature and
	// every response must be signed. SigningKey is the MAC key, and
	// ExpectedRequestSequenceNumber the number the next request must be signed
	// at.
	SigningActive                 bool
	SigningKey                    []byte
	ExpectedRequestSequenceNumber uint32

	// sessions are the authenticated sessions on this connection, by UID, and
	// uids allocates those identifiers.
	sessions map[uint16]*Session
	uids     *identifierAllocator

	// currentRequestFrame is the raw frame being handled, kept because a
	// signature covers the message as received: verifying one means hashing those
	// exact bytes rather than a re-marshalling of the decoded fields. It is valid
	// only for the duration of one handler call, which is safe because a single
	// goroutine owns the connection.
	currentRequestFrame []byte

	// pendingAuth holds the authentication exchanges part-way through, keyed by
	// the UID assigned when the challenge was issued.
	//
	// They are keyed rather than held in a single field because a connection may
	// carry several sessions: once one client identity is established, the next
	// session setup starts a new exchange, and a single field could not tell the
	// first leg of the second exchange from the second leg of the first.
	pendingAuth map[uint16]*spnego.AcceptContext
}

// newConnection binds an accepted transport to the server that accepted it.
func newConnection(srv *Server, t transport.Transport, remote net.Addr) *Connection {
	return &Connection{
		Server:      srv,
		Transport:   t,
		Remote:      remote,
		sessions:    make(map[uint16]*Session),
		uids:        newIdentifierAllocator(srv.config.MaxSessionsPerConnection),
		pendingAuth: make(map[uint16]*spnego.AcceptContext),
	}
}

// PendingAuth returns the authentication exchange a UID names while it is still
// in progress, or nil when the UID names no such exchange. A handler uses it to
// tell the second leg of a session setup from the first.
func (c *Connection) PendingAuth(uid uint16) *spnego.AcceptContext {
	if uid == 0 {
		return nil
	}
	return c.pendingAuth[uid]
}

// Close closes the connection's transport. It is safe to call more than once.
func (c *Connection) Close() error {
	if c.Transport == nil {
		return nil
	}
	return c.Transport.Close()
}

// serve runs the receive loop for one connection until the client disconnects,
// the frame stream desynchronizes, or the server is closed.
//
// A panic in a handler is contained here: it takes down this connection and
// nothing else. Without that, a defect reachable from the network would stop the
// whole listener.
func (c *Connection) serve() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("SMB1 server: panic while serving %s: %v", c.Remote, r)
		}
	}()
	defer c.Close()

	logger.Debugf("SMB1 server: serving %s", c.Remote)

	for {
		raw, err := c.Transport.Receive()
		if err != nil {
			// A closed connection, a timeout on an idle client and a truncated
			// frame all end the conversation, and none of them is worth more
			// than a debug line.
			logger.Debugf("SMB1 server: connection with %s ended: %v", c.Remote, err)
			return
		}

		if err := c.handleFrame(raw); err != nil {
			logger.Debugf("SMB1 server: dropping connection with %s: %v", c.Remote, err)
			return
		}
	}
}

// handleFrame decodes one received frame and answers it.
//
// It returns an error only when the frame stream cannot be trusted to continue —
// the frame is not an SMB message at all, or it claims to be a response. A frame
// whose header decodes but whose body does not is answered with an error status
// and the connection is kept, which is what a client that sent one malformed
// request expects.
func (c *Connection) handleFrame(raw []byte) error {
	if len(raw) < header.SMB_HEADER_SIZE {
		return fmt.Errorf("frame of %d bytes is shorter than the %d-byte SMB header", len(raw), header.SMB_HEADER_SIZE)
	}
	if !bytes.Equal(raw[:len(smbMagic)], smbMagic) {
		return fmt.Errorf("frame does not begin with the SMB protocol identifier (got % x)", raw[:len(smbMagic)])
	}

	// Decode the header on its own first. A body that will not decode still has
	// to be answered, and answering requires the identifiers from the header.
	requestHeader := header.NewHeader()
	if _, err := requestHeader.Unmarshal(raw[:header.SMB_HEADER_SIZE]); err != nil {
		return fmt.Errorf("failed to decode the SMB header: %v", err)
	}

	// A request with SMB_FLAGS_REPLY set is not a request. It matters beyond
	// tidiness: the message layer chooses request or response structures from
	// that bit, so honouring it on an inbound frame would let a client pick
	// which decoder runs against its bytes.
	if requestHeader.Flags.IsReply() {
		return fmt.Errorf("frame has SMB_FLAGS_REPLY set, so it is a response rather than a request")
	}

	// A connection that is signing accepts nothing unsigned, and each request
	// consumes the number the exchange has reached.
	var responseSequenceNumber uint32
	if c.SigningActive {
		expected := c.ExpectedRequestSequenceNumber
		if !signing.Verify(c.SigningKey, raw, expected) {
			return fmt.Errorf("request failed signature verification at sequence %d", expected)
		}
		responseSequenceNumber = signing.ResponseSequenceNumber(expected)
		c.ExpectedRequestSequenceNumber = signing.NextRequestSequenceNumber(expected)
	}

	request := message.NewMessage()
	if err := request.Unmarshal(raw); err != nil {
		// The header is intact, so the client gets a proper error response and
		// the connection survives. An unrecognized command code is a different
		// error from a body that does not match a recognized one.
		status := nt_status.NT_STATUS_INVALID_SMB
		if !isKnownCommand(requestHeader.Command) {
			status = nt_status.NT_STATUS_SMB_BAD_COMMAND
		}
		logger.Debugf("SMB1 server: failed to decode command 0x%02X from %s (%s): %v",
			uint8(requestHeader.Command), c.Remote, statusName(status), err)

		w := &responseWriter{conn: c, request: &message.Message{Header: requestHeader}}
		c.Server.writeError(c, w, status)
		return nil
	}

	w := &responseWriter{conn: c, request: request}
	if c.SigningActive {
		w.SignResponse(c.SigningKey, responseSequenceNumber)
	}

	// The raw frame is kept for the duration of this handler call: the exchange
	// that arms signing has to verify its own request, which it can only do after
	// deriving the key from that request's contents.
	c.currentRequestFrame = raw
	defer func() { c.currentRequestFrame = nil }()

	// Registered handlers see the request first, so a caller can observe or
	// intercept it before the built-in dispatch answers it.
	for _, handler := range c.Server.snapshotHandlers() {
		if handler.Run(c.Server, c, w, request) {
			return nil
		}
	}

	c.Server.dispatch(c, w, request)
	return nil
}
