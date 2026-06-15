package functions_test

import (
	"bytes"
	"testing"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// bindAckNDR64 is a bind_ack that mirrors a real Windows Server 2016 response to the
// two-context NDR64 bind: context 0 (NDR 2.0) rejected, context 1 (NDR64) accepted.
func bindAckNDR64(t *testing.T) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		Results: []pdu.PresentationResult{
			{Result: pdu.ResultProviderRejection, Reason: pdu.ReasonProposedTransferSyntaxesNotSupported},
			{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDR64TransferSyntax()},
		},
	}
	b, err := ack.Marshal()
	if err != nil {
		t.Fatalf("bind_ack marshal: %v", err)
	}
	return b
}

// TestLsarOpenPolicy_NDR64RequestMatchesCapture pins the NDR64 LsarOpenPolicy request
// stub to bytes captured from a live Windows Server 2016 exchange (192.168.1.31), which
// the server accepted. This anchors the NDR64 encoding to real wire bytes rather than a
// hand-computed expectation: a NULL [unique] SystemName referent (8 octets), an
// LSAPR_OBJECT_ATTRIBUTES with all-NULL pointer members (40 octets: Length + four
// 8-octet NULL referents + Attributes, 8-aligned), and the DesiredAccess 0x02000000.
func TestLsarOpenPolicy_NDR64RequestMatchesCapture(t *testing.T) {
	ft := &fakeTransport{}
	ft.queue(bindAckNDR64(t))
	c := client.NewClient(ft)
	c.PreferNDR64(true)
	if err := c.Bind(lsarpc.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if c.NegotiatedSyntax() != ndr.NDR64 {
		t.Fatalf("negotiated %s, want NDR64", c.NegotiatedSyntax())
	}

	// Canned success response: 20-byte handle + STATUS_SUCCESS (call_id 2).
	ft.queue(responsePDU(t, 2, append(make([]byte, 20), 0x00, 0x00, 0x00, 0x00)))

	if _, err := functions.LsarOpenPolicy(c, 0x02000000); err != nil {
		t.Fatalf("LsarOpenPolicy() under NDR64 error = %v", err)
	}

	if len(ft.sent) != 2 {
		t.Fatalf("sent %d PDUs, want 2 (bind + request)", len(ft.sent))
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != lsarpc.OpnumLsarOpenPolicy {
		t.Errorf("opnum = %d, want %d", req.Opnum, lsarpc.OpnumLsarOpenPolicy)
	}

	// Captured NDR64 stub: 56 zero octets followed by DesiredAccess 0x02000000 (LE).
	want := append(make([]byte, 56), 0x00, 0x00, 0x00, 0x02)
	if !bytes.Equal(req.Stub, want) {
		t.Errorf("NDR64 LsarOpenPolicy request stub:\n got  %x\nwant %x (captured from Windows Server 2016)", req.Stub, want)
	}
}
