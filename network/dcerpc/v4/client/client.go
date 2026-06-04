// Package client implements the client side of the connectionless (datagram) DCE/RPC
// protocol machine ([C706] chapter 10), layered on a datagram transport
// (network/dcerpc/v4/transport) and the connectionless PDU codec
// (network/dcerpc/v4/pdu).
//
// Unlike the connection-oriented client, there is no Bind: a call is established
// implicitly by the first request PDU, which carries the activity UUID, interface
// identity, and a per-activity sequence number. The client fragments the request
// stub into request PDUs, transmits them, and then drives a receive loop that
// handles the server's flow-control and status PDUs:
//
//   - fack    — flow-control acknowledgement; the client may send more fragments.
//   - working — the server is processing the call; the client keeps waiting.
//   - nocall  — the server has not seen the call; the client retransmits.
//   - response— output fragments, reassembled until the lastfrag PDU arrives.
//   - fault / reject — the call failed; returned as *FaultError / *RejectError.
//
// When no PDU arrives before the transport timeout, the client retransmits the
// request fragments (while still awaiting any response) or sends a ping (while
// awaiting further response fragments), giving up after the configured limits.
//
// NOT YET IMPLEMENTED (documented limitations):
//   - The conv callback interface (conv_who_are_you) used for at-most-once
//     verification of non-idempotent calls; mark calls idempotent for now.
//   - Windowed burst flow control: all request fragments are sent immediately rather
//     than throttled to the server's advertised fack window. This is correct for the
//     small, typically single-fragment requests connectionless RPC carries, and the
//     server's nocall/fack still drive retransmission.
//   - Broadcast and maybe (no-response) calls, and client-side cancel.
//
// References:
//   - [C706] chapter 10, Connectionless RPC Protocol Machines:
//     https://pubs.opengroup.org/onlinepubs/9629399/chap10.htm
package client

