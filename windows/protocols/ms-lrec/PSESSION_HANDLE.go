package mslrec

// PSESSION_HANDLE is the RPC context handle that references an active event session on
// the server ([MS-LREC] 2.2.1.1). The IDL declares it as [context_handle] void*; on the
// wire it is the 20-byte ndr_context_handle representation — a 4-byte attributes field
// followed by a 16-byte GUID ([MS-RPCE] 2.3.2.2) — transmitted by value, never as a
// referent pointer.
//
// RpcNetEventOpenSession and RpcNetEventCloseSession declare the parameter as
// PSESSION_HANDLE* (a pointer to the handle). Because a context handle is always
// transmitted as its 20-byte value, that extra indirection is a C-language artifact with
// no separate NDR referent, so both the by-value ([in]) and by-pointer ([out]/[in,out])
// parameters are modeled as this same value type.
type PSESSION_HANDLE [20]byte

// IsZero reports whether the handle is all zeros — for example, the nulled handle a
// server returns from RpcNetEventCloseSession after tearing the session down.
func (h PSESSION_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
