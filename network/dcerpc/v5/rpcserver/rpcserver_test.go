package rpcserver

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// testSyntax is an interface identifier used only here. It is not any real
// interface's, so a test cannot pass by accident against a registered one.
func testSyntax() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x11112222, B: 0x3333, C: 0x4444, D: 0x5555, E: 0x666677778888},
		MajorVersion: 2,
		MinorVersion: 1,
	}
}

// otherSyntax is a second identifier, for the interface a Dispatcher does not
// serve.
func otherSyntax() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x99990000, B: 0xaaaa, C: 0xbbbb, D: 0xcccc, E: 0xddddeeeeffff},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// stubService answers one opnum with a fixed stub and reports the errors a
// Service is allowed to report for the others.
type stubService struct {
	answer []byte

	// calls counts the calls made, to check that a refused PDU never reaches the
	// interface.
	mutex sync.Mutex
	calls int
}

func (s *stubService) AbstractSyntax() syntax.SyntaxID { return testSyntax() }

func (s *stubService) Call(opnum uint16, stub []byte) ([]byte, error) {
	s.mutex.Lock()
	s.calls++
	s.mutex.Unlock()

	switch opnum {
	case 0:
		return s.answer, nil
	case 1:
		return nil, ErrBadStub
	case 2:
		return nil, errors.New("something else went wrong")
	default:
		return nil, ErrUnknownOpnum
	}
}

func (s *stubService) callCount() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.calls
}

// bindRequest builds a bind PDU offering one context per abstract syntax given,
// each offering NDR20.
func bindRequest(t *testing.T, callID uint32, maxRecvFrag uint16, abstracts ...syntax.SyntaxID) []byte {
	t.Helper()

	contexts := make([]pdu.ContextElement, 0, len(abstracts))
	for i, abstract := range abstracts {
		contexts = append(contexts, pdu.ContextElement{
			ContextID:        uint16(i),
			AbstractSyntax:   abstract,
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		})
	}

	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID),
		MaxXmitFrag: maxRecvFrag,
		MaxRecvFrag: maxRecvFrag,
		ContextList: contexts,
	}
	encoded, err := bind.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the bind: %v", err)
	}
	return encoded
}

// callRequest builds a complete, unfragmented request PDU.
func callRequest(t *testing.T, callID uint32, opnum uint16, stub []byte) []byte {
	t.Helper()

	request := &pdu.Request{
		Header: pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID),
		Opnum:  opnum,
		Stub:   stub,
	}
	encoded, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}
	return encoded
}

// splitPDUs walks a reply that may hold several PDUs back to back, as a byte
// stream transport delivers them, and returns each one.
func splitPDUs(t *testing.T, stream []byte) [][]byte {
	t.Helper()

	pdus := [][]byte{}
	for offset := 0; offset < len(stream); {
		header, err := pdu.PeekHeader(stream[offset:])
		if err != nil {
			t.Fatalf("failed to read the header of the PDU at offset %d: %v", offset, err)
		}
		length := int(header.FragLength)
		if length < pdu.HeaderSize || offset+length > len(stream) {
			t.Fatalf("PDU at offset %d claims a length of %d, which does not fit in the %d bytes remaining",
				offset, length, len(stream)-offset)
		}
		pdus = append(pdus, stream[offset:offset+length])
		offset += length
	}
	return pdus
}

