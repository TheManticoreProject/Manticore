// Package mstrp holds the NDR wire structures for the Telephony Remote Protocol
// ([MS-TRP]): the RPC context-handle types shared by the tapsrv and remotesp interfaces.
//
// The interface descriptors and method stubs live under network/dcerpc/interfaces
// (keyed by UUID/version); this package holds only the data types, so the dependency
// direction stays functions -> structures and never the reverse.
package mstrp

// PCONTEXT_HANDLE_TYPE is the RPC context handle the tapsrv interface returns from
// ClientAttach and threads through ClientRequest / ClientDetach ([MS-TRP] 2.2.1,
// 3.2.4). The IDL types it as [context_handle] void *; on the wire it is a 20-octet
// context handle (a 4-octet attributes field followed by a 16-octet GUID,
// [MS-RPCE] 2.3.2.2), transmitted inline with no referent id.
type PCONTEXT_HANDLE_TYPE [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// ClientDetach, which the server returns nulled out).
func (h PCONTEXT_HANDLE_TYPE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
