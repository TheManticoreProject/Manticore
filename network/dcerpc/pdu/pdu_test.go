package pdu

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

func TestHeader_GoldenBytes(t *testing.T) {
	h := Header{
		RPCVersion:         5,
		RPCVersionMinor:    0,
		PacketType:         PacketTypeBind,
		PacketFlags:        PFCFirstFrag | PFCLastFrag,
		DataRepresentation: DataRepresentationLittleEndian,
		FragLength:         72,
		AuthLength:         0,
		CallID:             1,
	}
	want := []byte{
		0x05, 0x00, 0x0b, 0x03,
		0x10, 0x00, 0x00, 0x00,
		0x48, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
	}
	got, err := h.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("header bytes:\n got %x\nwant %x", got, want)
	}
}

func TestHeader_RejectsBigEndianDREP(t *testing.T) {
	data := []byte{
		0x05, 0x00, 0x0b, 0x03,
		0x00, 0x00, 0x00, 0x00, // big-endian DREP (high nibble 0)
		0x10, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
	}
	var h Header
	if _, err := h.Unmarshal(data); err == nil {
		t.Fatal("Unmarshal of big-endian DREP: error = nil, want non-nil")
	}
}

func TestHeader_RoundTrip(t *testing.T) {
	h := NewHeader(PacketTypeRequest, PFCFirstFrag, 42)
	h.FragLength = 100
	b, _ := h.Marshal()
	var got Header
	n, err := got.Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if n != HeaderSize {
		t.Errorf("consumed %d, want %d", n, HeaderSize)
	}
	if got != h {
		t.Errorf("round trip: got %+v, want %+v", got, h)
	}
}

func TestBind_RoundTripAndFraming(t *testing.T) {
	b := &Bind{
		MaxXmitFrag:  4280,
		MaxRecvFrag:  4280,
		AssocGroupID: 0,
		ContextList: []ContextElement{
			{
				ContextID:      0,
				AbstractSyntax: syntax.SyntaxID{UUID: guid.GUID{A: 0x12345678, B: 0x1234, C: 0xabcd, D: 0xef00, E: 0x0123456789ab}, MajorVersion: 1, MinorVersion: 0},
				TransferSyntaxes: []syntax.SyntaxID{
					syntax.NDRTransferSyntax(),
				},
			},
		},
	}
	b.Header.CallID = 1

	raw, err := b.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	// header(16) + xmit/recv/assoc(8) + ctxlist hdr(4) + p_cont_elem(4 + 20 + 20) = 72
	if len(raw) != 72 {
		t.Errorf("marshalled length = %d, want 72", len(raw))
	}
	if int(b.Header.FragLength) != len(raw) {
		t.Errorf("frag_length = %d, want %d", b.Header.FragLength, len(raw))
	}
	if PacketType(raw[2]) != PacketTypeBind {
		t.Errorf("ptype byte = %d, want %d (bind)", raw[2], PacketTypeBind)
	}

	var got Bind
	n, err := got.Unmarshal(raw)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if n != len(raw) {
		t.Errorf("consumed %d, want %d", n, len(raw))
	}
	if got.MaxXmitFrag != b.MaxXmitFrag || got.MaxRecvFrag != b.MaxRecvFrag {
		t.Errorf("frag sizes: got %d/%d", got.MaxXmitFrag, got.MaxRecvFrag)
	}
	if len(got.ContextList) != 1 || len(got.ContextList[0].TransferSyntaxes) != 1 {
		t.Fatalf("context list not round-tripped: %+v", got.ContextList)
	}
	if !got.ContextList[0].AbstractSyntax.Equal(b.ContextList[0].AbstractSyntax) {
		t.Error("abstract syntax not round-tripped")
	}
	if !got.ContextList[0].TransferSyntaxes[0].Equal(syntax.NDRTransferSyntax()) {
		t.Error("transfer syntax not round-tripped")
	}
}

func TestBind_RejectsEmptyContextList(t *testing.T) {
	b := &Bind{MaxXmitFrag: 4280, MaxRecvFrag: 4280}
	if _, err := b.Marshal(); err == nil {
		t.Fatal("Marshal of bind with no contexts: error = nil, want non-nil")
	}
}

func TestBindAck_RoundTripWithAlignmentPad(t *testing.T) {
	// "lsarpc" + NUL = 7 bytes, forcing a 3-byte alignment pad before the result list.
	ack := &BindAck{
		MaxXmitFrag:      4280,
		MaxRecvFrag:      4280,
		AssocGroupID:     0x11223344,
		SecondaryAddress: "lsarpc",
		Results: []PresentationResult{
			{Result: ResultAcceptance, Reason: 0, TransferSyntax: syntax.NDRTransferSyntax()},
		},
	}
	raw, err := ack.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	// 16 + 8 + 2(len) + 7(addr) + 3(pad) + 4(result list hdr) + 24(one result) = 64
	if len(raw) != 64 {
		t.Errorf("marshalled length = %d, want 64 (alignment pad expected)", len(raw))
	}

	var got BindAck
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.SecondaryAddress != "lsarpc" {
		t.Errorf("SecondaryAddress = %q, want %q", got.SecondaryAddress, "lsarpc")
	}
	if got.AssocGroupID != 0x11223344 {
		t.Errorf("AssocGroupID = %#x, want 0x11223344", got.AssocGroupID)
	}
	if len(got.Results) != 1 || got.Results[0].Result != ResultAcceptance {
		t.Fatalf("results not round-tripped: %+v", got.Results)
	}
	if !got.Accepted() {
		t.Error("Accepted() = false, want true")
	}
	if !got.Results[0].TransferSyntax.Equal(syntax.NDRTransferSyntax()) {
		t.Error("result transfer syntax not round-tripped")
	}
}

