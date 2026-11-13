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
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// requestHeaderOverhead is the fixed size of a request PDU before the stub data: the
// 16-byte common header plus the 8-byte request body (alloc_hint, p_cont_id, opnum).
const requestHeaderOverhead = pdu.HeaderSize + 8

// maxResponseStubSize caps the stub data readResponse reassembles for one call. A
// server that keeps sending response fragments without PFC_LAST_FRAG is otherwise
// buffered without limit; 64 MiB is well above legitimate responses (DRSUAPI
// GetNCChanges batches are a few MiB).
const maxResponseStubSize = 64 << 20

// Client is a connection-oriented DCE/RPC client bound to a single presentation
// context on a transport.
type Client struct {
	transport transport.Transport

	// callID is the call_id of the bind PDU (always 1). Per-call ids are allocated
	// from nextCallID by the demultiplexer.
	callID uint32

	// contextID is the presentation context id negotiated at Bind (0 for v1).
	contextID uint16

	// sendFragMax is the largest request fragment, in bytes, the server will accept,
	// derived from the bind_ack.
	sendFragMax uint16

	// recvFragMax is the largest response fragment, in bytes, the client accepts:
	// the larger of the max_recv_frag proposed at Bind and the max_xmit_frag the
	// server announced in its bind_ack. Windows announces and sends 2048-byte
	// fragments even when the client proposed less.
	recvFragMax uint16

	// preferNDR64 makes Bind additionally propose the NDR64 transfer syntax and prefer
	// it when the server accepts it. Default false: NDR20 only, identical to before.
	preferNDR64 bool

	// negotiatedSyntax is the transfer syntax the server accepted at Bind; Invoke
	// marshals and unmarshals stubs under it. Default NDR20.
	negotiatedSyntax ndr.Syntax

	// Response demultiplexer (demux.go): a single reader goroutine, started on the
	// first call after a successful bind, routes responses/faults to the call that
	// owns each call_id, so Call and Invoke are safe to use concurrently.
	readerOnce sync.Once  // starts the reader exactly once
	readerCond *sync.Cond // (on callsMu) wakes the parked reader when a request is on the wire or on close
	sendMu     sync.Mutex // serializes one call's request-fragment burst on the wire

	// serialCalls is set when the transport cannot carry more than one call in
	// flight at a time (transport.SerialCaller, e.g. an SMB named pipe). callSerialMu
	// then serializes a whole call exchange — request burst plus the wait for the
	// response — so concurrent Call/Invoke degrade to serial over such a transport
	// rather than interleaving request writes and response reads on one pipe (which
	// overruns it) or racing the per-connection SMB1 signing sequence number.
	serialCalls  bool
	callSerialMu sync.Mutex
	callsMu      sync.Mutex              // guards calls, nextCallID, readerErr, outstanding and closed
	calls        map[uint32]*pendingCall // in-flight calls, keyed by call_id (for routing)
	nextCallID   uint32                  // last allocated per-call id (bind used callID=1)
	outstanding  int                     // requests sent but not yet answered; gates the reader
	readerErr    error                   // terminal transport/protocol error, once set
	closed       bool                    // set by Close; stops the reader

	bound bool

	// relayPending records that BindRelayNegotiate has sent a relayed bind and is
	// awaiting the victim's AUTHENTICATE via BindRelayAuthenticate. See relay.go.
	relayPending bool

	// Authentication state, configured by SetAuth and established during Bind. When
	// authType is pdu.AuthTypeNone the bind and calls are anonymous, exactly as before.
	authType      uint8
	authLevel     uint8
	authContextID uint32
	workstation   string
	creds         *credentials.Credentials
	sec           SecurityContext // non-nil once an authenticated bind completes
	sessionKey    []byte          // NTLM exported session key, captured at completeAuth
	// bindToken is the auth_value carried in the bind PDU for a single-leg provider
	// (e.g. Netlogon) configured via SetAuthProvider; nil for the NTLM multi-leg flow.
	bindToken []byte
	// ntlmNegotiate is the NTLM NEGOTIATE message sent in the bind, kept as sent
	// because the AUTHENTICATE MIC is computed over it (MS-NLMP 3.1.5.1.2).
	ntlmNegotiate []byte
	// ntlmTargetName is the SPN the NTLM AUTHENTICATE asserts in MsvAvTargetName,
	// set by SetNTLMTargetName; empty selects "cifs/<DnsComputerName>".
	ntlmTargetName string

	// proposeFeatures and proposedFeatureBits, set by ProposeBindTimeFeatures,
	// offer bind-time feature negotiation; bindFeatures and bindFeaturesAcked
	// hold the server's answer.
	proposeFeatures     bool
	proposedFeatureBits uint8
	bindFeatures        uint16
	bindFeaturesAcked   bool

	// proposeConcMpx makes Bind set PFC_CONC_MPX in the bind's pfc_flags, offering
	// concurrent multiplexing on the association; concMpxAcked records whether the
	// server echoed PFC_CONC_MPX in the bind_ack's pfc_flags. Set via
	// EnableConcurrentMultiplexing. This only negotiates the capability — issuing
	// several calls at once is a separate step.
	proposeConcMpx bool
	concMpxAcked   bool
}

