package msrrp

// RPC_HKEY is an RPC context handle ([MS-RRP] 2.2.3): a 4-byte attributes field followed
// by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2). It is the handle to an open
// registry key returned by the Open*/BaseRegCreateKey/BaseRegOpenKey methods.
type RPC_HKEY [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// BaseRegCloseKey).
func (h RPC_HKEY) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
