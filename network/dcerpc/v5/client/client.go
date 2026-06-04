// Package client implements the high-level connection-oriented DCE/RPC client: it
// binds to an interface and issues calls over a transport.
//
// The flow follows [C706] chapter 12 and [MS-RPCE]:
//
//  1. Bind sends a bind PDU proposing the transport's max_xmit_frag / max_recv_frag,
//     a single presentation context (the target abstract syntax with the NDR 2.0
//     transfer syntax), and records the server's negotiated fragment sizes from the
//     bind_ack.
//  2. Call serializes a request, fragments it so that no fragment exceeds the
//     negotiated send size (PFC_FIRST_FRAG on the first, PFC_LAST_FRAG on the last,
//     both on a single fragment), writes every fragment, then reassembles the
//     response fragments until PFC_LAST_FRAG. A fault PDU is returned as an error.
//
// The call_id is constant across the fragments of one call and is incremented for
// each new call, as required by [C706] section 12.6.2.
//
// References:
//   - [C706] chapter 12 (connection-oriented protocol), "Fragmentation and
//     Reassembly": https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 3.3.1.5.3 Bind Time Feature Negotiation and 2.2.2 PDU definitions:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/87964b3c-1785-4aae-a993-734999441ed3
//   - [MS-RPCE] 2.1.1.2 SMB (NCACN_NP):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/7063c7bd-b48b-42e7-9154-3c2ec4113c0d
package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// requestHeaderOverhead is the fixed size of a request PDU before the stub data: the
// 16-byte common header plus the 8-byte request body (alloc_hint, p_cont_id, opnum).
const requestHeaderOverhead = pdu.HeaderSize + 8

// Client is a connection-oriented DCE/RPC client bound to a single presentation
// context on a transport.
type Client struct {
	transport transport.Transport

	// callID is the call_id of the most recent call; it is incremented per call.
	callID uint32

	// contextID is the presentation context id negotiated at Bind (0 for v1).
	contextID uint16

	// sendFragMax is the largest request fragment, in bytes, the server will accept,
	// derived from the bind_ack.
	sendFragMax uint16

	bound bool
}

// NewClient returns a DCE/RPC client over the supplied transport. The transport must
// be ready to Connect (for ncacn_np, its SMB session and IPC$ tree connect are
// already established).
func NewClient(t transport.Transport) *Client {
	return &Client{transport: t}
}

// Bind connects the transport and binds to the interface identified by
// abstractSyntax, negotiating the NDR 2.0 transfer syntax in a single presentation
// context. It returns an error if the server rejects the bind (bind_nak) or accepts
// no context.
func (c *Client) Bind(abstractSyntax syntax.SyntaxID) error {
	if err := c.transport.Connect(); err != nil {
		return fmt.Errorf("dcerpc bind: %w", err)
	}

	c.callID = 1
	c.contextID = 0

	bind := &pdu.Bind{
		MaxXmitFrag:  c.transport.MaxXmitFrag(),
		MaxRecvFrag:  c.transport.MaxRecvFrag(),
		AssocGroupID: 0,
		ContextList: []pdu.ContextElement{
			{
				ContextID:        c.contextID,
				AbstractSyntax:   abstractSyntax,
				TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
			},
		},
	}
	bind.Header = pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID)

	raw, err := bind.Marshal()
	if err != nil {
		return fmt.Errorf("dcerpc bind: marshal: %w", err)
	}
	if err := c.transport.Send(raw); err != nil {
		return fmt.Errorf("dcerpc bind: send: %w", err)
	}

	respFrag, err := c.readFragment(&fragmentReader{t: c.transport})
	if err != nil {
		return fmt.Errorf("dcerpc bind: recv: %w", err)
	}

	hdr, err := pdu.PeekHeader(respFrag)
	if err != nil {
		return fmt.Errorf("dcerpc bind: %w", err)
	}
	switch hdr.PacketType {
	case pdu.PacketTypeBindAck:
		var ack pdu.BindAck
		if _, err := ack.Unmarshal(respFrag); err != nil {
			return fmt.Errorf("dcerpc bind: parse bind_ack: %w", err)
		}
		if !ack.Accepted() {
			return fmt.Errorf("dcerpc bind: server accepted no presentation context: %s", ack.String())
		}
		c.sendFragMax = negotiateSendMax(c.transport.MaxXmitFrag(), ack.MaxXmitFrag, ack.MaxRecvFrag)
		c.bound = true
		return nil
	case pdu.PacketTypeBindNak:
		var nak pdu.BindNak
		if _, err := nak.Unmarshal(respFrag); err != nil {
			return fmt.Errorf("dcerpc bind: parse bind_nak: %w", err)
		}
		return fmt.Errorf("dcerpc bind rejected: reject_reason=%d", nak.RejectReason)
	default:
		return fmt.Errorf("dcerpc bind: unexpected response PDU type %s", hdr.PacketType)
	}
}

// Call invokes the method identified by opnum, sending stub as its marshalled
// arguments and returning the marshalled results. A fault PDU is returned as a
// *pdu.Fault error.
func (c *Client) Call(opnum uint16, stub []byte) ([]byte, error) {
	if !c.bound {
		return nil, fmt.Errorf("dcerpc call: not bound; call Bind first")
	}

	c.callID++

	if err := c.sendRequest(opnum, stub); err != nil {
		return nil, fmt.Errorf("dcerpc call (opnum %d): %w", opnum, err)
	}

	result, err := c.readResponse()
	if err != nil {
		return nil, fmt.Errorf("dcerpc call (opnum %d): %w", opnum, err)
	}
	return result, nil
}