// SessionKey returns the session key established during an authenticated Bind — the
// NTLM exported session key, or for Kerberos the GSS context key (the negotiated
// subkey) — or nil if the bind was anonymous or has not completed. It is the key DRSUAPI
// uses to decrypt replicated secrets (the OID/PEK unwrap in IDL_DRSGetNCChanges); most
// callers do not need it. The returned slice is the client's own copy — do not mutate it.
func (c *Client) SessionKey() []byte { return c.sessionKey }

// NewClient returns a DCE/RPC client over the supplied transport. The transport must
// be ready to Connect (for ncacn_np, its SMB session and IPC$ tree connect are
// already established).
func NewClient(t transport.Transport) *Client {
	// The demultiplexer's shared state is initialized here, before any goroutine
	// exists, so the reader goroutine, Call and Close all observe it without a data
	// race (lazy init inside the reader's sync.Once would race a concurrent Close
	// that reads readerCond under callsMu).
	c := &Client{transport: t, calls: make(map[uint32]*pendingCall)}
	c.readerCond = sync.NewCond(&c.callsMu)
	// A transport that cannot multiplex calls (an SMB named pipe) makes the client
	// serialize whole call exchanges; see call and callSerialMu.
	if sc, ok := t.(transport.SerialCaller); ok {
		c.serialCalls = sc.SerializesCalls()
	}
	return c
}

// PreferNDR64 controls whether Bind proposes the NDR64 transfer syntax in addition to
// NDR 2.0 and uses it when the server accepts it. It must be called before Bind. The
// default is false (NDR 2.0 only).
func (c *Client) PreferNDR64(enable bool) { c.preferNDR64 = enable }

// ProposeBindTimeFeatures makes Bind offer bind-time feature negotiation
// ([MS-RPCE] 3.3.1.5.3) with the given bits of the first bitmask octet
// (syntax.SecurityContextMultiplexingSupported,
// syntax.KeepConnectionOnOrphanSupported). A client sets only the bits of
// features it supports, since the server answers for those alone. This client
// implements neither feature yet, so the answer is informational: it tells the
// caller what the server supports and changes nothing about later calls. It
// must be called before Bind.
func (c *Client) ProposeBindTimeFeatures(bits uint8) {
	c.proposeFeatures, c.proposedFeatureBits = true, bits
}

// BindTimeFeatures reports what the server answered to bind-time feature
// negotiation: its BindTimeFeatureNegotiationResponseBitmask, and whether it
// acknowledged the negotiation at all. It is meaningful only after a successful
// Bind that proposed the negotiation.
func (c *Client) BindTimeFeatures() (bits uint16, acknowledged bool) {
	return c.bindFeatures, c.bindFeaturesAcked
}

// NegotiatedSyntax reports the transfer syntax the server accepted at Bind. It is
// meaningful only after a successful Bind.
func (c *Client) NegotiatedSyntax() ndr.Syntax { return c.negotiatedSyntax }

// EnableConcurrentMultiplexing makes Bind offer concurrent multiplexing by setting
// PFC_CONC_MPX in the bind's pfc_flags ([C706] 12.6.3.1). The server indicates
// support by echoing PFC_CONC_MPX in its bind_ack, which ConcurrentMultiplexingSupported
// then reports. It must be called before Bind. This only negotiates the capability;
// the client still issues one call at a time until the concurrent-call path lands.
func (c *Client) EnableConcurrentMultiplexing() { c.proposeConcMpx = true }

