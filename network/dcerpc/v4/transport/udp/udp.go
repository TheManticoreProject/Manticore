// Package udp implements the connectionless DCE/RPC datagram transport for the
// ncadg_ip_udp protocol sequence: RPC PDUs carried directly over UDP, with no
// intermediate protocol ([MS-RPCE] 2.1.1.1, [C706] Appendix I protocol identifier
// 0x08).
//
// The transport uses a connected UDP socket (net.DialUDP), so it only sends to and
// receives from the configured server. UDP preserves message boundaries, so each
// Send writes one PDU datagram and each Recv returns one PDU datagram. Reliability,
// fragmentation, and acknowledgement are handled by the protocol machine above this
// transport, not here.
//
// References:
//   - [MS-RPCE] 2.1.1.1 UDP (NCADG_IP_UDP):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/f3c9d073-1563-4d47-861a-14023ec4990e
package udp

import (
	"fmt"
	"net"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
)

// recvBufferSize is the size of the buffer used to read a single datagram. It is the
// maximum UDP payload so that any incoming PDU fits in one read; the maximum size of
// a PDU we will *send* is bounded separately by Transport.maxPDUSize.
const recvBufferSize = 65535

// DefaultTimeout is the read/write timeout applied to each operation when no explicit
// deadline has been configured via SetDeadline.
const DefaultTimeout = 10 * time.Second

// Transport is a connected-UDP implementation of transport.Transport for
// ncadg_ip_udp.
type Transport struct {
	host       string
	port       int
	maxPDUSize int
	timeout    time.Duration

	conn  *net.UDPConn
	raddr *net.UDPAddr
}

// compile-time assertion that *Transport satisfies the interface.
var _ transport.Transport = (*Transport)(nil)

// Option configures a Transport.
type Option func(*Transport)

// WithTimeout sets the per-operation read/write timeout. A non-positive duration
// disables the per-operation timeout, leaving deadline management entirely to
// SetDeadline.
func WithTimeout(d time.Duration) Option {
	return func(t *Transport) { t.timeout = d }
}

// WithMaxPDUSize sets the maximum datagram size, in bytes, that Send will transmit.
// Use transport.MaxPDUSizeLegacy when targeting Windows NT 4.0.
func WithMaxPDUSize(n int) Option {
	return func(t *Transport) { t.maxPDUSize = n }
}

// New returns an unconnected ncadg_ip_udp transport for the given host and UDP port.
// Call Connect before Send or Recv.
func New(host string, port int, opts ...Option) *Transport {
	t := &Transport{
		host:       host,
		port:       port,
		maxPDUSize: transport.MaxPDUSizeDefault,
		timeout:    DefaultTimeout,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Connect resolves the server address and opens the connected UDP socket. It is a
// no-op if already connected.
func (t *Transport) Connect() error {
	if t.conn != nil {
		return nil
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(t.host, fmt.Sprintf("%d", t.port)))
	if err != nil {
		return fmt.Errorf("ncadg_ip_udp: resolve %s:%d: %w", t.host, t.port, err)
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return fmt.Errorf("ncadg_ip_udp: dial %s: %w", raddr, err)
	}
	t.conn = conn
	t.raddr = raddr
	return nil
}

// Send writes datagram as a single UDP message. A datagram larger than MaxPDUSize is
// rejected, since fragmentation is the protocol machine's responsibility.
func (t *Transport) Send(datagram []byte) (int, error) {
	if t.conn == nil {
		return 0, fmt.Errorf("ncadg_ip_udp: send on closed transport")
	}
	if len(datagram) > t.maxPDUSize {
		return 0, fmt.Errorf("ncadg_ip_udp: datagram of %d bytes exceeds max PDU size %d", len(datagram), t.maxPDUSize)
	}
	if t.timeout > 0 {
		if err := t.conn.SetWriteDeadline(time.Now().Add(t.timeout)); err != nil {
			return 0, fmt.Errorf("ncadg_ip_udp: set write deadline: %w", err)
		}
	}
	n, err := t.conn.Write(datagram)
	if err != nil {
		return n, fmt.Errorf("ncadg_ip_udp: send: %w", err)
	}
	return n, nil
}

// Recv reads and returns a single datagram.
func (t *Transport) Recv() ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("ncadg_ip_udp: recv on closed transport")
	}
	if t.timeout > 0 {
		if err := t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
			return nil, fmt.Errorf("ncadg_ip_udp: set read deadline: %w", err)
		}
	}
	buf := make([]byte, recvBufferSize)
	n, err := t.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("ncadg_ip_udp: recv: %w", err)
	}
	return buf[:n], nil
}

// SetDeadline sets the absolute read/write deadline on the socket. When a per-
// operation timeout is configured (the default), Send and Recv override the read or
// write deadline on their next call; set the timeout to zero via WithTimeout to make
// SetDeadline authoritative.
func (t *Transport) SetDeadline(deadline time.Time) error {
	if t.conn == nil {
		return fmt.Errorf("ncadg_ip_udp: set deadline on closed transport")
	}
	return t.conn.SetDeadline(deadline)
}

// MaxPDUSize returns the maximum datagram size, in bytes, that Send will transmit.
func (t *Transport) MaxPDUSize() int { return t.maxPDUSize }

// RemoteAddr returns the resolved server address, or nil if not connected.
func (t *Transport) RemoteAddr() net.Addr {
	if t.raddr == nil {
		return nil
	}
	return t.raddr
}

// IsConnected reports whether the socket is open.
func (t *Transport) IsConnected() bool { return t.conn != nil }

// Close closes the UDP socket. It is safe to call on an already-closed transport.
func (t *Transport) Close() error {
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	return err
}
