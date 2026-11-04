package rpcserver

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
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

// DefaultMaxFragment is the reply fragment size used before a bind has named
// one, and when a bind names one too small to carry a PDU.
//
// 4280 is what Windows offers and what every client is prepared for.
const DefaultMaxFragment = 4280

// minFragmentStub is the smallest stub a fragment may carry. A fragment size that
// left no room for stub bytes would make progress impossible, so a client asking
// for one is given DefaultMaxFragment instead.
const minFragmentStub = 8

// responseBodyOverhead is the fixed part of a response PDU behind its header:
// alloc_hint(4) p_cont_id(2) cancel_count(1) reserved(1).
const responseBodyOverhead = 8

// MaxRequestStub bounds the stub a fragmented request may reassemble to. A
// request that grows past it is refused with a fault rather than buffered
// without limit. It is per association, and a client can open many, so it is
// kept well below what one connection could otherwise pin: the interfaces
// served here take requests of a few hundred bytes.
const MaxRequestStub = 1 << 20

// defaultAssocGroup is the association group handed out when a client asks for
// one.
const defaultAssocGroup = 0x00001234

// Dispatcher is the set of interfaces an endpoint answers for.
//
// It holds no per-call and no per-client state, so one Dispatcher serves every
// client of its endpoint and is safe for concurrent use. What a bind establishes
// belongs to an Association, which Open returns one of.
type Dispatcher struct {
	// services are the interfaces this endpoint answers for. Written once by New
	// and only read afterwards.
	services []Service

	// ntlm, set by EnableNTLM, configures authenticated binds. Nil refuses them.
	ntlm *NTLMAuth

	// onAcceptContext, set by OnAcceptContext, is invoked with each authenticated
	// bind's SPNEGO acceptor before its NEGOTIATE is answered, so a relay can drive
	// the exchange — setting ChallengeProvider to source the CHALLENGE from an
	// upstream target, and CredentialLookup to forward the captured AUTHENTICATE.
	// Nil leaves the acceptor driven solely by EnableNTLM. It mirrors the SMB
	// server's Config.OnAcceptContext.
	onAcceptContext func(accept *spnego.AcceptContext)
}

// New builds a Dispatcher for a set of interfaces.
//
// Parameters:
//   - services: the interfaces this endpoint answers for
//
// Returns:
//   - The dispatcher
func New(services ...Service) *Dispatcher {
	return &Dispatcher{services: services}
}

// Services returns the interfaces this endpoint answers for.
func (d *Dispatcher) Services() []Service {
	return append([]Service(nil), d.services...)
}

// Open starts an association: one client's conversation with this endpoint.
//
// Every transport that carries RPC has something an association corresponds to —
// one open of a named pipe, one TCP connection — and the caller is what knows
// which. A caller that handed PDUs straight to the Dispatcher would have nowhere
// to keep what a bind negotiated.
//
// Returns:
//   - A new association, with no presentation context bound yet
func (d *Dispatcher) Open() *Association {
	return &Association{
		dispatcher:  d,
		contexts:    map[uint16]Service{},
		maxFragment: DefaultMaxFragment,
	}
}

// Association is one client's conversation with an endpoint.
//
// It holds what a bind establishes and the requests after it are interpreted
// against: which interface each presentation context names, and the largest
// fragment this client will accept. Two clients of one endpoint have separate
// associations and negotiate separately.
//
// An association is guarded, and Handle processes one PDU at a time, so a
// transport that lets a client have several calls in flight cannot corrupt the
// context table or the security context: they are answered one after another.
type Association struct {
	// dispatcher is the endpoint this association is with.
	dispatcher *Dispatcher

	// mutex guards contexts and maxFragment.
	mutex sync.Mutex

	// handling serializes Handle. An authenticated association's security
	// context is stateful (RC4 streams and sequence numbers advance with every
	// PDU in wire order), so its PDUs must be processed one at a time and in the
	// order they arrived; a transport with several calls in flight gets them
	// answered one after another.
	handling sync.Mutex

	// contexts maps a presentation context identifier to the interface the
	// client bound it to. It is what makes a request's p_cont_id meaningful:
	// [C706] 12.6.4.9 has a request name its interface by the context it was
	// negotiated under, not by the interface's own identifier.
	contexts map[uint16]Service

	// maxFragment is the largest reply fragment this client said it can receive.
	maxFragment uint16

	// partial is the request being reassembled from its fragments, or nil. The
	// association does not multiplex calls, so at most one is in progress.
	partial *partialRequest

	// auth is what an authenticated bind established, or nil for an
	// unauthenticated association.
	auth *authState
}

