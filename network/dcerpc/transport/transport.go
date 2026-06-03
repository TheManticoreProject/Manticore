// Package transport defines the transport abstraction used by the DCE/RPC client.
//
// Unlike the SMB byte-stream transport (network/smb/smb_v10/transport), which frames
// raw bytes over a socket, a DCE/RPC transport carries whole PDUs as discrete
// messages: a connection-oriented call is strictly request -> response, so the
// interface exposes a single SendReceive exchange rather than independent Send and
// Receive primitives.
//
// Concrete implementations:
//   - network/dcerpc/transport/smb: ncacn_np, DCE/RPC over an SMB named pipe.
package transport

// Transport carries DCE/RPC PDUs between the client and the server.
type Transport interface {
	// Connect opens the underlying endpoint (for ncacn_np, the named pipe). It is
	// idempotent: calling Connect on an already-connected transport is a no-op.
	Connect() error

	// SendReceive writes a single outgoing PDU and returns the bytes of the response
	// PDU(s). The returned buffer holds up to MaxRecvFrag bytes read from the
	// endpoint; reassembling DCE/RPC fragments that span multiple reads is the
	// responsibility of the caller (the DCE/RPC client layer).
	SendReceive(pdu []byte) ([]byte, error)

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
