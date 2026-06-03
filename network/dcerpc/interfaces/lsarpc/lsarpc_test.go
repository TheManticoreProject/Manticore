package lsarpc

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
)

// fakeTransport is an in-memory transport.Transport for driving the client without a
// network.
type fakeTransport struct {
	sent      [][]byte
	recvQueue [][]byte
}

func (f *fakeTransport) Connect() error { return nil }
func (f *fakeTransport) Send(pdu []byte) error {
	f.sent = append(f.sent, append([]byte(nil), pdu...))
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

func (f *fakeTransport) queue(b []byte) { f.recvQueue = append(f.recvQueue, b) }

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

func responsePDU(t *testing.T, callID uint32, stub []byte) []byte {
	t.Helper()
	resp := &pdu.Response{Stub: stub}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID)
	b, err := resp.Marshal()
	if err != nil {
		t.Fatalf("response marshal: %v", err)
	}
	return b
}

// boundClient returns a client already bound to lsarpc over ft.
func boundClient(t *testing.T, ft *fakeTransport) *client.Client {
	t.Helper()
	ft.queue(bindAck(t))
	c := client.NewClient(ft)
	if err := c.Bind(SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return c
}

func TestOpenPolicy2_RequestMarshalling(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	// Canned success response: 20-byte handle + STATUS_SUCCESS. First call is call_id 2.
	wantHandle := PolicyHandle{0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	respStub := append(append([]byte(nil), wantHandle[:]...), 0x00, 0x00, 0x00, 0x00)
	ft.queue(responsePDU(t, 2, respStub))

	h, err := OpenPolicy2(c, MaximumAllowed)
	if err != nil {
		t.Fatalf("OpenPolicy2() error = %v", err)
	}
	if h != wantHandle {
		t.Errorf("handle = %x, want %x", h, wantHandle)
	}

	// Verify the request stub: NULL SystemName ptr (4) + zero ObjectAttributes (24) +
	// DesiredAccess (4, little-endian).
	if len(ft.sent) != 2 {
		t.Fatalf("sent %d PDUs, want 2 (bind + request)", len(ft.sent))
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != OpnumOpenPolicy2 {
		t.Errorf("opnum = %d, want %d", req.Opnum, OpnumOpenPolicy2)
	}
	wantStub := make([]byte, 0, 32)
	wantStub = append(wantStub, 0, 0, 0, 0)
	wantStub = append(wantStub, make([]byte, 24)...)
	wantStub = append(wantStub, 0x00, 0x00, 0x00, 0x02) // MAXIMUM_ALLOWED 0x02000000 LE
	if !bytes.Equal(req.Stub, wantStub) {
		t.Errorf("request stub:\n got %x\nwant %x", req.Stub, wantStub)
	}
}

func TestOpenPolicy2_AccessDenied(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	respStub := append(make([]byte, 20), 0x22, 0x00, 0x00, 0xc0) // STATUS_ACCESS_DENIED
	ft.queue(responsePDU(t, 2, respStub))

	_, err := OpenPolicy2(c, PolicyViewLocalInformation)
	if err == nil {
		t.Fatal("OpenPolicy2() with access denied: error = nil, want non-nil")
	}
}

func TestClose_RoundTrip(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	// LsarClose returns a zeroed handle + STATUS_SUCCESS.
	ft.queue(responsePDU(t, 2, make([]byte, 24)))

	in := PolicyHandle{0x00, 0x00, 0x00, 0x00, 0xaa, 0xbb}
	out, err := Close(c, in)
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !out.IsZero() {
		t.Errorf("handle after Close = %x, want zeroed", out)
	}

	// The request stub is exactly the 20-byte handle.
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != OpnumClose {
		t.Errorf("opnum = %d, want %d", req.Opnum, OpnumClose)
	}
	if !bytes.Equal(req.Stub, in[:]) {
		t.Errorf("close request stub = %x, want %x", req.Stub, in[:])
	}
}

func TestParseHandleResponse_TooShort(t *testing.T) {
	if _, _, err := parseHandleResponse(make([]byte, 23)); err == nil {
		t.Fatal("parseHandleResponse of short buffer: error = nil, want non-nil")
	}
}

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusAccessDenied); got != "STATUS_ACCESS_DENIED" {
		t.Errorf("StatusString(0xC0000022) = %q", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q", got)
	}
}
