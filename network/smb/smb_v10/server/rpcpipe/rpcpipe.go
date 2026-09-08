package rpcpipe

import (
	"fmt"
	"strings"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/rpcserver"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
)

// ShareEntry is one share as an enumeration reports it.
//
// It is this package's own shape rather than the SMB server's, so a caller can
// serve a share list that has nothing to do with a running server — a test, or a
// host whose shares live elsewhere — and the services below depend on a list and
// not on where the list came from.
type ShareEntry struct {
	// Name is the share name a client connects to.
	Name string

	// Comment is the remark reported beside it.
	Comment string

	// Type is what the share is, as an STYPE_* value.
	Type uint32
}

// Options configures a Handler.
type Options struct {
	// ServerName is the host name wkssvc reports. A client displays it.
	ServerName string

	// DomainName is the workgroup or domain wkssvc reports.
	DomainName string

	// VersionMajor and VersionMinor are the operating system version wkssvc
	// reports. Both zero reports 6.1, the version of Windows 7 and Server
	// 2008 R2, because a client displays the number and one of zero reads as
	// missing rather than as a version.
	VersionMajor uint32
	VersionMinor uint32

	// Shares supplies the shares an enumeration reports, called afresh on each
	// call so a share added later appears without the handler being told.
	//
	// Nil reports an empty list, which is what an endpoint with no share source
	// should say rather than failing.
	Shares func() []ShareEntry

	// Services are further RPC interfaces to serve, keyed by pipe name.
	//
	// A name that is one of the built-in pipes adds its interfaces to that pipe,
	// and any other name adds a pipe. A pipe may carry several interfaces: a bind
	// picks one of them and the requests after it are dispatched by the
	// presentation context it negotiated, so a caller can serve its own interface
	// over IPC$ without writing a PipeHandler.
	Services map[string][]rpcserver.Service
}

// defaultVersionMajor and defaultVersionMinor are the version reported when the
// caller names none. See Options.VersionMajor.
const (
	defaultVersionMajor = 6
	defaultVersionMinor = 1
)

// Handler and Session must satisfy the server's pipe contract, which is the whole
// point of the package: a mismatch is a compile error here rather than a share
// that refuses every pipe at run time.
var (
	_ server.PipeHandler = (*Handler)(nil)
	_ server.PipeSession = (*Session)(nil)
)

// Handler answers RPC calls on the pipes of an IPC$ share.
//
// It satisfies the PipeHandler interface of network/smb/smb_v10/server, so a
// share carrying one serves srvsvc and wkssvc:
//
//	srv.AddShare(&server.Share{
//	    Name:  "IPC$",
//	    Type:  server.ShareTypeNamedPipe,
//	    Pipes: rpcpipe.ForServer(srv, rpcpipe.Options{ServerName: "MANTICORE"}),
//	})
//
// A Handler is safe for concurrent use, which it has to be: one is reachable from
// every connection the server accepts.
type Handler struct {
	// endpoints maps a normalised pipe name to the dispatcher answering it. It is
	// written once by New and only read afterwards, which is what makes a Handler
	// safe to share: everything a client establishes lives on the Session its
	// open returns.
	endpoints map[string]*rpcserver.Dispatcher
}

// New builds a Handler serving srvsvc and wkssvc.
//
// Parameters:
//   - options: the names to report and where to read shares from
//
// Returns:
//   - The handler
func New(options Options) *Handler {
	shares := options.Shares
	if shares == nil {
		shares = func() []ShareEntry { return nil }
	}

	versionMajor, versionMinor := options.VersionMajor, options.VersionMinor
	if versionMajor == 0 && versionMinor == 0 {
		versionMajor, versionMinor = defaultVersionMajor, defaultVersionMinor
	}

	served := map[string][]rpcserver.Service{
		"srvsvc": {&srvsvcService{shares: shares}},
		"wkssvc": {&wkssvcService{
			serverName:   options.ServerName,
			domainName:   options.DomainName,
			versionMajor: versionMajor,
			versionMinor: versionMinor,
		}},
	}
	for name, services := range options.Services {
		endpoint := normalisePipeName(name)
		if endpoint == "" || len(services) == 0 {
			continue
		}
		served[endpoint] = append(served[endpoint], services...)
	}

	endpoints := make(map[string]*rpcserver.Dispatcher, len(served))
	for endpoint, services := range served {
		endpoints[endpoint] = rpcserver.New(services...)
	}
	return &Handler{endpoints: endpoints}
}

// ForServer builds a Handler whose share enumeration reports a server's own
// shares.
//
// It fills Options.Shares from the server, leaving anything the caller already
// set in place, so a share registered with AddShare appears in an enumeration
// without being described twice.
//
// Parameters:
//   - srv: the server whose shares an enumeration should report
//   - options: the rest of the configuration
//
// Returns:
//   - The handler
func ForServer(srv *server.Server, options Options) *Handler {
	if srv != nil && options.Shares == nil {
		options.Shares = func() []ShareEntry { return sharesOf(srv) }
	}
	return New(options)
}

