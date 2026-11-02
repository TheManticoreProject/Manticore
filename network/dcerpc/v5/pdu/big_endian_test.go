package pdu

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// beResponse hand-builds a big-endian response PDU: an all-zero DREP (integer order
// 0) with every integer field most-significant byte first.
func beResponse(callID uint32, allocHint uint32, contextID uint16, cancelCount uint8, stub []byte) []byte {
	body := make([]byte, 8)
	binary.BigEndian.PutUint32(body[0:4], allocHint)
	binary.BigEndian.PutUint16(body[4:6], contextID)
	body[6] = cancelCount
	body = append(body, stub...)

	hdr := make([]byte, HeaderSize)
	hdr[0], hdr[1] = 5, 0
	hdr[2] = byte(PacketTypeResponse)
	hdr[3] = byte(PFCFirstFrag | PFCLastFrag)
	// DREP left all-zero: big-endian integers, ASCII, IEEE float.
	binary.BigEndian.PutUint16(hdr[8:10], uint16(HeaderSize+len(body)))
	binary.BigEndian.PutUint32(hdr[12:16], callID)
	return append(hdr, body...)
}

func TestResponse_UnmarshalBigEndian(t *testing.T) {
	stub := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	raw := beResponse(0x01020304, 4, 0x0007, 0, stub)

	var r Response
	if _, err := r.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal big-endian response: %v", err)
	}
	if r.Header.CallID != 0x01020304 {
		t.Errorf("CallID = %#x, want 0x01020304", r.Header.CallID)
	}
	if r.AllocHint != 4 {
		t.Errorf("AllocHint = %d, want 4", r.AllocHint)
	}
	if r.ContextID != 7 {
		t.Errorf("ContextID = %d, want 7", r.ContextID)
	}
	if string(r.Stub) != string(stub) {
		t.Errorf("Stub = %x, want %x", r.Stub, stub)
	}
}

func TestFault_UnmarshalBigEndian(t *testing.T) {
	// Big-endian fault: alloc_hint 0, context_id 1, cancel_count 0, status 0x1C010002
	// (nca_s_fault_ndr), no stub.
	body := make([]byte, 16)
	binary.BigEndian.PutUint16(body[4:6], 1)           // context_id
	binary.BigEndian.PutUint32(body[8:12], 0x1C010002) // status
	hdr := make([]byte, HeaderSize)
	hdr[0] = 5
	hdr[2] = byte(PacketTypeFault)
	hdr[3] = byte(PFCFirstFrag | PFCLastFrag)
	binary.BigEndian.PutUint16(hdr[8:10], uint16(HeaderSize+len(body)))
	binary.BigEndian.PutUint32(hdr[12:16], 9)
	raw := append(hdr, body...)

	var f Fault
	if _, err := f.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal big-endian fault: %v", err)
	}
	if f.Header.CallID != 9 {
		t.Errorf("CallID = %d, want 9", f.Header.CallID)
	}
	if f.ContextID != 1 {
		t.Errorf("ContextID = %d, want 1", f.ContextID)
	}
	if f.Status != 0x1C010002 {
		t.Errorf("Status = %#x, want 0x1C010002", f.Status)
	}
}

func TestSecTrailer_UnmarshalWithOrder(t *testing.T) {
	// auth_context_id 0x11223344 reads big-endian from these octets and
	// little-endian from the reverse.
	be := []byte{AuthTypeNTLMSSP, AuthLevelPktPrivacy, 0, 0, 0x11, 0x22, 0x33, 0x44}
	var st SecTrailer
	if err := st.UnmarshalWithOrder(be, binary.BigEndian); err != nil {
		t.Fatalf("UnmarshalWithOrder: %v", err)
	}
	if st.AuthContextID != 0x11223344 {
		t.Errorf("big-endian AuthContextID = %#x, want 0x11223344", st.AuthContextID)
	}
	// The little-endian Unmarshal reads the same octets the other way, unchanged.
	var le SecTrailer
	if err := le.Unmarshal(be); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if le.AuthContextID != 0x44332211 {
		t.Errorf("little-endian AuthContextID = %#x, want 0x44332211", le.AuthContextID)
	}
}

