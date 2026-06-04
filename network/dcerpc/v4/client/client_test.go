package client

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// testActivity is the fixed activity UUID used so scripted server PDUs match the
// client's conversation.
var testActivity = guid.GUID{A: 0xaabbccdd, B: 0x1122, C: 0x3344, D: 0x5566, E: 0x778899aabbcc}

// timeoutError is a net.Error reporting a timeout, the signal the real transport
// gives when no datagram arrived before the deadline.
type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

var _ net.Error = timeoutError{}

// recvEvent is one scripted result of a Recv call.
type recvEvent struct {
	data []byte
	err  error
}

// scriptedConn is a transport.Transport whose Recv returns a scripted sequence of
// datagrams or errors, and which records everything sent. When the script is
// exhausted, Recv reports a timeout, exercising the client's retransmit/ping paths.
type scriptedConn struct {
	sent   [][]byte
	events []recvEvent
	maxPDU int
}

var _ transport.Transport = (*scriptedConn)(nil)

func (m *scriptedConn) Connect() error { return nil }

func (m *scriptedConn) Send(b []byte) (int, error) {
	m.sent = append(m.sent, append([]byte(nil), b...))
	return len(b), nil
}

func (m *scriptedConn) Recv() ([]byte, error) {
	if len(m.events) == 0 {
		return nil, timeoutError{}
	}
	e := m.events[0]
	m.events = m.events[1:]
	return e.data, e.err
}

func (m *scriptedConn) SetDeadline(time.Time) error { return nil }

func (m *scriptedConn) MaxPDUSize() int {
	if m.maxPDU == 0 {
		return transport.MaxPDUSizeDefault
	}
	return m.maxPDU
}

func (m *scriptedConn) RemoteAddr() net.Addr { return nil }
func (m *scriptedConn) IsConnected() bool    { return true }
func (m *scriptedConn) Close() error         { return nil }

// serverPDU builds a marshalled server PDU addressed to the test conversation.
func serverPDU(t *testing.T, pt pdu.PacketType, seq uint32, flags pdu.Flags1, fragnum uint16, body []byte) []byte {
	t.Helper()
	h := pdu.NewHeader(pt)
	h.ActivityID = testActivity
	h.SequenceNumber = seq
	h.Flags1 |= flags
	h.FragmentNumber = fragnum
	h.ServerBoot = 0x12345678
	raw, err := (&pdu.PDU{Header: h, Body: body}).Marshal()
	if err != nil {
		t.Fatalf("build server PDU: %v", err)
	}
	return raw
}

// decodeSent unmarshals a recorded outbound datagram.
func decodeSent(t *testing.T, raw []byte) pdu.PDU {
	t.Helper()
	var p pdu.PDU
	if _, err := p.Unmarshal(raw); err != nil {
		t.Fatalf("decode sent PDU: %v", err)
	}
	return p
}

// countSent returns how many recorded outbound PDUs have the given packet type.
func countSent(t *testing.T, conn *scriptedConn, pt pdu.PacketType) int {
	t.Helper()
	n := 0
	for _, raw := range conn.sent {
		if decodeSent(t, raw).Header.PacketType == pt {
			n++
		}
	}
	return n
}

func newTestClient(conn *scriptedConn, opts ...Option) *Client {
	return New(conn, append([]Option{WithActivityID(testActivity)}, opts...)...)
}

func TestCallSingleFragmentResponse(t *testing.T) {
	want := []byte("output parameters")
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, 0, 0, want)},
	}}
	c := newTestClient(conn)

	got, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, OpNum: 3, Stub: []byte("in"), Idempotent: true})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("response = % x, want % x", got, want)
	}
	if countSent(t, conn, pdu.PacketTypeRequest) != 1 {
		t.Fatalf("expected exactly 1 request PDU, sent %d", countSent(t, conn, pdu.PacketTypeRequest))
	}
	req := decodeSent(t, conn.sent[0])
	if req.Header.SequenceNumber != 0 || !req.Header.ActivityID.Equal(&testActivity) {
		t.Errorf("request header wrong: seq=%d act=%s", req.Header.SequenceNumber, req.Header.ActivityID.ToFormatD())
	}
	if req.Header.OpNum != 3 {
		t.Errorf("request opnum = %d, want 3", req.Header.OpNum)
	}
	if !req.Header.Flags1.Has(pdu.Flags1Idempotent) {
		t.Errorf("request should carry idempotent flag, got %s", req.Header.Flags1)
	}
}

