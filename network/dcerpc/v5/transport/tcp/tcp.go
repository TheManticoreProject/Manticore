// Package tcp implements the DCE/RPC ncacn_ip_tcp protocol sequence: DCE/RPC
// directly over a TCP connection.
//
// Unlike ncacn_np (network/dcerpc/v5/transport/smb), which carries PDUs as SMB named
// pipe writes and reads, ncacn_ip_tcp has no intermediate framing: each PDU is written
// straight to the socket and the receive side is driven by the common header's
// frag_length. This transport therefore implements Send as a single socket write of a
// complete PDU and Recv as a single socket read; reassembling fragments across reads is
// the DCE/RPC client's responsibility (its fragmentReader buffers by frag_length).
//
// The endpoint is an explicit host and port. The endpoint mapper (ept) on TCP port 135
// resolves the dynamic port a service listens on; see the epm interface
// (network/dcerpc/interfaces/...) for that lookup.
//
// References:
//   - [C706] chapter 12 (connection-oriented protocol) and Appendix I/L:
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.1.1.1 (RPC over TCP/IP, ncacn_ip_tcp):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/08c5dcde-be40-4d1a-aa8a-1c84acebabf0
package tcp

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// EndpointMapperPort is the well-known TCP port of the endpoint mapper (ept), used to
// resolve the dynamic port of a service before connecting to it.
const EndpointMapperPort = 135

// Default fragment sizes proposed at Bind time. 5840 (0x16D0) is the classic Windows
// default for RPC over TCP. Callers that want different fragments can override via
// SetMaxFrag before Connect.
const (
	DefaultMaxXmitFrag uint16 = 5840
	DefaultMaxRecvFrag uint16 = 5840
)

// socket is the subset of net.Conn the transport relies on. Expressing it as an
// interface keeps the transport unit-testable without a live server: tests substitute a
// net.Pipe end or a fake. *net.TCPConn (and any net.Conn) satisfies it.
type socket interface {
	io.ReadWriteCloser
	SetReadDeadline(t time.Time) error
}

// TCPTransport is a DCE/RPC transport over a TCP connection (ncacn_ip_tcp).
type TCPTransport struct {
	address string
	timeout time.Duration

	// dial opens the connection to address; it is overridable in tests.
	dial func(address string, timeout time.Duration) (socket, error)

	conn    socket
	maxXmit uint16
	maxRecv uint16
}

// Compile-time assertion that TCPTransport implements the DCE/RPC transport contract.
var _ dcerpctransport.Transport = (*TCPTransport)(nil)

// New creates a TCP transport that will connect to the given host and port. host may be
// an IPv4 address, IPv6 address, or hostname; the port is the explicit RPC endpoint
// (use EndpointMapperPort to talk to the endpoint mapper).
func New(host string, port int) *TCPTransport {
	return &TCPTransport{
		address: net.JoinHostPort(host, strconv.Itoa(port)),
		dial:    dialTCP,
		maxXmit: DefaultMaxXmitFrag,
		maxRecv: DefaultMaxRecvFrag,
	}
}

// dialTCP is the production dialer: a plain TCP connection bounded by timeout.
func dialTCP(address string, timeout time.Duration) (socket, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// SetTimeout bounds Connect and each subsequent Recv: Connect fails if the connection
// cannot be established within d, and Recv fails if no data arrives within d. A
// non-positive d removes the bound (blocking I/O). It must be called before Connect to
// affect the dial.
func (t *TCPTransport) SetTimeout(d time.Duration) {
	if d < 0 {
		d = 0
	}
	t.timeout = d
}

// SetMaxFrag overrides the transmit and receive fragment sizes proposed at Bind time.
// It must be called before Connect to take effect.
func (t *TCPTransport) SetMaxFrag(xmit, recv uint16) {
	t.maxXmit = xmit
	t.maxRecv = recv
}

// Connect opens the TCP connection. It is idempotent.
func (t *TCPTransport) Connect() error {
	if t.conn != nil {
		return nil
	}
	conn, err := t.dial(t.address, t.timeout)
	if err != nil {
		return fmt.Errorf("ncacn_ip_tcp: connect to %s: %w", t.address, err)
	}
	t.conn = conn
	return nil
}

// Send writes a complete PDU to the socket. A net.Conn write returns a non-nil error
// whenever fewer than len(pdu) bytes are written, so a single Write suffices.
func (t *TCPTransport) Send(pdu []byte) error {
	if t.conn == nil {
		return fmt.Errorf("ncacn_ip_tcp: not connected to %s: call Connect first", t.address)
	}
	if len(pdu) == 0 {
		return fmt.Errorf("ncacn_ip_tcp: refusing to send an empty PDU to %s", t.address)
	}
	n, err := t.conn.Write(pdu)
	if err != nil {
		return fmt.Errorf("ncacn_ip_tcp: write PDU to %s: %w", t.address, err)
	}
	if n != len(pdu) {
		return fmt.Errorf("ncacn_ip_tcp: short write to %s: wrote %d of %d bytes", t.address, n, len(pdu))
	}
	return nil
}

// Recv reads up to MaxRecvFrag bytes from the socket and returns them. The result may
// hold part of a PDU, exactly one PDU, or several PDUs; the caller reassembles by
// frag_length. A read that returns no data before the peer closes the connection is
// reported as an error.
func (t *TCPTransport) Recv() ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("ncacn_ip_tcp: not connected to %s: call Connect first", t.address)
	}

	var deadline time.Time
	if t.timeout > 0 {
		deadline = time.Now().Add(t.timeout)
	}
	if err := t.conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("ncacn_ip_tcp: set read deadline on %s: %w", t.address, err)
	}

	buf := make([]byte, t.maxRecv)
	for {
		n, err := t.conn.Read(buf)
		if n > 0 {
			return buf[:n], nil
		}
		if err != nil {
			return nil, fmt.Errorf("ncacn_ip_tcp: read from %s: %w", t.address, err)
		}
		// A compliant io.Reader returns (0, nil) only rarely; loop until data or error.
	}
}

// Close tears down the TCP connection. It is idempotent.
func (t *TCPTransport) Close() error {
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	if err != nil {
		return fmt.Errorf("ncacn_ip_tcp: close %s: %w", t.address, err)
	}
	return nil
}

// MaxXmitFrag returns the proposed maximum transmit fragment size.
func (t *TCPTransport) MaxXmitFrag() uint16 { return t.maxXmit }

// MaxRecvFrag returns the proposed maximum receive fragment size.
func (t *TCPTransport) MaxRecvFrag() uint16 { return t.maxRecv }
