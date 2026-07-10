package nbt

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios"
	"github.com/TheManticoreProject/Manticore/network/netbios/nbns"
)

const (
	// DefaultCalledName is the wildcard NetBIOS server name that modern SMB
	// servers answer to regardless of their configured machine name.
	DefaultCalledName = "*SMBSERVER"

	// fallbackCallingName is the CALLING name used when the local hostname
	// cannot be determined.
	fallbackCallingName = "MANTICORE"

	// NetBIOSSuffixWorkstation and NetBIOSSuffixServer are the one-byte service
	// suffixes appended to the 16th byte of a NetBIOS name: the calling name is
	// the workstation service (0x00) and the called name is the server service
	// (0x20).
	NetBIOSSuffixWorkstation byte = 0x00
	NetBIOSSuffixServer      byte = 0x20

	// maxRetargets bounds the RETARGET re-dial chain so a misbehaving server
	// cannot force an unbounded sequence of reconnects.
	maxRetargets = 5
)

// NBTTransport implements the Transport interface for NetBIOS over TCP
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/45170055-a0cd-4910-9228-801d5bf7ac84
type NBTTransport struct {
	conn    net.Conn
	timeout time.Duration

	// handshake enables the RFC 1002 4.3 session-establishment handshake on
	// Connect; it is on by default so the port-139 path stands up a NetBIOS
	// session before the first SESSION MESSAGE.
	handshake bool
	// calledName and callingName are the CALLED and CALLING NetBIOS names sent
	// in the SESSION REQUEST.
	calledName  string
	callingName string
}

// NewNBTTransport creates a new NetBIOS over TCP transport with the RFC 1002
// session-establishment handshake enabled and the CALLED name defaulting to the
// "*SMBSERVER" wildcard.
func NewNBTTransport() *NBTTransport {
	return &NBTTransport{
		handshake:   true,
		calledName:  DefaultCalledName,
		callingName: defaultCallingName(),
	}
}

// SetHandshakeEnabled controls whether Connect performs the RFC 1002
// session-establishment handshake. Disable it to talk to a modern server that
// accepts SESSION MESSAGEs on port 139 without a prior SESSION REQUEST.
func (n *NBTTransport) SetHandshakeEnabled(enabled bool) {
	n.handshake = enabled
}

// SetCalledName overrides the CALLED NetBIOS name sent in the SESSION REQUEST
// (defaults to "*SMBSERVER"). An empty value is ignored.
func (n *NBTTransport) SetCalledName(name string) {
	if name != "" {
		n.calledName = name
	}
}

// SetCallingName overrides the CALLING NetBIOS name sent in the SESSION REQUEST
// (defaults to the local hostname). An empty value is ignored.
func (n *NBTTransport) SetCallingName(name string) {
	if name != "" {
		n.callingName = name
	}
}

// Connect establishes a NetBIOS over TCP connection and, unless the handshake
// has been disabled, performs the RFC 1002 session-establishment handshake.
func (n *NBTTransport) Connect(ipaddr net.IP, port int) error {
	// Default NetBIOS port is 139 if not specified
	if port == 0 {
		port = 139
	}

	if err := n.dial(ipaddr, port); err != nil {
		return err
	}

	if n.handshake {
		if err := n.EstablishSession(n.calledName, n.callingName); err != nil {
			n.Close()
			return err
		}
	}

	return nil
}

