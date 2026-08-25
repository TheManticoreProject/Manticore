package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Config configures a Server.
//
// It holds only the settings the current phase honours; an exported knob that
// does nothing is worse than one that does not exist yet. Negotiation and
// authentication settings arrive with the phases that consume them.
type Config struct {
	// ServerName and DomainName are the NetBIOS names advertised during
	// authentication, and DNSComputerName and DNSDomainName their fully
	// qualified forms. They reach the client in the CHALLENGE TargetInfo, where a
	// client folds them into its response, so they are part of what it commits
	// to. Empty names are omitted rather than advertised as empty.
	ServerName      string
	DomainName      string
	DNSComputerName string
	DNSDomainName   string

	// NativeOS and NativeLanMan are the informational strings returned in the
	// session-setup response. They MUST NOT be empty: strict clients reject a
	// session setup that leaves them blank, so NewServer defaults them.
	NativeOS     string
	NativeLanMan string

	// MaxBufferSize is the largest SMB message the server accepts, and
	// MaxMpxCount the number of commands it will have outstanding. NewServer
	// defaults both.
	MaxBufferSize uint32
	MaxMpxCount   uint16

	// ServerGUID is advertised in the negotiate response under extended
	// security. The zero value means generate a random one at NewServer time; a
	// client uses it only to notice that two names resolve to one host, and it
	// is not a secure identifier.
	ServerGUID guid.GUID

	// Authenticator resolves a claimed identity to the account's NT hash, and
	// reports false for an identity it does not know. StaticAccounts builds one
	// from a fixed list.
	//
	// Nil means no logon can succeed: every attempt is refused, which is what a
	// server whose purpose is to harvest responses wants. Holding NT hashes
	// rather than passwords is deliberate — the hash is all verification needs.
	Authenticator func(domain, username string) (ntHash [16]byte, ok bool)

	// AllowGuest admits a logon whose identity the Authenticator does not know,
	// as a guest, reporting SMB_SETUP_GUEST so the client knows it was not
	// authenticated as itself.
	//
	// A guest session derives no key, so it cannot sign. It is therefore refused
	// outright when the signing policy requires signatures, rather than being
	// granted a session that cannot carry a single subsequent request.
	AllowGuest bool

	// AllowAnonymous admits a null session: a logon claiming no identity and
	// carrying no response. Like a guest session it has no key and cannot sign.
	AllowAnonymous bool

	// SigningPolicy selects whether message signing is unsupported, offered or
	// demanded. The zero value is SigningDisabled.
	SigningPolicy SigningPolicy

	// MaxSessionsPerConnection bounds the sessions one connection may establish.
	// Zero applies DefaultMaxSessionsPerConnection.
	MaxSessionsPerConnection int

	// Timeout bounds each read on a connection, so a client that opens a socket
	// and says nothing does not hold a goroutine forever. Zero means no bound.
	Timeout time.Duration

	// MaxConnections bounds the number of connections served at once. A
	// connection arriving while the server is at the limit is closed
	// immediately. Zero means unbounded.
	MaxConnections int
}

// Configuration defaults applied by NewServer when a field is left zero.
const (
	// DefaultNativeOS and DefaultNativeLanMan are informational only, but must
	// be non-empty for strict clients to accept the session setup.
	DefaultNativeOS     = "Unix"
	DefaultNativeLanMan = "Manticore"

	// DefaultMaxBufferSize is what Windows offers. [MS-CIFS] 2.2.4.52.2 requires
	// a multiple of 4 and suggests at least 4356.
	DefaultMaxBufferSize uint32 = 16644

	// DefaultMaxMpxCount is the number of outstanding commands advertised. The
	// server answers one request at a time, so this is what it will accept
	// rather than a promise of concurrency.
	DefaultMaxMpxCount uint16 = 50

	// DefaultMaxSessionsPerConnection bounds how many sessions one connection may
	// establish, so a client cannot exhaust the 16-bit identifier space.
	DefaultMaxSessionsPerConnection = 64
)

// SigningPolicy selects a server's stance on SMB message signing.
type SigningPolicy int

const (
	// SigningDisabled advertises no signing support and never signs. A client
	// that requires signatures will refuse to talk to the server.
	SigningDisabled SigningPolicy = iota

	// SigningEnabled advertises signing and uses it when the client asks for it,
	// leaving an unsigned session available to a client that does not.
	SigningEnabled

	// SigningRequired advertises signing as mandatory and refuses a session that
	// cannot sign, which includes every guest and anonymous session.
	SigningRequired
)

