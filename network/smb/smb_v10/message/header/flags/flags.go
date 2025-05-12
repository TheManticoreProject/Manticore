package flags

import "strings"

// SMB Header Flags
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
const (
	// This bit is set (1) in the SMB_COM_NEGOTIATE (0x72) Response (section 2.2.4.52.2) if the server supports
	// SMB_COM_LOCK_AND_READ (0x13) (section 2.2.4.20) and SMB_COM_WRITE_AND_UNLOCK (0x14) (section 2.2.4.21) commands.
	FLAGS_LOCK_AND_READ_OK = 0x01
	// Obsolete
	// When set (on an SMB request being sent to the server), the client guarantees that there is a receive buffer posted such that a send without
	// acknowledgment can be used by the server to respond to the client's request.
	// This behavior is specific to an obsolete transport. This bit MUST be set to zero by the client and MUST be ignored by the server.
	FLAGS_BUF_AVAIL = 0x02
	// This flag MUST be set to zero by the client and MUST be ignored by the server.
	FLAGS_RESERVED = 0x04
	// This flag MUST be set to zero by the client and MUST be ignored by the server.
	FLAGS_CASE_INSENSITIVE = 0x08
	// Obsolete
	// If this bit is set then all pathnames in the SMB SHOULD be treated as case-insensitive.
	FLAGS_CANONICALIZED_PATHS = 0x10
	// Obsolescent
	// When set in session setup, this bit indicates that all paths sent to the server are already in canonical format.
	// That is, all file and directory names are composed of valid file name characters in all upper-case, and that the path
	// segments are separated by backslash characters ('\').
	FLAGS_OPLOCK = 0x20
	// Obsolescent
	// This bit has meaning only in the deprecated SMB_COM_OPEN (0x02) Request (section 2.2.4.3.1), SMB_COM_CREATE (0x03) Request (section 2.2.4.4.1),
	// and SMB_COM_CREATE_NEW (0x0F) Request (section 2.2.4.16.1) messages, where it is used to indicate that the client is requesting an Exclusive OpLock.
	// It SHOULD be set to zero by the client, and ignored by the server, in all other SMB requests. If the server grants this OpLock request,
	// then this bit SHOULD remain set in the corresponding response SMB to indicate to the client that the OpLock request was granted.
	FLAGS_OPBATCH = 0x40
	// When on, this message is being sent from the server in response to a client request. The Command field usually contains the same value
	// in a protocol request from the client to the server as in the matching response from the server to the client. This bit unambiguously
	// distinguishes the message as a server response.
	FLAGS_REPLY = 0x80
)

type Flags uint16

func (f Flags) IsLockAndReadOk() bool {
	return f&FLAGS_LOCK_AND_READ_OK != 0
}

func (f Flags) IsBufAvail() bool {
	return f&FLAGS_BUF_AVAIL != 0
}

func (f Flags) IsReserved() bool {
	return f&FLAGS_RESERVED != 0
}

func (f Flags) IsCaseInsensitive() bool {
	return f&FLAGS_CASE_INSENSITIVE != 0
}

func (f Flags) IsCanonicalizedPaths() bool {
	return f&FLAGS_CANONICALIZED_PATHS != 0
}

func (f Flags) IsOplock() bool {
	return f&FLAGS_OPLOCK != 0
}

func (f Flags) IsOplockBatch() bool {
	return f&FLAGS_OPBATCH != 0
}

func (f Flags) IsReply() bool {
	return f&FLAGS_REPLY != 0
}

// String returns a string representation of the flags that are set,
// with each flag name in uppercase separated by a pipe character.
// Flags are listed in alphabetical order.
func (f Flags) String() string {
	var result string
	var flagList []string

	if f&FLAGS_BUF_AVAIL == FLAGS_BUF_AVAIL {
		flagList = append(flagList, "BUF_AVAIL")
	}

	if f&FLAGS_CANONICALIZED_PATHS == FLAGS_CANONICALIZED_PATHS {
		flagList = append(flagList, "CANONICALIZED_PATHS")
	}

	if f&FLAGS_CASE_INSENSITIVE == FLAGS_CASE_INSENSITIVE {
		flagList = append(flagList, "CASE_INSENSITIVE")
	}

	if f&FLAGS_LOCK_AND_READ_OK == FLAGS_LOCK_AND_READ_OK {
		flagList = append(flagList, "LOCK_AND_READ_OK")
	}

	if f&FLAGS_OPBATCH == FLAGS_OPBATCH {
		flagList = append(flagList, "OPBATCH")
	}

	if f&FLAGS_OPLOCK == FLAGS_OPLOCK {
		flagList = append(flagList, "OPLOCK")
	}

	if f&FLAGS_REPLY == FLAGS_REPLY {
		flagList = append(flagList, "REPLY")
	}

	if f&FLAGS_RESERVED == FLAGS_RESERVED {
		flagList = append(flagList, "RESERVED")
	}

	if len(flagList) == 0 {
		return "NONE"
	}

	result = strings.Join(flagList, "|")

	return result
}
