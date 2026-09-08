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
}

// defaultVersionMajor and defaultVersionMinor are the version reported when the
// caller names none. See Options.VersionMajor.
const (
	defaultVersionMajor = 6
	defaultVersionMinor = 1
)

// Handler must satisfy the server's pipe contract, which is the whole point of
// the package: a mismatch is a compile error here rather than a share that
// refuses every pipe at run time.
var _ server.PipeHandler = (*Handler)(nil)

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
	// written once by New and only read afterwards.
	endpoints map[string]*rpcserver.Dispatcher

	// mutex guards opened, the count of opens outstanding per pipe.
	mutex  sync.Mutex
	opened map[string]int
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

	return &Handler{
		endpoints: map[string]*rpcserver.Dispatcher{
			"srvsvc": rpcserver.New(&srvsvcService{shares: shares}),
			"wkssvc": rpcserver.New(&wkssvcService{
				serverName:   options.ServerName,
				domainName:   options.DomainName,
				versionMajor: versionMajor,
				versionMinor: versionMinor,
			}),
		},
		opened: map[string]int{},
	}
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

// OpenPipe reports whether a pipe exists under this handler.
func (h *Handler) OpenPipe(name string) error {
	endpoint := normalisePipeName(name)
	if _, served := h.endpoints[endpoint]; !served {
		return fmt.Errorf("pipe %q is not served", name)
	}

	h.mutex.Lock()
	h.opened[endpoint]++
	h.mutex.Unlock()

	logger.Debugf("rpcpipe: opened %q", endpoint)
	return nil
}

// Transact answers one RPC exchange on a pipe.
//
// Parameters:
//   - name: the pipe, without its leading separator or "PIPE" element
//   - input: the request PDU the client wrote
//   - maxOutput: the largest answer the caller will take in one exchange
//
// Returns:
//   - The reply PDUs, whether more of the answer remains, and an error only when
//     the pipe is not served or the request is not a PDU at all
func (h *Handler) Transact(name string, input []byte, maxOutput int) ([]byte, bool, error) {
	endpoint := normalisePipeName(name)
	dispatcher, served := h.endpoints[endpoint]
	if !served {
		return nil, false, fmt.Errorf("pipe %q is not served", name)
	}

	reply, err := dispatcher.Handle(input)
	if err != nil {
		return nil, false, fmt.Errorf("pipe %q: %w", endpoint, err)
	}

	// An answer past the caller's ceiling is cut and reported as incomplete,
	// which is the contract: the caller tells the client more remains and the
	// client reads the rest off the handle.
	if maxOutput > 0 && len(reply) > maxOutput {
		logger.Debugf("rpcpipe: %q answered with %d bytes, cut to the %d the caller allows",
			endpoint, len(reply), maxOutput)
		return reply[:maxOutput], true, nil
	}

	logger.Debugf("rpcpipe: %q answered %d request bytes with %d", endpoint, len(input), len(reply))
	return reply, false, nil
}

// ClosePipe releases what OpenPipe recorded.
func (h *Handler) ClosePipe(name string) error {
	endpoint := normalisePipeName(name)

	h.mutex.Lock()
	if h.opened[endpoint] > 0 {
		h.opened[endpoint]--
	}
	h.mutex.Unlock()
	return nil
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
	if upper := strings.ToUpper(trimmed); strings.HasPrefix(upper, "PIPE/") {
		trimmed = trimmed[len("PIPE/"):]
	}
	return strings.ToLower(strings.Trim(trimmed, "/"))
}