func TestBindAcceptsTheServedInterface(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	reply, err := association.Handle(bindRequest(t, 7, 4280, testSyntax()))
	if err != nil {
		t.Fatalf("Handle returned an error for a well-formed bind: %v", err)
	}

	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a bind_ack: %v", err)
	}
	if ack.Header.CallID != 7 {
		t.Errorf("the bind_ack carries call id %d, want the request's 7", ack.Header.CallID)
	}
	if !ack.Header.PacketFlags.Has(pdu.PFCFirstFrag) || !ack.Header.PacketFlags.Has(pdu.PFCLastFrag) {
		t.Errorf("the bind_ack is flagged %s, want both PFC_FIRST_FRAG and PFC_LAST_FRAG", ack.Header.PacketFlags)
	}
	if len(ack.Results) != 1 {
		t.Fatalf("the bind_ack carries %d results, want one per offered context (1)", len(ack.Results))
	}
	if ack.Results[0].Result != pdu.ResultAcceptance {
		t.Errorf("the context was answered with result %d, want acceptance (%d)",
			ack.Results[0].Result, pdu.ResultAcceptance)
	}
	if !ack.Results[0].TransferSyntax.Equal(syntax.NDRTransferSyntax()) {
		t.Errorf("the accepted transfer syntax is %v, want NDR20", ack.Results[0].TransferSyntax)
	}
	if ack.AssocGroupID == 0 {
		t.Error("the bind_ack reports association group 0, but a client that asked for one needs a non-zero group")
	}
	if !ack.Accepted() {
		t.Error("BindAck.Accepted reports the bind was not accepted")
	}
}

func TestBindRejectsTheContextNamingAnotherInterfaceAndKeepsTheServedOne(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	// Two contexts, the second naming an interface this endpoint does not serve.
	// Both get a result, positionally, per [C706] 12.6.4.4.
	reply, err := association.Handle(bindRequest(t, 1, 4280, testSyntax(), otherSyntax()))
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a bind_ack: %v", err)
	}
	if len(ack.Results) != 2 {
		t.Fatalf("the bind_ack carries %d results, want one per offered context (2)", len(ack.Results))
	}
	if ack.Results[0].Result != pdu.ResultAcceptance {
		t.Errorf("the first context was answered with result %d, want acceptance", ack.Results[0].Result)
	}
	if ack.Results[1].Result != pdu.ResultProviderRejection {
		t.Errorf("the second context was answered with result %d, want provider rejection (%d)",
			ack.Results[1].Result, pdu.ResultProviderRejection)
	}
	if ack.Results[1].Reason != pdu.ReasonAbstractSyntaxNotSupported {
		t.Errorf("the second context was rejected for reason %d, want abstract syntax not supported (%d)",
			ack.Results[1].Reason, pdu.ReasonAbstractSyntaxNotSupported)
	}
}

func TestBindNamingOnlyAnotherInterfaceIsRefused(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	reply, err := association.Handle(bindRequest(t, 2, 4280, otherSyntax()))
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	nak := &pdu.BindNak{}
	if _, err := nak.Unmarshal(reply); err != nil {
		t.Fatalf("a bind naming no served interface was not answered with a bind_nak: %v", err)
	}
	if nak.Header.CallID != 2 {
		t.Errorf("the bind_nak carries call id %d, want the request's 2", nak.Header.CallID)
	}
	if len(nak.Versions) == 0 {
		t.Error("the bind_nak names no protocol version, so a client cannot tell what to retry with")
	}
}

func TestBindOfferingOnlyNDR64IsRefused(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	// NDR64 ([MS-RPCE] 2.2.4.4). Accepting it and then encoding NDR20 would hand
	// the client a reply it decodes as the wrong shape.
	ndr64 := syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x71710533, B: 0xbeba, C: 0x4937, D: 0x8319, E: 0xb5dbef9ccc36},
		MajorVersion: 1,
		MinorVersion: 0,
	}

	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 3),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []pdu.ContextElement{{
			ContextID:        0,
			AbstractSyntax:   testSyntax(),
			TransferSyntaxes: []syntax.SyntaxID{ndr64},
		}},
	}
	encoded, err := bind.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the bind: %v", err)
	}

	reply, err := association.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}
	nak := &pdu.BindNak{}
	if _, err := nak.Unmarshal(reply); err != nil {
		t.Fatalf("a bind offering only NDR64 was not answered with a bind_nak: %v", err)
	}
}

func TestAuthenticatedBindIsRefused(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 4),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []pdu.ContextElement{{
			ContextID:        0,
			AbstractSyntax:   testSyntax(),
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
		SecTrailer: pdu.SecTrailer{AuthType: 10, AuthLevel: 6},
		AuthValue:  []byte("NTLMSSP\x00 a negotiate token"),
	}
	encoded, err := bind.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the bind: %v", err)
	}

	reply, err := association.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	nak := &pdu.BindNak{}
	if _, err := nak.Unmarshal(reply); err != nil {
		t.Fatalf("an authenticated bind was not answered with a bind_nak: %v", err)
	}
	if nak.RejectReason != bindNakAuthenticationTypeNotRecognized {
		t.Errorf("the bind_nak reports reason %d, want authentication type not recognized (%d)",
			nak.RejectReason, bindNakAuthenticationTypeNotRecognized)
	}
}

