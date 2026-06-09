package structures

// SC_RPC_HANDLE is an RPC context handle ([MS-SCMR] 2.2.4): a 4-byte attributes field
// followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2). It is the handle to an
// SCM database or service object returned by ROpenSCManager*/ROpenService*/RCreateService*.
type SC_RPC_HANDLE [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// RCloseServiceHandle).
func (h SC_RPC_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
