package capabilities

import "strings"

// Capabilities (4 bytes): A 32-bit field providing a set of server capability indicators.
// This bit field is used to indicate to the client which features are supported by the server.
// Any value not listed in the following table is unused. The server MUST set the unused bits
// to 0 in a response, and the client MUST ignore these bits.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a4229e1a-8a4e-489a-a2eb-11b7f360e60c
type Capabilities uint32

const (
	// CAP_RAW_MODE indicates the server supports SMB_COM_READ_RAW and SMB_COM_WRITE_RAW requests.
	// Raw mode is not supported over connectionless transports.
	CAP_RAW_MODE Capabilities = 0x00000001

	// CAP_MPX_MODE indicates the server supports SMB_COM_READ_MPX and SMB_COM_WRITE_MPX requests.
	// MPX mode is supported only over connectionless transports.
	CAP_MPX_MODE Capabilities = 0x00000002

	// CAP_UNICODE indicates the server supports UTF-16LE Unicode strings.
	CAP_UNICODE Capabilities = 0x00000004

	// CAP_LARGE_FILES indicates the server supports 64-bit file offsets.
	CAP_LARGE_FILES Capabilities = 0x00000008

	// CAP_NT_SMBS indicates the server supports SMB commands particular to the NT LAN Manager dialect.
	CAP_NT_SMBS Capabilities = 0x00000010

	// CAP_RPC_REMOTE_APIS indicates the server supports the use of Microsoft remote procedure call (MS-RPC) for remote API calls.
	CAP_RPC_REMOTE_APIS Capabilities = 0x00000020

	// CAP_STATUS32 indicates the server is capable of responding with 32-bit status codes in the Status field of the SMB Header.
	// CAP_STATUS32 is also sometimes referred to as CAP_NT_STATUS.
	CAP_STATUS32 Capabilities = 0x00000040

	// CAP_LEVEL_II_OPLOCKS indicates the server supports level II opportunistic locks (OpLocks).
	CAP_LEVEL_II_OPLOCKS Capabilities = 0x00000080

	// CAP_LOCK_AND_READ indicates the server supports the SMB_COM_LOCK_AND_READ command request.
	CAP_LOCK_AND_READ Capabilities = 0x00000100

	// CAP_NT_FIND indicates the server supports the TRANS2_FIND_FIRST2, TRANS2_FIND_NEXT2, and FIND_CLOSE2 command requests.
	CAP_NT_FIND Capabilities = 0x00000200

	// CAP_DFS indicates the server is aware of the DFS Referral Protocol and can respond to Microsoft DFS referral requests.
	CAP_DFS Capabilities = 0x00001000

	// CAP_LARGE_READX indicates the server supports large read operations.
	// This capability affects the maximum size, in bytes, of the server buffer for sending an SMB_COM_READ_ANDX response to the client.
	CAP_LARGE_READX Capabilities = 0x00004000
)

// String returns a string representation of the capabilities.
// The string is a bitmask of the capabilities that are set.
// The capabilities are listed in alphabetical order.
func (c Capabilities) String() string {
	var flagList []string

	if c&CAP_DFS == CAP_DFS {
		flagList = append(flagList, "CAP_DFS")
	}

	if c&CAP_LARGE_FILES == CAP_LARGE_FILES {
		flagList = append(flagList, "CAP_LARGE_FILES")
	}

	if c&CAP_LARGE_READX == CAP_LARGE_READX {
		flagList = append(flagList, "CAP_LARGE_READX")
	}

	if c&CAP_LEVEL_II_OPLOCKS == CAP_LEVEL_II_OPLOCKS {
		flagList = append(flagList, "CAP_LEVEL_II_OPLOCKS")
	}

	if c&CAP_LOCK_AND_READ == CAP_LOCK_AND_READ {
		flagList = append(flagList, "CAP_LOCK_AND_READ")
	}

	if c&CAP_MPX_MODE == CAP_MPX_MODE {
		flagList = append(flagList, "CAP_MPX_MODE")
	}

	if c&CAP_NT_FIND == CAP_NT_FIND {
		flagList = append(flagList, "CAP_NT_FIND")
	}

	if c&CAP_NT_SMBS == CAP_NT_SMBS {
		flagList = append(flagList, "CAP_NT_SMBS")
	}

	if c&CAP_RAW_MODE == CAP_RAW_MODE {
		flagList = append(flagList, "CAP_RAW_MODE")
	}

	if c&CAP_RPC_REMOTE_APIS == CAP_RPC_REMOTE_APIS {
		flagList = append(flagList, "CAP_RPC_REMOTE_APIS")
	}

	if c&CAP_STATUS32 == CAP_STATUS32 {
		flagList = append(flagList, "CAP_STATUS32")
	}

	if c&CAP_UNICODE == CAP_UNICODE {
		flagList = append(flagList, "CAP_UNICODE")
	}

	if len(flagList) == 0 {
		return "NONE"
	}

	return strings.Join(flagList, "|")
}