// partialRequest is a fragmented request received up to, but not including, its
// last fragment.
type partialRequest struct {
	callID    uint32
	contextID uint16
	opnum     uint16
	stub      []byte
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
func (a *Association) Handle(request []byte) ([]byte, error) {
	a.handling.Lock()
	defer a.handling.Unlock()

	header, err := pdu.PeekHeader(request)
	if err != nil {
		return nil, fmt.Errorf("not a DCE/RPC PDU: %w", err)
	}

	switch header.PacketType {
	case pdu.PacketTypeBind, pdu.PacketTypeAlterContext:
		// A bind or alter_context ends any request still being assembled.
		a.dropPartial()
		return a.handleBind(request, header)

	case pdu.PacketTypeRequest:
		return a.handleRequest(request, header)

	case pdu.PacketTypeOrphaned:
		// [C706] chapter 12: "The orphaned PDU is used by a client to notify a
		// server that it is aborting a request in progress". The request being
		// assembled is abandoned and nothing is sent back.
		a.dropPartial()
		return nil, nil

	case pdu.PacketTypeAuth3:
		return a.handleAuth3(request)

	default:
		// A packet type this endpoint does not answer for. A fault is a reply the
		// client can act on; silence would leave it waiting.
		logger.Debugf("rpcserver: refusing packet type %s", header.PacketType)
		return a.fault(header.CallID, 0, pdu.NCASProtoError)
	}
}

// handleBind answers a bind or an alter_context.
//
// The difference between the two is what happens to the contexts already bound:
// a bind starts the association's context table afresh, and an alter_context adds
// to it ([C706] 12.6.3.3, which has alter_context "negotiate a new presentation
// context ... on an existing association").
func (a *Association) handleBind(request []byte, header *pdu.Header) ([]byte, error) {
	// An alter_context PDU is wire-identical to a bind but for the packet type at
	// offset 2 of the common header, and Bind.Unmarshal refuses anything else.
	// Rewriting the type on a copy is how the client sends one; this is the same
	// move on the receiving side, and the copy leaves the caller's buffer alone.
	altering := header.PacketType == pdu.PacketTypeAlterContext
	body := request
	if altering {
		body = make([]byte, len(request))
		copy(body, request)
		body[2] = byte(pdu.PacketTypeBind)
	}

	bind := &pdu.Bind{}
	if _, err := bind.Unmarshal(body); err != nil {
		logger.Debugf("rpcserver: a bind PDU would not decode: %v", err)
		// [C706] defines no alter_context_nak PDU, so a malformed
		// alter_context is a protocol error on a live association.
		if altering {
			return a.fault(header.CallID, 0, pdu.NCASProtoError)
		}
		return a.bindNak(header.CallID, bindNakProtocolVersionNotSupported)
	}

	// An authenticated bind starts an NTLM exchange when the endpoint accepts
	// one; otherwise it is refused rather than half-answered, since a bind that
	// succeeded without a security context would leave every later request
	// unverifiable. An alter_context does not renegotiate security.
	var auth *authState
	var challengeToken []byte
	if header.AuthLength > 0 {
		trailer, token, err := pdu.ExtractAuthVerifier(request)
		if err == nil && trailer != nil && !altering {
			auth, challengeToken, err = a.dispatcher.startAuth(trailer, token)
		} else if err == nil {
			err = fmt.Errorf("no security context can be negotiated here")
		}
		if err != nil {
			logger.Debugf("rpcserver: refusing an authenticated bind (%d bytes of verifier): %v", header.AuthLength, err)
			if altering {
				return a.fault(header.CallID, 0, pdu.NCASFaultAccessDenied)
			}
			return a.bindNak(header.CallID, bindNakAuthenticationTypeNotRecognized)
		}
	}
	// An association that negotiated security cannot be rebound without it:
	// that would downgrade every later request to unverified.
	a.mutex.Lock()
	protected := a.auth != nil
	a.mutex.Unlock()
	if protected && auth == nil && !altering {
		logger.Debugf("rpcserver: refusing an unauthenticated bind on an authenticated association")
		return a.bindNak(header.CallID, bindNakAuthenticationTypeNotRecognized)
	}

	if len(bind.ContextList) == 0 {
		if altering {
			return a.fault(header.CallID, 0, pdu.NCASProtoError)
		}
		return a.bindNak(header.CallID, bindNakReasonNotSpecified)
	}

	// Each context the client offered gets a result, in order: [C706] 12.6.4.4
	// pairs them positionally with the contexts of the request.
	ndrSyntax := syntax.NDRTransferSyntax()
	results := make([]pdu.PresentationResult, 0, len(bind.ContextList))
	bound := map[uint16]Service{}

	for _, context := range bind.ContextList {
		service := a.dispatcher.serviceFor(context.AbstractSyntax)
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
		bound[context.ContextID] = service
	}

	// An alter_context has no nak of its own: one whose contexts are all rejected
	// is still answered with an alter_context_resp carrying the per-context
	// rejections, and the association and the contexts it already bound carry on
	// unchanged.
	var negotiated uint16
	switch {
	case len(bound) > 0:
		negotiated = a.negotiate(bound, bind.MaxRecvFrag, altering, auth != nil || protected)
	case altering:
		logger.Debugf("rpcserver: no presentation context in the alter_context could be accepted")
		negotiated = a.fragment()
	default:
		logger.Debugf("rpcserver: no presentation context in the bind could be accepted")
		return a.bindNak(header.CallID, bindNakReasonNotSpecified)
	}

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
	if !altering {
		// A bind starts the association afresh, its security included.
		a.mutex.Lock()
		a.auth = auth
		a.partial = nil
		a.mutex.Unlock()
	}
	if auth != nil {
		ack.SecTrailer = pdu.SecTrailer{AuthType: auth.authType, AuthLevel: auth.authLevel, AuthContextID: auth.contextID}
		ack.AuthValue = challengeToken
	}

	encoded, err := ack.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the bind_ack: %w", err)
	}

	// BindAck.Marshal forces PTYPE=bind_ack, and an alter_context_resp is
	// wire-identical to it but for the packet type, so the answer to an
	// alter_context is relabelled the same way the request was read.
	if altering {
		encoded[2] = byte(pdu.PacketTypeAlterContextResp)
	}
	return encoded, nil
}

