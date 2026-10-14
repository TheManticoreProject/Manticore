package hostserver

// EventKind identifies a traced connection event.
type EventKind uint8

const (
	EventFrameSent EventKind = iota + 1
	EventFrameReceived
	EventPoisoned
	EventClosed
)

func (k EventKind) String() string {
	switch k {
	case EventFrameSent:
		return "frame-sent"
	case EventFrameReceived:
		return "frame-received"
	case EventPoisoned:
		return "poisoned"
	case EventClosed:
		return "closed"
	}
	return "unknown"
}

// Event is connection metadata for diagnostics. It has no field that can
// carry frame bytes, field values, credentials or IBM message text.
type Event struct {
	Kind          EventKind
	ServerID      uint16
	RequestID     uint16
	CorrelationID uint32
	Bytes         int
	// Err is set for EventPoisoned. Errors from this package describe
	// offsets and lengths, never content.
	Err error
}

// Tracer receives Events. It is called synchronously and must not block or
// call back into the Conn.
type Tracer interface {
	Trace(Event)
}
