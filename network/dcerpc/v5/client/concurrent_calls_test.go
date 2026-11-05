package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// echoTransport is a blocking, Close-safe in-memory transport that answers each
// bind with a bind_ack and each request with a response echoing the request's
// call_id and stub. Recv blocks until a reply is queued or the transport is closed,
// modelling a real transport, so several concurrent Call goroutines can be driven
// against one Client and the demultiplexer's by-call_id routing checked under -race.
type echoTransport struct {
	mu     sync.Mutex
	cond   *sync.Cond
	queue  [][]byte
	closed bool
}

func newEchoTransport() *echoTransport {
	e := &echoTransport{}
	e.cond = sync.NewCond(&e.mu)
	return e
}

func (e *echoTransport) Connect() error      { return nil }
func (e *echoTransport) MaxXmitFrag() uint16 { return 4280 }
func (e *echoTransport) MaxRecvFrag() uint16 { return 4280 }

func (e *echoTransport) enqueue(b []byte) {
	e.mu.Lock()
	if !e.closed {
		e.queue = append(e.queue, b)
		e.cond.Signal()
	}
	e.mu.Unlock()
}

func (e *echoTransport) Send(p []byte) error {
	hdr, err := pdu.PeekHeader(p)
	if err != nil {
		return err
	}
	switch hdr.PacketType {
	case pdu.PacketTypeBind:
		ack := &pdu.BindAck{
			MaxXmitFrag:      4280,
			MaxRecvFrag:      4280,
			SecondaryAddress: "lsarpc",
			Results:          []pdu.PresentationResult{{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()}},
		}
		ack.Header = pdu.NewHeader(pdu.PacketTypeBindAck, pdu.PFCFirstFrag|pdu.PFCLastFrag, hdr.CallID)
		b, err := ack.Marshal()
		if err != nil {
			return err
		}
		e.enqueue(b)
	case pdu.PacketTypeRequest:
		var req pdu.Request
		if _, err := req.Unmarshal(p); err != nil {
			return err
		}
		// Echo the request stub back, tagged with the same call_id and context, so
		// the caller can confirm it received the answer to its own call.
		resp := &pdu.Response{ContextID: req.ContextID, Stub: req.Stub}
		resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, req.Header.CallID)
		b, err := resp.Marshal()
		if err != nil {
			return err
		}
		e.enqueue(b)
	default:
		return fmt.Errorf("echoTransport: unexpected PDU type %s", hdr.PacketType)
	}
	return nil
}

func (e *echoTransport) Recv() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for len(e.queue) == 0 && !e.closed {
		e.cond.Wait()
	}
	if len(e.queue) == 0 {
		return nil, errors.New("echoTransport closed")
	}
	b := e.queue[0]
	e.queue = e.queue[1:]
	return b, nil
}

func (e *echoTransport) Close() error {
	e.mu.Lock()
	e.closed = true
	e.cond.Broadcast()
	e.mu.Unlock()
	return nil
}

// TestConcurrentCallsRouteByCallID issues many Calls concurrently on one Client and
// checks each goroutine gets its own response back — the demultiplexer must route by
// call_id. Run under -race it also guards the shared call table, send path and
// reader against data races.
//
// Note: echoTransport is internally synchronized, so this test exercises by-call_id
// routing but not the `outstanding` reader gate's race-avoidance against a transport
// that produces its reply during Send. That property is validated by the live
// concurrent probe (concurrent SamrConnect2 over a real pipe); do not "simplify" the
// gate by moving the outstanding++ before the send on the strength of this test.
func TestConcurrentCallsRouteByCallID(t *testing.T) {
	c := NewClient(newEchoTransport())
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	defer c.Close()

	const workers = 64
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// A distinct 8-byte stub per worker; the echo server returns it verbatim.
			want := make([]byte, 8)
			binary.LittleEndian.PutUint64(want, uint64(0xA5A50000)+uint64(i))
			got, err := c.Call(uint16(i%16), want)
			if err != nil {
				errs <- fmt.Errorf("worker %d: Call: %w", i, err)
				return
			}
			if !bytes.Equal(got, want) {
				errs <- fmt.Errorf("worker %d: got % x, want % x (response routed to the wrong call)", i, got, want)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// TestCloseRacesFirstCall stresses Close() concurrent with the first Call(), the
// case that first touches the demultiplexer's shared state. It guards against the
// client's own state (calls, readerCond, bound) being read/written without
// synchronization — those are initialized in NewClient, before any goroutine, and
// bound is read and written under callsMu. Run under -race.
func TestCloseRacesFirstCall(t *testing.T) {
	for i := 0; i < 200; i++ {
		c := NewClient(newEchoTransport())
		if err := c.Bind(testSyntax()); err != nil {
			t.Fatalf("Bind: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = c.Call(0, []byte("x")) }()
		go func() { defer wg.Done(); _ = c.Close() }()
		wg.Wait()
	}
}
