package ms_icpr

import (
	"bytes"
	"errors"
	"testing"

	ICertPassage "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/91ae6020-9e3c-11cf-8d7c-00aa00c091be/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	mswcce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wcce"
)

// fakeTransport is an in-memory transport.Transport for driving the DCE/RPC client
// without a network.
type fakeTransport struct {
	sent      [][]byte
	recvQueue [][]byte
}

func (f *fakeTransport) Connect() error { return nil }
func (f *fakeTransport) Send(p []byte) error {
	f.sent = append(f.sent, append([]byte(nil), p...))
	return nil
}
func (f *fakeTransport) Recv() ([]byte, error) {
	if len(f.recvQueue) == 0 {
		return nil, errors.New("recv queue empty")
	}
	c := f.recvQueue[0]
	f.recvQueue = f.recvQueue[1:]
	return c, nil
}
func (f *fakeTransport) Close() error        { return nil }
func (f *fakeTransport) MaxXmitFrag() uint16 { return 4280 }
func (f *fakeTransport) MaxRecvFrag() uint16 { return 4280 }
func (f *fakeTransport) queue(b []byte)      { f.recvQueue = append(f.recvQueue, b) }

func bindAck(t *testing.T) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		Results:     []pdu.PresentationResult{{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()}},
	}
	b, err := ack.Marshal()
	if err != nil {
		t.Fatalf("bind_ack marshal: %v", err)
	}
	return b
}

func responsePDU(t *testing.T, callID uint32, stubBytes []byte) []byte {
	t.Helper()
	resp := &pdu.Response{Stub: stubBytes}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID)
	b, err := resp.Marshal()
	if err != nil {
		t.Fatalf("response marshal: %v", err)
	}
	return b
}

func stub(t *testing.T, v any) []byte {
	t.Helper()
	b, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("marshal response stub: %v", err)
	}
	return b
}

// boundClient returns a DCE/RPC client already bound to ICertPassage over ft.
func boundClient(t *testing.T, ft *fakeTransport) *dcerpcclient.Client {
	t.Helper()
	ft.queue(bindAck(t))
	c := dcerpcclient.NewClient(ft)
	if err := c.Bind(ICertPassage.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return c
}

// certReqResp mirrors the CertServerRequest [out] response structure, for marshalling a
// canned reply. Field order must match the generated certServerRequestResponse.
type certReqResp struct {
	PdwRequestId           ndr.DWORD
	PdwDisposition         ndr.DWORD
	PctbCert               mswcce.CERTTRANSBLOB
	PctbEncodedCert        mswcce.CERTTRANSBLOB
	PctbDispositionMessage mswcce.CERTTRANSBLOB
	Status                 ndr.DWORD `ndr:"retval"`
}

func blob(b []byte) mswcce.CERTTRANSBLOB {
	return mswcce.CERTTRANSBLOB{Cb: ndr.DWORD(len(b)), Pb: b}
}

// TestRequestIssued submits a request and projects a canned "issued" reply: the DER
// certificate, request id, disposition and message come back, and opnum 0 went on the wire
// carrying the CSR.
func TestRequestIssued(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	wantCert := []byte{0x30, 0x82, 0x01, 0x0a, 0xDE, 0xAD, 0xBE, 0xEF} // a stand-in DER cert
	csr := []byte{0x30, 0x81, 0xEE, 0x01, 0x02, 0x03}
	msg := encodeUTF16("Issued")
	ft.queue(responsePDU(t, 2, stub(t, &certReqResp{
		PdwRequestId:           42,
		PdwDisposition:         DispIssued,
		PctbEncodedCert:        blob(wantCert),
		PctbDispositionMessage: blob(msg),
		Status:                 0,
	})))

	res, err := request(c, "CORP-CA", csr, "CertificateTemplate:User")
	if err != nil {
		t.Fatalf("request() error = %v", err)
	}
	if !res.Issued() {
		t.Fatalf("Disposition = %d, want DispIssued (%d)", res.Disposition, DispIssued)
	}
	if res.RequestID != 42 {
		t.Errorf("RequestID = %d, want 42", res.RequestID)
	}
	if !bytes.Equal(res.Certificate, wantCert) {
		t.Errorf("Certificate = %x, want %x", res.Certificate, wantCert)
	}
	if res.DispositionMessage != "Issued" {
		t.Errorf("DispositionMessage = %q, want %q", res.DispositionMessage, "Issued")
	}

	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request PDU does not parse: %v", err)
	}
	if req.Opnum != ICertPassage.OpnumCertServerRequest {
		t.Errorf("opnum = %d, want %d", req.Opnum, ICertPassage.OpnumCertServerRequest)
	}
	if !bytes.Contains(req.Stub, csr) {
		t.Errorf("request stub does not carry the CSR")
	}
}

// TestRequestDenied checks a CA denial (S_OK status, DispDenied) comes back as a nil error
// with the denial disposition and reason, not as a hard error.
func TestRequestDenied(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	ft.queue(responsePDU(t, 2, stub(t, &certReqResp{
		PdwDisposition:         DispDenied,
		PctbDispositionMessage: blob(encodeUTF16("Denied by Policy Module")),
		Status:                 0,
	})))

	res, err := request(c, "CORP-CA", []byte{0x30, 0x01}, "")
	if err != nil {
		t.Fatalf("request() error = %v, want nil for a policy denial", err)
	}
	if res.Issued() {
		t.Fatal("Issued() = true for a denied request")
	}
	if res.Disposition != DispDenied {
		t.Errorf("Disposition = %d, want DispDenied (%d)", res.Disposition, DispDenied)
	}
	if res.DispositionMessage != "Denied by Policy Module" {
		t.Errorf("DispositionMessage = %q", res.DispositionMessage)
	}
}
