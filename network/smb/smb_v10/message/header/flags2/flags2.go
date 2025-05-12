package flags2

import "strings"

// SMB Header Flags2
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
const (
	// Long Names Allowed: Long file names are allowed in the response
	FLAGS2_LONG_NAMES_ALLOWED = 1
	// Extended Attributes: Extended attributes are supported
	FLAGS2_EXTENDED_ATTRIBUTES = 1 << 1
	// Security Signatures: Security signatures are supported
	FLAGS2_SECURITY_SIGNATURE = 1 << 2
	// Compressed: Compression is requested
	FLAGS2_COMPRESSED = 1 << 3
	// Reserved: Reserved for future use
	FLAGS2_RESERVED_4 = 1 << 4
	// Security Signatures Required: Security signatures are required
	FLAGS2_SECURITY_SIGNATURE_REQUIRED = 1 << 5
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

type Flags2 uint16

func (f Flags2) IsLongNamesAllowed() bool {
	return f&FLAGS2_LONG_NAMES_ALLOWED != 0
}

func (f Flags2) IsExtendedAttributes() bool {
	return f&FLAGS2_EXTENDED_ATTRIBUTES != 0
}

func (f Flags2) IsSecuritySignature() bool {
	return f&FLAGS2_SECURITY_SIGNATURE != 0
}

func (f Flags2) IsCompressed() bool {
	return f&FLAGS2_COMPRESSED != 0
}

func (f Flags2) IsSecuritySignatureRequired() bool {
	return f&FLAGS2_SECURITY_SIGNATURE_REQUIRED != 0
}

func (f Flags2) IsLongNamesUsed() bool {
	return f&FLAGS2_LONG_NAMES_USED != 0
}

func (f Flags2) IsReparsePathUsed() bool {
	return f&FLAGS2_REPARSE_PATH != 0
}

func (f Flags2) IsExtendedSecurity() bool {
	return f&FLAGS2_EXTENDED_SECURITY != 0
}

func (f Flags2) IsDfs() bool {
	return f&FLAGS2_DFS != 0
}

func (f Flags2) IsPagingIO() bool {
	return f&FLAGS2_PAGING_IO != 0
}

func (f Flags2) IsNTStatusErrorCodes() bool {
	return f&FLAGS2_NT_STATUS_ERROR_CODES != 0
}

func (f Flags2) IsUnicode() bool {
	return f&FLAGS2_UNICODE != 0
}

// String returns a string representation of the flags2 that are set,
// with each flag name in uppercase separated by a pipe character.
// Flags are listed in alphabetical order.
func (f Flags2) String() string {
	var flagList []string

	if f&FLAGS2_COMPRESSED == FLAGS2_COMPRESSED {
		flagList = append(flagList, "COMPRESSED")
	}

	if f&FLAGS2_DFS == FLAGS2_DFS {
		flagList = append(flagList, "DFS")
	}

	if f&FLAGS2_EXTENDED_ATTRIBUTES == FLAGS2_EXTENDED_ATTRIBUTES {
		flagList = append(flagList, "EXTENDED_ATTRIBUTES")
	}

	if f&FLAGS2_EXTENDED_SECURITY == FLAGS2_EXTENDED_SECURITY {
		flagList = append(flagList, "EXTENDED_SECURITY")
	}

	if f&FLAGS2_LONG_NAMES_ALLOWED == FLAGS2_LONG_NAMES_ALLOWED {
		flagList = append(flagList, "LONG_NAMES_ALLOWED")
	}

	if f&FLAGS2_LONG_NAMES_USED == FLAGS2_LONG_NAMES_USED {
		flagList = append(flagList, "LONG_NAMES_USED")
	}

	if f&FLAGS2_NT_STATUS_ERROR_CODES == FLAGS2_NT_STATUS_ERROR_CODES {
		flagList = append(flagList, "NT_STATUS_ERROR_CODES")
	}

	if f&FLAGS2_PAGING_IO == FLAGS2_PAGING_IO {
		flagList = append(flagList, "PAGING_IO")
	}

	if f&FLAGS2_REPARSE_PATH == FLAGS2_REPARSE_PATH {
		flagList = append(flagList, "REPARSE_PATH")
	}

	if f&FLAGS2_RESERVED_4 == FLAGS2_RESERVED_4 {
		flagList = append(flagList, "RESERVED_4")
	}

	if f&FLAGS2_RESERVED_6 == FLAGS2_RESERVED_6 {
		flagList = append(flagList, "RESERVED_6")
	}

	if f&FLAGS2_RESERVED_8 == FLAGS2_RESERVED_8 {
		flagList = append(flagList, "RESERVED_8")
	}

	if f&FLAGS2_RESERVED_9 == FLAGS2_RESERVED_9 {
		flagList = append(flagList, "RESERVED_9")
	}

	if f&FLAGS2_SECURITY_SIGNATURE == FLAGS2_SECURITY_SIGNATURE {
		flagList = append(flagList, "SECURITY_SIGNATURE")
	}

	if f&FLAGS2_SECURITY_SIGNATURE_REQUIRED == FLAGS2_SECURITY_SIGNATURE_REQUIRED {
		flagList = append(flagList, "SECURITY_SIGNATURE_REQUIRED")
	}

	if f&FLAGS2_UNICODE == FLAGS2_UNICODE {
		flagList = append(flagList, "UNICODE")
	}

	if len(flagList) == 0 {
		return "NONE"
	}

	return strings.Join(flagList, "|")
}
