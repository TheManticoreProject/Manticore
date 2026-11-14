package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// TestBindProposesHeaderSignWhenEnabled: EnableHeaderSigning makes the bind set
// PFC_SUPPORT_HEADER_SIGN in its pfc_flags, and a server that echoes it in the
// bind_ack is recorded as supporting header signing ([MS-RPCE] 2.2.2.3).
func TestBindProposesHeaderSignWhenEnabled(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckWithFlags(t, pdu.PFCSupportHeaderSign))

	c := NewClient(ft)
	c.EnableHeaderSigning()
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if !bind.Header.PacketFlags.Has(pdu.PFCSupportHeaderSign) {
		t.Errorf("sent bind pfc_flags = 0x%02x, want PFC_SUPPORT_HEADER_SIGN set", uint8(bind.Header.PacketFlags))
	}
	if !c.HeaderSigningSupported() {
		t.Error("HeaderSigningSupported() = false after the server echoed PFC_SUPPORT_HEADER_SIGN")
	}
}

// TestBindRecordsHeaderSignNotSupported: a server that does not echo
// PFC_SUPPORT_HEADER_SIGN leaves the capability unsupported even though the client
// proposed it.
func TestBindRecordsHeaderSignNotSupported(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckWithFlags(t, 0))

	c := NewClient(ft)
	c.EnableHeaderSigning()
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if !bind.Header.PacketFlags.Has(pdu.PFCSupportHeaderSign) {
		t.Errorf("sent bind pfc_flags = 0x%02x, want PFC_SUPPORT_HEADER_SIGN set", uint8(bind.Header.PacketFlags))
	}
	if c.HeaderSigningSupported() {
		t.Error("HeaderSigningSupported() = true though the server did not echo PFC_SUPPORT_HEADER_SIGN")
	}
}

// TestBindDoesNotProposeHeaderSignByDefault: without the opt-in the bind does not
// set PFC_SUPPORT_HEADER_SIGN, and a server that sets the bit unsolicited is not
// read as a negotiated capability.
func TestBindDoesNotProposeHeaderSignByDefault(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckWithFlags(t, pdu.PFCSupportHeaderSign)) // server sets it unsolicited

	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if bind.Header.PacketFlags.Has(pdu.PFCSupportHeaderSign) {
		t.Errorf("default bind set PFC_SUPPORT_HEADER_SIGN unexpectedly (pfc_flags = 0x%02x)", uint8(bind.Header.PacketFlags))
	}
	if c.HeaderSigningSupported() {
		t.Error("HeaderSigningSupported() = true without the client proposing PFC_SUPPORT_HEADER_SIGN")
	}
}
