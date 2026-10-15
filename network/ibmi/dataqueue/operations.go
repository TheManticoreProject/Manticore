package dataqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Attributes queries a queue's attributes.
func (s *Session) Attributes(ctx context.Context, q Queue) (*Attributes, error) {
	const op = "query attributes"
	o, err := s.object(q)
	if err != nil {
		return nil, err
	}
	corr, err := s.corr.Next()
	if err != nil {
		return nil, err
	}
	reply, err := s.exchange(ctx, op, message.QueryAttributesFrame(corr, o), 0, message.ReplyQueryAttributes, message.ReplyCommon)
	if err != nil {
		return nil, err
	}
	if reply.Header.RequestID == message.ReplyCommon {
		return nil, s.commonError(op, reply)
	}
	r, err := message.ParseAttributesReply(reply)
	if err != nil {
		return nil, s.replyOutcome(op, err)
	}
	desc, err := ccsid.DecodeFixed(s.codec, r.Description[:])
	if err != nil {
		return nil, s.replyOutcome(op, err)
	}
	return &Attributes{
		MaxEntryLength:  int(r.MaxEntry),
		Kind:            kindFromWire(r.Type),
		KeyLength:       int(r.KeyLength),
		SaveSenderInfo:  r.SaveSender,
		ForceAuxStorage: r.ForceAuxStorage,
		Description:     desc,
	}, nil
}

// Read removes and returns the next entry of a FIFO or LIFO queue, waiting
// up to wait (0 returns at once, WaitForever waits until ctx is done). It
// returns nil, nil when no entry arrived. A lost reply may have consumed
// an entry (ErrUncertainOutcome).
func (s *Session) Read(ctx context.Context, q Queue, wait time.Duration) (*Entry, error) {
	return s.read(ctx, "read", q, 0, nil, wait, false)
}

// Peek returns the next entry of a FIFO or LIFO queue without removing it.
// Another reader may remove it at any moment.
func (s *Session) Peek(ctx context.Context, q Queue, wait time.Duration) (*Entry, error) {
	return s.read(ctx, "peek", q, 0, nil, wait, true)
}

// ReadKey removes and returns the first entry of a keyed queue whose key
// compares to key as search says. Keys compare bytewise.
func (s *Session) ReadKey(ctx context.Context, q Queue, search Search, key []byte, wait time.Duration) (*Entry, error) {
	return s.read(ctx, "read", q, search, key, wait, false)
}

// PeekKey is ReadKey without removing the entry.
func (s *Session) PeekKey(ctx context.Context, q Queue, search Search, key []byte, wait time.Duration) (*Entry, error) {
	return s.read(ctx, "peek", q, search, key, wait, true)
}

func (s *Session) read(ctx context.Context, op string, q Queue, search Search, key []byte, wait time.Duration, peek bool) (*Entry, error) {
	o, err := s.object(q)
	if err != nil {
		return nil, err
	}
	seconds, err := waitSeconds(wait)
	if err != nil {
		return nil, err
	}
	req := message.ReadRequest{Object: o, Wait: seconds, Peek: peek}
	if search != 0 || key != nil {
		if !search.valid() {
			return nil, fmt.Errorf("dataqueue: %s: undefined search %s", op, search)
		}
		enc, err := ccsid.EncodeFixed(s.codec, search.String(), 2)
		if err != nil {
			return nil, err
		}
		req.Keyed, req.Key = true, key
		copy(req.Search[:], enc)
	}
	if req.Correlation, err = s.corr.Next(); err != nil {
		return nil, err
	}
	f, err := req.Frame()
	if err != nil {
		return nil, fmt.Errorf("dataqueue: %s: %w", op, err)
	}
	// The server may hold the request for the whole wait.
	timeout := time.Duration(0)
	switch {
	case wait == WaitForever:
		timeout = WaitForever
	case wait > 0:
		timeout = wait + s.conn.OperationTimeout()
	}
	reply, err := s.exchange(ctx, op, f, timeout, message.ReplyRead, message.ReplyCommon)
	if err != nil {
		return nil, err
	}
	if reply.Header.RequestID == message.ReplyCommon {
		err := s.common(op, reply, message.RCNoData)
		if err == nil {
			return nil, nil // no entry: a normal result
		}
		if e, ok := err.(*Error); ok && e.ReturnCode == message.RCSuccess {
			return nil, s.replyOutcome(op, &hostserver.MalformedError{What: "reply", Offset: 18, Reason: "common reply with success code in place of an entry"})
		}
		return nil, err
	}
	r, err := message.ParseReadReply(reply)
	if err != nil {
		return nil, s.replyOutcome(op, err)
	}
	e := &Entry{Data: r.Entry, Key: r.Key}
	if r.HasSender {
		if e.Sender, err = decodeSender(s.codec, r.Sender); err != nil {
			return nil, s.replyOutcome(op, err)
		}
	}
	return e, nil
}