func TestCallFragmentedResponseReassembly(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, pdu.Flags1Frag, 0, []byte("HELLO "))},
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, pdu.Flags1Frag|pdu.Flags1LastFrag, 1, []byte("WORLD"))},
	}}
	c := newTestClient(conn)

	got, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, OpNum: 0, Idempotent: true})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if string(got) != "HELLO WORLD" {
		t.Fatalf("reassembled = %q, want %q", got, "HELLO WORLD")
	}
	// Each response fragment had the frag flag set, so each should be fack'd.
	if facks := countSent(t, conn, pdu.PacketTypeFack); facks != 2 {
		t.Errorf("expected 2 facks for 2 response fragments, sent %d", facks)
	}
}

func TestCallMultiFragmentRequest(t *testing.T) {
	// maxBody = 4 forces a 10-byte stub into 3 request fragments.
	conn := &scriptedConn{
		maxPDU: pdu.HeaderSize + 4,
		events: []recvEvent{{data: serverPDU(t, pdu.PacketTypeResponse, 0, 0, 0, []byte("ok"))}},
	}
	c := newTestClient(conn)

	if _, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Stub: []byte("0123456789"), Idempotent: true}); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if reqs := countSent(t, conn, pdu.PacketTypeRequest); reqs != 3 {
		t.Fatalf("expected 3 request fragments, sent %d", reqs)
	}
	for i := 0; i < 3; i++ {
		p := decodeSent(t, conn.sent[i])
		if p.Header.FragmentNumber != uint16(i) {
			t.Errorf("request fragment %d has fragnum %d", i, p.Header.FragmentNumber)
		}
		if !p.Header.Flags1.Has(pdu.Flags1Frag) {
			t.Errorf("request fragment %d missing frag flag", i)
		}
		if last := i == 2; p.Header.Flags1.Has(pdu.Flags1LastFrag) != last {
			t.Errorf("request fragment %d lastfrag wrong", i)
		}
	}
}

func TestCallWorkingThenResponse(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeWorking, 0, 0, 0, nil)},
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, 0, 0, []byte("done"))},
	}}
	c := newTestClient(conn)

	got, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if string(got) != "done" {
		t.Fatalf("response = %q, want done", got)
	}
}

func TestCallNoCallTriggersRetransmit(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeNoCall, 0, 0, 0, nil)},
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, 0, 0, []byte("late"))},
	}}
	c := newTestClient(conn)

	got, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if string(got) != "late" {
		t.Fatalf("response = %q, want late", got)
	}
	if reqs := countSent(t, conn, pdu.PacketTypeRequest); reqs != 2 {
		t.Fatalf("nocall should trigger one retransmit (2 request sends), got %d", reqs)
	}
}

func TestCallRetransmitOnTimeoutThenSuccess(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{err: timeoutError{}},
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, 0, 0, []byte("ok"))},
	}}
	c := newTestClient(conn)

	if _, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true}); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if reqs := countSent(t, conn, pdu.PacketTypeRequest); reqs != 2 {
		t.Fatalf("timeout before any response should retransmit the request (2 sends), got %d", reqs)
	}
}

func TestCallGivesUpAfterMaxRequests(t *testing.T) {
	conn := &scriptedConn{} // empty script: Recv always times out
	c := newTestClient(conn, WithMaxRequests(2))

	_, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true})
	if !errors.Is(err, ErrNoResponse) {
		t.Fatalf("expected ErrNoResponse, got %v", err)
	}
	// 1 initial send + 2 retransmits before giving up.
	if reqs := countSent(t, conn, pdu.PacketTypeRequest); reqs != 3 {
		t.Fatalf("expected 3 request sends with MaxRequests=2, got %d", reqs)
	}
}

func TestCallPingsWhileAwaitingResponseFragments(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, pdu.Flags1Frag, 0, []byte("part1"))},
		{err: timeoutError{}},
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, pdu.Flags1Frag|pdu.Flags1LastFrag, 1, []byte("part2"))},
	}}
	c := newTestClient(conn)

	iface := guid.GUID{A: 1}
	got, err := c.Call(CallRequest{Interface: iface, InterfaceVersion: 1, OpNum: 9, Idempotent: true})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if string(got) != "part1part2" {
		t.Fatalf("reassembled = %q, want part1part2", got)
	}
	if pings := countSent(t, conn, pdu.PacketTypePing); pings != 1 {
		t.Fatalf("expected 1 ping while awaiting fragments, sent %d", pings)
	}
	// The ping must identify the call: activity, sequence number, interface, and opnum.
	for _, raw := range conn.sent {
		p := decodeSent(t, raw)
		if p.Header.PacketType != pdu.PacketTypePing {
			continue
		}
		if !p.Header.InterfaceID.Equal(&iface) || p.Header.OpNum != 9 || p.Header.SequenceNumber != 0 {
			t.Errorf("ping does not carry the call identity: iface=%s op=%d seq=%d",
				p.Header.InterfaceID.ToFormatD(), p.Header.OpNum, p.Header.SequenceNumber)
		}
	}
}

