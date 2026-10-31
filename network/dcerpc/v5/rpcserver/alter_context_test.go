package rpcserver

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// alterContextRequest builds an alter_context offering the given presentation
// contexts.
func alterContextRequest(t *testing.T, callID uint32, contexts ...pdu.ContextElement) []byte {
	t.Helper()

	alter := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: contexts,
	}
	encoded, err := alter.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the alter_context: %v", err)
	}
	encoded[2] = byte(pdu.PacketTypeAlterContext)
	return encoded
}

func TestAlterContextRejectingEveryContextKeepsTheAssociation(t *testing.T) {
	// An alter_context has no nak: one whose contexts are all refused is answered
	// with an alter_context_resp carrying a rejection for each, and what the
	// association already bound stays usable.
	service := &stubService{answer: []byte{0xAA, 0xBB, 0xCC, 0xDD}}
	association := New(service).Open()

	if _, err := association.Handle(bindRequest(t, 1, 4280, testSyntax())); err != nil {
		t.Fatalf("the bind failed: %v", err)
	}

	reply, err := association.Handle(alterContextRequest(t, 2, pdu.ContextElement{
		ContextID:        1,
		AbstractSyntax:   otherSyntax(),
		TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
	}))
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}

	header, err := pdu.PeekHeader(reply)
	if err != nil {
		t.Fatalf("failed to read the reply header: %v", err)
	}
	if header.PacketType != pdu.PacketTypeAlterContextResp {
		t.Fatalf("an alter_context whose only context was rejected was answered with %s, want %s",
			header.PacketType, pdu.PacketTypeAlterContextResp)
	}
	if header.CallID != 2 {
		t.Errorf("alter_context_resp call_id = %d, want 2", header.CallID)
	}

	relabelled := append([]byte(nil), reply...)
	relabelled[2] = byte(pdu.PacketTypeBindAck)
	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(relabelled); err != nil {
		t.Fatalf("the alter_context_resp would not decode: %v", err)
	}
	if len(ack.Results) != 1 {
		t.Fatalf("alter_context_resp carries %d results, want 1", len(ack.Results))
	}
	if ack.Results[0].Result != contextResultProviderRejection ||
		ack.Results[0].Reason != contextReasonAbstractSyntaxNotSupported {
		t.Errorf("result = %d reason %d, want provider rejection, abstract syntax not supported",
			ack.Results[0].Result, ack.Results[0].Reason)
	}

	// The context the bind negotiated survives the refused alter_context.
	if association.Bound() != 1 {
		t.Errorf("after the refused alter_context the association reports %d bound contexts, want 1",
			association.Bound())
	}
	call, err := association.Handle(callRequest(t, 3, 0, nil))
	if err != nil {
		t.Fatalf("the request after the alter_context failed: %v", err)
	}
	callHeader, err := pdu.PeekHeader(call)
	if err != nil {
		t.Fatalf("failed to read the response header: %v", err)
	}
	if callHeader.PacketType != pdu.PacketTypeResponse {
		t.Errorf("a request on the bound context after a refused alter_context was answered with %s, want %s",
			callHeader.PacketType, pdu.PacketTypeResponse)
	}
}

func TestMalformedAlterContextIsFaulted(t *testing.T) {
	// A malformed alter_context is a protocol error on a live association, and
	// the answer is a fault: bind_nak only answers a bind.
	association := New(&stubService{}).Open()

	encoded := alterContextRequest(t, 4, pdu.ContextElement{
		ContextID:        0,
		AbstractSyntax:   testSyntax(),
		TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
	})
	// Cut the PDU off inside its context list, keeping frag_length consistent.
	truncated := encoded[:pdu.HeaderSize+10]
	truncated[8] = byte(len(truncated))
	truncated[9] = 0

	reply, err := association.Handle(truncated)
	if err != nil {
		t.Fatalf("Handle returned an error: %v", err)
	}
	header, err := pdu.PeekHeader(reply)
	if err != nil {
		t.Fatalf("failed to read the reply header: %v", err)
	}
	if header.PacketType != pdu.PacketTypeFault {
		t.Fatalf("a malformed alter_context was answered with %s, want %s",
			header.PacketType, pdu.PacketTypeFault)
	}
	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil {
		t.Fatalf("the fault would not decode: %v", err)
	}
	if fault.Status != pdu.NCASProtoError {
		t.Errorf("fault status = %s, want %s", pdu.FaultStatus(fault.Status), pdu.FaultStatus(pdu.NCASProtoError))
	}
}
