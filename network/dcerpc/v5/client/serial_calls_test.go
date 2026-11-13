package client

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// serialModeEchoTransport wraps echoTransport, reports itself as a serial-call
// transport (transport.SerialCaller) like the SMB named-pipe transport, and tracks
// how many call exchanges are in flight at once. A request increments the gauge when
// it is sent and a response decrements it when the reader reads it; the reader side
// is slowed so that, on a transport the client fails to serialize, several requests
// pile up and the peak is visibly greater than one.
type serialModeEchoTransport struct {
	*echoTransport
	serial bool
	// atomic.Int64 carries its own alignment, so these are safe on 32-bit builds.
	inFlight  atomic.Int64
	maxFlight atomic.Int64
}

func newSerialModeEchoTransport(serial bool) *serialModeEchoTransport {
	return &serialModeEchoTransport{echoTransport: newEchoTransport(), serial: serial}
}

func (s *serialModeEchoTransport) SerializesCalls() bool { return s.serial }

func (s *serialModeEchoTransport) Send(p []byte) error {
	if hdr, err := pdu.PeekHeader(p); err == nil && hdr.PacketType == pdu.PacketTypeRequest {
		n := s.inFlight.Add(1)
		for {
			m := s.maxFlight.Load()
			if n <= m || s.maxFlight.CompareAndSwap(m, n) {
				break
			}
		}
	}
	return s.echoTransport.Send(p)
}

func (s *serialModeEchoTransport) Recv() ([]byte, error) {
	b, err := s.echoTransport.Recv()
	if err != nil {
		return b, err
	}
	if hdr, herr := pdu.PeekHeader(b); herr == nil && hdr.PacketType == pdu.PacketTypeResponse {
		// Widen the window a response spends "in flight": on a transport the client
		// does not serialize, other requests are sent meanwhile and the gauge climbs.
		time.Sleep(time.Millisecond)
		s.inFlight.Add(-1)
	}
	return b, nil
}

// TestSerialCallTransportRunsOneCallAtATime drives many concurrent Call goroutines
// over a transport that reports SerializesCalls() == true and asserts the client
// never has more than one call exchange in flight: callSerialMu must serialize the
// whole exchange (request burst plus the wait for the response). Run under -race it
// also guards the added lock.
func TestSerialCallTransportRunsOneCallAtATime(t *testing.T) {
	tr := newSerialModeEchoTransport(true)
	c := NewClient(tr)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if !c.serialCalls {
		t.Fatal("expected client to detect the serial-call transport")
	}

	const workers = 16
	var wg sync.WaitGroup
	var fail atomic.Bool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Call(0, []byte{0x01, 0x02, 0x03}); err != nil {
				fail.Store(true)
			}
		}()
	}
	wg.Wait()
	if fail.Load() {
		t.Fatal("a Call returned an error")
	}
	if got := tr.maxFlight.Load(); got != 1 {
		t.Fatalf("peak concurrent calls = %d over a serial-call transport, want 1", got)
	}
}

// TestConcurrentTransportAllowsOverlap is the control: the same harness on a
// transport that reports SerializesCalls() == false lets several calls be in flight
// at once, confirming the test actually measures overlap (and that serialization is
// not imposed on multiplexing transports).
func TestConcurrentTransportAllowsOverlap(t *testing.T) {
	tr := newSerialModeEchoTransport(false)
	c := NewClient(tr)
	if err := c.Bind(testSyntax()); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if c.serialCalls {
		t.Fatal("transport reporting false must leave the client concurrent")
	}

	const workers = 16
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Call(0, []byte{0x01})
		}()
	}
	wg.Wait()
	if got := tr.maxFlight.Load(); got < 2 {
		t.Fatalf("peak concurrent calls = %d over a multiplexing transport, want >= 2", got)
	}
}

// TestSerialModeFlagFollowsTransport confirms the client takes its serial-call
// behaviour from the transport.
func TestSerialModeFlagFollowsTransport(t *testing.T) {
	if !NewClient(newSerialModeEchoTransport(true)).serialCalls {
		t.Error("serial transport should set serialCalls")
	}
	if NewClient(newSerialModeEchoTransport(false)).serialCalls {
		t.Error("multiplexing transport should leave serialCalls false")
	}
	// A plain transport that does not implement SerialCaller stays concurrent.
	if NewClient(newEchoTransport()).serialCalls {
		t.Error("transport without SerialCaller should leave serialCalls false")
	}
}
