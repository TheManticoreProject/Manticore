package client

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// fakeTransport is an in-memory transport.Transport for driving the client without a
// network. Recv pops queued chunks in order, letting tests split a PDU across reads.
type fakeTransport struct {
	maxXmit, maxRecv uint16
	connectErr       error

	sent      [][]byte
	recvQueue [][]byte
}

func (f *fakeTransport) Connect() error { return f.connectErr }
func (f *fakeTransport) Send(pdu []byte) error {
	f.sent = append(f.sent, append([]byte(nil), pdu...))
	return nil
}
func (f *fakeTransport) Recv() ([]byte, error) {
	if len(f.recvQueue) == 0 {
		return nil, errors.New("recv queue empty")
	}
	chunk := f.recvQueue[0]
	f.recvQueue = f.recvQueue[1:]
	return chunk, nil
}
func (f *fakeTransport) Close() error        { return nil }
func (f *fakeTransport) MaxXmitFrag() uint16 { return f.maxXmit }
func (f *fakeTransport) MaxRecvFrag() uint16 { return f.maxRecv }

// queue appends one or more whole PDUs to the recv queue as a single read chunk.
func (f *fakeTransport) queue(chunks ...[]byte) {
	for _, c := range chunks {
		f.recvQueue = append(f.recvQueue, c)
	}
}

func testSyntax() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x12345678, B: 0x1234, C: 0xabcd, D: 0xef00, E: 0x0123456789ab},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

func mustMarshal(t *testing.T, m interface{ Marshal() ([]byte, error) }) []byte {
	t.Helper()
	b, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return b
}

func bindAckBytes(t *testing.T, xmit, recv uint16) []byte {
	ack := &pdu.BindAck{
		MaxXmitFrag:      xmit,
		MaxRecvFrag:      recv,
		SecondaryAddress: "lsarpc",
		Results:          []pdu.PresentationResult{{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()}},
	}
	return mustMarshal(t, ack)
}

func responseBytes(t *testing.T, callID uint32, flags pdu.PFCFlags, stub []byte) []byte {
	resp := &pdu.Response{Stub: stub}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, flags, callID)
	return mustMarshal(t, resp)
}

func TestBind_Success(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))

	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if !c.bound {
		t.Fatal("client not marked bound after successful Bind")
	}
	if c.sendFragMax != 4280 {
		t.Errorf("sendFragMax = %d, want 4280", c.sendFragMax)
	}

	// Verify the bind PDU we sent.
	if len(ft.sent) != 1 {
		t.Fatalf("sent %d PDUs, want 1", len(ft.sent))
	}
	var bind pdu.Bind
	if _, err := bind.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("sent bind does not parse: %v", err)
	}
	if bind.Header.CallID != 1 {
		t.Errorf("bind call_id = %d, want 1", bind.Header.CallID)
	}
	if len(bind.ContextList) != 1 || !bind.ContextList[0].AbstractSyntax.Equal(testSyntax()) {
		t.Errorf("bind abstract syntax not as expected: %+v", bind.ContextList)
	}
	if !bind.ContextList[0].TransferSyntaxes[0].Equal(syntax.NDRTransferSyntax()) {
		t.Error("bind did not propose NDR transfer syntax")
	}
}

func TestBind_Nak(t *testing.T) {
	nak := &pdu.BindNak{RejectReason: pdu.ReasonProposedTransferSyntaxesNotSupported, Versions: []pdu.ProtocolVersion{{Major: 5, Minor: 0}}}
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(mustMarshal(t, nak))

	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err == nil {
		t.Fatal("Bind() with bind_nak: error = nil, want non-nil")
	}
	if c.bound {
		t.Error("client marked bound after bind_nak")
	}
}

func TestBind_NoContextAccepted(t *testing.T) {
	ack := &pdu.BindAck{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		Results:     []pdu.PresentationResult{{Result: pdu.ResultProviderRejection, Reason: pdu.ReasonAbstractSyntaxNotSupported}},
	}
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(mustMarshal(t, ack))

	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err == nil {
		t.Fatal("Bind() with no accepted context: error = nil, want non-nil")
	}
}

func TestBind_ConnectError(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280, connectErr: errors.New("pipe open failed")}
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err == nil {
		t.Fatal("Bind() with connect error: error = nil, want non-nil")
	}
}