import (
	"errors"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Default flow-control limits ([C706] chapter 10: MAX_REQUESTS, MAX_PINGS).
const (
	// DefaultMaxRequests is how many times the request burst is retransmitted before
	// the client gives up waiting for the server to acknowledge the call.
	DefaultMaxRequests = 5
	// DefaultMaxPings is how many pings are sent, while a response is outstanding,
	// before the client gives up.
	DefaultMaxPings = 5
	// DefaultFackWindow is the receive window, in fragments, advertised in the facks
	// the client sends for inbound response fragments.
	DefaultFackWindow = 8
)

// ErrNoResponse is returned when the server never produces a response within the
// configured retransmission and ping limits.
var ErrNoResponse = errors.New("dcerpc(cl): no response within retransmission limits")

// FaultError reports a fault PDU returned by the server, carrying its status code.
type FaultError struct{ Status uint32 }

func (e *FaultError) Error() string { return fmt.Sprintf("dcerpc(cl) fault: status 0x%08x", e.Status) }

// RejectError reports a reject PDU returned by the server, carrying its status code.
type RejectError struct{ Status uint32 }

func (e *RejectError) Error() string {
	return fmt.Sprintf("dcerpc(cl) reject: status 0x%08x", e.Status)
}

// Client drives connectionless calls over a single activity (conversation).
type Client struct {
	transport transport.Transport

	activityID guid.GUID
	seq        uint32

	serverBoot     uint32
	haveServerBoot bool

	maxRequests int
	maxPings    int
	fackWindow  uint16
}

// Option configures a Client.
type Option func(*Client)

// WithActivityID sets a fixed activity UUID instead of a random one (useful for tests
// and for resuming a known conversation).
func WithActivityID(id guid.GUID) Option { return func(c *Client) { c.activityID = id } }

// WithMaxRequests overrides the request retransmission limit.
func WithMaxRequests(n int) Option { return func(c *Client) { c.maxRequests = n } }

// WithMaxPings overrides the ping limit.
func WithMaxPings(n int) Option { return func(c *Client) { c.maxPings = n } }

// New returns a connectionless client bound to the given datagram transport. Unless
// WithActivityID is supplied, a fresh random activity UUID is generated.
func New(t transport.Transport, opts ...Option) *Client {
	c := &Client{
		transport:   t,
		activityID:  *guid.NewGUID(),
		maxRequests: DefaultMaxRequests,
		maxPings:    DefaultMaxPings,
		fackWindow:  DefaultFackWindow,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ActivityID returns the activity UUID identifying this client's conversation.
func (c *Client) ActivityID() guid.GUID { return c.activityID }

// CallRequest describes one connectionless RPC call.
type CallRequest struct {
	// Object is the optional object UUID; leave zero for none.
	Object guid.GUID
	// Interface is the interface (abstract syntax) UUID.
	Interface guid.GUID
	// InterfaceVersion is the interface major version (the if_vers field).
	InterfaceVersion uint32
	// OpNum is the operation number within the interface.
	OpNum uint16
	// Stub is the NDR-marshalled input parameters.
	Stub []byte
	// Idempotent marks the call idempotent, allowing safe retransmission without
	// at-most-once verification.
	Idempotent bool
}

// Call performs a single connectionless RPC call and returns the reassembled response
// stub (the NDR-marshalled output parameters). A server fault or reject is returned
// as *FaultError or *RejectError.
func (c *Client) Call(req CallRequest) ([]byte, error) {
	if err := c.transport.Connect(); err != nil {
		return nil, err
	}

	seq := c.seq
	c.seq++

	tmpl := pdu.NewHeader(pdu.PacketTypeRequest)
	tmpl.ObjectID = req.Object
	tmpl.InterfaceID = req.Interface
	tmpl.InterfaceVersion = req.InterfaceVersion
	tmpl.ActivityID = c.activityID
	tmpl.SequenceNumber = seq
	tmpl.OpNum = req.OpNum
	if c.haveServerBoot {
		tmpl.ServerBoot = c.serverBoot
	}
	if req.Idempotent {
		tmpl.Flags1 |= pdu.Flags1Idempotent
	}

	frags, err := fragmentRequest(tmpl, req.Stub, c.transport.MaxPDUSize())
	if err != nil {
		return nil, err
	}
	if err := c.sendFragments(frags); err != nil {
		return nil, err
	}

	return c.awaitResponse(seq, frags)
}

// sendFragments transmits every request fragment.
func (c *Client) sendFragments(frags []pdu.PDU) error {
	for i := range frags {
		raw, err := frags[i].Marshal()
		if err != nil {
			return err
		}
		if _, err := c.transport.Send(raw); err != nil {
			return err
		}
	}
	return nil
}

// awaitResponse runs the receive loop until the response is fully reassembled, the
// server faults/rejects, or the retransmission/ping limits are exhausted.
func (c *Client) awaitResponse(seq uint32, reqFrags []pdu.PDU) ([]byte, error) {
	reasm := &responseReassembler{}
	gotResponseFrag := false
	requests := 0
	pings := 0

	for {
		raw, err := c.transport.Recv()
		if err != nil {
			if !isTimeout(err) {
				return nil, err
			}
			// No PDU arrived in time. If we have begun receiving the response, poll
			// with a ping; otherwise the server may not have seen the call, so
			// retransmit the request burst.
			if gotResponseFrag {
				if pings >= c.maxPings {
					return nil, ErrNoResponse
				}
				pings++
				if err := c.sendPing(seq); err != nil {
					return nil, err
				}
			} else {
				if requests >= c.maxRequests {
					return nil, ErrNoResponse
				}
				requests++
				if err := c.sendFragments(reqFrags); err != nil {
					return nil, err
				}
			}
			continue
		}

		var p pdu.PDU
		if _, err := p.Unmarshal(raw); err != nil {
			// Ignore undecodable datagrams rather than aborting the call.
			continue
		}
		// Only accept PDUs for this activity and call.
		if !p.Header.ActivityID.Equal(&c.activityID) || p.Header.SequenceNumber != seq {
			continue
		}
		c.adoptServerBoot(p.Header.ServerBoot)

		switch p.Header.PacketType {
		case pdu.PacketTypeFack:
			// Flow-control acknowledgement of our request fragments. With the current
			// send-all strategy there is nothing more to send; keep waiting.
			continue
		case pdu.PacketTypeWorking:
			// Server is processing; reset the retransmission count and keep waiting.
			requests = 0
			continue
		case pdu.PacketTypeNoCall:
			// Server has not seen the call; retransmit the request burst.
			if requests >= c.maxRequests {
				return nil, ErrNoResponse
			}
			requests++
			if err := c.sendFragments(reqFrags); err != nil {
				return nil, err
			}
			continue
		case pdu.PacketTypeResponse:
			gotResponseFrag = true
			reasm.add(p.Header, p.Body)
			// Acknowledge a received response fragment unless the server suppressed
			// facks for it.
			if p.Header.Flags1.Has(pdu.Flags1Frag) && !p.Header.Flags1.Has(pdu.Flags1NoFack) {
				if err := c.sendFack(p.Header); err != nil {
					return nil, err
				}
			}
			if reasm.complete() {
				return reasm.assemble(), nil
			}
			continue
		case pdu.PacketTypeFault:
			status, _ := pdu.UnmarshalStatusBody(p.Body)
			return nil, &FaultError{Status: status}
		case pdu.PacketTypeReject:
			status, _ := pdu.UnmarshalStatusBody(p.Body)
			return nil, &RejectError{Status: status}
		default:
			// ping/ack/cancel/cancel_ack are not expected inbound on the client; ignore.
			continue
		}
	}
}

// replyHeader builds a header for a client-originated control PDU (ping, fack) that
// targets the same call as the given request/response header.
func (c *Client) replyHeader(pt pdu.PacketType, seq uint32, object, iface guid.GUID, ifVers uint32, opnum uint16) pdu.Header {
	h := pdu.NewHeader(pt)
	h.ObjectID = object
	h.InterfaceID = iface
	h.InterfaceVersion = ifVers
	h.ActivityID = c.activityID
	h.SequenceNumber = seq
	h.OpNum = opnum
	if c.haveServerBoot {
		h.ServerBoot = c.serverBoot
	}
	return h
}

// sendPing transmits a ping PDU to poll the server for an outstanding call.
func (c *Client) sendPing(seq uint32) error {
	h := c.replyHeader(pdu.PacketTypePing, seq, guid.GUID{}, guid.GUID{}, 0, 0)
	p := pdu.PDU{Header: h}
	raw, err := p.Marshal()
	if err != nil {
		return err
	}
	_, err = c.transport.Send(raw)
	return err
}

// sendFack transmits a fack acknowledging an inbound response fragment.
func (c *Client) sendFack(resp pdu.Header) error {
	h := c.replyHeader(pdu.PacketTypeFack, resp.SequenceNumber, resp.ObjectID, resp.InterfaceID, resp.InterfaceVersion, resp.OpNum)
	h.FragmentNumber = resp.FragmentNumber
	h.SerialHi = resp.SerialHi
	h.SerialLo = resp.SerialLo
	body, err := (&pdu.FackBody{
		WindowSize:  c.fackWindow,
		MaxTSDU:     uint32(c.transport.MaxPDUSize()),
		MaxFragSize: uint32(c.transport.MaxPDUSize()),
		SerialNum:   resp.Serial(),
	}).Marshal()
	if err != nil {
		return err
	}
	p := pdu.PDU{Header: h, Body: body}
	raw, err := p.Marshal()
	if err != nil {
		return err
	}
	_, err = c.transport.Send(raw)
	return err
}

// adoptServerBoot records the server boot time from the first PDU that carries a
// non-zero value, and resets the sequence space if the server appears to have
// rebooted ([C706] chapter 10, BOOT_TIME_EQ).
func (c *Client) adoptServerBoot(boot uint32) {
	if boot == 0 {
		return
	}
	if !c.haveServerBoot {
		c.serverBoot = boot
		c.haveServerBoot = true
		return
	}
	if c.serverBoot != boot {
		// Server rebooted: adopt the new boot time. The sequence number keeps
		// increasing, which is the conservative choice for at-most-once servers.
		c.serverBoot = boot
	}
}

// isTimeout reports whether err is a network timeout (the signal the transport gives
// when no datagram arrived before the deadline).
func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
