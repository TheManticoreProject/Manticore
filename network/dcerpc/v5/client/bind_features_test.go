package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// TestBindTimeFeatureNegotiationWireForm checks the proposed transfer syntax is
// 6CB71C2C-9812-4540 followed by the eight-octet bitmask, version 1.0
// ([MS-RPCE] 3.3.1.5.3 and 2.2.2.14), as Windows parses it.
func TestBindTimeFeatureNegotiationWireForm(t *testing.T) {
	encoded, err := syntax.BindTimeFeatureNegotiation(0x03).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// GUID wire form: Data1-3 little-endian, Data4 as bytes; then version 1.0.
	want := []byte{0x2c, 0x1c, 0xb7, 0x6c, 0x12, 0x98, 0x40, 0x45, 0x03, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0}
	if !bytes.Equal(encoded, want) {
		t.Fatalf("encoded % x, want % x", encoded, want)
	}
	if !syntax.BindTimeFeatureNegotiation(0).IsBindTimeFeatureNegotiation() || syntax.NDRTransferSyntax().IsBindTimeFeatureNegotiation() {
		t.Error("IsBindTimeFeatureNegotiation misclassifies a syntax")
	}
}

// TestBindProposesAndRecordsFeatures binds with the negotiation proposed: the
// bind carries one more context with the bitmask, and the server's
// negotiate_ack reason is recorded while the NDR context is still selected.
func TestBindProposesAndRecordsFeatures(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckResults(t, 4280, 4280,
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()},
		pdu.PresentationResult{Result: pdu.ResultNegotiateAck, Reason: 0x0002},
	))
	c := NewClient(ft)
	c.ProposeBindTimeFeatures(syntax.KeepConnectionOnOrphanSupported)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var sent pdu.Bind
	if _, err := sent.Unmarshal(ft.sent[0]); err != nil {
		t.Fatal(err)
	}
	if len(sent.ContextList) != 2 {
		t.Fatalf("the bind carries %d contexts, want 2", len(sent.ContextList))
	}
	feature := sent.ContextList[1]
	if feature.ContextID != 1 || !feature.AbstractSyntax.Equal(testSyntax()) || len(feature.TransferSyntaxes) != 1 ||
		!feature.TransferSyntaxes[0].Equal(syntax.BindTimeFeatureNegotiation(0x02)) {
		t.Errorf("the negotiation context is %+v", feature)
	}

	bits, acked := c.BindTimeFeatures()
	if !acked || bits != 0x0002 {
		t.Errorf("BindTimeFeatures() = %#x, %v; want 0x2, true", bits, acked)
	}
	if c.contextID != 0 {
		t.Errorf("selected context %d, want the NDR context 0", c.contextID)
	}
}

// TestBindWithoutProposalOffersNoNegotiation keeps the default bind unchanged.
func TestBindWithoutProposalOffersNoNegotiation(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatal(err)
	}
	var sent pdu.Bind
	if _, err := sent.Unmarshal(ft.sent[0]); err != nil {
		t.Fatal(err)
	}
	for _, ctx := range sent.ContextList {
		if ctx.TransferSyntaxes[0].IsBindTimeFeatureNegotiation() {
			t.Error("a default bind offers feature negotiation")
		}
	}
	if _, acked := c.BindTimeFeatures(); acked {
		t.Error("an un-negotiated bind reports an acknowledgement")
	}
}

// TestBindSurvivesARejectedFeatureContext checks a server that rejects the
// negotiation context (it does not know it) still binds on the NDR64 context,
// with the negotiation context numbered after NDR (0) and NDR64 (1).
func TestBindSurvivesARejectedFeatureContext(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckResults(t, 4280, 4280,
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()},
		pdu.PresentationResult{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDR64TransferSyntax()},
		pdu.PresentationResult{Result: pdu.ResultProviderRejection, Reason: 2},
	))
	c := NewClient(ft)
	c.PreferNDR64(true)
	c.ProposeBindTimeFeatures(0x03)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	var sent pdu.Bind
	if _, err := sent.Unmarshal(ft.sent[0]); err != nil {
		t.Fatal(err)
	}
	if len(sent.ContextList) != 3 || sent.ContextList[2].ContextID != 2 {
		t.Fatalf("the bind's contexts are %+v", sent.ContextList)
	}
	if _, acked := c.BindTimeFeatures(); acked {
		t.Error("a rejected negotiation is reported as acknowledged")
	}
	if c.contextID != 1 {
		t.Errorf("selected context %d, want NDR64 context 1", c.contextID)
	}
}
