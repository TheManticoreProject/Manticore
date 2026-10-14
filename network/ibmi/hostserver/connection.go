package hostserver

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// DefaultOperationTimeout bounds each Send, Receive or Exchange when
// Options.OperationTimeout is 0.
const DefaultOperationTimeout = 30 * time.Second

var (
	// ErrPoisoned is matched (via errors.Is) by every operation on a Conn
	// whose stream can no longer be trusted. The Conn is already closed.
	ErrPoisoned = errors.New("hostserver: connection poisoned")
	// ErrClosed is returned by operations after Close.
	ErrClosed = errors.New("hostserver: connection closed")
	// ErrSendNotStarted is matched when Send gave up before writing any
	// byte. The Conn remains usable.
	ErrSendNotStarted = errors.New("hostserver: send not started")
	// ErrUnexpectedReply is matched when a reply's server ID, reply ID or
	// correlation ID differs from the expectation.
	ErrUnexpectedReply = errors.New("hostserver: unexpected reply")
)

// Options configures a Conn. Zero values select the defaults.
type Options struct {
	MaxFrame         uint32
	OperationTimeout time.Duration
	Tracer           Tracer
}

// Conn is a frame transport over one net.Conn (TCP or TLS). Operations are
// serialized: at most one is in flight. Any failure once a frame has
// started moving, and any malformed or unexpected reply, poisons and closes
// the Conn.
type Conn struct {
	conn     net.Conn
	maxFrame uint32
	timeout  time.Duration
	tracer   Tracer

	mu       sync.Mutex
	poisoned error
	closed   bool
}

// NewConn wraps an established connection. The Conn owns c and closes it.
func NewConn(c net.Conn, opts Options) (*Conn, error) {
	max, err := ResolveMaxFrame(opts.MaxFrame)
	if err != nil {
		return nil, err
	}
	if opts.OperationTimeout < 0 {
		return nil, fmt.Errorf("hostserver: negative operation timeout %s", opts.OperationTimeout)
	}
	timeout := opts.OperationTimeout
	if timeout == 0 {
		timeout = DefaultOperationTimeout
	}
	return &Conn{conn: c, maxFrame: max, timeout: timeout, tracer: opts.Tracer}, nil
}

// MaxFrame returns the effective maximum frame size.
func (c *Conn) MaxFrame() uint32 { return c.maxFrame }

// Send writes one whole frame.
func (c *Conn) Send(ctx context.Context, f Frame) (SendReceipt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.send(ctx, f)
}

// Receive reads one whole frame.
func (c *Conn) Receive(ctx context.Context) (Frame, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.receive(ctx)
}

// Expect describes the only acceptable reply to a request.
type Expect struct {
	ServerID      uint16
	ReplyIDs      []uint16
	CorrelationID uint32
}

// Exchange sends req and receives its reply under one lock, so no other
// operation can interleave. A reply that does not match expect poisons the
// Conn. The receipt is returned even on failure so callers can tell whether
// the request may have executed.
func (c *Conn) Exchange(ctx context.Context, req Frame, expect Expect) (Frame, SendReceipt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	receipt, err := c.send(ctx, req)
	if err != nil {
		return Frame{}, receipt, err
	}
	reply, err := c.receive(ctx)
	if err != nil {
		return Frame{}, receipt, err
	}
	if err := checkReply(reply.Header, expect); err != nil {
		return Frame{}, receipt, c.poison(err)
	}
	return reply, receipt, nil
}

func checkReply(h Header, expect Expect) error {
	if h.ServerID != expect.ServerID {
		return fmt.Errorf("%w: server ID 0x%04X, want 0x%04X", ErrUnexpectedReply, h.ServerID, expect.ServerID)
	}
	if h.CorrelationID != expect.CorrelationID {
		return fmt.Errorf("%w: correlation %d, want %d", ErrUnexpectedReply, h.CorrelationID, expect.CorrelationID)
	}
	for _, id := range expect.ReplyIDs {
		if h.RequestID == id {
			return nil
		}
	}
	return fmt.Errorf("%w: reply ID 0x%04X, want one of %04X", ErrUnexpectedReply, h.RequestID, expect.ReplyIDs)
}

// Close closes the connection. It is idempotent.
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	err := c.conn.Close()
	c.trace(Event{Kind: EventClosed})
	return err
}

// Err returns the poisoning cause, ErrClosed after Close, or nil.
func (c *Conn) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.usable()
}