func TestBindAck_EmptySecondaryAddress(t *testing.T) {
	ack := &BindAck{MaxXmitFrag: 4280, MaxRecvFrag: 4280}
	raw, err := ack.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got BindAck
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.SecondaryAddress != "" {
		t.Errorf("SecondaryAddress = %q, want empty", got.SecondaryAddress)
	}
	if got.Accepted() {
		t.Error("Accepted() = true on a bind_ack with no results")
	}
}

func TestBindNak_RoundTrip(t *testing.T) {
	nak := &BindNak{
		RejectReason: ReasonProposedTransferSyntaxesNotSupported,
		Versions:     []ProtocolVersion{{Major: 5, Minor: 0}},
	}
	raw, err := nak.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got BindNak
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.RejectReason != ReasonProposedTransferSyntaxesNotSupported {
		t.Errorf("RejectReason = %d", got.RejectReason)
	}
	if len(got.Versions) != 1 || got.Versions[0] != (ProtocolVersion{Major: 5, Minor: 0}) {
		t.Errorf("Versions = %+v", got.Versions)
	}
}

func TestRequest_RoundTrip(t *testing.T) {
	req := &Request{
		ContextID: 0,
		Opnum:     0x000f,
		Stub:      []byte{0xde, 0xad, 0xbe, 0xef},
	}
	req.Header.CallID = 7
	raw, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if PacketType(raw[2]) != PacketTypeRequest {
		t.Errorf("ptype = %d, want request", raw[2])
	}

	var got Request
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.Opnum != 0x000f || got.ContextID != 0 {
		t.Errorf("opnum/ctx = %d/%d", got.Opnum, got.ContextID)
	}
	if got.AllocHint != 4 {
		t.Errorf("AllocHint = %d, want 4 (defaulted to stub length)", got.AllocHint)
	}
	if !bytes.Equal(got.Stub, req.Stub) {
		t.Errorf("stub = %x, want %x", got.Stub, req.Stub)
	}
	if got.Header.CallID != 7 {
		t.Errorf("call_id = %d, want 7", got.Header.CallID)
	}
}

func TestRequest_WithObjectUUID(t *testing.T) {
	obj := &guid.GUID{A: 0xaabbccdd, B: 0x1122, C: 0x3344, D: 0x5566, E: 0x778899aabbcc}
	req := &Request{
		ContextID:  1,
		Opnum:      2,
		ObjectUUID: obj,
		Stub:       []byte{0x01},
	}
	raw, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !req.Header.PacketFlags.Has(PFCObjectUuid) {
		t.Error("PFCObjectUuid flag not set when ObjectUUID present")
	}

	var got Request
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.ObjectUUID == nil || !got.ObjectUUID.Equal(obj) {
		t.Errorf("ObjectUUID = %v, want %v", got.ObjectUUID, obj)
	}
	if !bytes.Equal(got.Stub, []byte{0x01}) {
		t.Errorf("stub = %x", got.Stub)
	}
}

func TestResponse_RoundTrip(t *testing.T) {
	resp := &Response{
		ContextID: 0,
		Stub:      []byte{0x11, 0x22, 0x33},
	}
	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got Response
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !bytes.Equal(got.Stub, resp.Stub) {
		t.Errorf("stub = %x, want %x", got.Stub, resp.Stub)
	}
	if got.AllocHint != 3 {
		t.Errorf("AllocHint = %d, want 3", got.AllocHint)
	}
}

func TestFault_RoundTripAndError(t *testing.T) {
	f := &Fault{
		ContextID: 0,
		Status:    NCASOpRngError,
	}
	raw, err := f.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got Fault
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.Status != NCASOpRngError {
		t.Errorf("Status = %#x, want %#x", got.Status, NCASOpRngError)
	}
	// Fault implements error.
	var asErr error = &got
	if asErr.Error() == "" {
		t.Error("Fault.Error() returned empty string")
	}
	if FaultStatus(got.Status).String() != "nca_s_op_rng_error" {
		t.Errorf("status string = %q", FaultStatus(got.Status).String())
	}
}

func TestPeekHeader_Dispatch(t *testing.T) {
	f := &Fault{Status: NCASFaultNDR}
	raw, _ := f.Marshal()
	h, err := PeekHeader(raw)
	if err != nil {
		t.Fatalf("PeekHeader() error = %v", err)
	}
	if h.PacketType != PacketTypeFault {
		t.Errorf("PacketType = %s, want fault", h.PacketType)
	}
}

func TestUnmarshal_WrongPacketType(t *testing.T) {
	b := &Bind{MaxXmitFrag: 1, MaxRecvFrag: 1, ContextList: []ContextElement{{TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()}}}}
	raw, _ := b.Marshal()
	var resp Response
	if _, err := resp.Unmarshal(raw); err == nil {
		t.Fatal("unmarshalling a bind as a response: error = nil, want non-nil")
	}
}
