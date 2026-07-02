package mscmpo

// PCONTEXT_HANDLE is the RPC context handle used by the IXnRemote methods
// ([MS-CMPO] 2.2.2; [C706]). The IDL declares it as [context_handle] void*; on the wire
// it is the 20-byte ndr_context_handle representation — a 4-byte attributes field
// followed by a 16-byte GUID ([MS-RPCE] 2.3.2.2) — transmitted by value, never as a
// referent pointer.
type PCONTEXT_HANDLE [20]byte

// PPCONTEXT_HANDLE is a pointer to a PCONTEXT_HANDLE (the IDL PCONTEXT_HANDLE*), used for
// the [out]/[in,out] handle parameters of BuildContext, BuildContextW, and
// TearDownContext. Because a context handle is always transmitted as its 20-byte value,
// the extra indirection is a C-language artifact with no separate NDR referent, so it is
// modeled as the same value type.
type PPCONTEXT_HANDLE = PCONTEXT_HANDLE

// IsZero reports whether the handle is all zeros — for example, after the server tears a
// context down and returns the nulled handle.
func (h PCONTEXT_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
