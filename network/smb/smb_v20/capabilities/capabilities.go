package capabilities

// Capabilities is the 4-byte protocol-capabilities field carried in SMB2
// NEGOTIATE Request/Response and SESSION_SETUP Request.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/63abf97c-0d09-47e2-88d6-6bfa552949a5
type Capabilities uint32

const (
	// SMB2_GLOBAL_CAP_DFS indicates support for the Distributed File System.
	SMB2_GLOBAL_CAP_DFS Capabilities = 0x00000001
	// SMB2_GLOBAL_CAP_LEASING indicates support for leasing (not valid for SMB 2.0.2).
	SMB2_GLOBAL_CAP_LEASING Capabilities = 0x00000002
	// SMB2_GLOBAL_CAP_LARGE_MTU indicates support for multi-credit operations (not valid for SMB 2.0.2).
	SMB2_GLOBAL_CAP_LARGE_MTU Capabilities = 0x00000004
	// SMB2_GLOBAL_CAP_MULTI_CHANNEL indicates support for multiple channels per session.
	SMB2_GLOBAL_CAP_MULTI_CHANNEL Capabilities = 0x00000008
	// SMB2_GLOBAL_CAP_PERSISTENT_HANDLES indicates support for persistent handles.
	SMB2_GLOBAL_CAP_PERSISTENT_HANDLES Capabilities = 0x00000010
	// SMB2_GLOBAL_CAP_DIRECTORY_LEASING indicates support for directory leasing.
	SMB2_GLOBAL_CAP_DIRECTORY_LEASING Capabilities = 0x00000020
	// SMB2_GLOBAL_CAP_ENCRYPTION indicates support for encryption.
	SMB2_GLOBAL_CAP_ENCRYPTION Capabilities = 0x00000040
	// SMB2_GLOBAL_CAP_NOTIFICATIONS indicates support for server-to-client notifications.
	SMB2_GLOBAL_CAP_NOTIFICATIONS Capabilities = 0x00000080
)

// Has reports whether all bits in c2 are set in c.
func (c Capabilities) Has(c2 Capabilities) bool {
	return c&c2 == c2
}
