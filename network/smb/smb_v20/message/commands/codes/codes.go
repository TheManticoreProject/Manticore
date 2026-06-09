package codes

import "fmt"

// CommandCode is the 2-byte command code carried in the SMB2 packet header.
//
// Unlike SMB 1.0 (where the command is a single byte, see
// network/smb/smb_v10/message/commands/codes), SMB2 uses a 16-bit little-endian
// Command field.
type CommandCode uint16

// SMB2 Commands
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/fb188936-5050-48d3-b350-dc43059638a4
const (
	SMB2_NEGOTIATE       CommandCode = 0x0000
	SMB2_SESSION_SETUP   CommandCode = 0x0001
	SMB2_LOGOFF          CommandCode = 0x0002
	SMB2_TREE_CONNECT    CommandCode = 0x0003
	SMB2_TREE_DISCONNECT CommandCode = 0x0004
	SMB2_CREATE          CommandCode = 0x0005
	SMB2_CLOSE           CommandCode = 0x0006
	SMB2_FLUSH           CommandCode = 0x0007
	SMB2_READ            CommandCode = 0x0008
	SMB2_WRITE           CommandCode = 0x0009
	SMB2_LOCK            CommandCode = 0x000A
	SMB2_IOCTL           CommandCode = 0x000B
	SMB2_CANCEL          CommandCode = 0x000C
	SMB2_ECHO            CommandCode = 0x000D
	SMB2_QUERY_DIRECTORY CommandCode = 0x000E
	SMB2_CHANGE_NOTIFY   CommandCode = 0x000F
	SMB2_QUERY_INFO      CommandCode = 0x0010
	SMB2_SET_INFO        CommandCode = 0x0011
	SMB2_OPLOCK_BREAK    CommandCode = 0x0012
)

// CommandCodeNames maps each SMB2 command code to its human-readable name.
var CommandCodeNames = map[CommandCode]string{
	SMB2_NEGOTIATE:       "NEGOTIATE",
	SMB2_SESSION_SETUP:   "SESSION_SETUP",
	SMB2_LOGOFF:          "LOGOFF",
	SMB2_TREE_CONNECT:    "TREE_CONNECT",
	SMB2_TREE_DISCONNECT: "TREE_DISCONNECT",
	SMB2_CREATE:          "CREATE",
	SMB2_CLOSE:           "CLOSE",
	SMB2_FLUSH:           "FLUSH",
	SMB2_READ:            "READ",
	SMB2_WRITE:           "WRITE",
	SMB2_LOCK:            "LOCK",
	SMB2_IOCTL:           "IOCTL",
	SMB2_CANCEL:          "CANCEL",
	SMB2_ECHO:            "ECHO",
	SMB2_QUERY_DIRECTORY: "QUERY_DIRECTORY",
	SMB2_CHANGE_NOTIFY:   "CHANGE_NOTIFY",
	SMB2_QUERY_INFO:      "QUERY_INFO",
	SMB2_SET_INFO:        "SET_INFO",
	SMB2_OPLOCK_BREAK:    "OPLOCK_BREAK",
}

// String returns the human-readable name of the command code, or a generic
// representation if the code is unknown.
func (c CommandCode) String() string {
	if name, ok := CommandCodeNames[c]; ok {
		return name
	}
	return fmt.Sprintf("CommandCode(0x%04x)", uint16(c))
}

// IsValid reports whether the command code is one of the 19 defined SMB2 commands.
func (c CommandCode) IsValid() bool {
	_, ok := CommandCodeNames[c]
	return ok
}