func TestAlterContextIsAnsweredWithAnAlterContextResponse(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 5),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []pdu.ContextElement{{
			ContextID:        1,
			AbstractSyntax:   testSyntax(),
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
	}
	encoded, err := bind.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the bind: %v", err)
	}
	// Rewrite the packet type in place: Bind.Marshal forces PacketTypeBind, and
	// an alter_context has the same body.
	encoded[2] = byte(pdu.PacketTypeAlterContext)

	reply, err := association.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	header, err := pdu.PeekHeader(reply)
	if err != nil {
		t.Fatalf("failed to read the reply header: %v", err)
	}
	if header.PacketType != pdu.PacketTypeAlterContextResp {
		t.Errorf("an alter_context was answered with %s, want %s",
			header.PacketType, pdu.PacketTypeAlterContextResp)
	}
}

func TestRequestReturnsTheInterfacesStub(t *testing.T) {
	service := &stubService{answer: []byte("0123456789abcdef")}
	endpoint := New(service)
	association := endpoint.Open()

	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	reply, err := association.Handle(callRequest(t, 9, 0, []byte("in")))
	if err != nil {
		t.Fatalf("Handle returned an error for a well-formed request: %v", err)
	}

	response := &pdu.Response{}
	if _, err := response.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a response: %v", err)
	}
	if response.Header.CallID != 9 {
		t.Errorf("the response carries call id %d, want the request's 9", response.Header.CallID)
	}
	if !bytes.Equal(response.Stub, service.answer) {
		t.Errorf("the response carries stub % x, want the interface's % x", response.Stub, service.answer)
	}
	if response.AllocHint != uint32(len(service.answer)) {
		t.Errorf("the response's alloc_hint is %d, want the stub length %d",
			response.AllocHint, len(service.answer))
	}
}

func TestRequestFaultStatuses(t *testing.T) {
	cases := []struct {
		name   string
		opnum  uint16
		status uint32
	}{
		{"an opnum the interface does not implement", 42, pdu.NCASOpRngError},
		{"a stub the interface cannot decode", 1, pdu.NCASFaultNDR},
		{"any other failure in the interface", 2, pdu.NCASProtoError},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			endpoint := New(&stubService{})
			association := endpoint.Open()
			if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
				t.Fatalf("the bind failed: %v", err)
			}

			reply, err := association.Handle(callRequest(t, 1, test.opnum, nil))
			if err != nil {
				t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
			}

			fault := &pdu.Fault{}
			if _, err := fault.Unmarshal(reply); err != nil {
				t.Fatalf("the reply is not a fault: %v", err)
			}
			if fault.Status != test.status {
				t.Errorf("the fault reports %s, want %s",
					pdu.FaultStatus(fault.Status), pdu.FaultStatus(test.status))
			}
		})
	}
}

func TestFragmentedRequestIsRefusedWithoutReachingTheInterface(t *testing.T) {
	service := &stubService{answer: []byte("unreachable")}
	endpoint := New(service)
	association := endpoint.Open()
	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	// PFC_FIRST_FRAG without PFC_LAST_FRAG: the first fragment of a request whose
	// stub is incomplete. Handing it to the interface would decode a fragment as
	// a whole stub.
	request := &pdu.Request{
		Header: pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag, 1),
		Opnum:  0,
		Stub:   []byte("half a st"),
	}
	encoded, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}

	reply, err := association.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
	}

	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("a fragmented request was not answered with a fault: %v", err)
	}
	if fault.Status != pdu.NCASProtoError {
		t.Errorf("the fault reports %s, want %s", pdu.FaultStatus(fault.Status), pdu.FaultStatus(pdu.NCASProtoError))
	}
	if service.callCount() != 0 {
		t.Errorf("the interface was called %d times for a fragmented request, want 0", service.callCount())
	}
}

