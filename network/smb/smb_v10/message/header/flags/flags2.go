package flags

// SMB Header Flags2
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
const (
	// Long Names Allowed: Long file names are allowed in the response
	FLAGS2_LONG_NAMES_ALLOWED = 1
	// Extended Attributes: Extended attributes are supported
	FLAGS2_EXTENDED_ATTRIBUTES = 1 << 1
	// Security Signatures: Security signatures are supported
	FLAGS2_SMB_SECURITY_SIGNATURE = 1 << 2
	// Compressed: Compression is requested
	FLAGS2_COMPRESSED = 1 << 3
	// Reserved: Reserved for future use
	FLAGS2_RESERVED_4 = 1 << 4
	// Security Signatures Required: Security signatures are required
	FLAGS2_SMB_SECURITY_SIGNATURE_REQUIRED = 1 << 5
	// Reserved: Reserved for future use
	FLAGS2_RESERVED_6 = 1 << 6
	// Long Names Used: Path names in request are long file names
	FLAGS2_LONG_NAMES_USED = 1 << 7
	// Reserved: Reserved for future use
	FLAGS2_RESERVED_8 = 1 << 8
	// Reserved: Reserved for future use
	FLAGS2_RESERVED_9 = 1 << 9
	// Reparse Path: The request uses a @GMT reparse path
	FLAGS2_REPARSE_PATH = 1 << 10
	// Extended Security Negotiation: Extended security negotiation is supported
	FLAGS2_EXTENDED_SECURITY = 1 << 11
	// Dfs: Resolve pathnames with Dfs
	FLAGS2_DFS = 1 << 12
	// Execute-only Reads: Don't permit reads if execute-only
	FLAGS2_PAGING_IO = 1 << 13
	// Error Code Type: Error codes are NT error codes
	FLAGS2_NT_STATUS_ERROR_CODES = 1 << 14
	// Unicode Strings: Strings are Unicode
	FLAGS2_UNICODE = 1 << 15
)
