package rpcserver

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// Service is one RPC interface a Dispatcher can answer for.
type Service interface {
	// AbstractSyntax identifies the interface and its version, which is what a
	// bind names and what a Dispatcher matches against.
	AbstractSyntax() syntax.SyntaxID

	// Call runs one method and returns its response stub.
	//
	// The returned error is turned into a fault: ErrUnknownOpnum becomes
	// nca_s_op_rng_error and ErrBadStub becomes nca_s_fault_ndr, so a Service can
	// report those two without knowing the fault codes. Anything else becomes
	// nca_s_proto_error.
	//
	// A method that fails at the protocol's own level — an unsupported
	// information level, say — is not an error here: it returns a stub carrying
	// the protocol's status code, because that is a successful call reporting a
	// failure and a client parses it as one.
	Call(opnum uint16, stub []byte) ([]byte, error)
}

// ErrUnknownOpnum reports a method the Service does not implement.
var ErrUnknownOpnum = errors.New("unknown opnum")

// ErrBadStub reports a request stub the Service could not decode.
var ErrBadStub = errors.New("request stub could not be decoded")

// DefaultMaxFragment is the reply fragment size used when a client's bind named
// none, or named one too small to carry a PDU header.
//
// 4280 is what Windows offers and what every client is prepared for.
const DefaultMaxFragment = 4280

// minFragmentStub is the smallest stub a fragment may carry. A fragment size that
// left no room for stub bytes would make progress impossible, so a client asking
// for one is given DefaultMaxFragment instead.
const minFragmentStub = 8

// Dispatcher answers RPC PDUs for a set of interfaces.
//
// One Dispatcher serves one endpoint — one named pipe, say — and is safe for
// concurrent use. It keeps no per-call state; the negotiated fragment size is the
// only thing a bind changes, and it is guarded, because a transport that cannot
// tell two opens of an endpoint apart delivers both clients' PDUs to one
// Dispatcher and may do so from two goroutines at once.
type Dispatcher struct {
	// services are the interfaces this endpoint answers for. Written once by New
	// and only read afterwards.
	services []Service

	// mutex guards maxFragment.
	mutex sync.Mutex

	// maxFragment is the largest reply fragment any client of this endpoint has
	// said it can receive.
	//
	// It is the smallest such value seen, not the latest: without per-open state
	// one size has to serve every client, and a fragment smaller than a client's
	// maximum is always acceptable to it while a larger one is not. So a client
	// asking for less shrinks the endpoint's fragments for everyone rather than
	// risking a reply another client cannot take.
	maxFragment uint16
}

// New builds a Dispatcher for a set of interfaces.
//
// Parameters:
//   - services: the interfaces this endpoint answers for
//
// Returns:
//   - The dispatcher
func New(services ...Service) *Dispatcher {
	return &Dispatcher{services: services, maxFragment: DefaultMaxFragment}
}

// Handle answers one received PDU and returns the reply.
//
// The reply may be several PDUs concatenated: a response larger than the client's
// fragment size is split, and the caller writes the whole thing back as one
// stream. That is what a byte-stream transport such as a named pipe wants — the
// client reads fragments off the pipe until it sees one with PFC_LAST_FRAG.
//
// Parameters:
//   - request: one complete received PDU
//
// Returns:
//   - The reply PDUs to send back, and an error only when no reply can be framed
func (d *Dispatcher) Handle(request []byte) ([]byte, error) {
	header, err := pdu.PeekHeader(request)
	if err != nil {
		return nil, fmt.Errorf("not a DCE/RPC PDU: %w", err)
	}

	switch header.PacketType {
	case pdu.PacketTypeBind, pdu.PacketTypeAlterContext:
		return d.handleBind(request, header)

	case pdu.PacketTypeRequest:
		return d.handleRequest(request, header)

	default:
		// A packet type this endpoint does not answer for. A fault is a reply the
		// client can act on; silence would leave it waiting.
		logger.Debugf("rpcserver: refusing packet type %s", header.PacketType)
		return d.fault(header.CallID, 0, pdu.NCASProtoError)
	}
}

