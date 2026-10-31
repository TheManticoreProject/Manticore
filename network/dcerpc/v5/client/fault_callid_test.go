package client

import (
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// TestCall_FaultCallIDMismatch checks that a fault answering some other call is not
// reported as the outcome of the current call: a stale fault left on the connection
// (call_id 7) must not satisfy a request issued with call_id 2.
func TestCall_FaultCallIDMismatch(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	fault := &pdu.Fault{Status: pdu.NCASOpRngError}
	fault.Header = pdu.NewHeader(pdu.PacketTypeFault, pdu.PFCFirstFrag|pdu.PFCLastFrag, 7)
	ft.queue(mustMarshal(t, fault))

	_, err := c.Call(0x0099, nil)
	if err == nil {
		t.Fatal("Call() with mismatched fault call_id: error = nil, want non-nil")
	}
	var fe *pdu.Fault
	if errors.As(err, &fe) {
		t.Fatalf("Call() returned the fault of call_id 7 as the result of call_id 2: %v", err)
	}
}

// TestCall_FaultCallIDZero checks that a fault with call_id 0, which a server sends
// for an error it cannot attribute to a specific call, is still surfaced as a fault.
func TestCall_FaultCallIDZero(t *testing.T) {
	ft := &fakeTransport{maxXmit: 4280, maxRecv: 4280}
	ft.queue(bindAckBytes(t, 4280, 4280))
	c := NewClient(ft)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	fault := &pdu.Fault{Status: pdu.NCASProtoError}
	fault.Header = pdu.NewHeader(pdu.PacketTypeFault, pdu.PFCFirstFrag|pdu.PFCLastFrag, 0)
	ft.queue(mustMarshal(t, fault))

	_, err := c.Call(0x0099, nil)
	var fe *pdu.Fault
	if !errors.As(err, &fe) {
		t.Fatalf("Call() error = %v, want *pdu.Fault", err)
	}
	if fe.Status != pdu.NCASProtoError {
		t.Errorf("fault status = %#x, want %#x", fe.Status, pdu.NCASProtoError)
	}
}
