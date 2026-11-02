package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// TestCall_ResponseContextMismatch checks that a Response naming another
// presentation context than the one the request used is refused: its stub is
// encoded for that context's transfer syntax, not the one the call expects.
func TestCall_ResponseContextMismatch(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	resp := &pdu.Response{ContextID: c.contextID + 1, Stub: []byte{1, 2, 3, 4}}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID+1)
	ft.queue(mustMarshal(t, resp))

	if stub, err := c.Call(0x0001, nil); err == nil {
		t.Fatalf("Call() accepted a response on context %d for a request on context %d: stub %x", resp.ContextID, c.contextID, stub)
	}
}

// TestCall_ResponseContextMatch checks that the matching context still passes.
func TestCall_ResponseContextMatch(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	resp := &pdu.Response{ContextID: c.contextID, Stub: []byte{1, 2, 3, 4}}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID+1)
	ft.queue(mustMarshal(t, resp))

	stub, err := c.Call(0x0001, nil)
	if err != nil || len(stub) != 4 {
		t.Fatalf("Call() = %x, %v; want the 4-byte stub", stub, err)
	}
}
