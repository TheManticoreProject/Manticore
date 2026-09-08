package server

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// AsyncResponder answers a request after the handler that received it has
// returned.
//
// It exists for the commands whose answer is not ready when they are asked:
// NT_TRANSACT_NOTIFY_CHANGE is answered when a directory changes, and a byte-range
// lock that offers to wait is answered when the range comes free. A handler takes
// one with ResponseWriter.Defer, returns success without writing, and completes
// the exchange from wherever the answer arrives.
//
// # What a responder may touch
//
// A responder is safe to use from any goroutine, and it is the only part of the
// connection that is. Everything else on a Connection — the open, tree, session
// and search tables — belongs to the goroutine running the receive loop and has no
// locking, so a responder must capture what it needs at Defer time rather than
// reading those tables when it fires.
//
// That constraint is the reason this is a narrow interface rather than a handle on
// the connection: what a deferred answer needs is a way to send bytes, and giving
// it anything more would invite a data race that no test on one goroutine would
// find.
//
// # Cancellation
//
// Cancelled closes when the client sends SMB_COM_NT_CANCEL for the request, or
// when the connection goes away. A responder that never fires must still be
// released with Done, or the connection keeps a record of it until it closes.
type AsyncResponder interface {
	// Respond sends a successful response and releases the responder.
	Respond(cmd command_interface.CommandInterface) error

	// RespondWithStatus sends a response carrying a status and releases the
	// responder.
	RespondWithStatus(cmd command_interface.CommandInterface, status nt_status.NT_STATUS) error

	// RespondWithError sends a payload-less error response and releases the
	// responder.
	RespondWithError(status nt_status.NT_STATUS) error

	// Cancelled is closed when the request is cancelled or the connection ends.
	// It never carries a value; a receive on it means stop.
	Cancelled() <-chan struct{}

	// Done releases the responder without answering. Calling it after a Respond
	// is harmless, so it is safe in a defer.
	Done()
}

// ErrResponderDone reports a responder used after it has answered or been
// released. It is a programming error rather than a protocol one: a request has
// one response, and sending a second would leave a reply on the connection that
// no request accounts for.
var ErrResponderDone = errors.New("the asynchronous responder has already been released")

// ErrUnsolicitedWhileSigning reports an attempt to send a server-initiated message
// on a signing connection.
//
// A signed message consumes a sequence number from the pair the two sides keep in
// step, and a message the client never asked for has no request whose number it
// can use. Getting that wrong desynchronises signing for the rest of the
// connection, and every subsequent request fails verification — so it is refused
// here rather than guessed at, until the numbering can be checked against a
// reference capture.
var ErrUnsolicitedWhileSigning = errors.New("cannot send a server-initiated message on a signing connection")

// asyncResponder is the AsyncResponder bound to one deferred request.
type asyncResponder struct {
	conn *Connection

	// request is the header being answered, kept so the reply can echo the
	// identifiers the client correlates on.
	request *header.Header

	// signKey and signSequence are the signing state reserved for this request's
	// response, captured at Defer time. They are the response's own numbers, so a
	// deferred answer signs exactly as an immediate one would.
	signKey      []byte
	signSequence uint32

	// uid and tid override the request's, for a handler that assigned one.
	uid    uint16
	uidSet bool
	tid    uint16
	tidSet bool

	// cancelled is closed by the receive loop on NT_CANCEL or on shutdown.
	cancelled chan struct{}

	// once guards the release, so a double answer is reported rather than sent.
	once     sync.Once
	released bool
	mutex    sync.Mutex
}

// Respond sends a successful response.
func (r *asyncResponder) Respond(cmd command_interface.CommandInterface) error {
	return r.RespondWithStatus(cmd, nt_status.NT_STATUS_SUCCESS)
}

// RespondWithStatus sends a response carrying a status.
func (r *asyncResponder) RespondWithStatus(
	cmd command_interface.CommandInterface,
	status nt_status.NT_STATUS,
) error {
	if cmd == nil {
		return fmt.Errorf("cannot answer with no command")
	}

	r.mutex.Lock()
	if r.released {
		r.mutex.Unlock()
		return ErrResponderDone
	}
	r.released = true
	r.mutex.Unlock()

	defer r.conn.releaseResponder(r)

	reply := message.NewMessage()
	reply.Header = replyHeader(r.request, status, len(r.signKey) > 0)
	if r.uidSet {
		reply.Header.UID = types.USHORT(r.uid)
	}
	if r.tidSet {
		reply.Header.TID = types.USHORT(r.tid)
	}

	cmd.SetUnicode(r.request.Flags2.IsUnicode())
	reply.AddCommand(cmd)
	reply.Header.Command = r.request.Command

	return r.conn.frame(reply, r.signKey, r.signSequence)
}