// repeatConn always returns the same PDU, simulating a server that floods
// non-terminal PDUs forever.
type repeatConn struct{ pdu []byte }

func (m *repeatConn) Connect() error              { return nil }
func (m *repeatConn) Send([]byte) (int, error)    { return 0, nil }
func (m *repeatConn) Recv() ([]byte, error)       { return append([]byte(nil), m.pdu...), nil }
func (m *repeatConn) SetDeadline(time.Time) error { return nil }
func (m *repeatConn) MaxPDUSize() int             { return transport.MaxPDUSizeDefault }
func (m *repeatConn) RemoteAddr() net.Addr        { return nil }
func (m *repeatConn) IsConnected() bool           { return true }
func (m *repeatConn) Close() error                { return nil }

func TestCallBoundsFloodOfNonTerminalPDUs(t *testing.T) {
	// A server that endlessly sends "working" never trips the retransmission or ping
	// limits (working resets requests and arrives before any timeout), so without the
	// receive backstop the loop would never terminate.
	conn := &repeatConn{pdu: serverPDU(t, pdu.PacketTypeWorking, 0, 0, 0, nil)}
	c := New(conn, WithActivityID(testActivity), WithMaxReceives(16))

	if _, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true}); !errors.Is(err, ErrNoResponse) {
		t.Fatalf("expected ErrNoResponse from the receive backstop, got %v", err)
	}
}

func TestCallFault(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeFault, 0, 0, 0, pdu.MarshalStatusBody(0x1c010003))},
	}}
	c := newTestClient(conn)

	_, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true})
	var fe *FaultError
	if !errors.As(err, &fe) {
		t.Fatalf("expected *FaultError, got %v", err)
	}
	if fe.Status != 0x1c010003 {
		t.Fatalf("fault status = 0x%08x, want 0x1c010003", fe.Status)
	}
}

func TestCallReject(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeReject, 0, 0, 0, pdu.MarshalStatusBody(0x1c000008))},
	}}
	c := newTestClient(conn)

	_, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true})
	var re *RejectError
	if !errors.As(err, &re) {
		t.Fatalf("expected *RejectError, got %v", err)
	}
	if re.Status != 0x1c000008 {
		t.Fatalf("reject status = 0x%08x, want 0x1c000008", re.Status)
	}
}

func TestCallIgnoresForeignActivity(t *testing.T) {
	// A PDU for a different activity must be ignored; the call then times out.
	foreign := pdu.NewHeader(pdu.PacketTypeResponse)
	foreign.ActivityID = guid.GUID{A: 0xdeadbeef}
	foreignRaw, _ := (&pdu.PDU{Header: foreign, Body: []byte("not mine")}).Marshal()

	conn := &scriptedConn{events: []recvEvent{{data: foreignRaw}}}
	c := newTestClient(conn, WithMaxRequests(0), WithMaxPings(0))

	if _, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true}); !errors.Is(err, ErrNoResponse) {
		t.Fatalf("expected ErrNoResponse after ignoring foreign activity, got %v", err)
	}
}

func TestSequenceNumberIncrementsPerCall(t *testing.T) {
	conn := &scriptedConn{events: []recvEvent{
		{data: serverPDU(t, pdu.PacketTypeResponse, 0, 0, 0, []byte("a"))},
		{data: serverPDU(t, pdu.PacketTypeResponse, 1, 0, 0, []byte("b"))},
	}}
	c := newTestClient(conn)

	for i, want := range []string{"a", "b"} {
		got, err := c.Call(CallRequest{Interface: guid.GUID{A: 1}, InterfaceVersion: 1, Idempotent: true})
		if err != nil {
			t.Fatalf("Call %d: %v", i, err)
		}
		if string(got) != want {
			t.Fatalf("Call %d response = %q, want %q", i, got, want)
		}
	}
	// The two requests must carry sequence numbers 0 and 1.
	if seq := decodeSent(t, conn.sent[0]).Header.SequenceNumber; seq != 0 {
		t.Errorf("first call seq = %d, want 0", seq)
	}
	last := conn.sent[len(conn.sent)-1]
	if seq := decodeSent(t, last).Header.SequenceNumber; seq != 1 {
		t.Errorf("second call seq = %d, want 1", seq)
	}
}