// reverseBytes reverses a byte slice in place and returns it.
func reverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

// toBigEndianSyntax takes the little-endian wire form of a syntax identifier (what
// SyntaxID.Marshal produces) and re-encodes it big-endian: the three integer UUID
// fields (time_low, time_mid, time_hi_and_version) and the 2-byte major and minor
// versions swap bytes, while the clock_seq and node octets (bytes 8..16) are byte
// strings and stay put.
func toBigEndianSyntax(le []byte) []byte {
	be := append([]byte(nil), le...)
	reverseBytes(be[0:4])   // time_low
	reverseBytes(be[4:6])   // time_mid
	reverseBytes(be[6:8])   // time_hi_and_version
	reverseBytes(be[16:18]) // major version
	reverseBytes(be[18:20]) // minor version
	return be
}

func TestBindAck_UnmarshalBigEndian(t *testing.T) {
	// A big-endian bind_ack that accepts the first context with the NDR transfer
	// syntax must decode its scalar fields and negotiate that syntax correctly under
	// the receiver-makes-right rule.
	ndr := syntax.NDRTransferSyntax()
	leSyntax, err := ndr.Marshal()
	if err != nil {
		t.Fatalf("marshal NDR syntax: %v", err)
	}
	beSyntax := toBigEndianSyntax(leSyntax)

	body := make([]byte, 8)
	binary.BigEndian.PutUint16(body[0:2], 0x16D0) // max_xmit_frag
	binary.BigEndian.PutUint16(body[2:4], 0x16D0) // max_recv_frag
	binary.BigEndian.PutUint32(body[4:8], 0x1234) // assoc_group_id

	// sec_addr: a NUL-terminated "135" (length includes the NUL).
	addr := append([]byte("135"), 0x00)
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(addr)))
	body = append(body, lenBuf...)
	body = append(body, addr...)
	// Pad to a 4-byte boundary measured from the start of the PDU.
	body = append(body, make([]byte, align4(HeaderSize+len(body)))...)

	// result list: n_results = 1, reserved, then one p_result_t.
	body = append(body, 0x01, 0x00, 0x00, 0x00)
	result := make([]byte, 4)
	binary.BigEndian.PutUint16(result[0:2], ResultAcceptance)
	binary.BigEndian.PutUint16(result[2:4], ReasonNotSpecified)
	result = append(result, beSyntax...)
	body = append(body, result...)

	hdr := make([]byte, HeaderSize)
	hdr[0] = 5
	hdr[2] = byte(PacketTypeBindAck)
	hdr[3] = byte(PFCFirstFrag | PFCLastFrag)
	// DREP left all-zero: big-endian integers.
	binary.BigEndian.PutUint16(hdr[8:10], uint16(HeaderSize+len(body)))
	binary.BigEndian.PutUint32(hdr[12:16], 42)
	raw := append(hdr, body...)

	var ack BindAck
	if _, err := ack.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal big-endian bind_ack: %v", err)
	}
	if ack.Header.CallID != 42 {
		t.Errorf("CallID = %d, want 42", ack.Header.CallID)
	}
	if ack.MaxXmitFrag != 0x16D0 || ack.MaxRecvFrag != 0x16D0 {
		t.Errorf("frag sizes = %d/%d, want 0x16D0/0x16D0", ack.MaxXmitFrag, ack.MaxRecvFrag)
	}
	if ack.AssocGroupID != 0x1234 {
		t.Errorf("AssocGroupID = %#x, want 0x1234", ack.AssocGroupID)
	}
	if ack.SecondaryAddress != "135" {
		t.Errorf("SecondaryAddress = %q, want %q", ack.SecondaryAddress, "135")
	}
	if !ack.Accepted() {
		t.Fatal("bind_ack not reported as accepted")
	}
	if len(ack.Results) != 1 {
		t.Fatalf("got %d results, want 1", len(ack.Results))
	}
	if !ack.Results[0].TransferSyntax.Equal(ndr) {
		t.Errorf("negotiated transfer syntax = %s, want %s", ack.Results[0].TransferSyntax, ndr)
	}
}

