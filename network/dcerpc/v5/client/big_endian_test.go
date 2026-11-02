package client

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// beResponseBytes hand-builds a big-endian response PDU on context 0: an all-zero
// DREP with every integer field most-significant byte first.
func beResponseBytes(callID uint32, stub []byte) []byte {
	body := make([]byte, 8)
	binary.BigEndian.PutUint32(body[0:4], uint32(len(stub))) // alloc_hint
	// context_id (body[4:6]) and cancel_count/reserved left zero.
	body = append(body, stub...)

	hdr := make([]byte, pdu.HeaderSize)
	hdr[0] = 5
	hdr[2] = byte(pdu.PacketTypeResponse)
	hdr[3] = byte(pdu.PFCFirstFrag | pdu.PFCLastFrag)
	// DREP all-zero => big-endian integers.
	binary.BigEndian.PutUint16(hdr[8:10], uint16(pdu.HeaderSize+len(body)))
	binary.BigEndian.PutUint32(hdr[12:16], callID)
	return append(hdr, body...)
}

// TestInvoke_DecodesBigEndianResponse drives a full call whose response arrives in
// big-endian and asserts the stub is decoded in that order (receiver-makes-right):
// the octets 0xCA 0xFE 0xBA 0xBE decode to the dword 0xCAFEBABE, where a little-
// endian decode would wrongly yield 0xBEBAFECA.
func TestInvoke_DecodesBigEndianResponse(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	// The first call after the bind uses call_id 2.
	ft.queue(beResponseBytes(2, []byte{0xCA, 0xFE, 0xBA, 0xBE}))

	var out invokeResp
	if err := c.Invoke(&invokeReq{A: 0x11, B: 0x22}, &out); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if out.X != 0xCAFEBABE {
		t.Errorf("out.X = %#x, want 0xcafebabe (big-endian stub decoded big-endian)", out.X)
	}
}

// TestInvoke_LittleEndianResponseUnchanged pins that an ordinary little-endian
// response still decodes as before: the same octets yield 0xCAFEBABE only because
// the little-endian stub holds them least-significant byte first.
func TestInvoke_LittleEndianResponseUnchanged(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	ft.queue(responseBytes(t, 2, pdu.PFCFirstFrag|pdu.PFCLastFrag, []byte{0xBE, 0xBA, 0xFE, 0xCA}))

	var out invokeResp
	if err := c.Invoke(&invokeReq{A: 0x11, B: 0x22}, &out); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if out.X != 0xCAFEBABE {
		t.Errorf("out.X = %#x, want 0xcafebabe", out.X)
	}
}