// RespondWithError sends a payload-less error response.
func (r *asyncResponder) RespondWithError(status nt_status.NT_STATUS) error {
	return r.RespondWithStatus(newErrorResponse(r.request.Command), status)
}

// Cancelled reports when the request no longer needs an answer.
func (r *asyncResponder) Cancelled() <-chan struct{} {
	return r.cancelled
}

// Done releases the responder without answering.
func (r *asyncResponder) Done() {
	r.mutex.Lock()
	already := r.released
	r.released = true
	r.mutex.Unlock()

	if !already {
		r.conn.releaseResponder(r)
	}
}

// cancel closes the responder's cancellation channel, once.
func (r *asyncResponder) cancel() {
	r.once.Do(func() { close(r.cancelled) })
}

// Compile-time assurance that asyncResponder satisfies the contract.
var _ AsyncResponder = (*asyncResponder)(nil)

// frame marshals, signs and sends one message, serialised against every other
// write on the connection.
//
// Every path that puts bytes on the transport goes through here — the immediate
// responses, the deferred ones and the server-initiated ones — because the
// transport is not safe for concurrent use and a deferred answer runs on a
// goroutine the receive loop knows nothing about.
//
// Parameters:
//   - reply: the message to send
//   - signKey: the MAC key, or nil for an unsigned message
//   - signSequence: the sequence number to sign at
//
// Returns:
//   - The error from marshalling or sending
func (c *Connection) frame(reply *message.Message, signKey []byte, signSequence uint32) error {
	marshalled, err := reply.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal the response: %v", err)
	}
	return c.send(marshalled, signKey, signSequence)
}

// send signs and writes one already-marshalled message, serialised against every
// other write on the connection.
//
// It is separate from frame because a batched response has to know its own size
// before it is sent — [MS-CIFS] section 2.2.3.4 caps a batch at the negotiated
// buffer — so that path marshals first and sends through here.
//
// Parameters:
//   - marshalled: the message as it will go on the wire
//   - signKey: the MAC key, or nil for an unsigned message
//   - signSequence: the sequence number to sign at
//
// Returns:
//   - The error from sending
func (c *Connection) send(marshalled []byte, signKey []byte, signSequence uint32) error {
	// The lock covers signing as well as sending: a signature is written into the
	// buffer being sent, so two goroutines signing and sending independently could
	// interleave a partial write.
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if len(signKey) > 0 {
		signing.Sign(signKey, marshalled, signSequence)
	}

	if _, err := c.Transport.Send(marshalled); err != nil {
		return fmt.Errorf("failed to send the response: %v", err)
	}
	return nil
}

// defer_ registers a deferred answer for the request being handled.
//
// The signing state is captured now rather than when the answer fires, because it
// belongs to this request: the sequence number the response signs at was reserved
// when the request arrived, and a later answer signs at the same number an
// immediate one would have.
func (c *Connection) deferResponse(
	request *header.Header,
	signKey []byte,
	signSequence uint32,
	uid uint16, uidSet bool,
	tid uint16, tidSet bool,
) AsyncResponder {
	responder := &asyncResponder{
		conn:         c,
		request:      request,
		signKey:      signKey,
		signSequence: signSequence,
		uid:          uid,
		uidSet:       uidSet,
		tid:          tid,
		tidSet:       tidSet,
		cancelled:    make(chan struct{}),
	}

	c.responderMutex.Lock()
	if c.responders == nil {
		c.responders = make(map[*asyncResponder]struct{})
	}
	c.responders[responder] = struct{}{}
	outstanding := len(c.responders)
	c.responderMutex.Unlock()

	logger.Debugf("SMB1 server: %s deferred command 0x%02X MID 0x%04X, %d outstanding",
		c.Remote, uint8(request.Command), uint16(request.MID), outstanding)
	return responder
}