// dial opens the raw TCP connection to ipaddr:port, replacing any existing
// connection. It is shared by Connect and the RETARGET re-dial path.
func (n *NBTTransport) dial(ipaddr net.IP, port int) error {
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

// EstablishSession performs the RFC 1002 4.3 session-establishment handshake on
// the current connection: it sends a SESSION REQUEST (0x81) carrying the
// second-level-encoded CALLED and CALLING NetBIOS names and interprets the
// reply. A POSITIVE SESSION RESPONSE (0x82) completes the handshake; a NEGATIVE
// SESSION RESPONSE (0x83) is returned as the mapped error-code error; and a
// RETARGET SESSION RESPONSE (0x84) re-dials the advertised IP/port and retries,
// up to maxRetargets times.
func (n *NBTTransport) EstablishSession(calledName, callingName string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected")
	}

	triedDiscovery := false
	for retargets := 0; ; {
		request, err := buildSessionRequest(calledName, callingName)
		if err != nil {
			return err
		}

		if _, err := n.conn.Write(request); err != nil {
			return fmt.Errorf("failed to send SESSION REQUEST: %v", err)
		}

		messageType, body, err := n.readSessionResponse()
		if err != nil {
			return fmt.Errorf("failed to read SESSION RESPONSE: %v", err)
		}

		switch netbios.SESSION_MESSAGE_TYPE(messageType) {
		case netbios.SESSION_POSITIVE_RESPONSE:
			return nil

		case netbios.SESSION_NEGATIVE_RESPONSE:
			if len(body) < 1 {
				return fmt.Errorf("malformed NEGATIVE SESSION RESPONSE: missing error code")
			}
			code := netbios.NEGATIVE_SESSION_ERROR(body[0])
			// A server that does not answer to the "*SMBSERVER" wildcard often
			// still listens on its real machine name. When the wildcard is
			// refused, resolve the host's File Server Service (<20>) name via a
			// NODE STATUS query and retry once against it.
			if !triedDiscovery && isWildcardCalledName(calledName) &&
				(code == netbios.NEGATIVE_SESSION_NOT_LISTENING_ON_CALLED_NAME ||
					code == netbios.NEGATIVE_SESSION_CALLED_NAME_NOT_PRESENT) {
				triedDiscovery = true
				// The server closes the TCP connection after a NEGATIVE
				// response, so capture the endpoint and reconnect before
				// retrying against the discovered name.
				if remote, ok := n.conn.RemoteAddr().(*net.TCPAddr); ok {
					if discovered, derr := n.discoverServerName(); derr == nil && discovered != "" {
						n.Close()
						if derr := n.dial(remote.IP, remote.Port); derr == nil {
							calledName = discovered
							continue
						}
					}
				}
			}
			return code

		case netbios.SESSION_RETARGET_RESPONSE:
			retargets++
			if retargets > maxRetargets {
				return fmt.Errorf("NetBIOS session establishment exceeded %d retargets", maxRetargets)
			}
			ip, port, err := parseRetarget(body)
			if err != nil {
				return err
			}
			// Re-dial the advertised endpoint and retry the SESSION REQUEST.
			n.Close()
			if err := n.dial(ip, port); err != nil {
				return fmt.Errorf("failed to re-dial after RETARGET to %s:%d: %v", ip.String(), port, err)
			}
			// The new endpoint may again need name discovery.
			triedDiscovery = false

		default:
			return fmt.Errorf("unexpected NetBIOS session response type 0x%02X (%s)", messageType, netbios.SESSION_MESSAGE_TYPE(messageType))
		}
	}
}

// isWildcardCalledName reports whether name is a wildcard called name (e.g. the
// "*SMBSERVER" convention), i.e. one a server may legitimately not be listening
// on even though it accepts SMB on its real machine name.
func isWildcardCalledName(name string) bool {
	return strings.HasPrefix(name, "*")
}

// discoverServerName issues a NODE STATUS query (RFC 1002 4.2.17, over UDP 137)
// to the currently connected host and returns its unique File Server Service
// (<20>) NetBIOS name, so a refused "*SMBSERVER" wildcard can be retried against
// the host's real name. It is best-effort: any failure (including UDP 137 being
// filtered) leaves the original NEGATIVE response to stand.
func (n *NBTTransport) discoverServerName() (string, error) {
	host, _, err := net.SplitHostPort(n.conn.RemoteAddr().String())
	if err != nil {
		return "", err
	}

	client := nbns.NewClient()
	if n.timeout > 0 {
		client.Timeout = n.timeout
	}

	result, err := client.NodeStatus(host)
	if err != nil {
		return "", err
	}
	for _, name := range result.Names {
		if name.Suffix == NetBIOSSuffixServer && !name.IsGroup() {
			return name.Name, nil
		}
	}
	return "", fmt.Errorf("no File Server Service name in NODE STATUS response")
}

// buildSessionRequest assembles an RFC 1002 4.3.2 SESSION REQUEST: a 4-byte
// session header (TYPE 0x81, FLAGS 0, LENGTH 68 for the two 34-byte encoded
// names) followed by the second-level-encoded CALLED name (server suffix 0x20)
// and CALLING name (workstation suffix 0x00).
func buildSessionRequest(calledName, callingName string) ([]byte, error) {
	called, err := nbns.EncodeSessionServiceName(calledName, NetBIOSSuffixServer)
	if err != nil {
		return nil, fmt.Errorf("failed to encode called name %q: %v", calledName, err)
	}
	calling, err := nbns.EncodeSessionServiceName(callingName, NetBIOSSuffixWorkstation)
	if err != nil {
		return nil, fmt.Errorf("failed to encode calling name %q: %v", callingName, err)
	}

	length := len(called) + len(calling) // 34 + 34 = 68

	packet := make([]byte, 0, 4+length)
	packet = append(packet, byte(netbios.SESSION_REQUEST)) // TYPE
	packet = append(packet, 0x00)                          // FLAGS (length fits in 16 bits)
	packet = append(packet, byte((length>>8)&0xFF), byte(length&0xFF))
	packet = append(packet, called...)
	packet = append(packet, calling...)

	return packet, nil
}