// handleRequest answers a request PDU by calling the interface its presentation
// context names.
func (a *Association) handleRequest(request []byte, header *pdu.Header) ([]byte, error) {
	incoming := &pdu.Request{}
	if _, err := incoming.Unmarshal(request); err != nil {
		logger.Debugf("rpcserver: a request PDU would not decode: %v", err)
		return a.fault(header.CallID, 0, pdu.NCASProtoError)
	}

	// On an authenticated association every fragment must arrive after a
	// verified AUTH3, and at the packet levels each carries a verifier that is
	// checked (and at the privacy level decrypted) before anything else reads
	// its stub.
	a.mutex.Lock()
	auth := a.auth
	a.mutex.Unlock()
	if auth != nil {
		if !auth.established {
			logger.Debugf("rpcserver: a request on an association whose authentication did not complete")
			return a.fault(header.CallID, incoming.ContextID, pdu.NCASFaultAccessDenied)
		}
		if auth.protectsMessages() {
			stub, err := auth.unprotectRequest(request, header, incoming.ObjectUUID != nil)
			if err != nil {
				logger.Debugf("rpcserver: a protected request was refused: %v", err)
				return a.fault(header.CallID, incoming.ContextID, pdu.NCASFaultAccessDenied)
			}
			incoming.Stub = stub
		} else if header.AuthLength > 0 {
			// At the connect level a verifier carries nothing to check but its
			// context, and its padding is not part of the stub.
			stub, err := auth.stripConnectVerifier(request, incoming.Stub)
			if err != nil {
				logger.Debugf("rpcserver: a connect-level verifier was refused: %v", err)
				return a.fault(header.CallID, incoming.ContextID, pdu.NCASFaultAccessDenied)
			}
			incoming.Stub = stub
		}
	}

	// A request in one fragment ends any request still being assembled, which
	// the client has evidently abandoned.
	if header.PacketFlags.Has(pdu.PFCFirstFrag) && header.PacketFlags.Has(pdu.PFCLastFrag) {
		a.dropPartial()
	}

	// A request in fragments is reassembled before any of it reaches a Service:
	// a partial stub would be decoded as a whole one. Every fragment but the
	// last is answered with nothing: the reply follows the whole request.
	if !header.PacketFlags.Has(pdu.PFCFirstFrag) || !header.PacketFlags.Has(pdu.PFCLastFrag) {
		stub, complete, status := a.reassemble(header, incoming)
		if status != pdu.NCASStatusOK {
			return a.fault(header.CallID, incoming.ContextID, status)
		}
		if !complete {
			return nil, nil
		}
		incoming.Stub = stub
	}

	// The context the request names is what says which interface it is for, so a
	// context this association never negotiated cannot be answered. That is a
	// request before a bind, or one carrying a context the bind rejected.
	service := a.serviceForContext(incoming.ContextID)
	if service == nil {
		logger.Debugf("rpcserver: request on presentation context %d, which this association has not bound",
			incoming.ContextID)
		return a.fault(header.CallID, incoming.ContextID, pdu.NCASFaultContextMismatch)
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
		return a.fault(header.CallID, incoming.ContextID, status)
	}

	return a.response(header.CallID, incoming.ContextID, stub)
}

