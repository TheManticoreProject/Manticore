package server

import (
	"net"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// TestRegisterHandlerConcurrentWithDispatchNoRace exercises RegisterHandler
// (which appends to Handlers under the write lock) concurrently with the
// per-packet dispatch path (which reads a snapshot under the read lock). Run
// under -race it proves the two paths no longer race on the Handlers slice.
func TestRegisterHandlerConcurrentWithDispatchNoRace(t *testing.T) {
	s := &Server{Closed: make(chan struct{})}

	// A no-op handler that keeps the chain going; it never touches the writer,
	// so a nil ResponseWriter is fine for the purposes of this race test.
	noop := HandlerFunc(func(*Server, net.Addr, ResponseWriter, *message.Message) bool {
		return true
	})

	const registrars = 32
	const dispatchers = 32
	const iterations = 200

	remote := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5355}
	msg := message.NewMessage()

	var wg sync.WaitGroup

	// Writers: keep appending handlers.
	for i := 0; i < registrars; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				s.RegisterHandler(noop)
			}
		}()
	}

	// Readers: keep dispatching, which snapshots the Handlers slice under lock.
	for i := 0; i < dispatchers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				s.processHandlers(s, remote, nil, msg)
			}
		}()
	}

	wg.Wait()

	// Every registration must have landed exactly once.
	if got := len(s.snapshotHandlers()); got != registrars*iterations {
		t.Errorf("registered handler count = %d, want %d", got, registrars*iterations)
	}
}

// TestWriteMessageNonUDPAddrReturnsError verifies the UDP ResponseWriter returns
// a descriptive error (instead of panicking) when its remote address is not a
// *net.UDPAddr, which is the checked-assertion hardening this test guards.
func TestWriteMessageNonUDPAddrReturnsError(t *testing.T) {
	// A non-UDP remote address; the assertion must fail cleanly rather than panic.
	w := &responseWriter{
		Server:     &Server{},
		RemoteAddr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5355},
	}

	err := w.WriteMessage(message.NewMessage())
	if err == nil {
		t.Fatal("WriteMessage() error = nil, want an error for a non-*net.UDPAddr remote address")
	}
}
