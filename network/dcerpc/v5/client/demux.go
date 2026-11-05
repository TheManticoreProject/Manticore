package client

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// This file holds the response demultiplexer that lets Call and Invoke be used
// from several goroutines at once. A single reader goroutine, started on the first
// call after a successful bind, reads PDUs off the transport and routes each
// response or fault to the in-flight call that owns its call_id ([C706] 12.5.1:
// call_id identifies a call on an association). Reassembly of a call's response
// fragments is owned entirely by that reader goroutine, so only the per-call result
// channel crosses between goroutines.
//
// Concurrency model (the transparent, always-on path):
//   - Call allocates a call_id under callsMu, registers a pendingCall, sends its
//     request-fragment burst under sendMu (so one call's fragments are never
//     interleaved mid-burst with another's), then blocks on the pendingCall's
//     buffered result channel.
//   - The reader goroutine owns pendingCall.stub/order; it delivers exactly one
//     result per call and removes the registration at the same time.
//   - A transport read error, a connection-level fault (call_id 0), or Close fans
//     the error out to every waiter.
//
// Per-message authentication under genuine concurrency (out-of-order responses)
// is handled separately in #1958; here a single reader unprotects fragments in
// arrival order, which keeps serial authenticated calls correct.

// pendingCall is one in-flight call awaiting its response. The reader goroutine
// owns stub and order; the waiting Call goroutine only reads the result channel.
type pendingCall struct {
	callID    uint32
	contextID uint16
	stub      []byte
	order     binary.ByteOrder
	done      chan callResult
}

// callResult is what the reader delivers to a waiting call.
type callResult struct {
	stub  []byte
	order binary.ByteOrder
	err   error
}

// ensureReader starts the demultiplexer reader goroutine once, on the first call
// after a successful bind. Bind performs its own inline reads (bind_ack, any auth
// continuation) and completes before any call runs, so the reader never competes
// with the bind handshake.
func (c *Client) ensureReader() {
	// calls and readerCond are initialized in NewClient; this only launches the
	// reader goroutine, exactly once.
	c.readerOnce.Do(func() {
		fr := &fragmentReader{t: c.transport}
		go c.readLoop(fr)
	})
}

// readLoop reads PDUs off the transport and routes them to in-flight calls. It
// reads only while at least one call is outstanding — a DCE/RPC server sends
// nothing a request did not ask for ([C706] 12.5), so between calls the reader
// parks rather than issuing a blocking read, which also lets the existing
// serial-call semantics (and tests) stand. It exits when the client is closed, the
// transport errors, or a connection-fatal protocol error occurs, failing every
// in-flight call on the way out.
func (c *Client) readLoop(fr *fragmentReader) {
	for {
		c.callsMu.Lock()
		for c.outstanding == 0 && !c.closed {
			c.readerCond.Wait()
		}
		closed := c.closed
		c.callsMu.Unlock()
		if closed {
			c.failAll(fmt.Errorf("dcerpc: client closed"))
			return
		}

		frag, err := fr.next()
		if err != nil {
			c.failAll(err)
			return
		}
		if err := c.route(frag); err != nil {
			c.failAll(err)
			return
		}
	}
}

