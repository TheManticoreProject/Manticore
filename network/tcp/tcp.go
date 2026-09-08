package tcp

import (
	"fmt"
	"io"
	"net"
	"time"
)

// MaxDirectTCPPayloadSize is the default cap on the payload accepted from a single
// Direct TCP frame, and is the largest the session service can describe: the length
// field is 24 bits wide ([MS-SMB2] 2.1).
//
// The transport carries SMB1, SMB2 and SMB3, whose negotiated limits differ by more
// than an order of magnitude — a Windows server advertises an 8 MiB MaxReadSize —
// so a smaller default here would refuse frames the peer was entitled to send after
// a successful negotiation. Bounding what a peer can induce this side to allocate is
// a policy the accepting side sets with SetMaxPayloadSize, not something a constant
// shared with the client can express.
const MaxDirectTCPPayloadSize = 0xFFFFFF

// TCPTransport implements the Transport interface for Direct TCP transport
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/f906c680-330c-43ae-9a71-f854e24aeee6
type TCPTransport struct {
	conn    net.Conn
	timeout time.Duration

	// maxPayload caps the payload accepted from a single frame. Zero means
	// MaxDirectTCPPayloadSize.
	maxPayload uint32
}

// SetMaxPayloadSize caps the payload this transport will accept from a single
// Direct TCP frame, so a listener can bound the allocation an unauthenticated peer
// can induce. A value of 0, or one above what the 24-bit length field can describe,
// restores the MaxDirectTCPPayloadSize default.
func (t *TCPTransport) SetMaxPayloadSize(n uint32) {
	if n == 0 || n > MaxDirectTCPPayloadSize {
		n = MaxDirectTCPPayloadSize
	}
	t.maxPayload = n
}

// maxPayloadSize returns the cap in effect for this transport.
func (t *TCPTransport) maxPayloadSize() uint32 {
	if t.maxPayload == 0 {
		return MaxDirectTCPPayloadSize
	}
	return t.maxPayload
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

	if cap := t.maxPayloadSize(); uint32(length) > cap {
		return nil, fmt.Errorf("Direct TCP payload length %d exceeds maximum %d", length, cap)
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
