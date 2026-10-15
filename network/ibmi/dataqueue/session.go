package dataqueue

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// Defaults and bounds of StartOptions.MaxMessage.
const (
	DefaultMaxMessage = 64 << 10
	MaxMaxMessage     = 1 << 20
)

// StartOptions configures Start.
type StartOptions struct {
	// SystemCCSID encodes names, search operators and descriptions; it is
	// the Sign-on system CCSID and must be supported by ibmi/ccsid.
	SystemCCSID uint32
	// Rand is the client-seed source; nil uses crypto/rand.
	Rand io.Reader
	// MaxMessage caps the message of an error reply (default 64 KiB, at
	// most 1 MiB).
	MaxMessage int
}

// Info is the non-secret state of a Session.
type Info struct {
	SystemCCSID uint32
	// ServerJob is the server job identity from 0xF002 CP 0x111F, or "".
	ServerJob string
}

// Session is one authenticated Data Queue server job.
type Session struct {
	conn       *hostserver.Conn
	corr       hostserver.Correlator
	codec      ccsid.Codec
	info       Info
	maxMessage int

	closeOnce sync.Once
	closeErr  error
}

var serverStart = startup.Server{ID: hostserver.ServerIDDataQueue, Prefix: "dataqueue", Stage: "data queue"}

// Start authenticates conn as a Data Queue server job for cred (a password,
// with a fresh seed pair and substitute, or a profile token) and exchanges
// attributes. family is the password family learned from Sign-on; it is
// ignored for a profile token. Start owns conn: on any error it is closed.
// A rejected credential returns a *signon.Error matching
// signon.ErrAuthenticationRejected and is never retried.
func Start(ctx context.Context, conn *hostserver.Conn, cred ibminet.Credential, family ibminet.PasswordFamily, opts StartOptions) (*Session, error) {
	s, err := start(ctx, conn, cred, family, opts)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return s, nil
}

func start(ctx context.Context, conn *hostserver.Conn, cred ibminet.Credential, family ibminet.PasswordFamily, opts StartOptions) (*Session, error) {
	s := &Session{conn: conn, maxMessage: opts.MaxMessage, info: Info{SystemCCSID: opts.SystemCCSID}}
	switch {
	case s.maxMessage == 0:
		s.maxMessage = DefaultMaxMessage
	case s.maxMessage < 0 || s.maxMessage > MaxMaxMessage:
		return nil, fmt.Errorf("dataqueue: MaxMessage %d outside 1-%d", s.maxMessage, MaxMaxMessage)
	}
	var err error
	if s.codec, err = ccsid.Lookup(ccsid.CCSID(opts.SystemCCSID)); err != nil {
		return nil, fmt.Errorf("dataqueue: system CCSID: %w", err)
	}
	started, err := startup.Authenticate(ctx, conn, serverStart, cred, family, opts.Rand)
	if err != nil {
		return nil, err
	}
	if started.HasJobIdentity {
		// The value starts with a 4-byte opaque prefix (docs/04).
		if s.info.ServerJob, err = ccsid.DecodeFixed(s.codec, started.JobIdentity[4:]); err != nil {
			return nil, fmt.Errorf("dataqueue: server job identity: %w", err)
		}
	}
	const op = "exchange attributes"
	corr, err := s.corr.Next()
	if err != nil {
		return nil, err
	}
	reply, err := s.exchange(ctx, op, message.ExchangeAttributesFrame(corr), 0, message.ReplyExchangeAttributes, message.ReplyCommon)
	if err != nil {
		return nil, err
	}
	if reply.Header.RequestID == message.ReplyCommon {
		return nil, s.commonError(op, reply)
	}
	return s, nil
}

// Info returns the session's non-secret state.
func (s *Session) Info() Info { return s.info }

// Err reports why the session can no longer be used, or nil.
func (s *Session) Err() error { return s.conn.Err() }

// Close closes the connection, which ends the server job; the protocol has
// no end request. It is idempotent.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		if err := s.conn.Close(); err != nil && !errors.Is(err, hostserver.ErrClosed) {
			s.closeErr = err
		}
	})
	return s.closeErr
}

// exchange sends f and returns a reply with one of ids. timeout 0 uses the
// connection's operation timeout; WaitForever sets none.
func (s *Session) exchange(ctx context.Context, op string, f hostserver.Frame, timeout time.Duration, ids ...uint16) (hostserver.Frame, error) {
	expect := hostserver.Expect{ServerID: hostserver.ServerIDDataQueue, ReplyIDs: ids, CorrelationID: f.Header.CorrelationID}
	var reply hostserver.Frame
	var receipt hostserver.SendReceipt
	var err error
	switch timeout {
	case 0:
		reply, receipt, err = s.conn.Exchange(ctx, f, expect)
	case WaitForever:
		reply, receipt, err = s.conn.ExchangeWithin(ctx, f, expect, 0)
	default:
		reply, receipt, err = s.conn.ExchangeWithin(ctx, f, expect, timeout)
	}
	if err != nil {
		return hostserver.Frame{}, exchangeOutcome(op, receipt, err)
	}
	return reply, nil
}

// replyOutcome reports an unusable reply: the server executed the request.
func (s *Session) replyOutcome(op string, err error) error {
	return &OutcomeError{Operation: op, MayHaveExecuted: true, Err: s.conn.Poison(err)}
}

// common parses a 0x8002 reply; it returns nil for success.
func (s *Session) common(op string, f hostserver.Frame, success uint16) error {
	r, err := message.ParseCommonReply(f, s.maxMessage)
	if err != nil {
		return s.replyOutcome(op, err)
	}
	if r.ReturnCode == success {
		return nil
	}
	e := &Error{Operation: op, ReturnCode: r.ReturnCode}
	if r.HasMessage {
		if text, err := s.codec.Decode(r.Message); err == nil {
			e.Text = strings.TrimRight(text, " \x00")
			e.MessageID = messageID.FindString(e.Text)
		}
	}
	return e
}

func (s *Session) commonError(op string, f hostserver.Frame) error {
	if err := s.common(op, f, message.RCSuccess); err != nil {
		return err
	}
	// Success where a specific reply was required.
	return s.replyOutcome(op, &hostserver.MalformedError{What: "reply", Offset: 18, Reason: "common reply with success code in place of the operation reply"})
}
