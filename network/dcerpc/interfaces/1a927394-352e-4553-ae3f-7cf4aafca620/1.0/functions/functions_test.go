package functions_test

import (
	"bytes"
	"testing"

	WdsRpcInterface "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a927394-352e-4553-ae3f-7cf4aafca620/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a927394-352e-4553-ae3f-7cf4aafca620/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// responder is an ndr.Invoker that records the marshalled request stub (so the on-the-wire
// NDR layout can be asserted) and replies with a canned response stub.
type responder struct {
	stub  []byte
	opnum uint16
	resp  []byte
}

func (r *responder) Invoke(in ndr.Call, out any) error {
	b, err := ndr.Request(in)
	if err != nil {
		return err
	}
	r.stub = b
	r.opnum = in.Opnum()
	if r.resp == nil {
		return nil
	}
	return ndr.Response(r.resp, out)
}

// replyStub is the response wire shape of WdsRpcMessage ([MS-WDSC] 3.1.4.1): the inline
// [out] puReplyPacketSize, then the [unique] pointer to the conformant reply byte array
// (pbReplyPacket), then the trailing Win32 return code. The size_is count is derived from
// the slice on marshal, exactly as the server encodes it.
type replyStub struct {
	PuReplyPacketSize uint32
	PbReplyPacket     []byte `ndr:"unique,size_is=PuReplyPacketSize"`
	Status            uint32 `ndr:"retval"`
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return b
}

// TestWdsRpcMessageRoundTrip drives a full request/response: the client sends a request
// packet and the server returns a reply packet. It pins the opnum, checks the request
// bytes appear in the request stub with the correct conformant layout, and asserts the
// reply packet is decoded intact with its size.
func TestWdsRpcMessageRoundTrip(t *testing.T) {
	request := []byte{0xca, 0xfe, 0xba, 0xbe, 0x01, 0x02, 0x03}
	reply := []byte{0xde, 0xad, 0xbe, 0xef, 0x11, 0x22}

	r := &responder{resp: mustMarshal(t, &replyStub{PbReplyPacket: reply})}
	size, got, err := functions.WdsRpcMessage(r, ndr.DWORD(len(request)), request)
	if err != nil {
		t.Fatalf("WdsRpcMessage: %v", err)
	}
	if r.opnum != WdsRpcInterface.OpnumWdsRpcMessage {
		t.Fatalf("opnum = %d, want %d", r.opnum, WdsRpcInterface.OpnumWdsRpcMessage)
	}
	if uint32(size) != uint32(len(reply)) {
		t.Fatalf("reply size = %d, want %d", size, len(reply))
	}
	if !bytes.Equal(got, reply) {
		t.Fatalf("reply packet = %x, want %x", got, reply)
	}
	// bRequestPacket is a [ref] pointer to a conformant array, so no referent id is
	// emitted: the request stub begins with uRequestPacketSize (7), then the conformant
	// max_count (7), then the request bytes.
	if len(r.stub) < 8 || r.stub[0] != 0x07 || r.stub[4] != 0x07 {
		t.Fatalf("expected uRequestPacketSize=7 and max_count=7 at start, got % x", r.stub[:min(8, len(r.stub))])
	}
	if !bytes.Contains(r.stub, request) {
		t.Fatalf("request packet %x not found in request stub %x", request, r.stub)
	}
}

// TestWdsRpcMessageErrorStatus verifies a nonzero Win32 return code is surfaced as an
// error and the (empty) reply packet is not returned as success.
func TestWdsRpcMessageErrorStatus(t *testing.T) {
	r := &responder{resp: mustMarshal(t, &replyStub{Status: WdsRpcInterface.ErrorAccessDenied})}
	_, got, err := functions.WdsRpcMessage(r, 1, []byte{0x01})
	if err == nil {
		t.Fatalf("expected error for return code 0x%08x, got nil (reply=%x)", WdsRpcInterface.ErrorAccessDenied, got)
	}
}
