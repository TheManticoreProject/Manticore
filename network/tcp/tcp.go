package tcp

import (
	"fmt"
	"io"
	"net"
	"time"
)

// MaxDirectTCPPayloadSize caps the payload size accepted from a single Direct TCP frame.
// The Direct TCP session service length field is 24 bits wide (up to ~16 MiB), but SMB1
// negotiates a MaxBufferSize of at most a few tens of kilobytes in practice. 1 MiB is a
// generous upper bound that rejects DoS-scale frames without blocking legitimate traffic.
const MaxDirectTCPPayloadSize = 1 * 1024 * 1024

// TCPTransport implements the Transport interface for Direct TCP transport
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/f906c680-330c-43ae-9a71-f854e24aeee6
type TCPTransport struct {
	conn    net.Conn
	timeout time.Duration
}

// NewTCPTransport creates a new Direct TCP transport
func NewTCPTransport() *TCPTransport {
	return &TCPTransport{}
}

// NewTCPTransportFromConn wraps an already-established connection — typically one
// returned by net.Listener.Accept — as a Direct TCP transport, so the server side
// of a connection uses the same framing as the client side. Connect MUST NOT be
// called on the result: the transport is connected from the outset, and calling
// Connect would dial a second connection and leak conn.
func NewTCPTransportFromConn(conn net.Conn) *TCPTransport {
	return &TCPTransport{conn: conn}
}

// Connect establishes a Direct TCP connection
func (t *TCPTransport) Connect(ipaddr net.IP, port int) error {
	// Default SMB port is 445 if not specified
	if port == 0 {
		port = 445
	}
	// Handle both IPv4 and IPv6 addresses
	var address string
	if ipaddr.To4() != nil {
		// IPv4 address
		address = fmt.Sprintf("%s:%d", ipaddr.String(), port)
	} else {
		// IPv6 address - needs square brackets
		address = fmt.Sprintf("[%s]:%d", ipaddr.String(), port)
	}

	conn, err := net.DialTimeout("tcp", address, t.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect via TCP: %v", err)
	}
	t.conn = conn

	return nil
}

// SetTimeout bounds Connect and each subsequent Receive: Connect fails if the
// TCP connection cannot be established within d, and Receive fails if a frame
// does not arrive within d. A non-positive d removes the bound (blocking I/O).
func (t *TCPTransport) SetTimeout(d time.Duration) {
	if d < 0 {
		d = 0
	}
	t.timeout = d
}

// Close terminates the Direct TCP connection
func (t *TCPTransport) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// Send transmits data over the Direct TCP connection with proper Direct TCP header
func (t *TCPTransport) Send(data []byte) (int, error) {
	if !t.IsConnected() {
		return 0, fmt.Errorf("not connected")
	}

	// Create Direct TCP header
	header := []byte{0x00} // First byte must be 0
	// Set length in big-endian format (3 bytes)
	length := len(data)
	header = append(header, byte((length>>16)&0xFF))
	header = append(header, byte((length>>8)&0xFF))
	header = append(header, byte(length&0xFF))

	packet := append(header, data...)

	// Send data
	return t.conn.Write(packet)
}

// Receive reads data from the Direct TCP connection, handling the Direct TCP header
func (t *TCPTransport) Receive() ([]byte, error) {
	if !t.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	// Apply the configured read deadline (a zero deadline blocks forever)
	var deadline time.Time
	if t.timeout > 0 {
		deadline = time.Now().Add(t.timeout)
	}
	if err := t.conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %v", err)
	}

	// Read Direct TCP header (4 bytes)
	header := make([]byte, 4)
	_, err := io.ReadFull(t.conn, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read Direct TCP header: %v", err)
	}

	// Verify first byte is 0x00
	if header[0] != 0x00 {
		return nil, fmt.Errorf("invalid Direct TCP header: first byte must be 0x00, got 0x%02x", header[0])
	}

	// Parse length from 3 bytes
	length := (int(header[1]) << 16) | (int(header[2]) << 8) | int(header[3])

	if length > MaxDirectTCPPayloadSize {
		return nil, fmt.Errorf("Direct TCP payload length %d exceeds maximum %d", length, MaxDirectTCPPayloadSize)
	}

	buffer := make([]byte, length)

	// Read the actual data
	_, err = io.ReadFull(t.conn, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read Direct TCP data: %v", err)
	}

	return buffer, nil
}

// IsConnected returns whether the Direct TCP transport is currently connected
func (t *TCPTransport) IsConnected() bool {
	return t.conn != nil
}