// releaseResponder forgets a responder that has answered or been released.
func (c *Connection) releaseResponder(responder *asyncResponder) {
	c.responderMutex.Lock()
	delete(c.responders, responder)
	c.responderMutex.Unlock()
}

// cancelResponders closes the cancellation channel of every deferred request
// matching a PID and MID, and reports how many it found.
//
// SMB_COM_NT_CANCEL names the request to cancel by the PID and MID in its own
// header, which is how a client refers to something it has not had an answer for.
func (c *Connection) cancelResponders(pid uint32, mid uint16) int {
	c.responderMutex.Lock()
	matching := []*asyncResponder{}
	for responder := range c.responders {
		if responder.request.GetPID() == pid && uint16(responder.request.MID) == mid {
			matching = append(matching, responder)
		}
	}
	c.responderMutex.Unlock()

	// Cancelled outside the lock: a responder that reacts immediately will call
	// back in to release itself.
	for _, responder := range matching {
		responder.cancel()
	}
	return len(matching)
}

// cancelAllResponders cancels every deferred request, which is what closing the
// connection does.
//
// A deferred answer that is still waiting has nowhere to send its reply once the
// transport is gone, so it is told to stop rather than left to fail on a write.
func (c *Connection) cancelAllResponders() {
	c.responderMutex.Lock()
	matching := make([]*asyncResponder, 0, len(c.responders))
	for responder := range c.responders {
		matching = append(matching, responder)
	}
	c.responderMutex.Unlock()

	for _, responder := range matching {
		responder.cancel()
	}
}

// OutstandingResponses reports how many deferred answers this connection is
// waiting to send, for a caller that wants to report on it and for the tests.
func (c *Connection) OutstandingResponses() int {
	c.responderMutex.Lock()
	defer c.responderMutex.Unlock()
	return len(c.responders)
}

// SendUnsolicited sends a message the client did not ask for.
//
// This is the one place in the protocol where the server sends a request rather
// than a response: [MS-CIFS] section 2.2.4.32.1 has the server send
// SMB_COM_LOCKING_ANDX with SMB_FLAGS_REPLY clear to break an oplock. The message
// carries a fresh MID, because there is no request to correlate it with.
//
// It is refused on a signing connection. A signed message consumes a sequence
// number from the pair the two sides keep in step, and a message with no request
// behind it has no number reserved for it; a wrong guess desynchronises signing
// for the rest of the connection and every later request fails verification.
// Refusing is better than a guess that cannot be checked here against a real
// client.
//
// Parameters:
//   - cmd: the command to send
//   - uid, tid, mid: the identifiers to carry
//
// Returns:
//   - ErrUnsolicitedWhileSigning on a signing connection, or the send error
func (c *Connection) SendUnsolicited(
	cmd command_interface.CommandInterface,
	uid, tid, mid uint16,
) error {
	if cmd == nil {
		return fmt.Errorf("cannot send no command")
	}
	if c.SigningActive {
		return ErrUnsolicitedWhileSigning
	}

	request := message.NewMessage()
	request.Header.Command = cmd.GetCommandCode()
	// SMB_FLAGS_REPLY stays clear: this is a request, and a client reads that bit
	// to decide which decoder to run.
	request.Header.Flags = flags.Flags(0)
	request.Header.Flags2 = c.negotiatedFlags2()
	request.Header.UID = types.USHORT(uid)
	request.Header.TID = types.USHORT(tid)
	request.Header.MID = types.USHORT(mid)

	cmd.SetUnicode(c.UseUnicode)
	request.AddCommand(cmd)
	request.Header.Command = cmd.GetCommandCode()

	logger.Debugf("SMB1 server: sending unsolicited command 0x%02X to %s",
		uint8(cmd.GetCommandCode()), c.Remote)
	return c.frame(request, nil, 0)
}

// negotiatedFlags2 rebuilds the SMB_FLAGS2 bits a server-initiated message
// carries, from what the connection agreed.
//
// A response mirrors them from the request it answers; an unsolicited message has
// no request to mirror, so they come from the connection instead.
func (c *Connection) negotiatedFlags2() flags2.Flags2 {
	var bits flags2.Flags2
	if c.UseUnicode {
		bits |= flags2.Flags2(flags2.FLAGS2_UNICODE)
	}
	if c.UseNTStatus {
		bits |= flags2.Flags2(flags2.FLAGS2_NT_STATUS_ERROR_CODES)
	}
	return bits
}
