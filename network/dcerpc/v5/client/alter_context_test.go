package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// alterContextRespBytes builds an alter_context_resp carrying one result.
func alterContextRespBytes(t *testing.T, callID uint32, result pdu.PresentationResult) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		Header:      pdu.NewHeader(pdu.PacketTypeBindAck, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		Results:     []pdu.PresentationResult{result},
	}
	b := mustMarshal(t, ack)
	b[2] = byte(pdu.PacketTypeAlterContextResp)
	return b
}

func continuationContexts() []pdu.ContextElement {
	return []pdu.ContextElement{{
		ContextID:        0,
		AbstractSyntax:   testSyntax(),
		TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
	}}
}

// TestSendAuthContinuation_ProviderRejection checks that an alter_context_resp
// refusing the bound presentation context fails the continuation: a server has no
// nak for an alter_context and reports the refusal through the per-context result.
func TestSendAuthContinuation_ProviderRejection(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(alterContextRespBytes(t, 1, pdu.PresentationResult{
		Result: pdu.ResultProviderRejection,
		Reason: pdu.ReasonAbstractSyntaxNotSupported,
	}))
	c := NewClient(ft)
	c.callID = 1

	if err := c.sendAuthContinuation(continuationContexts(), []byte{0x01}); err == nil {
		t.Fatal("sendAuthContinuation() with a provider_rejection result: error = nil, want non-nil")
	}
}

// TestSendAuthContinuation_Acceptance checks that an alter_context_resp accepting
// the bound presentation context completes the continuation.
func TestSendAuthContinuation_Acceptance(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(alterContextRespBytes(t, 1, pdu.PresentationResult{
		Result:         pdu.ResultAcceptance,
		TransferSyntax: syntax.NDRTransferSyntax(),
	}))
	c := NewClient(ft)
	c.callID = 1

	if err := c.sendAuthContinuation(continuationContexts(), []byte{0x01}); err != nil {
		t.Fatalf("sendAuthContinuation() with an acceptance result: error = %v", err)
	}
}
