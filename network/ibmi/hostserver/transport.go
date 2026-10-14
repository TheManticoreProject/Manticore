package hostserver

import "context"

// SendReceipt reports how many frame bytes the local connection accepted.
// It does not prove the peer received, parsed or executed anything.
type SendReceipt struct {
	BytesAccepted int
}

// Transport sends and receives whole, validated frames.
type Transport interface {
	Send(ctx context.Context, frame Frame) (SendReceipt, error)
	Receive(ctx context.Context) (Frame, error)
	Close() error
}

var _ Transport = (*Conn)(nil)
