// Package transport defines the datagram transport abstraction used by the
// connectionless (v4) DCE/RPC protocol machine.
//
// Unlike the connection-oriented transport in network/dcerpc/v5/transport, which
// exposes a byte-oriented Send/Recv pair over a reliable stream (an SMB named pipe),
// a connectionless transport is message-oriented: every Send transmits exactly one
// datagram and every Recv returns exactly one datagram, preserving message
// boundaries. Reliability, fragmentation, and acknowledgement are the responsibility
// of the protocol machine layered above this interface, not of the transport.
//
// The only concrete protocol sequence defined here is ncadg_ip_udp (RPC directly
// over UDP); its implementation lives in network/dcerpc/v4/transport/udp.
//
// References:
//   - [C706] section 12.5 (connectionless fragmentation and PDU sizing):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [C706] Appendix H (well-known endpoints) and Appendix I (protocol identifiers):
//     https://pubs.opengroup.org/onlinepubs/9629399/apdxh.htm
//   - [MS-RPCE] 2.1.1.1 UDP (NCADG_IP_UDP) — RPC directly over UDP, protocol
//     identifier 0x08, endpoint-mapper well-known endpoint 135:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/f3c9d073-1563-4d47-861a-14023ec4990e
//   - [MS-RPCE] Appendix B, product behavior note <17> (maximum PDU size over UDP):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/9039a59f-3075-4680-9517-19239bebf155
package transport

import (
	"net"
	"time"
)

const (
	// EndpointMapperPort is the well-known UDP port of the RPC endpoint mapper
	// ([MS-RPCE] 2.1.1.1, [C706] Appendix H).
	EndpointMapperPort = 135

	// ProtocolIdentifier is the ncadg_ip_udp protocol-tower floor identifier
	// ([MS-RPCE] 2.1.1.1, [C706] Appendix I).
	ProtocolIdentifier = 0x08

	// MaxPDUSizeDefault is the maximum connectionless PDU size, in bytes, that
	// Windows (other than NT 4.0) accepts over UDP ([MS-RPCE] Appendix B note <17>).
	// The protocol machine must fragment calls into PDUs no larger than this.
	MaxPDUSizeDefault = 4096

	// MaxPDUSizeLegacy is the maximum connectionless PDU size, in bytes, accepted by
	// Windows NT 4.0 over UDP ([MS-RPCE] Appendix B note <17>).
	MaxPDUSizeLegacy = 1024
)

// Transport carries connectionless DCE/RPC PDUs as datagrams between client and
// server. Implementations are not required to be safe for concurrent use.
type Transport interface {
	// Connect opens the underlying datagram endpoint. It is idempotent: calling
	// Connect on an already-connected transport is a no-op.
	Connect() error

	// Send transmits datagram as a single message and returns the number of bytes
	// written. A datagram larger than MaxPDUSize is rejected rather than truncated or
	// fragmented; fragmentation is the protocol machine's responsibility.
	Send(datagram []byte) (int, error)

	// Recv reads and returns the bytes of a single datagram. The returned slice is
	// owned by the caller.
	Recv() ([]byte, error)

	// SetDeadline sets the absolute read/write deadline on the underlying endpoint, as
	// net.Conn.SetDeadline does. The protocol machine uses it to drive retransmission
	// and call timers. A zero time clears the deadline.
	SetDeadline(t time.Time) error

	// MaxPDUSize is the largest datagram, in bytes, that Send will transmit and that
	// the protocol machine should produce when fragmenting.
	MaxPDUSize() int

	// RemoteAddr returns the address of the peer, or nil if not connected.
	RemoteAddr() net.Addr

	// IsConnected reports whether the transport currently has an open endpoint.
	IsConnected() bool

	// Close tears down the endpoint.
	Close() error
}