// ConcurrentMultiplexingSupported reports whether the server acknowledged
// PFC_CONC_MPX in its bind_ack. It is meaningful only after a successful Bind that
// proposed it via EnableConcurrentMultiplexing.
func (c *Client) ConcurrentMultiplexingSupported() bool { return c.concMpxAcked }

// Bind connects the transport and binds to the interface identified by
// abstractSyntax. It proposes the NDR 2.0 transfer syntax in presentation context 0,
// and — when PreferNDR64 is enabled — NDR64 in a second context (the established
// negotiation pattern: one transfer syntax per context, [MS-RPCE] 2.2.2.4). The
// transfer syntax the server accepts is recorded for subsequent calls, preferring
// NDR64 when both are accepted. It returns an error if the server rejects the bind
// (bind_nak) or accepts no context.
func (c *Client) Bind(abstractSyntax syntax.SyntaxID) error {
	if err := c.transport.Connect(); err != nil {
		return fmt.Errorf("dcerpc bind: %w", err)
	}

	c.callID = 1
	c.nextCallID = 1 // per-call ids start at 2; the bind used 1
	c.bindFeatures, c.bindFeaturesAcked = 0, false
	c.concMpxAcked = false

	// Context 0 is always NDR 2.0; an optional context 1 offers NDR64. Each context
	// carries a single transfer syntax so the bind_ack's per-context result identifies
	// exactly which one the server chose.
	contexts := []pdu.ContextElement{
		{
			ContextID:        0,
			AbstractSyntax:   abstractSyntax,
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		},
	}
	if c.preferNDR64 {
		contexts = append(contexts, pdu.ContextElement{
			ContextID:        1,
			AbstractSyntax:   abstractSyntax,
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDR64TransferSyntax()},
		})
	}
	// The negotiation rides in one more context naming the same interface,
	// whose single transfer syntax carries the proposed bits.
	featureContext := -1
	if c.proposeFeatures {
		featureContext = len(contexts)
		contexts = append(contexts, pdu.ContextElement{
			ContextID:        uint16(len(contexts)),
			AbstractSyntax:   abstractSyntax,
			TransferSyntaxes: []syntax.SyntaxID{syntax.BindTimeFeatureNegotiation(c.proposedFeatureBits)},
		})
	}

	bind := &pdu.Bind{
		MaxXmitFrag:  c.transport.MaxXmitFrag(),
		MaxRecvFrag:  c.transport.MaxRecvFrag(),
		AssocGroupID: 0,
		ContextList:  contexts,
	}
	bindFlags := pdu.PFCFirstFrag | pdu.PFCLastFrag
	if c.proposeConcMpx {
		bindFlags |= pdu.PFCConcMpx
	}
	bind.Header = pdu.NewHeader(pdu.PacketTypeBind, bindFlags, c.callID)

	// For an authenticated bind, carry an NTLM NEGOTIATE token in the bind's auth
	// verifier; the server's CHALLENGE comes back in the bind_ack and is completed with
	// an auth3 below.
	if c.authConfigured() {
		negToken, err := c.negotiateToken()
		if err != nil {
			return fmt.Errorf("dcerpc bind: ntlm negotiate: %w", err)
		}
		bind.SecTrailer = pdu.SecTrailer{AuthType: c.authType, AuthLevel: c.authLevel, AuthContextID: c.authContextID}
		bind.AuthValue = negToken
	}

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
		if featureContext >= 0 && featureContext < len(ack.Results) && ack.Results[featureContext].Result == pdu.ResultNegotiateAck {
			c.bindFeatures, c.bindFeaturesAcked = ack.Results[featureContext].Reason, true
		}
		// The server advertises concurrent-multiplexing support by echoing
		// PFC_CONC_MPX in the bind_ack's pfc_flags ([C706] 12.6.3.1). Record it only
		// when we proposed it, so a server that sets the bit unsolicited does not
		// read as a negotiated capability.
		if c.proposeConcMpx && hdr.PacketFlags.Has(pdu.PFCConcMpx) {
			c.concMpxAcked = true
		}
		ctxID, negotiated, ok := selectContext(bind.ContextList, ack.Results)
		if !ok {
			return fmt.Errorf("dcerpc bind: server accepted no presentation context: %s", ack.String())
		}
		c.contextID = ctxID
		c.negotiatedSyntax = negotiated
		c.sendFragMax = negotiateSendMax(c.transport.MaxXmitFrag(), ack.MaxXmitFrag, ack.MaxRecvFrag)
		c.recvFragMax = max(c.transport.MaxRecvFrag(), ack.MaxXmitFrag)
		if c.authConfigured() {
			continuation, err := c.completeAuth(respFrag)
			if err != nil {
				return fmt.Errorf("dcerpc bind: auth: %w", err)
			}
			// Kerberos GSS_C_DCE_STYLE requires a third leg: the initiator's AP-REP,
			// carried in an alter_context PDU, before the context is usable.
			if len(continuation) > 0 {
				if err := c.sendAuthContinuation(bind.ContextList, continuation); err != nil {
					return fmt.Errorf("dcerpc bind: auth continuation: %w", err)
				}
			}
			// A provider whose key is only final once the handshake completes
			// (Kerberos adopts the acceptor's subkey from its AP-REP) reports it
			// now, for callers such as DRSUAPI that encrypt with the session key.
			if keyed, ok := c.sec.(sessionKeyer); ok {
				c.sessionKey = append([]byte(nil), keyed.SessionKey()...)
			}
		}
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
// *pdu.Fault error. Call is safe to use from several goroutines at once: each call
// gets its own call_id and its response is routed back by the demultiplexer
// (demux.go).
func (c *Client) Call(opnum uint16, stub []byte) ([]byte, error) {
	result, _, err := c.call(opnum, stub)
	return result, err
}