// Write adds data to a FIFO or LIFO queue. A lost reply leaves it unknown
// whether the entry was added (ErrUncertainOutcome); it is never resent.
func (s *Session) Write(ctx context.Context, q Queue, data []byte) error {
	return s.write(ctx, q, nil, data)
}

// WriteKey adds data with key to a keyed queue. The key must be exactly
// the queue's key length.
func (s *Session) WriteKey(ctx context.Context, q Queue, key, data []byte) error {
	if key == nil {
		key = []byte{}
	}
	return s.write(ctx, q, key, data)
}

func (s *Session) write(ctx context.Context, q Queue, key, data []byte) error {
	const op = "write"
	o, err := s.object(q)
	if err != nil {
		return err
	}
	corr, err := s.corr.Next()
	if err != nil {
		return err
	}
	f, err := message.WriteRequest{Correlation: corr, Object: o, Entry: data, Key: key}.Frame()
	if err != nil {
		return fmt.Errorf("dataqueue: %s: %w", op, err)
	}
	return s.mutate(ctx, op, f)
}

// Clear removes every entry of a queue. It is destructive and never
// retried.
func (s *Session) Clear(ctx context.Context, q Queue) error {
	return s.clear(ctx, q, nil)
}

// ClearKey removes the entries of a keyed queue whose key equals key.
func (s *Session) ClearKey(ctx context.Context, q Queue, key []byte) error {
	if key == nil {
		key = []byte{}
	}
	return s.clear(ctx, q, key)
}

func (s *Session) clear(ctx context.Context, q Queue, key []byte) error {
	const op = "clear"
	o, err := s.object(q)
	if err != nil {
		return err
	}
	corr, err := s.corr.Next()
	if err != nil {
		return err
	}
	f, err := message.ClearFrame(corr, o, key)
	if err != nil {
		return fmt.Errorf("dataqueue: %s: %w", op, err)
	}
	return s.mutate(ctx, op, f)
}

// Create creates a queue. An existing queue fails with ErrAlreadyExists.
func (s *Session) Create(ctx context.Context, q Queue, opts CreateOptions) error {
	const op = "create"
	o, err := s.object(q)
	if err != nil {
		return err
	}
	kind, ok := opts.Kind.wire()
	if !ok {
		return fmt.Errorf("dataqueue: %s: undefined kind %s", op, opts.Kind)
	}
	authority, ok := opts.Authority.wire()
	if !ok {
		return fmt.Errorf("dataqueue: %s: undefined authority %s", op, opts.Authority)
	}
	if opts.MaxEntryLength < 1 || opts.MaxEntryLength > message.MaxEntryLength {
		return fmt.Errorf("dataqueue: %s: maximum entry length %d outside 1-%d", op, opts.MaxEntryLength, message.MaxEntryLength)
	}
	if opts.KeyLength < 0 || opts.KeyLength > message.MaxKeyLength {
		return fmt.Errorf("dataqueue: %s: key length %d outside 0-%d", op, opts.KeyLength, message.MaxKeyLength)
	}
	desc, err := ccsid.EncodeFixed(s.codec, opts.Description, message.DescriptionSize)
	if err != nil {
		return fmt.Errorf("dataqueue: %s: description: %w", op, err)
	}
	req := message.CreateRequest{
		Object: o, MaxEntry: uint32(opts.MaxEntryLength), Authority: authority, SaveSender: opts.SaveSenderInfo,
		Type: kind, KeyLength: uint16(opts.KeyLength), ForceAuxStorage: opts.ForceAuxStorage,
	}
	copy(req.Description[:], desc)
	if req.Correlation, err = s.corr.Next(); err != nil {
		return err
	}
	f, err := req.Frame()
	if err != nil {
		return fmt.Errorf("dataqueue: %s: %w", op, err)
	}
	return s.mutate(ctx, op, f)
}

// Delete deletes a queue and its entries. It is destructive and never
// retried; a lost reply does not mean the queue survived.
func (s *Session) Delete(ctx context.Context, q Queue) error {
	const op = "delete"
	o, err := s.object(q)
	if err != nil {
		return err
	}
	corr, err := s.corr.Next()
	if err != nil {
		return err
	}
	return s.mutate(ctx, op, message.DeleteFrame(corr, o))
}

// mutate sends a request answered only by the common reply.
func (s *Session) mutate(ctx context.Context, op string, f hostserver.Frame) error {
	reply, err := s.exchange(ctx, op, f, 0, message.ReplyCommon)
	clear(f.Variable) // entry and key bytes
	if err != nil {
		return err
	}
	return s.common(op, reply, message.RCSuccess)
}
