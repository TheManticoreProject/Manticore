package server

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
)

// Config configures a Server. Zero values select the defaults.
type Config struct {
	// ServerName, DomainName, DNSComputerName and DNSDomainName are advertised
	// in the NTLM CHALLENGE. Empty names are omitted.
	ServerName      string
	DomainName      string
	DNSComputerName string
	DNSDomainName   string

	// ServerGUID identifies the server in NEGOTIATE. Zero generates a random one.
	ServerGUID [16]byte

	// Dialects are the dialects offered. Nil offers 2.0.2 through 3.1.1.
	Dialects []dialects.Dialect

	// Authenticator resolves an NTLM identity to its NT hash. Nil verifies no
	// NTLM logon, which only AllowGuest can then admit.
	Authenticator func(domain, username string) (ntHash [16]byte, ok bool)

	// Kerberos, when set, accepts Kerberos logons with these acceptor options.
	Kerberos *gssapi.AcceptOptions

	// AllowGuest admits an identity that cannot be verified as a guest, and
	// AllowAnonymous a null session. Neither has a key, so neither can sign.
	AllowGuest     bool
	AllowAnonymous bool

	// RequireSigning demands signed requests on every session.
	RequireSigning bool

	// RequireEncryption encrypts every SMB 3.x session and refuses clients that
	// cannot encrypt.
	RequireEncryption bool

	// MaxReadSize and MaxWriteSize bound a single READ or WRITE. Zero selects
	// DefaultMaxIOSize.
	MaxReadSize  uint32
	MaxWriteSize uint32

	// MaxSessionsPerConnection, MaxTreesPerSession, MaxOpensPerConnection and
	// MaxNotifiesPerConnection bound what one client may hold, so it cannot
	// exhaust identifiers, handles or the watches behind pending CHANGE_NOTIFY
	// requests. Zero applies the default.
	MaxSessionsPerConnection int
	MaxTreesPerSession       int
	MaxOpensPerConnection    int
	MaxNotifiesPerConnection int

	// MaxConnections bounds the connections served at once; one arriving at the
	// limit is closed. Zero applies the default and a negative value means
	// unbounded.
	MaxConnections int

	// HandshakeTimeout bounds each read on a connection that has no established
	// session yet, so a peer that never authenticates is dropped. IdleTimeout
	// bounds each read on a connection that holds no open, the way Windows
	// disconnects idle sessions; a client with a file open or a CHANGE_NOTIFY
	// pending may stay quiet for as long as it likes. Zero applies the default
	// and a negative value means no bound.
	HandshakeTimeout time.Duration
	IdleTimeout      time.Duration

	// Timeout bounds every read on a connection, whatever its state; it caps
	// HandshakeTimeout and IdleTimeout. Zero means no bound.
	Timeout time.Duration
}

// Configuration defaults applied by NewServer when a field is left zero.
const (
	DefaultMaxIOSize                = 1 << 20
	DefaultMaxSessionsPerConnection = 64
	DefaultMaxTreesPerSession       = 64
	DefaultMaxOpensPerConnection    = 1024
	DefaultMaxNotifiesPerConnection = 64
	DefaultMaxConnections           = 1024
	DefaultHandshakeTimeout         = time.Minute
	DefaultIdleTimeout              = 15 * time.Minute
)

// Share is a disk share. Its FileSystem comes from the SMB1 server package,
// whose backends both servers share.
type Share struct {
	// Name is matched case-insensitively.
	Name string

	// FS is the storage behind the share.
	FS smb1.FileSystem

	// ReadOnly refuses every modifying operation.
	ReadOnly bool
}

// Server is an SMB 2/3 server.
type Server struct {
	config     Config
	oplocks    *oplockTable
	shareModes *shareModes

	mutex     sync.RWMutex
	handlers  []Handler
	shares    map[string]*Share
	listeners []transport.Listener
	conns     map[*Connection]struct{}
	closed    bool

	// slots holds one token per connection served when MaxConnections is positive.
	slots chan struct{}

	nextSessionID atomic.Uint64
	wg            sync.WaitGroup
}

