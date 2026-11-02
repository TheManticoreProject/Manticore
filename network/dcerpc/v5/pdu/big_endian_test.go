package pdu

import (
	"encoding/binary"
	"testing"
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

func TestBindFamily_RejectsBigEndian(t *testing.T) {
	// A big-endian bind-family PDU is refused with a clear error rather than
	// mis-parsed, because its presentation syntax identifiers are not yet decoded
	// byte-order-aware. The data path (response, fault) accepts big-endian.
	for _, pt := range []PacketType{PacketTypeBind, PacketTypeBindAck, PacketTypeBindNak} {
		hdr := make([]byte, HeaderSize)
		hdr[0] = 5
		hdr[2] = byte(pt)
		hdr[3] = byte(PFCFirstFrag | PFCLastFrag)
		// all-zero DREP = big-endian; pad with zeroes for the body.
		binary.BigEndian.PutUint16(hdr[8:10], uint16(HeaderSize+16))
		raw := append(hdr, make([]byte, 16)...)

		var err error
		switch pt {
		case PacketTypeBind:
			var b Bind
			_, err = b.Unmarshal(raw)
		case PacketTypeBindAck:
			var b BindAck
			_, err = b.Unmarshal(raw)
		case PacketTypeBindNak:
			var b BindNak
			_, err = b.Unmarshal(raw)
		}
		if err == nil {
			t.Errorf("%s: big-endian bind-family PDU was accepted, want an error", pt)
		}
	}
}