// String renders the policy for a log line.
func (p SigningPolicy) String() string {
	switch p {
	case SigningDisabled:
		return "disabled"
	case SigningEnabled:
		return "enabled"
	case SigningRequired:
		return "required"
	}
	return "unknown"
}

// Account is one credential the server will authenticate against.
type Account struct {
	// Domain and Username are matched against what a client claims. The domain
	// is compared exactly, because NTLMv2 folds it into the response as sent, so
	// a client claiming a different spelling produced a different response.
	Domain   string
	Username string

	// NTHash is the account's NT hash, which is all that verification needs.
	NTHash [16]byte
}

// StaticAccounts returns an Authenticator backed by a fixed list of accounts.
//
// The username is matched case-insensitively, as Windows does; the domain is
// matched exactly, because it is folded into the client's response as sent and a
// different spelling is a different credential.
//
// Parameters:
//   - accounts: the credentials to accept
//
// Returns:
//   - An Authenticator over those credentials
func StaticAccounts(accounts ...Account) func(domain, username string) ([16]byte, bool) {
	stored := make([]Account, len(accounts))
	copy(stored, accounts)

	return func(domain, username string) ([16]byte, bool) {
		for _, account := range stored {
			if account.Domain == domain && strings.EqualFold(account.Username, username) {
				return account.NTHash, true
			}
		}
		return [16]byte{}, false
	}
}

// Server is an SMB 1.0 server. It is safe for concurrent use: handlers and
// listeners may be registered before or during serving, and Close may be called
// from any goroutine.
type Server struct {
	config Config

	mutex     sync.RWMutex
	handlers  []Handler
	listeners []transport.Listener
	conns     map[*Connection]struct{}
	closed    bool

	// slots bounds concurrent connections when MaxConnections is set. It is nil
	// when the limit is unbounded.
	slots chan struct{}

	wg sync.WaitGroup
}

// NewServer creates a server from a configuration.
//
// Parameters:
//   - config: the server configuration; the zero value is usable
//
// Returns:
//   - The server
//   - An error if the configuration is invalid
func NewServer(config Config) (*Server, error) {
	if config.MaxConnections < 0 {
		return nil, fmt.Errorf("MaxConnections cannot be negative (got %d)", config.MaxConnections)
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("Timeout cannot be negative (got %s)", config.Timeout)
	}

	if config.NativeOS == "" {
		config.NativeOS = DefaultNativeOS
	}
	if config.NativeLanMan == "" {
		config.NativeLanMan = DefaultNativeLanMan
	}
	if config.MaxBufferSize == 0 {
		config.MaxBufferSize = DefaultMaxBufferSize
	}
	if config.MaxBufferSize%4 != 0 {
		return nil, fmt.Errorf("MaxBufferSize must be a multiple of 4 (got %d)", config.MaxBufferSize)
	}
	if config.MaxMpxCount == 0 {
		config.MaxMpxCount = DefaultMaxMpxCount
	}
	if config.MaxSessionsPerConnection == 0 {
		config.MaxSessionsPerConnection = DefaultMaxSessionsPerConnection
	}
	if config.MaxSessionsPerConnection < 0 {
		return nil, fmt.Errorf("MaxSessionsPerConnection cannot be negative (got %d)", config.MaxSessionsPerConnection)
	}
	// A policy that demands signatures cannot be served by a configuration that
	// can only produce sessions without a key.
	if config.SigningPolicy == SigningRequired && config.Authenticator == nil {
		return nil, fmt.Errorf("SigningRequired needs an Authenticator: a session that cannot be verified derives no key to sign with")
	}
	if config.ServerGUID == (guid.GUID{}) {
		config.ServerGUID = *guid.NewGUID()
	}

	s := &Server{
		config: config,
		conns:  make(map[*Connection]struct{}),
	}
	if config.MaxConnections > 0 {
		s.slots = make(chan struct{}, config.MaxConnections)
	}
	return s, nil
}

// Config returns the server's configuration.
func (s *Server) Config() Config {
	return s.config
}

