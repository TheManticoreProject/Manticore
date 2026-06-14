package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// bindAckResults builds a bind_ack with the given per-context results.
func bindAckResults(t *testing.T, xmit, recv uint16, results ...pdu.PresentationResult) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag:      xmit,
		MaxRecvFrag:      recv,
		SecondaryAddress: "lsarpc",
		Results:          results,
	}
	return mustMarshal(t, ack)
}

func TestBind_DefaultProposesOnlyNDR20(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))

	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if c.NegotiatedSyntax() != ndr.NDR20 {
		t.Errorf("negotiated syntax = %v, want NDR20", c.NegotiatedSyntax())
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if len(bind.ContextList) != 1 {
		t.Fatalf("default bind proposed %d contexts, want 1", len(bind.ContextList))
	}
}

func TestBind_NegotiatesNDR64(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	// Server accepts both contexts; the client must prefer NDR64 (context 1).
	ft.queue(bindAckResults(t, 4280, 4280,
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()},
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDR64TransferSyntax()},
	))

	c := NewClient(ft)
	c.PreferNDR64(true)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if c.NegotiatedSyntax() != ndr.NDR64 {
		t.Errorf("negotiated syntax = %v, want NDR64", c.NegotiatedSyntax())
	}
	if c.contextID != 1 {
		t.Errorf("context id = %d, want 1 (the NDR64 context)", c.contextID)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if len(bind.ContextList) != 2 {
		t.Fatalf("NDR64 bind proposed %d contexts, want 2", len(bind.ContextList))
	}
	if !bind.ContextList[0].TransferSyntaxes[0].Equal(syntax.NDRTransferSyntax()) {
		t.Error("context 0 should propose NDR 2.0")
	}
	if !bind.ContextList[1].TransferSyntaxes[0].Equal(syntax.NDR64TransferSyntax()) {
		t.Error("context 1 should propose NDR64")
	}
}

func TestBind_NDR64FallsBackToNDR20(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	// Server accepts NDR20 (context 0) and rejects NDR64 (context 1).
	ft.queue(bindAckResults(t, 4280, 4280,
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()},
		pdu.PresentationResult{Result: pdu.ResultProviderRejection, Reason: pdu.ReasonProposedTransferSyntaxesNotSupported},
	))

	c := NewClient(ft)
	c.PreferNDR64(true)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if c.NegotiatedSyntax() != ndr.NDR20 {
		t.Errorf("negotiated syntax = %v, want NDR20 (fallback)", c.NegotiatedSyntax())
	}
	if c.contextID != 0 {
		t.Errorf("context id = %d, want 0 (the NDR 2.0 context)", c.contextID)
	}
}

// confReq is an [in] parameter with a conformant array, whose NDR64 encoding (8-octet
// maximum_count) differs from its NDR20 encoding (4-octet), so the request stub reveals
// which transfer syntax the client used.
type confReq struct {
	Data []uint32 `ndr:"conformant"`
}

func (*confReq) Opnum() uint16 { return 9 }

func TestInvoke_UsesNegotiatedNDR64(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckResults(t, 4280, 4280,
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()},
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDR64TransferSyntax()},
	))
	c := NewClient(ft)
	c.PreferNDR64(true)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	ft.queue(responseBytes(t, 2, pdu.PFCFirstFrag|pdu.PFCLastFrag, nil))

	req := &confReq{Data: []uint32{0xAAAA, 0xBBBB}}
	if err := c.Invoke(req, nil); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	var sent pdu.Request
	if _, err := sent.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("sent request does not parse: %v", err)
	}
	if sent.ContextID != 1 {
		t.Errorf("request context id = %d, want 1 (NDR64)", sent.ContextID)
	}

	wantNDR64, err := ndr.MarshalAs(req, ndr.NDR64)
	if err != nil {
		t.Fatalf("MarshalAs NDR64: %v", err)
	}
	if !bytes.Equal(sent.Stub, wantNDR64) {
		t.Errorf("request stub = %x, want NDR64 encoding %x", sent.Stub, wantNDR64)
	}
	// And it must not be the NDR20 encoding (the 8-octet count makes it longer).
	ndr20, _ := ndr.Marshal(req)
	if bytes.Equal(sent.Stub, ndr20) {
		t.Error("request stub is the NDR20 encoding; negotiated NDR64 was not used")
	}
}