// call is the shared implementation of Call and Invoke. It also returns the
// response's byte order so Invoke can decode the stub under it without a shared
// field. It allocates a call_id, registers the pending call, sends the request
// fragments under the send lock, and waits for the demultiplexer to deliver the
// response, fault, or a terminal connection error.
func (c *Client) call(opnum uint16, stub []byte) ([]byte, binary.ByteOrder, error) {
	c.callsMu.Lock()
	bound := c.bound
	c.callsMu.Unlock()
	if !bound {
		return nil, nil, fmt.Errorf("dcerpc call: not bound; call Bind first")
	}
	c.ensureReader()

	// Over a transport that cannot multiplex calls (an SMB named pipe), run one call
	// at a time: hold callSerialMu across the whole exchange so another call's request
	// writes never interleave with this call's response reads on the single pipe, and
	// the SMB1 signing sequence number is advanced by one goroutine at a time. On a
	// multiplexing transport (TCP, SMB2) serialCalls is false and calls run
	// concurrently. callSerialMu is the outermost lock (callsMu and sendMu nest under
	// it, and the reader goroutine never takes it), so it cannot deadlock.
	if c.serialCalls {
		c.callSerialMu.Lock()
		defer c.callSerialMu.Unlock()
	}

	pc := &pendingCall{done: make(chan callResult, 1)}

	c.callsMu.Lock()
	if c.readerErr != nil {
		err := c.readerErr
		c.callsMu.Unlock()
		return nil, nil, fmt.Errorf("dcerpc call (opnum %d): connection failed: %w", opnum, err)
	}
	c.nextCallID++
	pc.callID = c.nextCallID
	pc.contextID = c.contextID
	c.calls[pc.callID] = pc
	c.callsMu.Unlock()

	// One call's request fragments must reach the wire as an uninterrupted burst,
	// so another concurrent call cannot interleave its fragments mid-sequence.
	c.sendMu.Lock()
	err := c.sendRequest(pc.callID, opnum, stub)
	c.sendMu.Unlock()
	if err != nil {
		// The request never went out whole; drop the registration so a stray later
		// fragment is ignored and the reader is not left holding it.
		c.callsMu.Lock()
		delete(c.calls, pc.callID)
		c.callsMu.Unlock()
		return nil, nil, fmt.Errorf("dcerpc call (opnum %d): %w", opnum, err)
	}

	// Only now is the request — and so the response the reader will read — on the
	// wire, so mark it outstanding and wake the reader. Gating the reader on
	// "request sent" rather than "call registered" keeps it from racing into an
	// empty read against a synchronous transport that produces the reply during
	// Send. outstanding counts sent-but-unanswered calls; it balances one reply per
	// increment, so the reader never reads more PDUs than were produced.
	c.callsMu.Lock()
	c.outstanding++
	c.readerCond.Signal()
	c.callsMu.Unlock()

	res := <-pc.done
	if res.err != nil {
		return nil, nil, fmt.Errorf("dcerpc call (opnum %d): %w", opnum, res.err)
	}
	return res.stub, res.order, nil
}

