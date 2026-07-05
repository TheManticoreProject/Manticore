package msscmr

// SC_NOTIFY_RPC_HANDLE is an RPC context handle ([MS-SCMR] 2.2.6): a 4-byte attributes
// field followed by a 16-byte GUID, 20 bytes total ([MS-RPCE] 2.3.2.2). It identifies a
// status-change notification registration returned by RNotifyServiceStatusChange and closed
// by RCloseNotifyHandle.
type SC_NOTIFY_RPC_HANDLE [20]byte

// IsZero reports whether the handle is all zeros (for example, after a successful
// RCloseNotifyHandle).
func (h SC_NOTIFY_RPC_HANDLE) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
