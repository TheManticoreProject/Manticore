package mspcq

// RPC_HQUERY is the PerflibV2 query context handle ([MS-PCQ] 2.2.1). The IDL declares it
// as [context_handle] HANDLE; on the wire an RPC context handle is a 4-byte attributes
// field followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2).
type RPC_HQUERY [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// PerflibV2CloseQueryHandle, which the server zeroes on return).
func (h RPC_HQUERY) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