// dropPartial abandons the request being reassembled, if any.
func (a *Association) dropPartial() {
	a.mutex.Lock()
	a.partial = nil
	a.mutex.Unlock()
}

// reassemble adds one fragment of a request to the request in progress, and
// returns the whole stub once the last fragment has arrived.
//
// The fragments of one call carry its call_id, and the first fixes the context
// and opnum the rest must repeat. A fragment that breaks
// either rule, a first fragment arriving while another call is incomplete, or a
// stub growing past MaxRequestStub abandons the request in progress and is
// answered with nca_s_proto_error.
//
// Parameters:
//   - header: the fragment's common header
//   - incoming: the decoded fragment
//
// Returns:
//   - The reassembled stub and true once complete, or false while fragments
//     remain; a fault status other than NCASStatusOK when the fragment is refused
func (a *Association) reassemble(header *pdu.Header, incoming *pdu.Request) ([]byte, bool, uint32) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if header.PacketFlags.Has(pdu.PFCFirstFrag) {
		// A context this association never bound cannot be answered, so the
		// request is refused before any of it is buffered.
		if _, bound := a.contexts[incoming.ContextID]; !bound {
			logger.Debugf("rpcserver: call %d starts on presentation context %d, which is not bound", header.CallID, incoming.ContextID)
			a.partial = nil
			return nil, false, pdu.NCASFaultContextMismatch
		}
		if a.partial != nil {
			logger.Debugf("rpcserver: call %d started while call %d was incomplete", header.CallID, a.partial.callID)
			a.partial = nil
			return nil, false, pdu.NCASProtoError
		}
		a.partial = &partialRequest{
			callID:    header.CallID,
			contextID: incoming.ContextID,
			opnum:     incoming.Opnum,
			stub:      append([]byte(nil), incoming.Stub...),
		}
		return nil, false, pdu.NCASStatusOK
	}

	partial := a.partial
	if partial == nil || partial.callID != header.CallID ||
		partial.contextID != incoming.ContextID || partial.opnum != incoming.Opnum {
		logger.Debugf("rpcserver: a fragment of call %d does not continue the call in progress", header.CallID)
		a.partial = nil
		return nil, false, pdu.NCASProtoError
	}
	if len(partial.stub)+len(incoming.Stub) > MaxRequestStub {
		logger.Debugf("rpcserver: call %d grew past %d bytes", header.CallID, MaxRequestStub)
		a.partial = nil
		return nil, false, pdu.NCASProtoError
	}
	partial.stub = append(partial.stub, incoming.Stub...)
	if !header.PacketFlags.Has(pdu.PFCLastFrag) {
		return nil, false, pdu.NCASStatusOK
	}
	a.partial = nil
	return partial.stub, true, pdu.NCASStatusOK
}

