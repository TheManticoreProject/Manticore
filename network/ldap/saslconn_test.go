package ldap

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

// xorCipher is a reversible stand-in for the GSS-API per-message layer: it lets
// the SASL framing (length prefixing, send-side chunking against maxSend, and
// receive-side reassembly across reads) be exercised without a Kerberos context.
type xorCipher struct{ max int }

func (c xorCipher) transform(b []byte) []byte {
	out := make([]byte, len(b))
	for i := range b {
		out[i] = b[i] ^ 0x5A
	}
	return out
}
func (c xorCipher) wrap(pt []byte) ([]byte, error)    { return c.transform(pt), nil }
func (c xorCipher) unwrap(buf []byte) ([]byte, error) { return c.transform(buf), nil }
func (c xorCipher) maxSend() int                      { return c.max }

func TestSASLConnFramedRoundtrip(t *testing.T) {
	// A max buffer just above the wrap overhead forces the send side to split a
	// large LDAP PDU into several SASL buffers; the receive side must reassemble.
	cipher := xorCipher{max: gssWrapOverhead + 72}
	clientRaw, serverRaw := net.Pipe()

	client := newSASLConn(clientRaw)
	client.activate(cipher)
	server := newSASLConn(serverRaw)
	server.activate(cipher)

	pdu := bytes.Repeat([]byte("LDAP-search-request-pdu;"), 40) // ~960 bytes

	go func() {
		if _, err := client.Write(pdu); err != nil {
			t.Errorf("client write: %v", err)
		}
	}()

	got := make([]byte, 0, len(pdu))
	buf := make([]byte, 100) // small reads exercise plaintext spanning Read calls
	for len(got) < len(pdu) {
		n, err := server.Read(buf)
		if err != nil {
			t.Fatalf("server read: %v", err)
		}
		got = append(got, buf[:n]...)
	}
	if !bytes.Equal(got, pdu) {
		t.Errorf("reassembled PDU mismatch: got %d bytes, want %d", len(got), len(pdu))
	}
}

func TestSASLConnWireIsWrapped(t *testing.T) {
	// The bytes on the wire must be length-prefixed and transformed (not the
	// plaintext), confirming the security layer actually engaged.
	cipher := xorCipher{max: 0} // 0 = no send limit: one buffer
	clientRaw, wireRaw := net.Pipe()
	client := newSASLConn(clientRaw)
	client.activate(cipher)

	pdu := []byte("plaintext-ldap-pdu")
	go func() {
		_, _ = client.Write(pdu)
	}()

	var hdr [4]byte
	if _, err := io.ReadFull(wireRaw, hdr[:]); err != nil {
		t.Fatalf("read length prefix: %v", err)
	}
	n := binary.BigEndian.Uint32(hdr[:])
	body := make([]byte, n)
	if _, err := io.ReadFull(wireRaw, body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if int(n) != len(pdu) {
		t.Errorf("buffer length = %d, want %d", n, len(pdu))
	}
	if bytes.Equal(body, pdu) {
		t.Error("wire buffer carries plaintext; security layer did not wrap it")
	}
	if !bytes.Equal(cipher.transform(body), pdu) {
		t.Error("wire buffer does not unwrap back to the plaintext PDU")
	}
}

func TestSASLConnPassThroughBeforeActivate(t *testing.T) {
	clientRaw, serverRaw := net.Pipe()
	client := newSASLConn(clientRaw)

	msg := []byte("cleartext-bind-exchange")
	go func() {
		_, _ = client.Write(msg)
	}()

	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(serverRaw, buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(buf, msg) {
		t.Errorf("pass-through altered bytes: got %q, want %q", buf, msg)
	}
}

func TestSASLConnActivateNudgesBlockedRead(t *testing.T) {
	// A read blocked in pass-through mode must hand off to the framed path once
	// activate flips the flag, without surfacing the deadline as an error.
	cipher := xorCipher{max: 0}
	localRaw, remoteRaw := net.Pipe()
	local := newSASLConn(localRaw)

	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		buf := make([]byte, 64)
		n, err := local.Read(buf) // blocks in pass-through until activate + a frame
		done <- result{append([]byte{}, buf[:n]...), err}
	}()

	// Let the reader block, then activate (which nudges it via a read deadline).
	time.Sleep(50 * time.Millisecond)
	local.activate(cipher)

	// Now send a framed, wrapped buffer from the peer.
	pdu := []byte("post-activation-pdu")
	go func() {
		var hdr [4]byte
		wrapped := cipher.transform(pdu)
		binary.BigEndian.PutUint32(hdr[:], uint32(len(wrapped)))
		_, _ = remoteRaw.Write(append(hdr[:], wrapped...))
	}()

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("read after activation errored: %v", r.err)
		}
		if !bytes.Equal(r.data, pdu) {
			t.Errorf("read after activation = %q, want %q", r.data, pdu)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("read did not complete after activation")
	}
}
