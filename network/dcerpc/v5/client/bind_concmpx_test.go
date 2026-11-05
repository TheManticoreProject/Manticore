package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// bindAckWithFlags builds a bind_ack that accepts the NDR 2.0 context and carries
// the given pfc_flags in its header, so a test can model a server that does (or
// does not) echo PFC_CONC_MPX.
func bindAckWithFlags(t *testing.T, flags pdu.PFCFlags) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag:      4280,
		MaxRecvFrag:      4280,
		SecondaryAddress: "lsarpc",
		Results: []pdu.PresentationResult{
			{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()},
		},
	}
	ack.Header = pdu.NewHeader(pdu.PacketTypeBindAck, pdu.PFCFirstFrag|pdu.PFCLastFrag|flags, 1)
	return mustMarshal(t, ack)
}

// TestBindProposesConcMpxWhenEnabled: EnableConcurrentMultiplexing makes the bind
// set PFC_CONC_MPX in its pfc_flags, and a server that echoes it is recorded as
// supporting concurrent multiplexing ([C706] 12.6.3.1).
func TestBindProposesConcMpxWhenEnabled(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckWithFlags(t, pdu.PFCConcMpx))

	c := NewClient(ft)
	c.EnableConcurrentMultiplexing()
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if !bind.Header.PacketFlags.Has(pdu.PFCConcMpx) {
		t.Errorf("sent bind pfc_flags = 0x%02x, want PFC_CONC_MPX set", uint8(bind.Header.PacketFlags))
	}
	if !c.ConcurrentMultiplexingSupported() {
		t.Error("ConcurrentMultiplexingSupported() = false after the server echoed PFC_CONC_MPX")
	}
}

// TestBindRecordsConcMpxNotSupported: a server that does not echo PFC_CONC_MPX
// leaves the capability unsupported even though the client proposed it.
func TestBindRecordsConcMpxNotSupported(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckWithFlags(t, 0))

	c := NewClient(ft)
	c.EnableConcurrentMultiplexing()
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if !bind.Header.PacketFlags.Has(pdu.PFCConcMpx) {
		t.Errorf("sent bind pfc_flags = 0x%02x, want PFC_CONC_MPX set", uint8(bind.Header.PacketFlags))
	}
	if c.ConcurrentMultiplexingSupported() {
		t.Error("ConcurrentMultiplexingSupported() = true though the server did not echo PFC_CONC_MPX")
	}
}

// TestBindDoesNotProposeConcMpxByDefault: without the opt-in the bind does not set
// PFC_CONC_MPX, and a server that sets the bit unsolicited is not read as a
// negotiated capability.
func TestBindDoesNotProposeConcMpxByDefault(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckWithFlags(t, pdu.PFCConcMpx)) // server sets it unsolicited

	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if bind.Header.PacketFlags.Has(pdu.PFCConcMpx) {
		t.Errorf("default bind set PFC_CONC_MPX unexpectedly (pfc_flags = 0x%02x)", uint8(bind.Header.PacketFlags))
	}
	if c.ConcurrentMultiplexingSupported() {
		t.Error("ConcurrentMultiplexingSupported() = true without the client proposing PFC_CONC_MPX")
	}
}
