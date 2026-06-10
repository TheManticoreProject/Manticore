package nbt

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios"
)

// NBTTransport implements the Transport interface for NetBIOS over TCP
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/45170055-a0cd-4910-9228-801d5bf7ac84
type NBTTransport struct {
	conn    net.Conn
	timeout time.Duration
}

// NewNBTTransport creates a new NetBIOS over TCP transport
func NewNBTTransport() *NBTTransport {
	return &NBTTransport{}
}

// Connect establishes a NetBIOS over TCP connection
func (n *NBTTransport) Connect(ipaddr net.IP, port int) error {
	// Default NetBIOS port is 139 if not specified
	if port == 0 {
		port = 139
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

	conn, err := net.DialTimeout("tcp", address, n.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect via TCP: %v", err)
	}
	n.conn = conn

	return nil
}

// SetTimeout bounds Connect and each subsequent Receive: Connect fails if the
// TCP connection cannot be established within d, and Receive fails if a frame
// does not arrive within d. A non-positive d removes the bound (blocking I/O).
func (n *NBTTransport) SetTimeout(d time.Duration) {
	if d < 0 {
		d = 0
	}
	n.timeout = d
}

// Close terminates the NetBIOS over TCP connection
func (n *NBTTransport) Close() error {
	if n.conn != nil {
		return n.conn.Close()
	}
	return nil
}

// Send transmits data over the NetBIOS over TCP connection with proper NetBIOS header
func (n *NBTTransport) Send(data []byte) (int, error) {
	if !n.IsConnected() {
		return 0, fmt.Errorf("not connected")
	}

	// Create NetBIOS header
	header := []byte{}
	// Set message type to Session Message (0x00)
	header = append(header, byte(netbios.SESSION_MESSAGE))
	// Set length in big-endian format
	length := len(data)
	header = append(header, 0x00)
	header = append(header, byte((length>>8)&0xFF))
	header = append(header, byte(length&0xFF))

	newPacket := append(header, data...)

	// Send data
	return n.conn.Write(newPacket)
}

// Receive reads data from the NetBIOS over TCP connection, handling the NetBIOS header
func (n *NBTTransport) Receive() ([]byte, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	// Apply the configured read deadline (a zero deadline blocks forever)
	var deadline time.Time
	if n.timeout > 0 {
		deadline = time.Now().Add(n.timeout)
	}
	if err := n.conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %v", err)
	}

	// Read NetBIOS header
	header := make([]byte, 4)
	_, err := io.ReadFull(n.conn, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read NetBIOS header: %v", err)
	}

	// Parse message type and length
	messageType := header[0]
	length := (int(header[2]) << 8) | int(header[3])

	// Verify message type is Session Message (0x00)
	if messageType != 0x00 {
		return nil, fmt.Errorf("unexpected NetBIOS message type: %d", messageType)
	}

	buffer := make([]byte, length)

	// Ensure buffer is large enough
	if len(buffer) < length {
		return nil, fmt.Errorf("buffer too small for message of length %d", length)
	}

	// Read the actual data
	_, err = io.ReadFull(n.conn, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read NetBIOS data: %v", err)
	}

	return buffer, nil
}

// IsConnected returns whether the NetBIOS transport is currently connected
func (n *NBTTransport) IsConnected() bool {
	return n.conn != nil
}
