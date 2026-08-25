package transport

import (
	"fmt"
	"net"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
	"github.com/TheManticoreProject/Manticore/network/tcp"
)

// Default SMB listening ports: 445 carries SMB directly over TCP, while 139
// carries it inside a NetBIOS session (RFC 1002).
const (
	DefaultDirectTCPPort = "445"
	DefaultNBTPort       = "139"
)

// Listener is the accept side of the SMB transport layer, the counterpart of
// Transport's Connect. Accept yields a Transport that is already connected and
// whose transport-level handshake, if the transport has one, is already
// complete, so a server above it can treat both transports identically.
type Listener interface {
	// Accept blocks until a client connects and returns it as a Transport
	// together with its remote address. A connection whose transport-level
	// handshake fails is discarded rather than returned, so Accept only ever
	// yields a Transport ready to carry SMB messages.
	Accept() (Transport, net.Addr, error)

	// Close stops listening. It does not close the connections already handed
	// out by Accept; those are owned by their callers.
	Close() error

	// Addr returns the address the listener is bound to, which is how a caller
	// discovers the port when it asked for port 0.
	Addr() net.Addr
}

// ListenTCP listens for Direct TCP SMB connections on addr. An addr with no port
// (including the empty string) is bound on DefaultDirectTCPPort.
//
// Parameters:
//   - addr: the address to bind, e.g. "0.0.0.0:445", ":445", "" or "127.0.0.1:0"
//
// Returns:
//   - A Listener yielding Direct TCP transports
//   - An error if the address cannot be bound
func ListenTCP(addr string) (Listener, error) {
	ln, err := net.Listen("tcp", withDefaultPort(addr, DefaultDirectTCPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen for Direct TCP SMB connections: %v", err)
	}
	return &tcpListener{listener: ln}, nil
}

// ListenNBT listens for NetBIOS-over-TCP SMB connections on addr, completing the
// RFC 1002 4.3 session handshake on each accepted connection before returning it.
// An addr with no port (including the empty string) is bound on DefaultNBTPort.
//
// acceptedNames are the CALLED NetBIOS names this endpoint answers to; an empty
// list answers to any name, and the "*SMBSERVER" wildcard is always answered.
//
// Parameters:
//   - addr: the address to bind, e.g. "0.0.0.0:139", ":139", "" or "127.0.0.1:0"
//   - acceptedNames: the CALLED names to serve, or nil to serve any
//
// Returns:
//   - A Listener yielding NetBIOS over TCP transports
//   - An error if the address cannot be bound
func ListenNBT(addr string, acceptedNames []string) (Listener, error) {
	ln, err := net.Listen("tcp", withDefaultPort(addr, DefaultNBTPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen for NetBIOS over TCP SMB connections: %v", err)
	}
	names := make([]string, len(acceptedNames))
	copy(names, acceptedNames)
	return &nbtListener{listener: ln, acceptedNames: names}, nil
}

// withDefaultPort appends defaultPort to addr when addr does not already carry
// one, so callers can pass a bare host or an empty string.
func withDefaultPort(addr, defaultPort string) string {
	if addr == "" {
		return ":" + defaultPort
	}
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	// An IPv6 literal has to be bracketed before a port can be appended to it.
	if strings.Count(addr, ":") > 1 && !strings.HasPrefix(addr, "[") {
		return "[" + addr + "]:" + defaultPort
	}
	return addr + ":" + defaultPort
}

// tcpListener adapts a net.Listener into a Listener yielding Direct TCP
// transports. Direct TCP has no transport-level handshake, so an accepted
// connection is usable immediately.
type tcpListener struct {
	listener net.Listener
}

func (l *tcpListener) Accept() (Transport, net.Addr, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, nil, err
	}
	return tcp.NewTCPTransportFromConn(conn), conn.RemoteAddr(), nil
}

func (l *tcpListener) Close() error { return l.listener.Close() }

func (l *tcpListener) Addr() net.Addr { return l.listener.Addr() }

// nbtListener adapts a net.Listener into a Listener yielding NetBIOS over TCP
// transports, performing the session handshake on each connection as it arrives.
type nbtListener struct {
	listener      net.Listener
	acceptedNames []string
}

// Accept accepts connections until one completes the NetBIOS session handshake.
// A connection whose handshake fails is logged and closed rather than surfaced
// as an error: a client addressing the wrong CALLED name, or speaking something
// other than the session service, must not take the listener down. The loop ends
// when the underlying listener itself fails, which is also how Close unblocks it.
func (l *nbtListener) Accept() (Transport, net.Addr, error) {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return nil, nil, err
		}

		remote := conn.RemoteAddr()
		nbtTransport := nbt.NewNBTTransportFromConn(conn)
		called, calling, err := nbtTransport.AcceptSession(l.acceptedNames)
		if err != nil {
			logger.Debugf("NetBIOS session handshake with %s failed: %v", remote, err)
			nbtTransport.Close()
			continue
		}

		logger.Debugf("NetBIOS session established with %s (called %q, calling %q)", remote, called, calling)
		return nbtTransport, remote, nil
	}
}

func (l *nbtListener) Close() error { return l.listener.Close() }

func (l *nbtListener) Addr() net.Addr { return l.listener.Addr() }

// Compile-time assurance that both listeners satisfy the Listener contract.
var (
	_ Listener = (*tcpListener)(nil)
	_ Listener = (*nbtListener)(nil)
)