// response frames a response stub, splitting it across fragments if it exceeds
// what the client said it can receive.
func (a *Association) response(callID uint32, contextID uint16, stub []byte) ([]byte, error) {
	a.mutex.Lock()
	auth := a.auth
	a.mutex.Unlock()
	protect := auth != nil && auth.established && auth.protectsMessages()

	budget := int(a.fragment()) - pdu.HeaderSize - responseBodyOverhead
	if protect {
		budget -= authOverhead
	}
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

		if protect {
			encoded, err := auth.protectResponse(pdu.NewHeader(pdu.PacketTypeResponse, flags, callID),
				uint32(len(stub)-sent), contextID, stub[sent:sent+chunk])
			if err != nil {
				return nil, fmt.Errorf("failed to protect a response fragment: %w", err)
			}
			out = append(out, encoded...)
			sent += chunk
			continue
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

// fault frames a fault PDU.
func (a *Association) fault(callID uint32, contextID uint16, status uint32) ([]byte, error) {
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
func (a *Association) bindNak(callID uint32, reason uint16) ([]byte, error) {
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

// negotiate records what a bind established and returns the fragment size the
// association will send at.
//
// Parameters:
//   - bound: the contexts the bind accepted, by context identifier
//   - requested: the client's max_recv_frag
//   - altering: true for an alter_context, which adds to what is bound rather
//     than replacing it
//
// Returns:
//   - The fragment size replies will be framed at
func (a *Association) negotiate(bound map[uint16]Service, requested uint16, altering, authenticated bool) uint16 {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !altering {
		a.contexts = map[uint16]Service{}
	}
	for id, service := range bound {
		a.contexts[id] = service
	}
	a.maxFragment = negotiatedFragment(requested)
	// A protected fragment also carries padding, a sec_trailer and a signature;
	// a size too small for them falls back like any other unusable size, so the
	// bind_ack announces the size replies are really framed at.
	if authenticated && int(a.maxFragment) < pdu.HeaderSize+responseBodyOverhead+minFragmentStub+authOverhead {
		a.maxFragment = DefaultMaxFragment
	}
	return a.maxFragment
}

// serviceForContext returns the interface a presentation context names, or nil
// when this association has not bound that context.
func (a *Association) serviceForContext(contextID uint16) Service {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.contexts[contextID]
}

// fragment returns the size replies are framed at.
func (a *Association) fragment() uint16 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.maxFragment
}

// Bound reports how many presentation contexts this association has negotiated.
func (a *Association) Bound() int {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return len(a.contexts)
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

// offersSyntax reports whether a list of transfer syntaxes includes one.
func offersSyntax(offered []syntax.SyntaxID, wanted syntax.SyntaxID) bool {
	for _, candidate := range offered {
		if candidate.Equal(wanted) {
			return true
		}
	}
	return false
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