// handleBind answers a bind or an alter_context.
//
// The two are answered the same way: alter_context re-negotiates presentation
// contexts on an existing association, and with no authentication to renegotiate
// there is nothing that makes it differ from a bind here.
func (d *Dispatcher) handleBind(request []byte, header *pdu.Header) ([]byte, error) {
	// An alter_context PDU is wire-identical to a bind but for the packet type at
	// offset 2 of the common header, and Bind.Unmarshal refuses anything else.
	// Rewriting the type on a copy is how the client sends one; this is the same
	// move on the receiving side, and the copy leaves the caller's buffer alone.
	body := request
	if header.PacketType == pdu.PacketTypeAlterContext {
		body = make([]byte, len(request))
		copy(body, request)
		body[2] = byte(pdu.PacketTypeBind)
	}

	bind := &pdu.Bind{}
	if _, err := bind.Unmarshal(body); err != nil {
		logger.Debugf("rpcserver: a bind PDU would not decode: %v", err)
		return d.bindNak(header.CallID, bindNakProtocolVersionNotSupported)
	}

	// An authenticated bind is refused rather than half-answered. See the package
	// documentation: completing one needs a security context this does not keep,
	// and a bind that succeeded without one would leave every later request
	// unverifiable.
	if header.AuthLength > 0 || len(bind.AuthValue) > 0 {
		logger.Debugf("rpcserver: refusing an authenticated bind (%d bytes of verifier)", header.AuthLength)
		return d.bindNak(header.CallID, bindNakAuthenticationTypeNotRecognized)
	}

	if len(bind.ContextList) == 0 {
		return d.bindNak(header.CallID, bindNakReasonNotSpecified)
	}

	// Each context the client offered gets a result, in order: [C706] 12.6.4.4
	// pairs them positionally with the contexts of the request.
	ndrSyntax := syntax.NDRTransferSyntax()
	results := make([]pdu.PresentationResult, 0, len(bind.ContextList))
	accepted := 0

	for _, context := range bind.ContextList {
		service := d.serviceFor(context.AbstractSyntax)
		if service == nil {
			results = append(results, pdu.PresentationResult{
				Result: contextResultProviderRejection,
				Reason: contextReasonAbstractSyntaxNotSupported,
			})
			continue
		}

		if !offersSyntax(context.TransferSyntaxes, ndrSyntax) {
			// NDR64 or nothing recognisable. Declining here is the only useful
			// answer: accepting and then encoding NDR would produce a reply the
			// client decodes as the wrong shape.
			results = append(results, pdu.PresentationResult{
				Result: contextResultProviderRejection,
				Reason: contextReasonTransferSyntaxesNotSupported,
			})
			continue
		}

		results = append(results, pdu.PresentationResult{
			Result:         contextResultAcceptance,
			TransferSyntax: ndrSyntax,
		})
		accepted++
	}

	if accepted == 0 {
		logger.Debugf("rpcserver: no presentation context in the bind could be accepted")
		return d.bindNak(header.CallID, bindNakReasonNotSpecified)
	}

	negotiated := d.observeFragment(bind.MaxRecvFrag)

	ack := &pdu.BindAck{
		Header:      pdu.NewHeader(pdu.PacketTypeBindAck, pdu.PFCFirstFrag|pdu.PFCLastFrag, header.CallID),
		MaxXmitFrag: negotiated,
		MaxRecvFrag: negotiated,
		// The association group is the client's when it named one, and 0x1234
		// otherwise; the value is opaque and only has to be non-zero and stable.
		AssocGroupID: assocGroup(bind.AssocGroupID),
		// The secondary address is the endpoint the association continues on.
		// Over a named pipe there is no second port to name, and an empty string
		// is what a pipe endpoint reports.
		SecondaryAddress: "",
		Results:          results,
	}
	encoded, err := ack.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the bind_ack: %w", err)
	}

	// BindAck.Marshal forces PTYPE=bind_ack, and an alter_context_resp is
	// wire-identical to it but for the packet type, so the answer to an
	// alter_context is relabelled the same way the request was read.
	if header.PacketType == pdu.PacketTypeAlterContext {
		encoded[2] = byte(pdu.PacketTypeAlterContextResp)
	}
	return encoded, nil
}