func TestBind_UnmarshalBigEndian(t *testing.T) {
	// Round-trip the abstract syntax through a big-endian bind: one context proposing
	// an interface UUID with a candidate NDR transfer syntax.
	abstract := syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x12345678, B: 0x9abc, C: 0xdef0, D: 0x1122, E: 0x334455667788},
		MajorVersion: 3,
		MinorVersion: 1,
	}
	leAbs, _ := abstract.Marshal()
	leTs, _ := syntax.NDRTransferSyntax().Marshal()

	body := make([]byte, 8)
	binary.BigEndian.PutUint16(body[0:2], 0x16D0)
	binary.BigEndian.PutUint16(body[2:4], 0x16D0)
	binary.BigEndian.PutUint32(body[4:8], 0)
	body = append(body, 0x01, 0x00, 0x00, 0x00) // n_context_elem = 1
	// context element: context_id 0, n_transfer_syn 1, reserved
	ce := make([]byte, 4)
	binary.BigEndian.PutUint16(ce[0:2], 0)
	ce[2] = 1
	ce = append(ce, toBigEndianSyntax(leAbs)...)
	ce = append(ce, toBigEndianSyntax(leTs)...)
	body = append(body, ce...)

	hdr := make([]byte, HeaderSize)
	hdr[0] = 5
	hdr[2] = byte(PacketTypeBind)
	hdr[3] = byte(PFCFirstFrag | PFCLastFrag)
	binary.BigEndian.PutUint16(hdr[8:10], uint16(HeaderSize+len(body)))
	binary.BigEndian.PutUint32(hdr[12:16], 7)
	raw := append(hdr, body...)

	var b Bind
	if _, err := b.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal big-endian bind: %v", err)
	}
	if len(b.ContextList) != 1 {
		t.Fatalf("got %d contexts, want 1", len(b.ContextList))
	}
	if !b.ContextList[0].AbstractSyntax.Equal(abstract) {
		t.Errorf("abstract syntax = %s, want %s", b.ContextList[0].AbstractSyntax, abstract)
	}
	if n := len(b.ContextList[0].TransferSyntaxes); n != 1 {
		t.Fatalf("got %d transfer syntaxes, want 1", n)
	}
	if !b.ContextList[0].TransferSyntaxes[0].Equal(syntax.NDRTransferSyntax()) {
		t.Errorf("transfer syntax = %s, want NDR", b.ContextList[0].TransferSyntaxes[0])
	}
}

func TestBindNak_UnmarshalBigEndian(t *testing.T) {
	// A big-endian bind_nak reads its reject reason in the sender's order.
	body := make([]byte, 3)
	binary.BigEndian.PutUint16(body[0:2], ReasonProposedTransferSyntaxesNotSupported)
	body[2] = 1 // one protocol version
	body = append(body, 5, 0)

	hdr := make([]byte, HeaderSize)
	hdr[0] = 5
	hdr[2] = byte(PacketTypeBindNak)
	hdr[3] = byte(PFCFirstFrag | PFCLastFrag)
	binary.BigEndian.PutUint16(hdr[8:10], uint16(HeaderSize+len(body)))
	raw := append(hdr, body...)

	var nak BindNak
	if _, err := nak.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal big-endian bind_nak: %v", err)
	}
	if nak.RejectReason != ReasonProposedTransferSyntaxesNotSupported {
		t.Errorf("RejectReason = %d, want %d", nak.RejectReason, ReasonProposedTransferSyntaxesNotSupported)
	}
	if len(nak.Versions) != 1 || nak.Versions[0] != (ProtocolVersion{Major: 5, Minor: 0}) {
		t.Errorf("Versions = %+v, want [{5 0}]", nak.Versions)
	}
}