func (c *Conn) usable() error {
	if c.poisoned != nil {
		return fmt.Errorf("%w: %w", ErrPoisoned, c.poisoned)
	}
	if c.closed {
		return ErrClosed
	}
	return nil
}

// poison records cause, closes the socket and returns the poisoned error.
func (c *Conn) poison(cause error) error {
	if c.poisoned == nil {
		c.poisoned = cause
		c.closed = true
		_ = c.conn.Close()
		c.trace(Event{Kind: EventPoisoned, Err: cause})
	}
	return fmt.Errorf("%w: %w", ErrPoisoned, cause)
}

func (c *Conn) trace(e Event) {
	if c.tracer != nil {
		c.tracer.Trace(e)
	}
}

// arm applies the earlier of the context deadline and the operation timeout
// to the socket, and forces an immediate deadline if ctx is cancelled while
// I/O is blocked. The returned function disarms it and clears the deadline.
func (c *Conn) arm(ctx context.Context) func() {
	deadline := time.Now().Add(c.timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = c.conn.SetDeadline(deadline)
	fired := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(fired)
		_ = c.conn.SetDeadline(time.Unix(1, 0))
	})
	return func() {
		if !stop() {
			<-fired
		}
		_ = c.conn.SetDeadline(time.Time{})
	}
}

// ioError prefers the context's error when cancellation forced the deadline.
func ioError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("%w (%w)", ctxErr, err)
	}
	return err
}

func (c *Conn) send(ctx context.Context, f Frame) (SendReceipt, error) {
	if err := c.usable(); err != nil {
		return SendReceipt{}, err
	}
	b, err := f.Marshal(c.maxFrame)
	if err != nil {
		return SendReceipt{}, err // nothing written; still usable
	}
	if err := ctx.Err(); err != nil {
		return SendReceipt{}, fmt.Errorf("%w: %w", ErrSendNotStarted, err)
	}
	disarm := c.arm(ctx)
	total := 0
	for total < len(b) {
		n, werr := c.conn.Write(b[total:])
		total += n
		if werr != nil {
			disarm()
			// Even a zero-byte failure may have corrupted TLS state.
			return SendReceipt{BytesAccepted: total}, c.poison(fmt.Errorf("send %d of %d bytes: %w", total, len(b), ioError(ctx, werr)))
		}
		if n == 0 {
			disarm()
			return SendReceipt{BytesAccepted: total}, c.poison(fmt.Errorf("send %d of %d bytes: %w", total, len(b), io.ErrShortWrite))
		}
	}
	disarm()
	c.trace(Event{Kind: EventFrameSent, ServerID: f.Header.ServerID, RequestID: f.Header.RequestID, CorrelationID: f.Header.CorrelationID, Bytes: total})
	return SendReceipt{BytesAccepted: total}, nil
}

func (c *Conn) receive(ctx context.Context) (Frame, error) {
	if err := c.usable(); err != nil {
		return Frame{}, err
	}
	// A reply may already be in flight, so any failure from here on leaves
	// the stream position unknown and poisons the Conn.
	if err := ctx.Err(); err != nil {
		return Frame{}, c.poison(fmt.Errorf("receive not started: %w", err))
	}
	disarm := c.arm(ctx)
	defer disarm()

	var hb [HeaderLength]byte
	if n, err := io.ReadFull(c.conn, hb[:]); err != nil {
		return Frame{}, c.poison(fmt.Errorf("receive header, %d of %d bytes: %w", n, HeaderLength, ioError(ctx, err)))
	}
	// Check the advertised length against the cap before allocating.
	if l := binary.BigEndian.Uint32(hb[:]); l > c.maxFrame {
		return Frame{}, c.poison(fmt.Errorf("%w: header advertises %d bytes, maximum %d", ErrFrameTooLarge, l, c.maxFrame))
	}
	h, err := UnmarshalHeader(hb[:])
	if err != nil {
		return Frame{}, c.poison(err)
	}
	body := make([]byte, h.Length-HeaderLength)
	if n, err := io.ReadFull(c.conn, body); err != nil {
		return Frame{}, c.poison(fmt.Errorf("receive body, %d of %d bytes: %w", n, len(body), ioError(ctx, err)))
	}
	c.trace(Event{Kind: EventFrameReceived, ServerID: h.ServerID, RequestID: h.RequestID, CorrelationID: h.CorrelationID, Bytes: int(h.Length)})
	t := int(h.TemplateLength)
	return Frame{Header: h, Template: body[:t:t], Variable: body[t:]}, nil
}
