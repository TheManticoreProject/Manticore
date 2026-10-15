package rmtcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

// DefaultCleanupTimeout bounds the 0x1004 end request when no timeout is
// given.
const DefaultCleanupTimeout = 5 * time.Second

// StartOptions configures Start.
type StartOptions struct {
	// NLV is the four-digit national-language version; empty means 2924.
	NLV string
	// Rand is the client-seed source; nil uses crypto/rand.
	Rand io.Reader
	// JobCCSIDOverride, when nonzero, replaces the negotiated job CCSID for
	// encoding names, commands and messages. It always adds a warning.
	JobCCSIDOverride uint32
	// Limits bound program-call and command replies; zero fields select the
	// defaults.
	Limits Limits
}

// Info is the non-secret state negotiated by Start.
type Info struct {
	// NegotiatedJobCCSID is the 0x8001 value; JobCCSID is the one in use
	// (the override, if any).
	NegotiatedJobCCSID uint32
	JobCCSID           uint32
	Capabilities       Capabilities
	// ServerJob is the server job identity from 0xF002 CP 0x111F, decoded
	// in the job CCSID, or "" if the server did not return it.
	ServerJob string
	Warnings  []string
}

// Session is one authenticated Remote Command server job.
type Session struct {
	conn   *hostserver.Conn
	corr   hostserver.Correlator
	info   Info
	codec  ccsid.Codec
	limits Limits

	closeOnce sync.Once
	closeErr  error
}

// Start authenticates conn as a Remote Command server job for cred (a
// password, with a fresh seed pair and substitute, or a profile token) and
// negotiates its attributes. family is the password family learned from
// Sign-on; it is ignored for a profile token. Start owns conn: on any error it is closed.
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

var (
	expectSeeds = hostserver.Expect{ServerID: hostserver.ServerIDRemoteCommand, ReplyIDs: []uint16{message.ReplyExchangeSeeds}}
	expectStart = hostserver.Expect{ServerID: hostserver.ServerIDRemoteCommand, ReplyIDs: []uint16{message.ReplyStartServer}}
)

func start(ctx context.Context, conn *hostserver.Conn, cred ibminet.Credential, family ibminet.PasswordFamily, opts StartOptions) (*Session, error) {
	if cred == nil {
		return nil, errors.New("rmtcmd: nil credential")
	}
	nlv := opts.NLV
	if nlv == "" {
		nlv = message.DefaultNLV
	}
	limits, err := opts.Limits.Resolve()
	if err != nil {
		return nil, err
	}
	clientSeed, err := auth.NewSeed(opts.Rand)
	if err != nil {
		return nil, err
	}

	attributes := message.SeedAttributes // the V5R4 golden value
	if _, isPassword := cred.(*ibminet.PasswordCredential); isPassword && family == auth.FamilySHA512 {
		attributes = message.SeedAttributesLevel4
	}
	f, _, err := conn.Exchange(ctx, message.ExchangeSeedsFrameWithAttributes(clientSeed, attributes), expectSeeds)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: exchange seeds: %w", err)
	}
	seeds, err := message.ParseExchangeSeedsReply(f)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: exchange seeds: %w", err)
	}
	if seeds.ReturnCode != 0 {
		return nil, &signon.Error{Stage: "remote command exchange seeds", Code: signon.ReturnCode(seeds.ReturnCode)}
	}

	sub, err := auth.Derive(cred, family, clientSeed, seeds.ServerSeed)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %w", err)
	}
	req, err := message.StartServerRequest{CredentialType: sub.CredentialType, Substitute: sub.Value, Profile: sub.Profile}.Frame()
	sub.Wipe()
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %w", err)
	}
	f, _, err = conn.Exchange(ctx, req, expectStart)
	clear(req.Variable) // holds the substitute and profile
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: start server: %w", err)
	}
	started, err := message.ParseStartServerReply(f)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: start server: %w", err)
	}
	if started.ReturnCode != 0 {
		return nil, &signon.Error{Stage: "remote command start server", Code: signon.ReturnCode(started.ReturnCode)}
	}

	s := &Session{conn: conn, limits: limits}
	corr, err := s.corr.Next()
	if err != nil {
		return nil, err
	}
	req, err = message.ExchangeAttributesFrame(corr, nlv)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: %w", err)
	}
	f, _, err = conn.Exchange(ctx, req, hostserver.Expect{
		ServerID: hostserver.ServerIDRemoteCommand, ReplyIDs: []uint16{message.ReplyExchangeAttributes}, CorrelationID: corr,
	})
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: exchange attributes: %w", err)
	}
	attrs, err := message.ParseExchangeAttributesReply(f)
	if err != nil {
		return nil, fmt.Errorf("rmtcmd: exchange attributes: %w", err)
	}
	if attrs.ReturnCode != 0 {
		text, warning := attributesWarnings[attrs.ReturnCode]
		if !warning {
			return nil, &AttributesError{Code: attrs.ReturnCode}
		}
		s.info.Warnings = append(s.info.Warnings, fmt.Sprintf("exchange attributes returned 0x%04X (%s)", attrs.ReturnCode, text))
	}

	s.info.NegotiatedJobCCSID = attrs.JobCCSID
	s.info.JobCCSID = attrs.JobCCSID
	if opts.JobCCSIDOverride != 0 {
		s.info.JobCCSID = opts.JobCCSIDOverride
		s.info.Warnings = append(s.info.Warnings, fmt.Sprintf("using job CCSID %d instead of the negotiated %d", opts.JobCCSIDOverride, attrs.JobCCSID))
	}
	if s.codec, err = ccsid.Lookup(ccsid.CCSID(s.info.JobCCSID)); err != nil {
		return nil, fmt.Errorf("rmtcmd: job CCSID: %w", err)
	}
	s.info.Capabilities = CapabilitiesForLevel(attrs.DataStreamLevel)
	if started.HasJobIdentity {
		// The value starts with a 4-byte opaque prefix (docs/04).
		if s.info.ServerJob, err = ccsid.DecodeFixed(s.codec, started.JobIdentity[4:]); err != nil {
			return nil, fmt.Errorf("rmtcmd: server job identity: %w", err)
		}
	}
	return s, nil
}

// Info returns a copy of the negotiated state.
func (s *Session) Info() Info {
	info := s.info
	info.Warnings = append([]string(nil), s.info.Warnings...)
	return info
}

// Capabilities returns the negotiated capabilities.
func (s *Session) Capabilities() Capabilities { return s.info.Capabilities }

// Codec returns the job-CCSID codec.
func (s *Session) Codec() ccsid.Codec { return s.codec }

// Err reports why the session can no longer be used (poisoned or closed),
// or nil.
func (s *Session) Err() error { return s.conn.Err() }

// Close ends the server job with the one-way 0x1004 request, bounded by
// timeout (0 means DefaultCleanupTimeout), and closes the socket. It skips
// 0x1004 on a poisoned or closed connection. It is idempotent and returns
// the first call's result.
func (s *Session) Close(timeout time.Duration) error {
	s.closeOnce.Do(func() {
		if timeout <= 0 {
			timeout = DefaultCleanupTimeout
		}
		var sendErr error
		if s.conn.Err() == nil {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			_, sendErr = s.conn.Send(ctx, message.EndServerFrame())
			cancel()
			if sendErr != nil {
				sendErr = fmt.Errorf("rmtcmd: end server: %w", sendErr)
			}
		}
		closeErr := s.conn.Close()
		s.closeErr = errors.Join(sendErr, closeErr)
	})
	return s.closeErr
}