// Invoke is the declarative counterpart to Call: it marshals the NDR request
// structure in (whose Opnum selects the method and whose exported fields are the [in]
// parameters), issues the call, and unmarshals the response into out (a pointer to
// the [out] parameter structure). out may be nil when the response carries no data.
// A fault PDU is returned as a *pdu.Fault error, as with Call.
func (c *Client) Invoke(in ndr.Call, out any) error {
	stub, err := ndr.Request(in)
	if err != nil {
		return fmt.Errorf("dcerpc invoke (opnum %d): marshal request: %w", in.Opnum(), err)
	}
	resp, err := c.Call(in.Opnum(), stub)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := ndr.Response(resp, out); err != nil {
		return fmt.Errorf("dcerpc invoke (opnum %d): unmarshal response: %w", in.Opnum(), err)
	}
	return nil
}

// sendRequest fragments stub into request PDUs no larger than the negotiated send
// size and writes each one.
func (c *Client) sendRequest(opnum uint16, stub []byte) error {
	// Largest stub chunk per fragment.
	budget := int(c.sendFragMax) - requestHeaderOverhead
	if budget <= 0 {
		return fmt.Errorf("negotiated fragment size %d too small for a request", c.sendFragMax)
	}

	// A zero-length stub still requires one (empty) fragment.
	for first, offset := true, 0; first || offset < len(stub); first = false {
		end := offset + budget
		if end > len(stub) {
			end = len(stub)
		}
		chunk := stub[offset:end]

		var flags pdu.PFCFlags
		if first {
			flags |= pdu.PFCFirstFrag
		}
		if end >= len(stub) {
			flags |= pdu.PFCLastFrag
		}

		req := &pdu.Request{
			ContextID: c.contextID,
			Opnum:     opnum,
			AllocHint: uint32(len(stub) - offset), // bytes remaining from this fragment on
			Stub:      chunk,
		}
		req.Header = pdu.NewHeader(pdu.PacketTypeRequest, flags, c.callID)

		raw, err := req.Marshal()
		if err != nil {
			return fmt.Errorf("marshal request fragment: %w", err)
		}
		if err := c.transport.Send(raw); err != nil {
			return fmt.Errorf("send request fragment: %w", err)
		}

		offset = end
	}
	return nil
}

// readResponse reassembles response fragments until PFC_LAST_FRAG, returning the
// concatenated stub data. A fault PDU is returned as a *pdu.Fault error.
func (c *Client) readResponse() ([]byte, error) {
	fr := &fragmentReader{t: c.transport}
	var stub []byte

	for {
		frag, err := c.readFragment(fr)
		if err != nil {
			return nil, err
		}

		hdr, err := pdu.PeekHeader(frag)
		if err != nil {
			return nil, err
		}

		switch hdr.PacketType {
		case pdu.PacketTypeResponse:
			var resp pdu.Response
			if _, err := resp.Unmarshal(frag); err != nil {
				return nil, fmt.Errorf("parse response: %w", err)
			}
			if resp.Header.CallID != c.callID {
				return nil, fmt.Errorf("response call_id %d does not match request call_id %d", resp.Header.CallID, c.callID)
			}
			stub = append(stub, resp.Stub...)
			if hdr.PacketFlags.Has(pdu.PFCLastFrag) {
				return stub, nil
			}
		case pdu.PacketTypeFault:
			var fault pdu.Fault
			if _, err := fault.Unmarshal(frag); err != nil {
				return nil, fmt.Errorf("parse fault: %w", err)
			}
			return nil, &fault
		default:
			return nil, fmt.Errorf("unexpected response PDU type %s", hdr.PacketType)
		}
	}
}

// readFragment reads one complete PDU from fr.
func (c *Client) readFragment(fr *fragmentReader) ([]byte, error) {
	return fr.next()
}

// Close closes the underlying transport.
func (c *Client) Close() error {
	c.bound = false
	return c.transport.Close()
}

// negotiateSendMax derives the maximum request fragment size from our proposal and
// the server's bind_ack values. A request fragment must fit within what the server
// will receive, so the smallest of the relevant limits is used.
func negotiateSendMax(ourXmit, ackXmit, ackRecv uint16) uint16 {
	m := ourXmit
	if ackXmit != 0 && ackXmit < m {
		m = ackXmit
	}
	if ackRecv != 0 && ackRecv < m {
		m = ackRecv
	}
	return m
}

// fragmentReader turns the transport's read-sized Recv into a stream of complete
// PDUs, buffering bytes that span or straddle reads.
type fragmentReader struct {
	t   transport.Transport
	buf []byte
}

// next returns the next complete PDU, reading from the transport as needed.
func (fr *fragmentReader) next() ([]byte, error) {
	if err := fr.fill(pdu.HeaderSize); err != nil {
		return nil, err
	}
	hdr, err := pdu.PeekHeader(fr.buf)
	if err != nil {
		return nil, err
	}
	fragLen := int(hdr.FragLength)
	if fragLen < pdu.HeaderSize {
		return nil, fmt.Errorf("invalid frag_length %d", fragLen)
	}
	if err := fr.fill(fragLen); err != nil {
		return nil, err
	}
	frag := fr.buf[:fragLen]
	fr.buf = fr.buf[fragLen:]
	return frag, nil
}

// fill reads from the transport until the buffer holds at least n bytes.
func (fr *fragmentReader) fill(n int) error {
	for len(fr.buf) < n {
		chunk, err := fr.t.Recv()
		if err != nil {
			return err
		}
		if len(chunk) == 0 {
			return fmt.Errorf("connection closed after %d of %d expected bytes", len(fr.buf), n)
		}
		fr.buf = append(fr.buf, chunk...)
	}
	return nil
}