func TestUnexpectedPacketTypeIsFaulted(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	// An auth3 PDU, which only an authenticated association uses and which this
	// endpoint never negotiates.
	auth3 := &pdu.Auth3{
		Header:     pdu.NewHeader(pdu.PacketTypeAuth3, pdu.PFCFirstFrag|pdu.PFCLastFrag, 11),
		SecTrailer: pdu.SecTrailer{AuthType: 10, AuthLevel: 6},
		AuthValue:  []byte("a token"),
	}
	encoded, err := auth3.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the auth3: %v", err)
	}

	reply, err := association.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
	}

	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("an auth3 was not answered with a fault: %v", err)
	}
	if fault.Header.CallID != 11 {
		t.Errorf("the fault carries call id %d, want the request's 11", fault.Header.CallID)
	}
}

func TestSomethingThatIsNotAPDUIsRejected(t *testing.T) {
	endpoint := New(&stubService{})
	association := endpoint.Open()

	if _, err := association.Handle([]byte("not a PDU")); err == nil {
		t.Error("Handle accepted nine bytes that are not a PDU")
	}
}

func TestLargeResponseIsFragmentedAndReassembles(t *testing.T) {
	// A stub several times the fragment size, with a recognisable pattern so a
	// reassembly that loses or duplicates a chunk shows up as a mismatch rather
	// than as the right length of the wrong bytes.
	answer := make([]byte, 3*DefaultMaxFragment+37)
	for i := range answer {
		answer[i] = byte(i * 7)
	}

	service := &stubService{answer: answer}
	endpoint := New(service)
	association := endpoint.Open()
	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	reply, err := association.Handle(callRequest(t, 2, 0, nil))
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	fragments := splitPDUs(t, reply)
	if len(fragments) < 4 {
		t.Fatalf("a %d-byte stub was sent in %d fragments at a %d-byte fragment size, want at least 4",
			len(answer), len(fragments), DefaultMaxFragment)
	}

	reassembled := []byte{}
	for i, fragment := range fragments {
		response := &pdu.Response{}
		if _, err := response.Unmarshal(fragment); err != nil {
			t.Fatalf("fragment %d is not a response: %v", i, err)
		}
		if len(fragment) > DefaultMaxFragment {
			t.Errorf("fragment %d is %d bytes, past the %d-byte fragment size the bind agreed",
				i, len(fragment), DefaultMaxFragment)
		}

		first := response.Header.PacketFlags.Has(pdu.PFCFirstFrag)
		last := response.Header.PacketFlags.Has(pdu.PFCLastFrag)
		if want := i == 0; first != want {
			t.Errorf("fragment %d has PFC_FIRST_FRAG %v, want %v", i, first, want)
		}
		if want := i == len(fragments)-1; last != want {
			t.Errorf("fragment %d has PFC_LAST_FRAG %v, want %v", i, last, want)
		}
		if want := uint32(len(answer) - len(reassembled)); response.AllocHint != want {
			t.Errorf("fragment %d reports alloc_hint %d, want the %d bytes still to come",
				i, response.AllocHint, want)
		}
		if !last && len(response.Stub)%8 != 0 {
			t.Errorf("fragment %d carries %d stub bytes, which is not a multiple of 8 and so moves the alignment of what follows",
				i, len(response.Stub))
		}
		reassembled = append(reassembled, response.Stub...)
	}

	if !bytes.Equal(reassembled, answer) {
		t.Errorf("the reassembled stub is %d bytes and does not match the %d the interface returned",
			len(reassembled), len(answer))
	}
}

