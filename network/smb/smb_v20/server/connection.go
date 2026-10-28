package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

var (
	smb1ProtocolID = []byte{0xFF, 'S', 'M', 'B'}

	// relatedFileID is the FileId a compounded request uses to name the handle
	// the previous request of the chain opened.
	relatedFileID = types.SMB2_FILEID{Persistent: ^uint64(0), Volatile: ^uint64(0)}
)

// Connection is one client transport connection. Its state belongs to the
// goroutine running serve; only send and the oplock table are shared.
type Connection struct {
	server    *Server
	transport transport.Transport
	remote    net.Addr

	// Negotiated state. dialect is zero until a NEGOTIATE has been answered.
	dialect            dialects.Dialect
	clientGUID         [16]byte
	clientCapabilities capabilities.Capabilities
	clientSecurityMode securitymode.SecurityMode
	clientDialects     []dialects.Dialect
	capabilities       capabilities.Capabilities
	securityMode       securitymode.SecurityMode
	cipher             uint16
	signingAlgorithm   int
	preauthHash        []byte

	credits    creditWindow
	sessions   map[uint64]*Session
	opens      map[uint64]*Open
	nextTreeID uint32
	nextFileID uint64
	// notifies are the CHANGE_NOTIFY requests still pending, by AsyncId.
	notifies    map[uint64]*pendingNotify
	nextAsyncID uint64
	// terminate is set by a handler whose request ends the connection.
	terminate bool

	writeMutex sync.Mutex
}

func newConnection(s *Server, t transport.Transport, remote net.Addr) *Connection {
	return &Connection{
		server:           s,
		transport:        t,
		remote:           remote,
		signingAlgorithm: -1,
		credits:          newCreditWindow(),
		sessions:         make(map[uint64]*Session),
		opens:            make(map[uint64]*Open),
		notifies:         make(map[uint64]*pendingNotify),
	}
}

// request is one decoded request of a frame, with the state it resolved to.
type request struct {
	header    *header.Header
	raw       []byte
	command   command_interface.CommandInterface
	encrypted bool

	session *Session
	tree    *Tree
	// fileID is what a related compounded request inherits as relatedFileID.
	fileID types.SMB2_FILEID
}

// response is the answer to one request.
type response struct {
	header  *header.Header
	command command_interface.CommandInterface
	session *Session

	// asyncID, when set, makes this the interim response of an asynchronous
	// request.
	asyncID uint64
	// sign signs the response with session's key, when it has one.
	sign bool
	// onSent receives the response bytes as sent, for the pre-auth hash.
	onSent func(wire []byte)
}

// serve runs the receive loop until the transport fails or a frame breaks the
// protocol badly enough that the connection has to go.
func (c *Connection) serve() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("SMB2 server: panic while serving %s: %v", c.remote, r)
		}
		c.release()
		c.transport.Close()
	}()

	for {
		frame, err := c.transport.Receive()
		if err != nil {
			logger.Debugf("SMB2 server: connection with %s ended: %v", c.remote, err)
			return
		}
		if err := c.handleFrame(frame); err != nil {
			logger.Debugf("SMB2 server: dropping %s: %v", c.remote, err)
			return
		}
	}
}

// release closes everything the connection still holds.
func (c *Connection) release() {
	for _, open := range c.opens {
		c.closeOpen(open)
	}
}

// send writes one frame. It is the only method safe to call from any goroutine.
func (c *Connection) send(frame []byte) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	_, err := c.transport.Send(frame)
	return err
}

// handleFrame answers one transport frame. An error drops the connection.
func (c *Connection) handleFrame(frame []byte) error {
	if bytes.HasPrefix(frame, smb1ProtocolID) {
		return c.handleSMB1Negotiate(frame)
	}

	var encryptedBy *Session
	if security.IsTransformHeader(frame) {
		session := c.sessions[security.TransformSessionID(frame)]
		if session == nil || session.decryptionKey == nil {
			return fmt.Errorf("encrypted frame for an unknown session")
		}
		plaintext, err := security.Decrypt(c.cipher, session.decryptionKey, frame)
		if err != nil {
			return err
		}
		frame, encryptedBy = plaintext, session
	}

	if !bytes.HasPrefix(frame, header.SMB2ProtocolId[:]) {
		return fmt.Errorf("frame is not an SMB2 message")
	}

	var responses []*response
	var previous *request
	previousStatus := nt_status.NT_STATUS_SUCCESS
	for offset := 0; offset < len(frame); {
		r, next, err := c.decodeRequest(frame[offset:], encryptedBy != nil)
		if err != nil {
			return err
		}

		resp, status, err := c.process(r, previous, previousStatus)
		if err != nil {
			return err
		}
		if c.terminate {
			return fmt.Errorf("%s ended the connection", r.header.Command)
		}
		if resp != nil {
			responses = append(responses, resp)
		}
		previous, previousStatus = r, status

		if next == 0 {
			break
		}
		offset += next
	}

	if len(responses) == 0 {
		return nil
	}
	return c.sendResponses(responses, encryptedBy)
}