// Invoke is the declarative counterpart to Call: it marshals the NDR request
// structure in (whose Opnum selects the method and whose exported fields are the [in]
// parameters), issues the call, and unmarshals the response into out (a pointer to
// the [out] parameter structure). out may be nil when the response carries no data.
// A fault PDU is returned as a *pdu.Fault error, as with Call.
func (c *Client) Invoke(in ndr.Call, out any) error {
	stub, err := ndr.RequestAs(in, c.negotiatedSyntax)
	if err != nil {
		return fmt.Errorf("dcerpc invoke (opnum %d): marshal request: %w", in.Opnum(), err)
	}
	resp, order, err := c.call(in.Opnum(), stub)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if order == nil {
		order = binary.LittleEndian
	}
	if err := ndr.ResponseAsOrder(resp, out, c.negotiatedSyntax, order); err != nil {
		return fmt.Errorf("dcerpc invoke (opnum %d): unmarshal response: %w", in.Opnum(), err)
	}
	return nil
}

// selectContext chooses the accepted presentation context from a bind_ack, preferring
// NDR64 when the server accepted it. It returns the context id to use on subsequent
// requests and the negotiated transfer syntax. The result list is positional with the
// proposed context list ([C706] section 12.6.3.1), so result i answers context i.
func selectContext(contexts []pdu.ContextElement, results []pdu.PresentationResult) (uint16, ndr.Syntax, bool) {
	var ndr20Ctx, ndr64Ctx uint16
	var ndr20OK, ndr64OK bool
	for i, r := range results {
		if r.Result != pdu.ResultAcceptance || i >= len(contexts) {
			continue
		}
		switch {
		case r.TransferSyntax.Equal(syntax.NDR64TransferSyntax()):
			ndr64Ctx, ndr64OK = contexts[i].ContextID, true
		case r.TransferSyntax.Equal(syntax.NDRTransferSyntax()):
			ndr20Ctx, ndr20OK = contexts[i].ContextID, true
		}
	}
	if ndr64OK {
		return ndr64Ctx, ndr.NDR64, true
	}
	if ndr20OK {
		return ndr20Ctx, ndr.NDR20, true
	}
	return 0, ndr.NDR20, false
}

// sendRequest fragments stub into request PDUs no larger than the negotiated send
// size, stamps each with callID, and writes them. The caller holds sendMu so the
// burst is not interleaved with another call's fragments.
func (c *Client) sendRequest(callID uint32, opnum uint16, stub []byte) error {
	// Largest stub chunk per fragment, reserving room for the auth verifier when each
	// request is signed or sealed.
	overhead := requestHeaderOverhead
	if c.protectsRequests() {
		overhead += c.authVerifierOverhead()
	}
	budget := int(c.sendFragMax) - overhead
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
		req.Header = pdu.NewHeader(pdu.PacketTypeRequest, flags, callID)

		var raw []byte
		var err error
		if c.protectsRequests() {
			raw, err = c.marshalProtectedRequest(req)
		} else {
			raw, err = req.Marshal()
		}
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

// readFragment reads one complete PDU from fr.
func (c *Client) readFragment(fr *fragmentReader) ([]byte, error) {
	return fr.next()
}

// Close closes the underlying transport and stops the demultiplexer reader. A
// reader parked while idle is woken by the closed flag; one blocked in a read is
// released by the transport close. Any in-flight call is failed by the reader on
// its way out.
func (c *Client) Close() error {
	c.callsMu.Lock()
	c.bound = false
	c.closed = true
	if c.readerCond != nil {
		c.readerCond.Signal()
	}
	c.callsMu.Unlock()
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
