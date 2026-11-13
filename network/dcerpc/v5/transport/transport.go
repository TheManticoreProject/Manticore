// Package transport defines the transport abstraction used by the DCE/RPC client.
//
// Unlike the SMB byte-stream transport (network/smb/common/transport), which frames
// raw bytes over a socket, a DCE/RPC transport carries PDUs. A connection-oriented
// call may span several fragments in each direction, so the interface exposes
// independent Send and Recv primitives: the DCE/RPC client writes all request
// fragments with Send, then reassembles the response by reading with Recv. This
// matches the named-pipe model in [MS-RPCE] 2.1.1.2, where PDUs are written and read
// as named pipe writes and reads.
//
// Concrete implementations:
//   - network/dcerpc/v5/transport/smb: ncacn_np, DCE/RPC over an SMB named pipe.
package transport

// Transport carries DCE/RPC PDUs between the client and the server.
type Transport interface {
	// Connect opens the underlying endpoint (for ncacn_np, the named pipe). It is
	// idempotent: calling Connect on an already-connected transport is a no-op.
	Connect() error

	// Send writes a single complete PDU to the endpoint.
	Send(pdu []byte) error

	// Recv reads from the endpoint and returns the bytes of a single read (up to
	// MaxRecvFrag bytes). The returned buffer may contain part of a PDU, exactly one
	// PDU, or several PDUs; reassembling fragments across reads is the caller's
	// responsibility. An empty result indicates the peer wrote nothing (for example,
	// the pipe was closed).
	Recv() ([]byte, error)

	// Close tears down the endpoint. For ncacn_np it closes the pipe handle but
	// leaves the underlying SMB session and tree connect intact, since their
	// lifecycle is owned by the caller that supplied the SMB client.
	Close() error

	// MaxXmitFrag is the largest fragment size, in bytes, that this transport is
	// willing to send. It is proposed in the Bind PDU; the value the server accepts
	// is returned in the Bind_Ack and may be smaller.
	MaxXmitFrag() uint16

	// MaxRecvFrag is the largest fragment size, in bytes, that this transport is
	// willing to receive. It is proposed in the Bind PDU alongside MaxXmitFrag.
	MaxRecvFrag() uint16
}

// SerialCaller is an optional interface a Transport implements when it cannot carry
// more than one call in flight at a time. Its send and receive share a single ordered
// channel — for ncacn_np, one SMB named-pipe handle — so interleaving one call's
// request writes with another call's response reads overruns the pipe; and over SMB1
// with signing active, the per-connection signing sequence number pairs each request
// with its response and cannot be multiplexed at all. When a transport reports true,
// the DCE/RPC client serializes whole call exchanges over it, degrading concurrent
// Call/Invoke to serial while keeping them correct and race-free. A transport whose
// send and receive are independent (ncacn_ip_tcp, or an SMB2 pipe with per-message
// MessageId signing) does not implement this interface, or returns false, and its
// calls run concurrently.
type SerialCaller interface {
	// SerializesCalls reports whether the client must run one call at a time over
	// this transport.
	SerializesCalls() bool
}

// MinFragSize is the smallest fragment size a connection-oriented DCE/RPC peer may
// propose for max_xmit_frag or max_recv_frag: MustRecvFragSize in C706 12.6.4.1.
// Every implementation must accept fragments of this size.
const MinFragSize uint16 = 1432

// NormalizeFragSize returns the fragment size a transport proposes at Bind for a
// requested size: def when size is 0, MinFragSize when size is below it, and size
// otherwise.
func NormalizeFragSize(size, def uint16) uint16 {
	switch {
	case size == 0:
		return def
	case size < MinFragSize:
		return MinFragSize
	default:
		return size
	}
}