func TestCall_SingleFragment(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	// First call uses call_id 2 (Bind used 1).
	wantStub := []byte{0xca, 0xfe, 0xba, 0xbe}
	ft.queue(responseBytes(t, 2, pdu.PFCFirstFrag|pdu.PFCLastFrag, wantStub))

	got, err := c.Call(0x000c, []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if !bytes.Equal(got, wantStub) {
		t.Errorf("Call() = %x, want %x", got, wantStub)
	}

	// The request PDU is sent after the bind PDU.
	if len(ft.sent) != 2 {
		t.Fatalf("sent %d PDUs, want 2 (bind + request)", len(ft.sent))
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("sent request does not parse: %v", err)
	}
	if req.Opnum != 0x000c {
		t.Errorf("opnum = %d, want 12", req.Opnum)
	}
	if req.Header.CallID != 2 {
		t.Errorf("request call_id = %d, want 2", req.Header.CallID)
	}
	if !req.Header.PacketFlags.Has(pdu.PFCFirstFrag) || !req.Header.PacketFlags.Has(pdu.PFCLastFrag) {
		t.Errorf("single-fragment request flags = %s, want first|last", req.Header.PacketFlags)
	}
}

func TestCall_Fault(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	fault := &pdu.Fault{Status: pdu.NCASOpRngError}
	fault.Header = pdu.NewHeader(pdu.PacketTypeFault, pdu.PFCFirstFrag|pdu.PFCLastFrag, 2)
	ft.queue(mustMarshal(t, fault))

	_, err := c.Call(0x0099, nil)
	if err == nil {
		t.Fatal("Call() returning a fault: error = nil, want non-nil")
	}
	var fe *pdu.Fault
	if !errors.As(err, &fe) {
		t.Fatalf("error is not *pdu.Fault: %v", err)
	}
	if fe.Status != pdu.NCASOpRngError {
		t.Errorf("fault status = %#x, want %#x", fe.Status, pdu.NCASOpRngError)
	}
}

func TestCall_MultiFragmentResponseReassembly(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	part1 := []byte{0x11, 0x22, 0x33}
	part2 := []byte{0x44, 0x55}
	frag1 := responseBytes(t, 2, pdu.PFCFirstFrag, part1)
	frag2 := responseBytes(t, 2, pdu.PFCLastFrag, part2)

	// Concatenate both response fragments and deliver them split at an arbitrary,
	// non-PDU-aligned boundary to exercise the fragment reader's buffering.
	all := append(append([]byte(nil), frag1...), frag2...)
	cut := len(frag1) - 2
	ft.queue(all[:cut], all[cut:])

	got, err := c.Call(0x0001, nil)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	want := append(append([]byte(nil), part1...), part2...)
	if !bytes.Equal(got, want) {
		t.Errorf("reassembled stub = %x, want %x", got, want)
	}
}

func TestCall_MultiFragmentRequest(t *testing.T) {
	// Negotiate a tiny fragment size so the request must be split. Budget per
	// fragment = sendFragMax - requestHeaderOverhead.
	const frag = uint16(requestHeaderOverhead + 4) // 4 stub bytes per fragment
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, frag, frag))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if c.sendFragMax != frag {
		t.Fatalf("sendFragMax = %d, want %d", c.sendFragMax, frag)
	}

	stub := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // 10 bytes -> 4 + 4 + 2
	ft.queue(responseBytes(t, 2, pdu.PFCFirstFrag|pdu.PFCLastFrag, []byte{0xff}))

	if _, err := c.Call(0x0005, stub); err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	// Request fragments are everything sent after the bind PDU.
	reqFrags := ft.sent[1:]
	if len(reqFrags) != 3 {
		t.Fatalf("request split into %d fragments, want 3", len(reqFrags))
	}

	var reassembled []byte
	for i, raw := range reqFrags {
		if len(raw) > int(frag) {
			t.Errorf("fragment %d is %d bytes, exceeds negotiated %d", i, len(raw), frag)
		}
		var req pdu.Request
		if _, err := req.Unmarshal(raw); err != nil {
			t.Fatalf("fragment %d does not parse: %v", i, err)
		}
		first := req.Header.PacketFlags.Has(pdu.PFCFirstFrag)
		last := req.Header.PacketFlags.Has(pdu.PFCLastFrag)
		switch i {
		case 0:
			if !first || last {
				t.Errorf("fragment 0 flags = %s, want first only", req.Header.PacketFlags)
			}
		case len(reqFrags) - 1:
			if first || !last {
				t.Errorf("last fragment flags = %s, want last only", req.Header.PacketFlags)
			}
		default:
			if first || last {
				t.Errorf("middle fragment %d flags = %s, want none", i, req.Header.PacketFlags)
			}
		}
		reassembled = append(reassembled, req.Stub...)
	}
	if !bytes.Equal(reassembled, stub) {
		t.Errorf("reassembled request stub = %x, want %x", reassembled, stub)
	}
}

func TestCall_NotBound(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	c := NewClient(ft)
	if _, err := c.Call(1, nil); err == nil {
		t.Fatal("Call() before Bind: error = nil, want non-nil")
	}
}

func TestCall_CallIDMismatch(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	// Response carries the wrong call_id (3 instead of 2).
	ft.queue(responseBytes(t, 3, pdu.PFCFirstFrag|pdu.PFCLastFrag, []byte{0x01}))
	if _, err := c.Call(1, nil); err == nil {
		t.Fatal("Call() with mismatched response call_id: error = nil, want non-nil")
	}
}
