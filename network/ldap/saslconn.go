package ldap

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
)

// gssWrapOverhead is a generous upper bound on the number of bytes a GSS-API
// per-message (CFX/RC4) Wrap token adds to its plaintext: the 16-byte token
// header, a cipher-block confounder, the trailing encrypted header, and the
// checksum. Subtracting it from the peer's advertised maximum buffer size yields
// a safe plaintext chunk size that never produces an over-long SASL buffer.
const gssWrapOverhead = 128

// saslConn wraps a network connection to apply the RFC 4752 SASL GSSAPI security
// layer once it has been negotiated. Before activation it is a transparent
// pass-through, so the plaintext SASL bind exchange flows unmodified. After
// activate is called, every write is GSS-wrapped into a length-prefixed SASL
// buffer and every read reassembles and GSS-unwraps such buffers, giving the LDAP
// layer above a continuous plaintext byte stream (RFC 4752 §3.1 / RFC 2222 §3).
type saslConn struct {
	net.Conn

	active atomic.Bool
	cipher saslCipher

	// readBuf holds plaintext recovered from the last unwrapped SASL buffer that
	// has not yet been consumed by a Read.
	readBuf []byte
}

// saslCipher wraps and unwraps the payload of a single SASL security-layer buffer.
type saslCipher interface {
	// wrap turns a plaintext chunk into the octets of one SASL buffer.
	wrap(plaintext []byte) ([]byte, error)
	// unwrap recovers the plaintext carried by one SASL buffer.
	unwrap(buffer []byte) ([]byte, error)
	// maxSend is the largest SASL buffer the peer will accept, or 0 for no limit.
	maxSend() int
}

// gssSASLCipher is the production saslCipher: it wraps/unwraps SASL buffers with
// GSS-API per-message tokens over an established Kerberos security context, using
// confidentiality (seal) for the RFC 4752 confidentiality layer and integrity
// only for the integrity layer.
type gssSASLCipher struct {
	ctx  *gssapi.SecContext
	seal bool
	max  int
}

func (c *gssSASLCipher) wrap(plaintext []byte) ([]byte, error) { return c.ctx.Wrap(plaintext, c.seal) }

func (c *gssSASLCipher) unwrap(buffer []byte) ([]byte, error) {
	pt, _, err := c.ctx.Unwrap(buffer)
	return pt, err
}

func (c *gssSASLCipher) maxSend() int { return c.max }

// newSASLConn wraps conn as a pass-through; call activate to switch on the
// security layer once it has been negotiated during the bind.
func newSASLConn(conn net.Conn) *saslConn {
	return &saslConn{Conn: conn}
}

// activate switches the connection into SASL security-layer mode. It is called
// after the GSSAPI bind has negotiated a layer, while the LDAP reader goroutine is
// blocked in a pass-through Read; a zero read deadline nudges that blocked read so
// it re-checks the active flag and switches to framed reads without disturbing the
// LDAP layer.
func (c *saslConn) activate(cipher saslCipher) {
	c.cipher = cipher
	c.active.Store(true)
	// Interrupt any read currently blocked in pass-through mode so it notices the
	// flag flip and resumes as a framed read.
	_ = c.Conn.SetReadDeadline(time.Now())
}

// Write applies the SASL security layer to outgoing LDAP data once active,
// splitting p into chunks small enough that each wrapped buffer fits the peer's
// advertised maximum, and framing each as a 4-octet big-endian length followed by
// the GSS-wrapped buffer. Before activation it writes through unchanged.
func (c *saslConn) Write(p []byte) (int, error) {
	if !c.active.Load() {
		return c.Conn.Write(p)
	}

	chunk := len(p)
	if max := c.cipher.maxSend(); max > gssWrapOverhead && max-gssWrapOverhead < chunk {
		chunk = max - gssWrapOverhead
	}

	total := 0
	for total < len(p) {
		end := total + chunk
		if end > len(p) {
			end = len(p)
		}
		buf, err := c.cipher.wrap(p[total:end])
		if err != nil {
			return total, fmt.Errorf("ldap sasl: wrap: %w", err)
		}
		var hdr [4]byte
		binary.BigEndian.PutUint32(hdr[:], uint32(len(buf)))
		if _, err := c.Conn.Write(append(hdr[:], buf...)); err != nil {
			return total, err
		}
		total = end
	}
	return total, nil
}

// Read reassembles the SASL security layer for incoming LDAP data once active,
// reading one length-prefixed buffer at a time, GSS-unwrapping it, and returning
// the recovered plaintext across as many Read calls as the caller needs. Before
// activation it reads through unchanged; the pass-through path tolerates the read
// deadline that activate uses to hand off cleanly.
func (c *saslConn) Read(p []byte) (int, error) {
	for {
		if c.active.Load() {
			return c.framedRead(p)
		}
		n, err := c.Conn.Read(p)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				// The deadline set by activate fired: clear it and re-check active.
				_ = c.Conn.SetReadDeadline(time.Time{})
				continue
			}
			return n, err
		}
		return n, nil
	}
}

// framedRead serves plaintext from the current SASL buffer, reading and
// unwrapping the next buffer from the wire when the pending plaintext is empty.
func (c *saslConn) framedRead(p []byte) (int, error) {
	if len(c.readBuf) == 0 {
		// Clear any lingering deadline from the activation nudge before blocking.
		_ = c.Conn.SetReadDeadline(time.Time{})

		var hdr [4]byte
		if _, err := io.ReadFull(c.Conn, hdr[:]); err != nil {
			return 0, err
		}
		n := binary.BigEndian.Uint32(hdr[:])
		if n == 0 {
			return 0, nil
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(c.Conn, buf); err != nil {
			return 0, err
		}
		pt, err := c.cipher.unwrap(buf)
		if err != nil {
			return 0, fmt.Errorf("ldap sasl: unwrap: %w", err)
		}
		c.readBuf = pt
	}

	n := copy(p, c.readBuf)
	c.readBuf = c.readBuf[n:]
	return n, nil
}
