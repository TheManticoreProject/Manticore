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
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// sentRequestStub drives LsarQueryInformationPolicy over ft and returns the marshalled
// request stub. The call's result is ignored: the request is sent before the response is
// read, and the [out] union response is not the subject of this test.
func sentRequestStub(t *testing.T, c *client.Client, ft *fakeTransport) []byte {
	t.Helper()
	var handle mslsad.LSAPR_HANDLE // zero (20-octet) handle
	// A null PolicyInformation referent + zero status, valid under both syntaxes.
	ft.queue(responsePDU(t, 2, make([]byte, 12)))
	_, _ = functions.LsarQueryInformationPolicy(c, handle, mslsad.PolicyLsaServerRoleInformation)
	if len(ft.sent) < 2 {
		t.Fatalf("sent %d PDUs, want >= 2 (bind + request)", len(ft.sent))
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != lsarpc.OpnumLsarQueryInformationPolicy {
		t.Fatalf("opnum = %d, want %d", req.Opnum, lsarpc.OpnumLsarQueryInformationPolicy)
	}
	return req.Stub
}

// TestLsarQueryInformationPolicy_EnumWidthBySyntax pins the InformationClass enum width
// to the transfer syntax: 2 octets under NDR20 (unchanged), 4 octets under NDR64. The
// NDR64 stub matches the request captured from a live Windows Server 2016 exchange,
// which the server accepted (a 2-octet enum was rejected with nca_s_fault_ndr).
func TestLsarQueryInformationPolicy_EnumWidthBySyntax(t *testing.T) {
	// NDR20: 20-octet handle + 2-octet enum (level 6) = 22 octets.
	ft20 := &fakeTransport{}
	c20 := boundClient(t, ft20)
	got20 := sentRequestStub(t, c20, ft20)
	want20 := append(make([]byte, 20), 0x06, 0x00)
	if !bytes.Equal(got20, want20) {
		t.Errorf("NDR20 request stub:\n got  %x\nwant %x", got20, want20)
	}

	// NDR64: 20-octet handle + 4-octet enum (level 6) = 24 octets.
	ft64 := &fakeTransport{}
	ack := &pdu.BindAck{
		MaxXmitFrag: 4280, MaxRecvFrag: 4280,
		Results: []pdu.PresentationResult{
			{Result: pdu.ResultProviderRejection, Reason: pdu.ReasonProposedTransferSyntaxesNotSupported},
			{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDR64TransferSyntax()},
		},
	}
	ackBytes, err := ack.Marshal()
	if err != nil {
		t.Fatalf("bind_ack marshal: %v", err)
	}
	ft64.queue(ackBytes)
	c64 := client.NewClient(ft64)
	c64.PreferNDR64(true)
	if err := c64.Bind(lsarpc.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if c64.NegotiatedSyntax() != ndr.NDR64 {
		t.Fatalf("negotiated %s, want NDR64", c64.NegotiatedSyntax())
	}
	got64 := sentRequestStub(t, c64, ft64)
	want64 := append(make([]byte, 20), 0x06, 0x00, 0x00, 0x00) // captured from Windows Server 2016
	if !bytes.Equal(got64, want64) {
		t.Errorf("NDR64 request stub:\n got  %x\nwant %x (captured from Windows Server 2016)", got64, want64)
	}
}