// sharesOf converts a server's registered shares into enumeration entries.
func sharesOf(srv *server.Server) []ShareEntry {
	registered := srv.Shares()
	entries := make([]ShareEntry, 0, len(registered))
	for _, share := range registered {
		if share == nil {
			continue
		}
		entries = append(entries, ShareEntry{
			Name:    share.Name,
			Comment: share.Comment,
			Type:    ShareTypeOf(share.Name, share.Type),
		})
	}
	return entries
}

// OpenPipe opens one instance of a pipe.
//
// The session it returns owns this client's RPC association: what its bind
// negotiates is remembered for the requests that follow, and two opens of the
// same pipe negotiate independently.
//
// Parameters:
//   - name: the pipe, in any of the forms a client names one
//
// Returns:
//   - The session, or an error naming the pipe when this handler does not serve
//     it
func (h *Handler) OpenPipe(name string) (server.PipeSession, error) {
	endpoint := normalisePipeName(name)
	dispatcher, served := h.endpoints[endpoint]
	if !served {
		// "not found" is the wording the server maps to
		// STATUS_OBJECT_NAME_NOT_FOUND, which is what a client asking for a pipe
		// that does not exist should be told.
		return nil, fmt.Errorf("pipe %q not found", name)
	}

	logger.Debugf("rpcpipe: opened %q", endpoint)
	return &Session{name: endpoint, association: dispatcher.Open()}, nil
}

// Session is one open of a pipe served by a Handler.
//
// It holds the RPC association for that open, which is the whole reason a session
// exists: a bind establishes the presentation contexts and the fragment size the
// requests after it are read against, and those belong to one client's open rather
// than to the pipe.
type Session struct {
	// name is the pipe this session is an open of, for the log lines and errors
	// that name it.
	name string

	// association is this open's RPC conversation with the endpoint.
	association *rpcserver.Association

	// mutex guards closed. A session belongs to one open on one connection, so
	// the server does not use one from two goroutines; the guard is here so a
	// caller that does still gets an error rather than a race.
	mutex  sync.Mutex
	closed bool
}

// Transact answers one RPC exchange on this open.
//
// Parameters:
//   - input: the request PDU the client wrote
//   - maxOutput: the largest answer the caller will take in one exchange
//
// Returns:
//   - The reply PDUs, whether more of the answer remains, and an error only when
//     the session is closed or the request is not a PDU at all
func (s *Session) Transact(input []byte, maxOutput int) ([]byte, bool, error) {
	if s.isClosed() {
		return nil, false, fmt.Errorf("pipe %q: the session is closed", s.name)
	}

	reply, err := s.association.Handle(input)
	if err != nil {
		return nil, false, fmt.Errorf("pipe %q: %w", s.name, err)
	}

	// An answer past the caller's ceiling is cut and reported as incomplete,
	// which is the contract: the caller tells the client more remains and the
	// client reads the rest off the handle.
	if maxOutput > 0 && len(reply) > maxOutput {
		logger.Debugf("rpcpipe: %q answered with %d bytes, cut to the %d the caller allows",
			s.name, len(reply), maxOutput)
		return reply[:maxOutput], true, nil
	}

	logger.Debugf("rpcpipe: %q answered %d request bytes with %d", s.name, len(input), len(reply))
	return reply, false, nil
}

// Close releases the session. It is safe to call more than once.
func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.closed = true
	return nil
}

// Bound reports how many presentation contexts this open has negotiated, which is
// zero until its client binds.
func (s *Session) Bound() int {
	return s.association.Bound()
}

// isClosed reports whether Close has been called.
func (s *Session) isClosed() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.closed
}

// Pipes returns the pipe names this handler serves, sorted for a stable listing.
func (h *Handler) Pipes() []string {
	names := make([]string, 0, len(h.endpoints))
	for name := range h.endpoints {
		names = append(names, name)
	}
	sortStrings(names)
	return names
}

// sortStrings sorts in place. It is an insertion sort because the slice holds two
// elements and a dependency on sort for that would be noise.
func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

// normalisePipeName reduces the forms a client names a pipe in to the endpoint
// name.
//
// The SMB server already strips the leading separator and the PIPE element, but
// a caller may hand over any of the forms a client sends, and the comparison is
// case-insensitive because pipe names are.
func normalisePipeName(name string) string {
	trimmed := strings.Trim(strings.ReplaceAll(name, `\`, "/"), "/")

	upper := strings.ToUpper(trimmed)
	switch {
	case upper == "PIPE":
		// The pipe directory itself, which names no pipe. Without this it would
		// normalise to "pipe" and a caller could register an endpoint under it.
		return ""
	case strings.HasPrefix(upper, "PIPE/"):
		trimmed = trimmed[len("PIPE/"):]
	}
	return strings.ToLower(strings.Trim(trimmed, "/"))
}