// handleRequest answers a request PDU by calling the interface it names.
func (d *Dispatcher) handleRequest(request []byte, header *pdu.Header) ([]byte, error) {
	incoming := &pdu.Request{}
	if _, err := incoming.Unmarshal(request); err != nil {
		logger.Debugf("rpcserver: a request PDU would not decode: %v", err)
		return d.fault(header.CallID, 0, pdu.NCASProtoError)
	}

	// A request arriving in fragments is refused rather than half-assembled.
	// Reassembly needs per-call state keyed by call ID, and a partial stub handed
	// to a Service would be decoded as a whole one.
	if !header.PacketFlags.Has(pdu.PFCFirstFrag) || !header.PacketFlags.Has(pdu.PFCLastFrag) {
		logger.Debugf("rpcserver: refusing a fragmented request (flags %s)", header.PacketFlags)
		return d.fault(header.CallID, incoming.ContextID, pdu.NCASProtoError)
	}

	// The endpoint decides the interface. A Dispatcher serves one endpoint, and
	// the context identifier a request carries was assigned by a bind this
	// Dispatcher may not have seen — a transport that cannot tell two opens apart
	// cannot keep per-bind state — so a single-interface endpoint dispatches by
	// what it serves rather than by what the client numbered it.
	service := d.only()
	if service == nil {
		logger.Debugf("rpcserver: request on an endpoint serving %d interfaces, which needs a bind context",
			len(d.services))
		return d.fault(header.CallID, incoming.ContextID, pdu.NCASUnkIf)
	}

	stub, err := service.Call(incoming.Opnum, incoming.Stub)
	if err != nil {
		status := pdu.NCASProtoError
		switch {
		case errors.Is(err, ErrUnknownOpnum):
			status = pdu.NCASOpRngError
		case errors.Is(err, ErrBadStub):
			status = pdu.NCASFaultNDR
		}
		logger.Debugf("rpcserver: opnum %d failed (%v), answering %s",
			incoming.Opnum, err, pdu.FaultStatus(status))
		return d.fault(header.CallID, incoming.ContextID, status)
	}

	return d.response(header.CallID, incoming.ContextID, stub)
}

// response frames a response stub, splitting it across fragments if it exceeds
// what the client said it can receive.
func (d *Dispatcher) response(callID uint32, contextID uint16, stub []byte) ([]byte, error) {
	budget := int(d.fragment()) - pdu.HeaderSize - responseBodyOverhead
	if budget < minFragmentStub {
		budget = DefaultMaxFragment - pdu.HeaderSize - responseBodyOverhead
	}
	// Fragments are kept to a multiple of eight so a stub's own alignment is not
	// disturbed by where a fragment happens to end.
	budget -= budget % 8

	out := []byte{}
	for first, sent := true, 0; first || sent < len(stub); first = false {
		chunk := len(stub) - sent
		if chunk > budget {
			chunk = budget
		}

		flags := pdu.PFCFlags(0)
		if first {
			flags |= pdu.PFCFirstFrag
		}
		if sent+chunk >= len(stub) {
			flags |= pdu.PFCLastFrag
		}

		reply := &pdu.Response{
			Header:    pdu.NewHeader(pdu.PacketTypeResponse, flags, callID),
			ContextID: contextID,
			// AllocHint on the first fragment tells the client the whole size, so
			// it can allocate once rather than growing per fragment.
			AllocHint: uint32(len(stub) - sent),
			Stub:      stub[sent : sent+chunk],
		}
		encoded, err := reply.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal a response fragment: %w", err)
		}
		out = append(out, encoded...)
		sent += chunk
	}
	return out, nil
}

// responseBodyOverhead is the fixed part of a response PDU behind its header:
// alloc_hint(4) p_cont_id(2) cancel_count(1) reserved(1).
const responseBodyOverhead = 8

