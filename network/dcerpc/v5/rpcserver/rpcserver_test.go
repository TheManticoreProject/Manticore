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
	dispatcher := New(&stubService{})

	reply, err := dispatcher.Handle(bindRequest(t, 7, 4280, testSyntax()))
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
	dispatcher := New(&stubService{})

	// Two contexts, the second naming an interface this endpoint does not serve.
	// Both get a result, positionally, per [C706] 12.6.4.4.
	reply, err := dispatcher.Handle(bindRequest(t, 1, 4280, testSyntax(), otherSyntax()))
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
	dispatcher := New(&stubService{})

	reply, err := dispatcher.Handle(bindRequest(t, 2, 4280, otherSyntax()))
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
	dispatcher := New(&stubService{})

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

	reply, err := dispatcher.Handle(encoded)
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}
	nak := &pdu.BindNak{}
	if _, err := nak.Unmarshal(reply); err != nil {
		t.Fatalf("a bind offering only NDR64 was not answered with a bind_nak: %v", err)
	}
}

func TestAuthenticatedBindIsRefused(t *testing.T) {
	dispatcher := New(&stubService{})

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

	reply, err := dispatcher.Handle(encoded)
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
	dispatcher := New(&stubService{})

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

	reply, err := dispatcher.Handle(encoded)
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
	dispatcher := New(service)

	if _, err := dispatcher.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	reply, err := dispatcher.Handle(callRequest(t, 9, 0, []byte("in")))
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
			dispatcher := New(&stubService{})

			reply, err := dispatcher.Handle(callRequest(t, 1, test.opnum, nil))
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
	dispatcher := New(service)

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

	reply, err := dispatcher.Handle(encoded)
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
	dispatcher := New(&stubService{})

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

	reply, err := dispatcher.Handle(encoded)
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
	dispatcher := New(&stubService{})

	if _, err := dispatcher.Handle([]byte("not a PDU")); err == nil {
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
	dispatcher := New(service)
	if _, err := dispatcher.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	reply, err := dispatcher.Handle(callRequest(t, 2, 0, nil))
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

func TestASmallerFragmentSizeIsHonouredForEveryClientOfTheEndpoint(t *testing.T) {
	answer := make([]byte, 4096)
	service := &stubService{answer: answer}
	dispatcher := New(service)

	// One client says it can only take 1024-byte fragments. The endpoint cannot
	// tell two opens apart, so that bound has to apply to the whole endpoint: a
	// fragment smaller than a client's maximum is always acceptable, a larger one
	// is not.
	if _, err := dispatcher.Handle(bindRequest(t, 1, 1024, testSyntax())); err != nil {
		t.Fatalf("the first bind failed: %v", err)
	}
	if _, err := dispatcher.Handle(bindRequest(t, 2, 4280, testSyntax())); err != nil {
		t.Fatalf("the second bind failed: %v", err)
	}

	reply, err := dispatcher.Handle(callRequest(t, 3, 0, nil))
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	for i, fragment := range splitPDUs(t, reply) {
		if len(fragment) > 1024 {
			t.Errorf("fragment %d is %d bytes, past the 1024 the smaller client asked for", i, len(fragment))
		}
	}
}

func TestAnAbsurdFragmentSizeFallsBackToTheDefault(t *testing.T) {
	dispatcher := New(&stubService{answer: make([]byte, 8192)})

	// A max_recv_frag of 4 leaves no room for a PDU header, let alone a stub. It
	// is a client that filled the field in wrongly, and answering at that size
	// would make progress impossible.
	reply, err := dispatcher.Handle(bindRequest(t, 1, 4, testSyntax()))
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

func TestMultiInterfaceEndpointRefusesARequest(t *testing.T) {
	// Two interfaces on one endpoint. A request names its context by an
	// identifier assigned in a bind, and without per-open state there is no
	// table to look it up in, so the request cannot be attributed.
	other := &secondService{}
	dispatcher := New(&stubService{answer: []byte("x")}, other)

	reply, err := dispatcher.Handle(callRequest(t, 1, 0, nil))
	if err != nil {
		t.Fatalf("Handle returned an error instead of framing a fault: %v", err)
	}

	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("the reply is not a fault: %v", err)
	}
	if fault.Status != pdu.NCASUnkIf {
		t.Errorf("the fault reports %s, want %s", pdu.FaultStatus(fault.Status), pdu.FaultStatus(pdu.NCASUnkIf))
	}
}

// secondService is a second interface, used to build an endpoint that serves
// more than one.
type secondService struct{}

func (s *secondService) AbstractSyntax() syntax.SyntaxID { return otherSyntax() }
func (s *secondService) Call(opnum uint16, stub []byte) ([]byte, error) {
	return nil, ErrUnknownOpnum
}

func TestConcurrentClientsOfOneEndpoint(t *testing.T) {
	// One Dispatcher serves every open of its endpoint, and the SMB server calls
	// a pipe handler on the goroutine of whichever connection is asking. Binds
	// and requests therefore overlap, which is what this checks under -race.
	dispatcher := New(&stubService{answer: make([]byte, 2048)})

	var waiting sync.WaitGroup
	for client := 0; client < 8; client++ {
		waiting.Add(1)
		go func(client int) {
			defer waiting.Done()
			for round := 0; round < 20; round++ {
				callID := uint32(client*100 + round)
				if _, err := dispatcher.Handle(bindRequest(t, callID, uint16(1024+client*64), testSyntax())); err != nil {
					t.Errorf("client %d: the bind failed: %v", client, err)
					return
				}
				if _, err := dispatcher.Handle(callRequest(t, callID, 0, nil)); err != nil {
					t.Errorf("client %d: the request failed: %v", client, err)
					return
				}
			}
		}(client)
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