// RegisterHandler appends a handler to the chain. Handlers run in registration
// order, before the built-in command dispatch, and the first one to report that
// it handled a request stops the chain.
func (s *Server) RegisterHandler(handler Handler) {
	if handler == nil {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handlers = append(s.handlers, handler)
}

// snapshotHandlers copies the handler chain, so a request is dispatched against
// a stable list even if another goroutine registers a handler mid-flight.
func (s *Server) snapshotHandlers() []Handler {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if len(s.handlers) == 0 {
		return nil
	}
	handlers := make([]Handler, len(s.handlers))
	copy(handlers, s.handlers)
	return handlers
}

// ListenAndServe listens for Direct TCP connections on addr and serves them. An
// addr with no port uses port 445. It blocks until the server is closed or the
// listener fails.
func (s *Server) ListenAndServe(addr string) error {
	listener, err := transport.ListenTCP(addr)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// ListenAndServeNBT listens for NetBIOS over TCP connections on addr and serves
// them, answering to the given CALLED NetBIOS names (nil answers to any). An
// addr with no port uses port 139. It blocks until the server is closed or the
// listener fails.
func (s *Server) ListenAndServeNBT(addr string, acceptedNames []string) error {
	listener, err := transport.ListenNBT(addr, acceptedNames)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// Serve accepts connections from a listener until the server is closed or the
// listener fails, serving each on its own goroutine. The listener is closed when
// Serve returns.
//
// Parameters:
//   - listener: the listener to accept from
//
// Returns:
//   - nil if the server was closed, or the listener's error otherwise
func (s *Server) Serve(listener transport.Listener) error {
	if listener == nil {
		return fmt.Errorf("cannot serve a nil listener")
	}
	if !s.trackListener(listener) {
		listener.Close()
		return fmt.Errorf("server is closed")
	}
	defer s.untrackListener(listener)

	logger.Infof("SMB1 server listening on %s", listener.Addr())

	for {
		conn, remote, err := listener.Accept()
		if err != nil {
			if s.isClosed() {
				return nil
			}
			return fmt.Errorf("failed to accept an SMB connection: %v", err)
		}

		if !s.acquireSlot() {
			logger.Warnf("SMB1 server: refusing %s, already serving the maximum of %d connections",
				remote, s.config.MaxConnections)
			conn.Close()
			continue
		}

		if s.config.Timeout > 0 {
			conn.SetTimeout(s.config.Timeout)
		}

		connection := newConnection(s, conn, remote)
		if !s.trackConnection(connection) {
			// The server was closed between Accept and here.
			s.releaseSlot()
			connection.Close()
			return nil
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer s.releaseSlot()
			defer s.untrackConnection(connection)
			connection.serve()
		}()
	}
}

// Listening reports whether the server currently has a listener.
func (s *Server) Listening() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.listeners) > 0
}

// Addr returns the address of the server's first listener, or nil if it is not
// listening.
func (s *Server) Addr() net.Addr {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if len(s.listeners) == 0 {
		return nil
	}
	return s.listeners[0].Addr()
}

// Connections returns the number of connections currently being served.
func (s *Server) Connections() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.conns)
}

// Close stops the server: it closes every listener so the accept loops return,
// closes every live connection so its receive loop returns, and waits for the
// connection goroutines to finish. It is safe to call more than once.
func (s *Server) Close() error {
	s.mutex.Lock()
	if s.closed {
		s.mutex.Unlock()
		s.wg.Wait()
		return nil
	}
	s.closed = true

	listeners := s.listeners
	s.listeners = nil
	conns := make([]*Connection, 0, len(s.conns))
	for conn := range s.conns {
		conns = append(conns, conn)
	}
	s.mutex.Unlock()

	var firstErr error
	for _, listener := range listeners {
		if err := listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	// Closing the transports is what unblocks the connection goroutines, which
	// are otherwise parked in a read that may have no deadline.
	for _, conn := range conns {
		conn.Close()
	}

	s.wg.Wait()
	return firstErr
}

// isClosed reports whether Close has been called.
func (s *Server) isClosed() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.closed
}

// trackListener records a listener so Close can stop it, reporting false if the
// server is already closed.
func (s *Server) trackListener(listener transport.Listener) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return false
	}
	s.listeners = append(s.listeners, listener)
	return true
}

// untrackListener forgets a listener whose Serve has returned.
func (s *Server) untrackListener(listener transport.Listener) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for i, tracked := range s.listeners {
		if tracked == listener {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			break
		}
	}
}

// trackConnection records a live connection so Close can tear it down,
// reporting false if the server is already closed.
func (s *Server) trackConnection(conn *Connection) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return false
	}
	s.conns[conn] = struct{}{}
	return true
}

// untrackConnection forgets a connection whose serve loop has returned.
func (s *Server) untrackConnection(conn *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.conns, conn)
}

// acquireSlot takes a connection slot, reporting false when the server is
// already at MaxConnections. It always succeeds when the limit is unbounded.
func (s *Server) acquireSlot() bool {
	if s.slots == nil {
		return true
	}
	select {
	case s.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// releaseSlot returns a connection slot.
func (s *Server) releaseSlot() {
	if s.slots == nil {
		return
	}
	select {
	case <-s.slots:
	default:
	}
}