// decodeRequest decodes the request at the start of data and returns the offset
// of the next compounded one, or zero for the last.
func (c *Connection) decodeRequest(data []byte, encrypted bool) (*request, int, error) {
	h := header.NewHeader()
	if _, err := h.Unmarshal(data); err != nil {
		return nil, 0, err
	}
	if h.ProtocolId != header.SMB2ProtocolId || h.StructureSize != header.SMB2_HEADER_STRUCTURE_SIZE || h.IsResponse() {
		return nil, 0, fmt.Errorf("malformed SMB2 request header")
	}

	end := len(data)
	if next := int(h.NextCommand); next != 0 {
		if next < header.SMB2_HEADER_SIZE || next > len(data) || next%8 != 0 {
			return nil, 0, fmt.Errorf("compounded request offset %d is invalid", next)
		}
		end = next
	}

	r := &request{header: h, raw: data[:end], encrypted: encrypted}
	if command, err := commands.CreateRequestCommand(h.Command); err == nil {
		if _, err := command.Unmarshal(r.raw[header.SMB2_HEADER_SIZE:]); err == nil {
			r.command = command
		}
	}
	return r, int(h.NextCommand), nil
}

// process validates a request against the connection state, runs its handler
// and builds the response. A nil response means none is sent.
func (c *Connection) process(r, previous *request, previousStatus nt_status.NT_STATUS) (*response, nt_status.NT_STATUS, error) {
	h := r.header
	if h.Command == codes.SMB2_CANCEL {
		// A CANCEL consumes no credit and gets no response of its own.
		c.cancel(h)
		return nil, nt_status.NT_STATUS_SUCCESS, nil
	}
	if !c.credits.consume(h.MessageId, h.CreditCharge) {
		return nil, 0, fmt.Errorf("MessageId %d is outside the granted credits", h.MessageId)
	}
	if (c.dialect == 0) != (h.Command == codes.SMB2_NEGOTIATE) {
		return nil, 0, fmt.Errorf("%s before or after the negotiation", h.Command)
	}

	if h.Flags.IsRelatedOperations() && previous != nil {
		h.SessionId = previous.header.SessionId
		h.TreeId = previous.header.TreeId
		r.fileID = previous.fileID
	}

	resp := &response{header: h}
	var status nt_status.NT_STATUS
	if handled, w := c.runHandlers(r); handled {
		if !w.written {
			return nil, nt_status.NT_STATUS_SUCCESS, nil
		}
		resp.command, status = w.command, w.status
		if w.sessionID != 0 {
			h.SessionId = w.sessionID
		}
		if session := c.sessions[h.SessionId]; session != nil && session.established {
			r.session = session
		}
	} else {
		status = c.authorize(r)
		if h.Flags.IsRelatedOperations() && previous != nil && !previousStatus.IsSuccess() {
			// A related request fails with the one it depends on.
			status = previousStatus
		}
		if status.IsSuccess() {
			status = c.dispatch(r, resp)
		}
	}
	if resp.session == nil {
		resp.session = r.session
	}
	resp.sign = resp.sign || h.Flags.IsSigned() || (r.session != nil && r.session.signingRequired)

	if resp.command == nil || (!status.IsSuccess() && !carriesBody(status)) {
		resp.command = commands.NewErrorResponse(h.Command)
	}
	// Warnings such as STATUS_NO_MORE_FILES are normal outcomes, not refusals.
	if status>>30 == 3 {
		logger.Debugf("SMB2 server: %s %s refused with %s", c.remote, h.Command, status)
	}
	resp.header = c.responseHeader(r, resp, status)
	return resp, status, nil
}

