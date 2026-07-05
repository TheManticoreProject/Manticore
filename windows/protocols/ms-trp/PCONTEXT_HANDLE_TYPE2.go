package mstrp

// PCONTEXT_HANDLE_TYPE2 is the RPC context handle the remotesp (reverse/callback)
// interface returns from RemoteSPAttach and threads through RemoteSPEventProc /
// RemoteSPDetach ([MS-TRP] 2.2.2, 3.1.4). The IDL types it as [context_handle] void *;
// on the wire it is a 20-octet context handle (a 4-octet attributes field followed by a
// 16-octet GUID, [MS-RPCE] 2.3.2.2), transmitted inline with no referent id.
type PCONTEXT_HANDLE_TYPE2 [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// RemoteSPDetach, which the server returns nulled out).
func (h PCONTEXT_HANDLE_TYPE2) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
