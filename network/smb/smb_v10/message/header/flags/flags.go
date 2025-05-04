package flags

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
	// If this bit is set then all pathnames in the SMB SHOULD be treated as case-insensitive.<26>
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
