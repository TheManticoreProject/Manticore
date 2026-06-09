package flags

import "strings"

// Flags is the 4-byte Flags field of the SMB2 packet header. Unlike SMB 1.0,
// which splits feature bits across a 1-byte Flags and a 2-byte Flags2 field,
// SMB2 carries a single 32-bit little-endian Flags field.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/fb188936-5050-48d3-b350-dc43059638a4
type Flags uint32

// SMB2 Header Flags
const (
	// SMB2_FLAGS_SERVER_TO_REDIR - When set, indicates the message is a response,
	// rather than a request. Set on responses from the server to the client and
	// MUST NOT be set on requests.
	SMB2_FLAGS_SERVER_TO_REDIR Flags = 0x00000001

	// SMB2_FLAGS_ASYNC_COMMAND - When set, indicates that this is an ASYNC SMB2
	// header. This flag MUST NOT be set when using the SYNC SMB2 header.
	SMB2_FLAGS_ASYNC_COMMAND Flags = 0x00000002

	// SMB2_FLAGS_RELATED_OPERATIONS - When set in a request, indicates that this
	// request is a related operation in a compounded request chain.
	SMB2_FLAGS_RELATED_OPERATIONS Flags = 0x00000004

	// SMB2_FLAGS_SIGNED - When set, indicates that this packet has been signed.
	SMB2_FLAGS_SIGNED Flags = 0x00000008

	// SMB2_FLAGS_PRIORITY_MASK - Only valid for the SMB 3.1.1 dialect. A mask for
	// the requested I/O priority of the request, a value in the range 0 to 7.
	SMB2_FLAGS_PRIORITY_MASK Flags = 0x00000070

	// SMB2_FLAGS_DFS_OPERATIONS - When set, indicates that this command is a DFS
	// operation.
	SMB2_FLAGS_DFS_OPERATIONS Flags = 0x10000000

	// SMB2_FLAGS_REPLAY_OPERATION - Only valid for the SMB 3.x dialect family.
	// When set, indicates that this command is a replay operation.
	SMB2_FLAGS_REPLAY_OPERATION Flags = 0x20000000
)

// IsServerToRedir reports whether the message is a response (SMB2_FLAGS_SERVER_TO_REDIR).
func (f Flags) IsServerToRedir() bool {
	return f&SMB2_FLAGS_SERVER_TO_REDIR != 0
}

// IsAsync reports whether the header uses the ASYNC form (SMB2_FLAGS_ASYNC_COMMAND).
func (f Flags) IsAsync() bool {
	return f&SMB2_FLAGS_ASYNC_COMMAND != 0
}

// IsRelatedOperations reports whether this request is part of a compounded chain
// of related operations (SMB2_FLAGS_RELATED_OPERATIONS).
func (f Flags) IsRelatedOperations() bool {
	return f&SMB2_FLAGS_RELATED_OPERATIONS != 0
}

// IsSigned reports whether the packet is signed (SMB2_FLAGS_SIGNED).
func (f Flags) IsSigned() bool {
	return f&SMB2_FLAGS_SIGNED != 0
}

// IsDfsOperations reports whether this command is a DFS operation (SMB2_FLAGS_DFS_OPERATIONS).
func (f Flags) IsDfsOperations() bool {
	return f&SMB2_FLAGS_DFS_OPERATIONS != 0
}

// IsReplayOperation reports whether this command is a replay operation (SMB2_FLAGS_REPLAY_OPERATION).
func (f Flags) IsReplayOperation() bool {
	return f&SMB2_FLAGS_REPLAY_OPERATION != 0
}

// Priority returns the requested I/O priority (0..7) encoded in the
// SMB2_FLAGS_PRIORITY_MASK bits. Only meaningful for the SMB 3.1.1 dialect.
func (f Flags) Priority() uint8 {
	return uint8((f & SMB2_FLAGS_PRIORITY_MASK) >> 4)
}

// String returns a pipe-separated list of the set flag names in alphabetical
// order, or "NONE" if no flags are set.
func (f Flags) String() string {
	var flagList []string

	if f&SMB2_FLAGS_ASYNC_COMMAND != 0 {
		flagList = append(flagList, "ASYNC_COMMAND")
	}
	if f&SMB2_FLAGS_DFS_OPERATIONS != 0 {
		flagList = append(flagList, "DFS_OPERATIONS")
	}
	if f&SMB2_FLAGS_RELATED_OPERATIONS != 0 {
		flagList = append(flagList, "RELATED_OPERATIONS")
	}
	if f&SMB2_FLAGS_REPLAY_OPERATION != 0 {
		flagList = append(flagList, "REPLAY_OPERATION")
	}
	if f&SMB2_FLAGS_SERVER_TO_REDIR != 0 {
		flagList = append(flagList, "SERVER_TO_REDIR")
	}
	if f&SMB2_FLAGS_SIGNED != 0 {
		flagList = append(flagList, "SIGNED")
	}

	if len(flagList) == 0 {
		return "NONE"
	}

	return strings.Join(flagList, "|")
}
