package nbt_test

import (
	"bytes"
	"testing"
)

// TestNBTTransport_ReceiveSkipsKeepAlive verifies that Receive transparently
// discards SESSION KEEP ALIVE (0x85) frames, which RFC 1002 4.3.7 allows either
// end to send at any time, and returns the SESSION MESSAGE that follows.
func TestNBTTransport_ReceiveSkipsKeepAlive(t *testing.T) {
	tr, srv, cleanup := connectedPair(t)
	defer cleanup()

	go func() {
		srv.Write([]byte{0x85, 0x00, 0x00, 0x00})
		srv.Write([]byte{0x85, 0x00, 0x00, 0x00})
		srv.Write([]byte{0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'})
	}()

	got, err := tr.Receive()
	if err != nil {
		t.Fatalf("NBTTransport.Receive() error = %v", err)
	}
	if !bytes.Equal(got, []byte("abc")) {
		t.Fatalf("Receive() = %q, want %q", got, "abc")
	}
}

// TestNBTTransport_ReceiveRejectsNonSessionMessage verifies that a
// session-service frame other than SESSION MESSAGE or SESSION KEEP ALIVE is
// still reported as an error.
func TestNBTTransport_ReceiveRejectsNonSessionMessage(t *testing.T) {
	tr, srv, cleanup := connectedPair(t)
	defer cleanup()

	go func() {
		srv.Write([]byte{0x82, 0x00, 0x00, 0x00}) // POSITIVE SESSION RESPONSE
	}()

	if _, err := tr.Receive(); err == nil {
		t.Fatal("Receive() accepted a POSITIVE SESSION RESPONSE frame")
	}
}
