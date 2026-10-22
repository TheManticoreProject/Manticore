package server

import (
	"net"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// The rules below are shared with the SMB 2/3 server, which serves the same
// backends, so both protocols apply one containment check and one mapping.

// ResolvePath converts a client path into the share-relative form a FileSystem
// is promised, refusing anything that could escape the share.
func ResolvePath(raw string) (string, error) {
	return resolvePath(raw)
}

// OpenFlagsFor translates a create request's access, disposition and options
// into what a backend needs.
func OpenFlagsFor(desiredAccess, createDisposition, createOptions uint32) (OpenFlags, nt_status.NT_STATUS) {
	return openFlagsFor(desiredAccess, createDisposition, createOptions)
}

// StatusForFSError maps a backend error to the status a client is answered with.
func StatusForFSError(err error) nt_status.NT_STATUS {
	return statusForFSError(err)
}

// Record captures an NTLM exchange another server answered, with this handler's
// reporting and deduplication.
func (h *CaptureHandler) Record(remote net.Addr, accept *spnego.AcceptContext) {
	h.record(remote, accept)
}