// NewServer validates config and returns a server ready to serve.
func NewServer(config Config) (*Server, error) {
	if config.Dialects == nil {
		config.Dialects = []dialects.Dialect{
			dialects.SMB2_DIALECT_2_0_2, dialects.SMB2_DIALECT_2_1_0,
			dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2, dialects.SMB2_DIALECT_3_1_1,
		}
	}
	for _, d := range config.Dialects {
		if _, ok := dialectRank[d]; !ok {
			return nil, fmt.Errorf("unsupported dialect %s", d)
		}
	}
	for name, limit := range map[string]*int{
		"MaxSessionsPerConnection": &config.MaxSessionsPerConnection,
		"MaxTreesPerSession":       &config.MaxTreesPerSession,
		"MaxOpensPerConnection":    &config.MaxOpensPerConnection,
		"MaxNotifiesPerConnection": &config.MaxNotifiesPerConnection,
	} {
		if *limit < 0 {
			return nil, fmt.Errorf("%s cannot be negative (got %d)", name, *limit)
		}
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("Timeout cannot be negative")
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = DefaultMaxConnections
	}
	if config.HandshakeTimeout == 0 {
		config.HandshakeTimeout = DefaultHandshakeTimeout
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = DefaultIdleTimeout
	}
	if config.MaxSessionsPerConnection == 0 {
		config.MaxSessionsPerConnection = DefaultMaxSessionsPerConnection
	}
	if config.MaxTreesPerSession == 0 {
		config.MaxTreesPerSession = DefaultMaxTreesPerSession
	}
	if config.MaxOpensPerConnection == 0 {
		config.MaxOpensPerConnection = DefaultMaxOpensPerConnection
	}
	if config.MaxNotifiesPerConnection == 0 {
		config.MaxNotifiesPerConnection = DefaultMaxNotifiesPerConnection
	}
	if config.MaxReadSize == 0 {
		config.MaxReadSize = DefaultMaxIOSize
	}
	if config.MaxWriteSize == 0 {
		config.MaxWriteSize = DefaultMaxIOSize
	}
	if config.ServerGUID == ([16]byte{}) {
		if _, err := rand.Read(config.ServerGUID[:]); err != nil {
			return nil, fmt.Errorf("failed to generate a server GUID: %v", err)
		}
	}

	s := &Server{
		config:     config,
		oplocks:    newOplockTable(),
		shareModes: newShareModes(),
		shares:     make(map[string]*Share),
		conns:      make(map[*Connection]struct{}),
	}
	if config.MaxConnections > 0 {
		s.slots = make(chan struct{}, config.MaxConnections)
	}
	// SessionIds are unpredictable so a client cannot guess another's.
	var seed [8]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return nil, fmt.Errorf("failed to seed session identifiers: %v", err)
	}
	s.nextSessionID.Store(binary.LittleEndian.Uint64(seed[:]))
	return s, nil
}

// newSessionID returns a fresh, non-zero SessionId.
func (s *Server) newSessionID() uint64 {
	for {
		if id := s.nextSessionID.Add(1); id != 0 {
			return id
		}
	}
}

// AddShare registers a disk share.
func (s *Server) AddShare(share *Share) error {
	if share == nil || share.Name == "" || share.FS == nil {
		return fmt.Errorf("a share needs a name and a file system")
	}
	if strings.ContainsAny(share.Name, `\/:*?"<>|`) || strings.EqualFold(share.Name, ipcShareName) {
		return fmt.Errorf("invalid share name %q", share.Name)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := strings.ToUpper(share.Name)
	if _, taken := s.shares[key]; taken {
		return fmt.Errorf("a share named %q is already registered", share.Name)
	}
	s.shares[key] = share
	return nil
}

// share returns the share registered under name, or nil.
func (s *Server) share(name string) *Share {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.shares[strings.ToUpper(name)]
}

// ListenAndServe serves Direct TCP on addr (port 445 when omitted).
func (s *Server) ListenAndServe(addr string) error {
	listener, err := transport.ListenTCP(addr)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// ListenAndServeNBT serves NetBIOS over TCP on addr (port 139 when omitted).
func (s *Server) ListenAndServeNBT(addr string, acceptedNames []string) error {
	listener, err := transport.ListenNBT(addr, acceptedNames)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// Serve accepts connections on listener until it fails or the server closes.
func (s *Server) Serve(listener transport.Listener) error {
	s.mutex.Lock()
	if s.closed {
		s.mutex.Unlock()
		listener.Close()
		return fmt.Errorf("server is closed")
	}
	s.listeners = append(s.listeners, listener)
	s.mutex.Unlock()

	logger.Infof("SMB2 server listening on %s", listener.Addr())
	for {
		t, remote, err := listener.Accept()
		if err != nil {
			if s.isClosed() {
				return nil
			}
			return fmt.Errorf("failed to accept an SMB connection: %v", err)
		}
		go func() {
			if err := s.ServeConn(t, remote); err != nil {
				logger.Debugf("SMB2 server: refusing %s: %v", remote, err)
			}
		}()
	}
}

// ServeConn serves one already established transport until it closes, which
// lets a caller that owns the socket, such as a proxy, hand it to the server.
func (s *Server) ServeConn(t transport.Transport, remote net.Addr) error {
	if s.slots != nil {
		select {
		case s.slots <- struct{}{}:
			defer func() { <-s.slots }()
		default:
			t.Close()
			return fmt.Errorf("already serving the maximum of %d connections", s.config.MaxConnections)
		}
	}
	conn := newConnection(s, t, remote)
	s.mutex.Lock()
	if s.closed {
		s.mutex.Unlock()
		t.Close()
		return fmt.Errorf("server is closed")
	}
	s.conns[conn] = struct{}{}
	s.wg.Add(1)
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.conns, conn)
		s.mutex.Unlock()
		s.wg.Done()
	}()
	conn.serve()
	return nil
}

// Addr returns the address of the first listener, or nil.
func (s *Server) Addr() net.Addr {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if len(s.listeners) == 0 {
		return nil
	}
	return s.listeners[0].Addr()
}

// Close stops every listener and connection and waits for them to finish.
func (s *Server) Close() error {
	s.mutex.Lock()
	s.closed = true
	listeners := s.listeners
	s.listeners = nil
	for conn := range s.conns {
		conn.transport.Close()
	}
	s.mutex.Unlock()

	var firstErr error
	for _, listener := range listeners {
		if err := listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	s.wg.Wait()
	return firstErr
}

func (s *Server) isClosed() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.closed
}