func TestEachAssociationNegotiatesItsOwnFragmentSize(t *testing.T) {
	// Two clients of one endpoint, one with a small receive buffer and one with
	// the usual size. Each is answered at the size it asked for: what a bind
	// negotiates belongs to the association, so neither client's buffer bounds
	// the other's replies.
	endpoint := New(&stubService{answer: make([]byte, 4096)})

	small := endpoint.Open()
	if _, err := small.Handle(bindRequest(t, 1, 1024, testSyntax())); err != nil {
		t.Fatalf("the small client's bind failed: %v", err)
	}
	large := endpoint.Open()
	if _, err := large.Handle(bindRequest(t, 2, 4280, testSyntax())); err != nil {
		t.Fatalf("the large client's bind failed: %v", err)
	}

	smallReply, err := small.Handle(callRequest(t, 3, 0, nil))
	if err != nil {
		t.Fatalf("the small client's request failed: %v", err)
	}
	largeReply, err := large.Handle(callRequest(t, 4, 0, nil))
	if err != nil {
		t.Fatalf("the large client's request failed: %v", err)
	}

	smallFragments := splitPDUs(t, smallReply)
	for i, fragment := range smallFragments {
		if len(fragment) > 1024 {
			t.Errorf("the small client's fragment %d is %d bytes, past the 1024 it asked for", i, len(fragment))
		}
	}

	largeFragments := splitPDUs(t, largeReply)
	for i, fragment := range largeFragments {
		if len(fragment) > 4280 {
			t.Errorf("the large client's fragment %d is %d bytes, past the 4280 it asked for", i, len(fragment))
		}
	}

	// The point of the test: the small client's bind did not shrink the large
	// client's replies.
	if len(largeFragments) >= len(smallFragments) {
		t.Errorf("the large client's reply came in %d fragments and the small client's in %d, so the two are not negotiating separately",
			len(largeFragments), len(smallFragments))
	}
}

func TestAnAbsurdFragmentSizeFallsBackToTheDefault(t *testing.T) {
	endpoint := New(&stubService{answer: make([]byte, 8192)})
	association := endpoint.Open()

	// A max_recv_frag of 4 leaves no room for a PDU header, let alone a stub. It
	// is a client that filled the field in wrongly, and answering at that size
	// would make progress impossible.
	reply, err := association.Handle(bindRequest(t, 1, 4, testSyntax()))
	if err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a bind_ack: %v", err)
	}
	if ack.MaxRecvFrag != DefaultMaxFragment {
		t.Errorf("the bind_ack agreed a fragment size of %d, want the default %d", ack.MaxRecvFrag, DefaultMaxFragment)
	}
}

func TestMultiInterfaceEndpointDispatchesByPresentationContext(t *testing.T) {
	// Two interfaces on one endpoint, bound in one bind under two context ids.
	// Which interface answers a request is decided by the context it names, so
	// both are reachable over the same association and neither is guessed at.
	first := &stubService{answer: []byte("from the first")}
	second := &secondService{answer: []byte("from the second")}
	endpoint := New(first, second)
	association := endpoint.Open()

	reply, err := association.Handle(bindRequest(t, 1, 4280, testSyntax(), otherSyntax()))
	if err != nil {
		t.Fatalf("the bind failed: %v", err)
	}
	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a bind_ack: %v", err)
	}
	for i, result := range ack.Results {
		if result.Result != pdu.ResultAcceptance {
			t.Fatalf("context %d was answered with result %d, want acceptance", i, result.Result)
		}
	}
	if got := association.Bound(); got != 2 {
		t.Fatalf("the association reports %d bound contexts, want 2", got)
	}

	// bindRequest numbers the contexts by their position, so context 0 is the
	// first interface and context 1 the second.
	for contextID, want := range map[uint16][]byte{0: first.answer, 1: second.answer} {
		request := &pdu.Request{
			Header:    pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, 2),
			ContextID: contextID,
			Opnum:     0,
		}
		encoded, err := request.Marshal()
		if err != nil {
			t.Fatalf("failed to marshal the request on context %d: %v", contextID, err)
		}

		reply, err := association.Handle(encoded)
		if err != nil {
			t.Fatalf("the request on context %d failed: %v", contextID, err)
		}
		response := &pdu.Response{}
		if _, err := response.Unmarshal(reply); err != nil {
			t.Fatalf("the reply on context %d is not a response: %v", contextID, err)
		}
		if !bytes.Equal(response.Stub, want) {
			t.Errorf("context %d was answered by the wrong interface: stub %q, want %q",
				contextID, response.Stub, want)
		}
		if response.ContextID != contextID {
			t.Errorf("the response to context %d carries context %d", contextID, response.ContextID)
		}
	}
}

