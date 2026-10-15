package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

type state uint8

const (
	stateNew state = iota
	stateConnecting
	stateReady
	statePoisoned
	stateClosing
	stateClosed
)

func (s state) String() string {
	return [...]string{"new", "connecting", "ready", "poisoned", "closing", "closed"}[s]
}

var (
	// ErrAlreadyConnected is returned by Connect on a client that already
	// connected or attempted to.
	ErrAlreadyConnected = errors.New("client: Connect already called; create a new Client")
	// ErrNotReady is matched when an operation needs a ready session.
	ErrNotReady = errors.New("client: session not ready")
	// ErrSingleUseToken is returned by Connect for a single-use profile
	// token, which cannot authenticate both Sign-on and Remote Command.
	ErrSingleUseToken = errors.New("client: Connect needs a multiple-use profile token")
	// ErrReservationReleased is returned when a callback-scoped handle is
	// used after its callback returned.
	ErrReservationReleased = errors.New("client: exclusive reservation used after its callback returned")
)

// Client is one IBM i connection. It is safe for concurrent use.
type Client struct {
	cfg Config
	// lifecycle is a one-slot semaphore serializing Connect, Close and
	// session operations; unlike a mutex its acquisition honors contexts.
	lifecycle chan struct{}

	// mu guards state, session and info for readers that must not wait
	// behind a long operation (SessionInfo).
	mu      sync.Mutex
	state   state
	session *rmtcmd.Session
	info    SessionInfo
}

// New validates config and applies defaults. It performs no I/O.
func New(config Config) (*Client, error) {
	cfg, err := config.resolve()
	if err != nil {
		return nil, err
	}
	return &Client{cfg: cfg, lifecycle: make(chan struct{}, 1)}, nil
}

func (c *Client) acquire(ctx context.Context) error {
	select {
	case c.lifecycle <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) release() { <-c.lifecycle }

func (c *Client) setState(s state) {
	c.mu.Lock()
	c.state = s
	c.info.Connected = s == stateReady
	c.mu.Unlock()
}

func (c *Client) getState() state {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// DiscoverServices resolves the Sign-on and Remote Command endpoints under
// the configured transport policy. Explicit ports skip their lookup. It may
// return a partial catalog with an error; it never substitutes a cleartext
// service for a failed secure one.
func (c *Client) DiscoverServices(ctx context.Context) (ibminet.ServiceCatalog, error) {
	return (&servicemapper.Client{Dial: c.cfg.Dial, Timeout: c.cfg.DialTimeout}).Discover(ctx, servicemapper.DiscoverOptions{
		Host:              c.cfg.Host,
		MapperPort:        c.cfg.MapperPort,
		Policy:            c.cfg.Transport,
		SignonPort:        c.cfg.SignonPort,
		RemoteCommandPort: c.cfg.RemoteCommandPort,
	})
}

// Connect validates cred with the Sign-on server, then starts the Remote
// Command server job. cred is a *PasswordCredential or a multiple-use
// *ProfileToken (a single-use token is rejected with ErrSingleUseToken
// because Connect authenticates twice). It may be called once; on failure
// every socket it opened is closed and the Client becomes closed. cred is
// not retained.
func (c *Client) Connect(ctx context.Context, cred ibminet.Credential) error {
	if cred == nil {
		return errors.New("client: nil credential")
	}
	// Connect authenticates twice (Sign-on, then Remote Command), so a
	// single-use token would be spent by the first and refused by the
	// second. Reject it before any I/O.
	if t, ok := cred.(*ibminet.ProfileToken); ok && t.Type() == ibminet.TokenSingleUse {
		return ErrSingleUseToken
	}
	if err := c.acquire(ctx); err != nil {
		return err
	}
	defer c.release()
	if c.getState() != stateNew {
		return ErrAlreadyConnected
	}
	c.setState(stateConnecting)

	info, session, err := c.connect(ctx, cred)
	if err != nil {
		c.setState(stateClosed)
		return err
	}
	c.mu.Lock()
	c.session, c.info = session, info
	c.mu.Unlock()
	c.setState(stateReady)
	return nil
}

func (c *Client) connect(ctx context.Context, cred ibminet.Credential) (SessionInfo, *rmtcmd.Session, error) {
	catalog, err := c.DiscoverServices(ctx)
	if err != nil {
		return SessionInfo{}, nil, fmt.Errorf("client: discover services: %w", err)
	}
	dialer := c.cfg.dialer()
	so, err := signon.Validate(ctx, catalog.Signon, cred, signon.Options{Dialer: dialer, CleanupTimeout: c.cfg.CleanupTimeout})
	if err != nil {
		return SessionInfo{}, nil, err
	}
	conn, err := dialer.DialEndpoint(ctx, catalog.RemoteCommand)
	if err != nil {
		return SessionInfo{}, nil, err
	}
	session, err := rmtcmd.Start(ctx, conn, cred, so.PasswordFamily, rmtcmd.StartOptions{NLV: c.cfg.NLV, JobCCSIDOverride: c.cfg.JobCCSIDOverride, Limits: c.cfg.limits()})
	if err != nil {
		return SessionInfo{}, nil, err // Start closed conn
	}
	ri := session.Info()
	info := SessionInfo{
		SignonEndpoint:     catalog.Signon,
		RemoteEndpoint:     catalog.RemoteCommand,
		ServerVersion:      so.ServerVersion,
		VRM:                so.VRM,
		PasswordLevel:      so.PasswordLevel,
		PasswordFamily:     so.PasswordFamily,
		SystemCCSID:        so.SystemCCSID,
		JobCCSID:           ri.JobCCSID,
		NegotiatedJobCCSID: ri.NegotiatedJobCCSID,
		RemoteCommandLevel: ri.Capabilities.RemoteCommandLevel,
		Capabilities:       ri.Capabilities,
		ServerJob:          ri.ServerJob,
		Warnings:           append(append([]string(nil), so.Warnings...), ri.Warnings...),
	}
	return info, session, nil
}

// SessionInfo returns a snapshot of the negotiated, non-secret metadata.
// It does not wait for in-flight operations.
func (c *Client) SessionInfo() SessionInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.info.clone()
}

// Close ends the Remote Command server job (skipped if the session is
// poisoned) and closes its socket. It waits for any in-flight operation and
// is idempotent.
func (c *Client) Close() error {
	_ = c.acquire(context.Background()) // never fails without cancellation
	defer c.release()
	switch c.getState() {
	case stateClosed:
		return nil
	case stateNew:
		c.setState(stateClosed)
		return nil
	}
	c.setState(stateClosing)
	c.mu.Lock()
	session := c.session
	c.session = nil
	c.mu.Unlock()
	var err error
	if session != nil {
		err = session.Close(c.cfg.CleanupTimeout)
	}
	c.setState(stateClosed)
	return err
}

// GenerateProfileToken authenticates cred with the Sign-on server and asks
// it to issue a profile token (see signon.GenerateProfileToken). It needs
// no connected session and does not change the Client's state. Use a
// multiple-use token with Connect; a single-use token suits one
// authentication elsewhere.
func (c *Client) GenerateProfileToken(ctx context.Context, cred *ibminet.PasswordCredential, req signon.TokenRequest) (*ibminet.ProfileToken, error) {
	catalog, err := c.DiscoverServices(ctx)
	if err != nil && catalog.Signon.Port == 0 {
		return nil, fmt.Errorf("client: discover services: %w", err)
	}
	return signon.GenerateProfileToken(ctx, catalog.Signon, cred, req, signon.Options{Dialer: c.cfg.dialer(), CleanupTimeout: c.cfg.CleanupTimeout})
}