// fault frames a fault PDU.
func (d *Dispatcher) fault(callID uint32, contextID uint16, status uint32) ([]byte, error) {
	f := &pdu.Fault{
		Header:    pdu.NewHeader(pdu.PacketTypeFault, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID),
		ContextID: contextID,
		Status:    status,
	}
	encoded, err := f.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal a fault: %w", err)
	}
	return encoded, nil
}

// bindNak frames a bind_nak carrying a reject reason.
func (d *Dispatcher) bindNak(callID uint32, reason uint16) ([]byte, error) {
	nak := &pdu.BindNak{
		Header:       pdu.NewHeader(pdu.PacketTypeBindNak, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID),
		RejectReason: reason,
		// The versions this endpoint speaks, which is what a client uses to
		// decide whether to retry differently.
		Versions: []pdu.ProtocolVersion{{Major: 5, Minor: 0}},
	}
	encoded, err := nak.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal a bind_nak: %w", err)
	}
	return encoded, nil
}

// serviceFor finds the interface a bind's abstract syntax names.
func (d *Dispatcher) serviceFor(abstract syntax.SyntaxID) Service {
	for _, service := range d.services {
		if service.AbstractSyntax().Equal(abstract) {
			return service
		}
	}
	return nil
}

// only returns the single interface this endpoint serves, or nil when it serves
// more than one and a request cannot be attributed without bind state.
func (d *Dispatcher) only() Service {
	if len(d.services) == 1 {
		return d.services[0]
	}
	return nil
}

// offersSyntax reports whether a list of transfer syntaxes includes one.
func offersSyntax(offered []syntax.SyntaxID, wanted syntax.SyntaxID) bool {
	for _, candidate := range offered {
		if candidate.Equal(wanted) {
			return true
		}
	}
	return false
}

// observeFragment records what a binding client can receive and returns the size
// the endpoint will use for it.
//
// Parameters:
//   - requested: the client's max_recv_frag
//
// Returns:
//   - The fragment size the endpoint will send at
func (d *Dispatcher) observeFragment(requested uint16) uint16 {
	usable := negotiatedFragment(requested)

	d.mutex.Lock()
	defer d.mutex.Unlock()
	if usable < d.maxFragment {
		d.maxFragment = usable
	}
	return d.maxFragment
}

// fragment returns the size replies are fragmented at.
func (d *Dispatcher) fragment() uint16 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.maxFragment
}

// negotiatedFragment bounds the client's requested fragment size.
//
// A client asking for less than a PDU header plus a usable stub is asking for
// something no reply fits in, so it is given the default instead: it is a client
// that filled the field in wrongly rather than one with a tiny buffer.
func negotiatedFragment(requested uint16) uint16 {
	if int(requested) < pdu.HeaderSize+responseBodyOverhead+minFragmentStub {
		return DefaultMaxFragment
	}
	return requested
}

// assocGroup echoes the client's association group, or invents one when it asked
// the server to.
//
// A client sends zero to mean "give me a group"; the value is opaque and only has
// to be non-zero, since this endpoint keeps no per-group state to look up.
func assocGroup(requested uint32) uint32 {
	if requested == 0 {
		return defaultAssocGroup
	}
	return requested
}

// defaultAssocGroup is the association group handed out when a client asks for
// one.
const defaultAssocGroup = 0x00001234

// The bind_nak reject reasons ([C706] 12.6.4.5).
const (
	bindNakReasonNotSpecified              uint16 = 0
	bindNakProtocolVersionNotSupported     uint16 = 4
	bindNakAuthenticationTypeNotRecognized uint16 = 8
)

// The presentation context results and rejection reasons ([C706] 12.6.4.4).
const (
	contextResultAcceptance        uint16 = 0
	contextResultProviderRejection uint16 = 2

	contextReasonAbstractSyntaxNotSupported   uint16 = 1
	contextReasonTransferSyntaxesNotSupported uint16 = 2
)
