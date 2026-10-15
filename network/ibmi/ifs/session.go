package ifs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/ifs/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// Defaults and bounds of StartOptions.
const (
	DefaultMaxChunk = 64 << 10
	MaxMaxChunk     = 16_000_000
	DefaultMaxChain = 256
	MaxMaxChain     = 1 << 16
)

// StartOptions configures Start.
type StartOptions struct {
	// SystemCCSID decodes the server job identity; 0 leaves Info.ServerJob
	// empty.
	SystemCCSID uint32
	// Rand is the client-seed source; nil uses crypto/rand.
	Rand io.Reader
	// MaxChunk caps the length of one read request (default 64 KiB, at
	// most 16,000,000 bytes).
	MaxChunk int
	// MaxChain caps the reply packets of one read (default 256).
	MaxChain int
}

// Info is the non-secret state negotiated by Start.
type Info struct {
	// DataStreamLevel is the level the server returned (at most 15).
	DataStreamLevel uint16
	// MaxBlock is the server's advertised maximum data block.
	MaxBlock uint32
	// PathCCSID encodes pathnames.
	PathCCSID uint16
	// ServerJob is the server job identity from 0xF002 CP 0x111F, or "".
	ServerJob string
}

// Session is one authenticated File server job.
type Session struct {
	// mu serializes operations: a read's reply chain must be consumed
	// before any other request is sent.
	mu       sync.Mutex
	conn     *hostserver.Conn
	corr     hostserver.Correlator
	info     Info
	maxChunk int
	maxChain int

	closeOnce sync.Once
	closeErr  error
}

var serverStart = startup.Server{ID: hostserver.ServerIDFile, Prefix: "ifs", Stage: "file server"}

// Start authenticates conn as a File server job for cred (a password, with
// a fresh seed pair and substitute, or a profile token) and exchanges
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

func resolve(name string, v, def, max int) (int, error) {
	switch {
	case v == 0:
		return def, nil
	case v < 0 || v > max:
		return 0, fmt.Errorf("ifs: %s %d outside 1-%d", name, v, max)
	}
	return v, nil
}

func start(ctx context.Context, conn *hostserver.Conn, cred ibminet.Credential, family ibminet.PasswordFamily, opts StartOptions) (*Session, error) {
	s := &Session{conn: conn}
	var err error
	if s.maxChunk, err = resolve("MaxChunk", opts.MaxChunk, DefaultMaxChunk, MaxMaxChunk); err != nil {
		return nil, err
	}
	if s.maxChain, err = resolve("MaxChain", opts.MaxChain, DefaultMaxChain, MaxMaxChain); err != nil {
		return nil, err
	}
	var codec ccsid.Codec
	if opts.SystemCCSID != 0 {
		if codec, err = ccsid.Lookup(ccsid.CCSID(opts.SystemCCSID)); err != nil {
			return nil, fmt.Errorf("ifs: system CCSID: %w", err)
		}
	}
	started, err := startup.Authenticate(ctx, conn, serverStart, cred, family, opts.Rand)
	if err != nil {
		return nil, err
	}
	if started.HasJobIdentity && codec != nil {
		// The value starts with a 4-byte opaque prefix (docs/04).
		if s.info.ServerJob, err = ccsid.DecodeFixed(codec, started.JobIdentity[4:]); err != nil {
			return nil, fmt.Errorf("ifs: server job identity: %w", err)
		}
	}

	const op = "exchange attributes"
	corr, err := s.corr.Next()
	if err != nil {
		return nil, err
	}
	req, err := message.ExchangeAttributesFrame(corr, message.DataStreamLevel, message.PreferredCCSIDs)
	if err != nil {
		return nil, err
	}
	reply, err := s.exchange(ctx, op, req, message.ReplyExchangeAttributes, message.ReplyReturnCode)
	if err != nil {
		return nil, err
	}
	if reply.Header.RequestID == message.ReplyReturnCode {
		return nil, s.returnCode(op, reply)
	}
	attrs, err := message.ParseAttributesReply(reply)
	if err != nil {
		return nil, s.replyOutcome(op, err)
	}
	if attrs.Level >= 16 {
		return nil, fmt.Errorf("%w: data-stream level %d (only levels below 16 are implemented)", ErrUnsupported, attrs.Level)
	}
	if c := attrs.CCSIDs[0]; c != 1200 && c != 13488 {
		return nil, fmt.Errorf("%w: pathname CCSID %d", ErrUnsupported, c)
	}
	s.info.DataStreamLevel, s.info.MaxBlock, s.info.PathCCSID = attrs.Level, attrs.MaxBlock, attrs.CCSIDs[0]
	return s, nil
}

// Info returns the negotiated, non-secret state.
func (s *Session) Info() Info { return s.info }

// Err reports why the session can no longer be used, or nil.
func (s *Session) Err() error { return s.conn.Err() }

// Close closes the connection, which ends the server job and invalidates
// every open File. It waits for an in-flight operation and is idempotent.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if err := s.conn.Close(); err != nil && !errors.Is(err, hostserver.ErrClosed) {
			s.closeErr = err
		}
	})
	return s.closeErr
}

func (s *Session) exchange(ctx context.Context, op string, f hostserver.Frame, ids ...uint16) (hostserver.Frame, error) {
	reply, receipt, err := s.conn.Exchange(ctx, f, hostserver.Expect{ServerID: hostserver.ServerIDFile, ReplyIDs: ids, CorrelationID: f.Header.CorrelationID})
	if err != nil {
		return hostserver.Frame{}, exchangeOutcome(op, receipt, err)
	}
	return reply, nil
}

// replyOutcome reports an unusable reply: the server executed the request.
func (s *Session) replyOutcome(op string, err error) error {
	return &OutcomeError{Operation: op, MayHaveExecuted: true, Err: s.conn.Poison(err)}
}

// returnCode turns a 0x8001 reply into an *Error (or nil for success).
func (s *Session) returnCode(op string, f hostserver.Frame) error {
	rc, err := message.ParseReturnCode(f)
	if err != nil {
		return s.replyOutcome(op, err)
	}
	if rc == RCSuccess {
		return nil
	}
	return &Error{Operation: op, ReturnCode: rc}
}

// encodePath encodes p as UTF-16BE code units, no BOM or terminator.
func encodePath(p string) ([]byte, error) {
	if p == "" {
		return nil, errors.New("ifs: empty path")
	}
	if !utf8.ValidString(p) {
		return nil, errors.New("ifs: path is not valid UTF-8")
	}
	var b []byte
	for _, u := range utf16.Encode([]rune(p)) {
		if u == 0 {
			return nil, errors.New("ifs: path contains NUL")
		}
		b = append(b, byte(u>>8), byte(u))
	}
	return b, nil
}