// route dispatches one PDU to the call that owns its call_id. It returns a non-nil
// error only for a connection-fatal condition (a parse failure, an oversized
// fragment, an unexpected PDU type, or a connection-level fault), which tears the
// reader down and fails all calls. A per-call error is delivered to that call and
// route returns nil.
func (c *Client) route(frag []byte) error {
	hdr, err := pdu.PeekHeader(frag)
	if err != nil {
		return err
	}
	// [C706] 12.6.3.1 bounds response fragments by the client's max_recv_frag; the
	// larger of our proposal and the server's announced max_xmit_frag (recorded at
	// bind) is accepted, matching the serial path.
	maxRecv := c.recvFragMax
	if maxRecv == 0 {
		maxRecv = c.transport.MaxRecvFrag()
	}
	if maxRecv != 0 && hdr.FragLength > maxRecv {
		return fmt.Errorf("response frag_length %d exceeds the max_recv_frag %d negotiated at bind", hdr.FragLength, maxRecv)
	}

	switch hdr.PacketType {
	case pdu.PacketTypeResponse:
		var resp pdu.Response
		if _, err := resp.Unmarshal(frag); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		pc := c.lookup(resp.Header.CallID)
		if pc == nil {
			// A response for no in-flight call (a late or duplicate fragment after
			// the call was already answered or abandoned). Drop it.
			return nil
		}
		// The stub is encoded in the transfer syntax of the context the response
		// names; one answering on another context cannot be decoded as this call's
		// result.
		if resp.ContextID != pc.contextID {
			c.deliver(pc, callResult{err: fmt.Errorf("response p_cont_id %d does not match request context %d", resp.ContextID, pc.contextID)})
			return nil
		}
		// Record the sender's byte order (receiver-makes-right); all fragments of one
		// response carry the same data representation.
		pc.order = resp.Header.ByteOrder()
		chunk := resp.Stub
		if c.protectsRequests() {
			// At PKT and above every response fragment must carry an auth verifier.
			if resp.Header.AuthLength == 0 {
				c.deliver(pc, callResult{err: fmt.Errorf("response fragment has no auth verifier at auth level %d", c.authLevel)})
				return nil
			}
			chunk, err = c.unprotectResponseStub(frag)
			if err != nil {
				c.deliver(pc, callResult{err: fmt.Errorf("unprotect response: %w", err)})
				return nil
			}
		}
		if len(pc.stub)+len(chunk) > maxResponseStubSize {
			// A server reassembling past the cap (e.g. an endless fragment stream
			// with no PFC_LAST_FRAG) is abusive and gives no clean resumption point,
			// so tear the connection down rather than drop and keep reading. failAll
			// delivers this error to the call.
			return fmt.Errorf("reassembled response exceeds %d bytes", maxResponseStubSize)
		}
		pc.stub = append(pc.stub, chunk...)
		if hdr.PacketFlags.Has(pdu.PFCLastFrag) {
			order := pc.order
			if order == nil {
				order = binary.LittleEndian
			}
			c.deliver(pc, callResult{stub: pc.stub, order: order})
		}
		return nil

	case pdu.PacketTypeFault:
		var fault pdu.Fault
		if _, err := fault.Unmarshal(frag); err != nil {
			return fmt.Errorf("parse fault: %w", err)
		}
		// call_id 0 is a connection-level fault the server sends when it cannot
		// attribute the error to a call (e.g. before tearing the association down):
		// fail every in-flight call.
		if fault.Header.CallID == 0 {
			return &fault
		}
		pc := c.lookup(fault.Header.CallID)
		if pc == nil {
			return nil
		}
		f := fault
		c.deliver(pc, callResult{err: &f})
		return nil

	default:
		return fmt.Errorf("unexpected response PDU type %s", hdr.PacketType)
	}
}

// lookup returns the in-flight call with the given call_id, or nil.
func (c *Client) lookup(callID uint32) *pendingCall {
	c.callsMu.Lock()
	defer c.callsMu.Unlock()
	return c.calls[callID]
}

// deliver hands one result to a call and removes its registration, at most once.
// A later fragment for the same call_id finds no registration and is dropped.
func (c *Client) deliver(pc *pendingCall, res callResult) {
	c.callsMu.Lock()
	if _, ok := c.calls[pc.callID]; !ok {
		c.callsMu.Unlock()
		return
	}
	delete(c.calls, pc.callID)
	c.outstanding-- // this call is answered; balances the increment made after its send
	c.callsMu.Unlock()
	pc.done <- res // buffered (cap 1): never blocks
}

// failAll records the terminal error and fails every in-flight call with it. A
// later Call sees readerErr and returns immediately rather than blocking forever.
func (c *Client) failAll(err error) {
	c.callsMu.Lock()
	if c.readerErr == nil {
		c.readerErr = err
	}
	pending := c.calls
	c.calls = make(map[uint32]*pendingCall)
	c.outstanding = 0
	c.callsMu.Unlock()
	for _, pc := range pending {
		pc.done <- callResult{err: err}
	}
}
