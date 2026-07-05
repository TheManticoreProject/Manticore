package msswn

// PCONTEXT_HANDLE is the RPC context handle returned by WitnessrRegister /
// WitnessrRegisterEx that identifies the client's registration on the Witness server
// ([MS-SWN] 2.2.1.1). The IDL types it as [context_handle] void *; on the wire it is a
// 20-octet context handle (a 4-byte attributes field followed by a 16-byte GUID,
// [MS-RPCE] 2.3.2.2), transmitted inline with no referent id.
type PCONTEXT_HANDLE [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// WitnessrUnRegister, which the server returns nulled out).
func (h PCONTEXT_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