// authorize resolves the session and tree a request runs in and enforces the
// session's signing and encryption requirements.
func (c *Connection) authorize(r *request) nt_status.NT_STATUS {
	h := r.header
	if r.command == nil {
		if _, known := handlers[h.Command]; !known {
			return nt_status.NT_STATUS_NOT_SUPPORTED
		}
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	switch h.Command {
	case codes.SMB2_NEGOTIATE, codes.SMB2_ECHO:
		return nt_status.NT_STATUS_SUCCESS
	case codes.SMB2_SESSION_SETUP:
		// The handler resolves its own session: one being set up has no key yet.
		return nt_status.NT_STATUS_SUCCESS
	}

	session := c.sessions[h.SessionId]
	if session == nil || !session.established {
		return nt_status.NT_STATUS_USER_SESSION_DELETED
	}
	r.session = session

	if session.encryptData && !r.encrypted {
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	if !r.encrypted && session.signingKey != nil {
		if h.Flags.IsSigned() {
			if !security.Verify(c.dialect, c.signingAlgorithm, session.signingKey, r.raw) {
				return nt_status.NT_STATUS_ACCESS_DENIED
			}
		} else if session.signingRequired {
			return nt_status.NT_STATUS_ACCESS_DENIED
		}
	}

	switch h.Command {
	case codes.SMB2_LOGOFF, codes.SMB2_TREE_CONNECT:
		return nt_status.NT_STATUS_SUCCESS
	}
	r.tree = session.trees[h.TreeId]
	if r.tree == nil {
		return nt_status.NT_STATUS_NETWORK_NAME_DELETED
	}
	return nt_status.NT_STATUS_SUCCESS
}

// carriesBody reports whether a non-success status still comes with the
// command's own response rather than an ERROR response.
func carriesBody(status nt_status.NT_STATUS) bool {
	return status == nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED || status == nt_status.NT_STATUS_BUFFER_OVERFLOW
}

// responseHeader builds the header correlated to a request.
func (c *Connection) responseHeader(r *request, resp *response, status nt_status.NT_STATUS) *header.Header {
	h := header.NewHeader()
	h.CreditCharge = r.header.CreditCharge
	h.Status = uint32(status)
	h.Command = r.header.Command
	h.Credit = c.credits.grant(r.header.Credit)
	h.Flags = flags.SMB2_FLAGS_SERVER_TO_REDIR | r.header.Flags&(flags.SMB2_FLAGS_RELATED_OPERATIONS|flags.SMB2_FLAGS_PRIORITY_MASK)
	h.MessageId = r.header.MessageId
	h.TreeId = r.header.TreeId
	h.SessionId = r.header.SessionId
	if resp.session != nil {
		h.SessionId = resp.session.id
	}
	if resp.asyncID != 0 {
		h.Flags |= flags.SMB2_FLAGS_ASYNC_COMMAND
		h.AsyncId = resp.asyncID
	}
	return h
}

// sendResponses frames the responses as one, possibly compounded, message: each
// segment is padded to 8 bytes and signed on its own, and the whole message is
// encrypted when the requests were.
func (c *Connection) sendResponses(responses []*response, encryptedBy *Session) error {
	var frame []byte
	segments := make([][]byte, len(responses))
	for i, resp := range responses {
		body, err := resp.command.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal the %s response: %v", resp.header.Command, err)
		}
		length := header.SMB2_HEADER_SIZE + len(body)
		if i < len(responses)-1 {
			length = (length + 7) &^ 7
			resp.header.NextCommand = uint32(length)
		}
		hdr, err := resp.header.Marshal()
		if err != nil {
			return err
		}
		segment := make([]byte, length)
		copy(segment, hdr)
		copy(segment[header.SMB2_HEADER_SIZE:], body)

		if encryptedBy == nil && resp.sign && resp.session != nil && resp.session.signingKey != nil {
			security.Sign(c.dialect, c.signingAlgorithm, resp.session.signingKey, segment)
		}
		segments[i] = segment
		frame = append(frame, segment...)
	}

	if encryptedBy != nil {
		sealed, err := security.Encrypt(c.cipher, encryptedBy.encryptionKey, encryptedBy.id, encryptedBy.nonce.Add(1), frame)
		if err != nil {
			return err
		}
		frame = sealed
	}
	if err := c.send(frame); err != nil {
		return err
	}

	for i, resp := range responses {
		if resp.onSent != nil {
			resp.onSent(segments[i])
		}
	}
	return nil
}

// handleSMB1Negotiate answers the SMB1 NEGOTIATE a multi-protocol client opens
// with, by upgrading it to SMB2 (MS-SMB2 3.3.5.3).
func (c *Connection) handleSMB1Negotiate(frame []byte) error {
	if c.dialect != 0 || !c.credits.consume(0, 1) {
		return fmt.Errorf("SMB1 message after the negotiation")
	}
	offered, err := smb1DialectStrings(frame)
	if err != nil {
		return err
	}

	var revision dialects.Dialect
	switch {
	case offered["SMB 2.???"] && c.server.highestDialect() > dialects.SMB2_DIALECT_2_0_2:
		// The client follows up with an SMB2 NEGOTIATE.
		revision = dialects.SMB2_DIALECT_WILDCARD
	case offered["SMB 2.002"] && c.server.offers(dialects.SMB2_DIALECT_2_0_2):
		revision = dialects.SMB2_DIALECT_2_0_2
	default:
		return fmt.Errorf("client offers no SMB2 dialect")
	}

	resp := &response{command: c.negotiateResponse(revision, nil)}
	resp.header = header.NewHeader()
	resp.header.Command = codes.SMB2_NEGOTIATE
	resp.header.Flags = flags.SMB2_FLAGS_SERVER_TO_REDIR
	resp.header.Credit = c.credits.grant(1)
	if revision != dialects.SMB2_DIALECT_WILDCARD {
		// An SMB1 NEGOTIATE carries no GUID, capabilities or security mode, so the
		// secure negotiate validation expects zeros for them.
		c.dialect, c.clientDialects = revision, []dialects.Dialect{revision}
	}
	return c.sendResponses([]*response{resp}, nil)
}

// smb1DialectStrings returns the dialect strings of an SMB1 NEGOTIATE request.
func smb1DialectStrings(frame []byte) (map[string]bool, error) {
	const smb1HeaderSize, smb1Negotiate = 32, 0x72
	if len(frame) < smb1HeaderSize+3 || frame[4] != smb1Negotiate {
		return nil, fmt.Errorf("SMB1 message is not a NEGOTIATE")
	}
	body := frame[smb1HeaderSize+1:] // WordCount is 0
	byteCount := int(binary.LittleEndian.Uint16(body))
	if 2+byteCount > len(body) {
		return nil, fmt.Errorf("SMB1 NEGOTIATE is truncated")
	}

	offered := make(map[string]bool)
	for _, entry := range bytes.Split(body[2:2+byteCount], []byte{0}) {
		if len(entry) > 1 && entry[0] == 0x02 {
			offered[string(entry[1:])] = true
		}
	}
	return offered, nil
}

// creditWindow is the connection's command sequence window (MS-SMB2 3.3.1.1):
// the MessageIds granted to the client, each of which it may use once, in any
// order.
type creditWindow struct {
	low     uint64              // MessageIds below this were used or expired
	granted uint64              // MessageIds below this have been granted
	used    map[uint64]struct{} // MessageIds at or above low already used
}

// maxCredits bounds the credits a client may hold at once.
const maxCredits = 512

// maxWindow bounds the span from the oldest unused MessageId to the newest
// granted one, so a MessageId the client never uses cannot keep the window, and
// the record of used MessageIds above it, growing.
const maxWindow = 8 * maxCredits

func newCreditWindow() creditWindow {
	return creditWindow{granted: 1, used: make(map[uint64]struct{})}
}

// available is the number of granted MessageIds not yet used.
func (w *creditWindow) available() uint64 {
	return w.granted - w.low - uint64(len(w.used))
}

// consume charges a request against the window: every MessageId it spans must
// be granted and not used before.
func (w *creditWindow) consume(messageID uint64, charge uint16) bool {
	n := uint64(max(charge, 1))
	if messageID < w.low || messageID+n > w.granted || messageID+n < messageID {
		return false
	}
	for id := messageID; id < messageID+n; id++ {
		if _, seen := w.used[id]; seen {
			return false
		}
	}
	for id := messageID; id < messageID+n; id++ {
		w.used[id] = struct{}{}
	}
	for {
		if _, seen := w.used[w.low]; !seen {
			break
		}
		delete(w.used, w.low)
		w.low++
	}
	return true
}

// grant hands the client up to requested credits, never leaving it with none.
func (w *creditWindow) grant(requested uint16) uint16 {
	var n uint64
	if available := w.available(); available < maxCredits {
		n = min(uint64(max(requested, 1)), maxCredits-available)
	}
	w.granted += n
	// The oldest MessageIds the client skipped expire once the window is full.
	for w.granted-w.low > maxWindow {
		delete(w.used, w.low)
		w.low++
	}
	return uint16(n)
}