func TestRequestBeforeABindIsFaulted(t *testing.T) {
	// A request names its interface by a presentation context, and a context is
	// only meaningful once a bind has negotiated it. Answering one anyway would
	// mean guessing which interface the client meant.
	service := &stubService{answer: []byte("unreachable")}
	endpoint := New(service)
	association := endpoint.Open()

	if got := association.Bound(); got != 0 {
		t.Errorf("a fresh association reports %d bound contexts, want 0", got)
	}

	reply, err := association.Handle(callRequest(t, 1, 0, nil))
	if err != nil {
		t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
	}
	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a fault: %v", err)
	}
	if fault.Status != pdu.NCASFaultContextMismatch {
		t.Errorf("the fault reports %s, want %s",
			pdu.FaultStatus(fault.Status), pdu.FaultStatus(pdu.NCASFaultContextMismatch))
	}
	if service.callCount() != 0 {
		t.Errorf("the interface was called %d times before a bind, want 0", service.callCount())
	}
}

func TestRequestOnAContextTheBindRejectedIsFaulted(t *testing.T) {
	// The bind offers two contexts and only the first is accepted, so a request
	// on the second names a context that was refused.
	endpoint := New(&stubService{answer: []byte("x")})
	association := endpoint.Open()

	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax(), otherSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}
	if got := association.Bound(); got != 1 {
		t.Fatalf("the association reports %d bound contexts, want only the accepted one", got)
	}

	request := &pdu.Request{
		Header:    pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, 2),
		ContextID: 1,
		Opnum:     0,
	}
	encoded, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}

	reply, err := association.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
	}
	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a fault: %v", err)
	}
	if fault.Status != pdu.NCASFaultContextMismatch {
		t.Errorf("the fault reports %s, want %s",
			pdu.FaultStatus(fault.Status), pdu.FaultStatus(pdu.NCASFaultContextMismatch))
	}
}

func TestAlterContextAddsToWhatIsBoundAndABindReplacesIt(t *testing.T) {
	// [C706] 12.6.3.3: alter_context negotiates a new presentation context on an
	// existing association, so what was bound stays bound. A bind starts the
	// association's context table again.
	first := &stubService{answer: []byte("from the first")}
	second := &secondService{answer: []byte("from the second")}
	endpoint := New(first, second)
	association := endpoint.Open()

	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}
	if got := association.Bound(); got != 1 {
		t.Fatalf("after the bind the association reports %d bound contexts, want 1", got)
	}

	// An alter_context naming the second interface, under context id 1.
	alter := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 2),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []pdu.ContextElement{{
			ContextID:        1,
			AbstractSyntax:   otherSyntax(),
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
	}
	encoded, err := alter.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the alter_context: %v", err)
	}
	encoded[2] = byte(pdu.PacketTypeAlterContext)

	if _, err := association.Handle(encoded); err != nil {
		t.Fatalf("the alter_context failed: %v", err)
	}
	if got := association.Bound(); got != 2 {
		t.Errorf("after the alter_context the association reports %d bound contexts, want 2 — the first was dropped",
			got)
	}

	// A fresh bind naming only the second interface leaves that one context.
	rebind := bindRequest(t, 3, 4280, otherSyntax())
	if _, err := association.Handle(rebind); err != nil {
		t.Fatalf("the second bind failed: %v", err)
	}
	if got := association.Bound(); got != 1 {
		t.Errorf("after a second bind the association reports %d bound contexts, want 1", got)
	}
}