// parseRetarget decodes the 6-byte RETARGET SESSION RESPONSE body (RFC 1002
// 4.3.5): a 4-byte IPv4 address followed by a 2-byte TCP port, both big-endian.
func parseRetarget(body []byte) (net.IP, int, error) {
	if len(body) < 6 {
		return nil, 0, fmt.Errorf("malformed RETARGET SESSION RESPONSE: need 6 bytes, got %d", len(body))
	}
	ip := net.IPv4(body[0], body[1], body[2], body[3])
	port := int(binary.BigEndian.Uint16(body[4:6]))
	return ip, port, nil
}

// defaultCallingName derives this client's CALLING NetBIOS name from the local
// hostname: the first label, uppercased and truncated to the 15-character
// NetBIOS limit. It falls back to a fixed name when the hostname is unavailable.
func defaultCallingName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return fallbackCallingName
	}
	if i := strings.IndexByte(host, '.'); i >= 0 {
		host = host[:i]
	}
	host = strings.ToUpper(host)
	if len(host) > 15 {
		host = host[:15]
	}
	if host == "" {
		return fallbackCallingName
	}
	return host
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

	// RFC 1002 4.3.1: the session header is TYPE(8) | FLAGS(8) | LENGTH(16).
	// Bit 0 of the FLAGS byte (E) is the length-extension bit and acts as the
	// high-order 17th bit of LENGTH, so the payload may range from 0 to 131071.
	length := len(data)
	if length > 0x1FFFF {
		return 0, fmt.Errorf("NetBIOS session message too large: %d bytes (max %d)", length, 0x1FFFF)
	}

	// Create NetBIOS header
	header := []byte{}
	// Set message type to Session Message (0x00)
	header = append(header, byte(netbios.SESSION_MESSAGE))
	// Set the FLAGS byte: bit 0 carries the 17th (high-order) length bit
	header = append(header, byte((length>>16)&0x01))
	// Set the low 16 bits of the length in big-endian format
	header = append(header, byte((length>>8)&0xFF))
	header = append(header, byte(length&0xFF))

	newPacket := append(header, data...)

	// Send data
	return n.conn.Write(newPacket)
}

// Receive reads a SESSION MESSAGE from the NetBIOS over TCP connection, handling
// the NetBIOS header, and rejects any other session-service message type.
func (n *NBTTransport) Receive() ([]byte, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	messageType, body, err := n.readFrame()
	if err != nil {
		return nil, err
	}

	// Verify message type is Session Message (0x00)
	if messageType != byte(netbios.SESSION_MESSAGE) {
		return nil, fmt.Errorf("unexpected NetBIOS message type: %d", messageType)
	}

	return body, nil
}

// readSessionResponse reads one NetBIOS session-service response, transparently
// skipping any SESSION KEEP ALIVE (0x85) frames, and returns the message type
// byte together with the frame body.
func (n *NBTTransport) readSessionResponse() (byte, []byte, error) {
	for {
		messageType, body, err := n.readFrame()
		if err != nil {
			return 0, nil, err
		}
		if netbios.SESSION_MESSAGE_TYPE(messageType) == netbios.SESSION_KEEP_ALIVE {
			continue
		}
		return messageType, body, nil
	}
}

// readFrame reads a single NetBIOS session-service frame (4-byte header plus
// body) subject to the configured read deadline and returns the message type
// byte and the body. RFC 1002 4.3.1: bit 0 of the FLAGS byte (header[1]) is the
// length-extension bit, forming the 17th (high-order) bit of the LENGTH field
// carried in header[2..3].
func (n *NBTTransport) readFrame() (byte, []byte, error) {
	// Apply the configured read deadline (a zero deadline blocks forever)
	var deadline time.Time
	if n.timeout > 0 {
		deadline = time.Now().Add(n.timeout)
	}
	if err := n.conn.SetReadDeadline(deadline); err != nil {
		return 0, nil, fmt.Errorf("failed to set read deadline: %v", err)
	}

	// Read NetBIOS header
	header := make([]byte, 4)
	if _, err := io.ReadFull(n.conn, header); err != nil {
		return 0, nil, fmt.Errorf("failed to read NetBIOS header: %v", err)
	}

	messageType := header[0]
	length := (int(header[1]&0x01) << 16) | (int(header[2]) << 8) | int(header[3])

	body := make([]byte, length)
	if _, err := io.ReadFull(n.conn, body); err != nil {
		return 0, nil, fmt.Errorf("failed to read NetBIOS data: %v", err)
	}

	return messageType, body, nil
}

// IsConnected returns whether the NetBIOS transport is currently connected
func (n *NBTTransport) IsConnected() bool {
	return n.conn != nil
}
