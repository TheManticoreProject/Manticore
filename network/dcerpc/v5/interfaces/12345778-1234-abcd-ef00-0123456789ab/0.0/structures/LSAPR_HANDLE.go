package structures

// LSAPR_HANDLE is an RPC context handle (LSAPR_HANDLE): a 4-byte attributes field
// followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2, [MS-LSAD] 2.2.2.1).
type LSAPR_HANDLE [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// LsarClose).
func (h LSAPR_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
