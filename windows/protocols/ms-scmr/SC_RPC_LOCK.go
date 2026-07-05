package msscmr

// SC_RPC_LOCK is an RPC context handle ([MS-SCMR] 2.2.5): a 4-byte attributes field
// followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2). It represents a lock on
// the SCM database acquired by RLockServiceDatabase and released by RUnlockServiceDatabase.
type SC_RPC_LOCK [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// RUnlockServiceDatabase).
func (h SC_RPC_LOCK) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