func TestTwoAssociationsOfOneEndpointBindIndependently(t *testing.T) {
	// The endpoint's interfaces are shared; what a bind establishes is not. One
	// client binding must not make the other client's requests answerable, which
	// is the whole reason an association exists.
	endpoint := New(&stubService{answer: []byte("answer")})

	bound := endpoint.Open()
	if _, err := bound.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}
	unbound := endpoint.Open()

	if got := unbound.Bound(); got != 0 {
		t.Errorf("the second association reports %d bound contexts after the first one bound, want 0", got)
	}

	reply, err := unbound.Handle(callRequest(t, 2, 0, nil))
	if err != nil {
		t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
	}
	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("the second association's request was answered rather than faulted: %v", err)
	}
	if fault.Status != pdu.NCASFaultContextMismatch {
		t.Errorf("the fault reports %s, want %s",
			pdu.FaultStatus(fault.Status), pdu.FaultStatus(pdu.NCASFaultContextMismatch))
	}

	// The association that did bind still works.
	if _, err := bound.Handle(callRequest(t, 3, 0, nil)); err != nil {
		t.Fatalf("the bound association's request failed: %v", err)
	}
}

func TestServicesReportsTheEndpointsInterfaces(t *testing.T) {
	endpoint := New(&stubService{}, &secondService{})

	services := endpoint.Services()
	if len(services) != 2 {
		t.Fatalf("Services reports %d interfaces, want 2", len(services))
	}
	// The slice is a copy, so a caller cannot reach into the endpoint's own.
	services[0] = nil
	if again := endpoint.Services(); again[0] == nil {
		t.Error("Services handed out the endpoint's own slice, so a caller can empty it")
	}
}

// secondService is a second interface, used to build an endpoint that serves
// more than one.
type secondService struct {
	answer []byte
}

func (s *secondService) AbstractSyntax() syntax.SyntaxID { return otherSyntax() }
func (s *secondService) Call(opnum uint16, stub []byte) ([]byte, error) {
	if opnum != 0 {
		return nil, ErrUnknownOpnum
	}
	return s.answer, nil
}

func TestConcurrentClientsOfOneEndpoint(t *testing.T) {
	// One Dispatcher serves every client of its endpoint, and the SMB server
	// calls a pipe handler on the goroutine of whichever connection is asking.
	// Each client has its own association; the endpoint's interface table is
	// shared, which is what this checks under -race.
	endpoint := New(&stubService{answer: make([]byte, 2048)})

	var waiting sync.WaitGroup
	for client := 0; client < 8; client++ {
		waiting.Add(1)
		go func(client int) {
			defer waiting.Done()

			association := endpoint.Open()
			for round := 0; round < 20; round++ {
				callID := uint32(client*100 + round)
				if _, err := association.Handle(bindRequest(t, callID, uint16(1024+client*64), testSyntax())); err != nil {
					t.Errorf("client %d: the bind failed: %v", client, err)
					return
				}
				if _, err := association.Handle(callRequest(t, callID, 0, nil)); err != nil {
					t.Errorf("client %d: the request failed: %v", client, err)
					return
				}
			}
		}(client)
	}
	waiting.Wait()
}

func TestOneAssociationUsedFromSeveralGoroutines(t *testing.T) {
	// A transport that let a client have several calls in flight would use one
	// association from several goroutines. The server does not, but the guard is
	// what makes that a supported thing to do rather than a corrupted context
	// table, and this is what checks it under -race.
	endpoint := New(&stubService{answer: make([]byte, 512)})
	association := endpoint.Open()
	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	var waiting sync.WaitGroup
	for caller := 0; caller < 8; caller++ {
		waiting.Add(1)
		go func(caller int) {
			defer waiting.Done()
			for round := 0; round < 20; round++ {
				if _, err := association.Handle(callRequest(t, uint32(caller*100+round), 0, nil)); err != nil {
					t.Errorf("caller %d: the request failed: %v", caller, err)
					return
				}
			}
		}(caller)
	}
	waiting.Wait()
}

func TestErrorsAreDistinguishable(t *testing.T) {
	// A Service reports the two conditions the Dispatcher maps to specific faults
	// by wrapping these sentinels, so wrapping has to survive errors.Is.
	wrapped := fmt.Errorf("decoding the parameters: %w", ErrBadStub)
	if !errors.Is(wrapped, ErrBadStub) {
		t.Error("a wrapped ErrBadStub is not recognised by errors.Is")
	}
	if errors.Is(wrapped, ErrUnknownOpnum) {
		t.Error("a wrapped ErrBadStub is mistaken for ErrUnknownOpnum")
	}
}
