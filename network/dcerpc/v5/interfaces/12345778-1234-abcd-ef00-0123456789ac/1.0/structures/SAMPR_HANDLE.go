package structures

// SAMPR_HANDLE is an RPC context handle ([MS-SAMR] 2.2.3.2): a 4-byte attributes
// field followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2).
type SAMPR_HANDLE [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// SamrCloseHandle).
func (h SAMPR_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
