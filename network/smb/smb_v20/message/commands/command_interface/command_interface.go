package command_interface

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
)

// CommandInterface is implemented by every SMB2 command body (request or
// response).
//
// Unlike SMB 1.0 — where a command is a WordCount/Words parameter block plus a
// ByteCount/Bytes data block, and batching is expressed with AndX links inside
// the command — an SMB2 command body is a single fixed-size structure (whose
// first field is a 2-byte StructureSize) followed by a variable section that the
// fixed fields reference by offset and length. Compounding is expressed at the
// message level (each compounded command carries its own SMB2 header, chained
// via the header NextCommand field), not inside the command, so the command
// interface carries no chaining methods.
type CommandInterface interface {
	// GetCommandCode returns the SMB2 command code this body belongs to.
	GetCommandCode() codes.CommandCode

	// SetCommandCode sets the SMB2 command code this body belongs to.
	SetCommandCode(codes.CommandCode)

	// GetStructureSize returns the fixed StructureSize value the command writes
	// as the first 2 bytes of its body. It identifies the structure variant and
	// is fixed per command per direction (for example, an SMB2 NEGOTIATE Request
	// uses 36).
	GetStructureSize() uint16

	// Marshal serializes the command body (including its leading StructureSize)
	// into a byte slice.
	Marshal() ([]byte, error)

	// Unmarshal deserializes the command body from the byte slice, which begins
	// at the command's StructureSize field (immediately after the SMB2 header)
	// and runs to the end of this command's region.
	Unmarshal([]byte) (int, error)
}

// Command is the base type embedded by every concrete SMB2 command. It carries
// the command code and provides default no-op Marshal/Unmarshal implementations
// that concrete commands override.
type Command struct {
	// CommandCode is the SMB2 command code this body belongs to.
	CommandCode codes.CommandCode

	// StructureSize is the fixed StructureSize value for the concrete command;
	// concrete constructors set it.
	StructureSize uint16
}

// GetCommandCode returns the SMB2 command code.
func (c *Command) GetCommandCode() codes.CommandCode {
	return c.CommandCode
}

// SetCommandCode sets the SMB2 command code.
func (c *Command) SetCommandCode(commandCode codes.CommandCode) {
	c.CommandCode = commandCode
}

// GetStructureSize returns the fixed StructureSize value for the command.
func (c *Command) GetStructureSize() uint16 {
	return c.StructureSize
}

// Marshal is the default no-op serialization; concrete commands override it.
func (c *Command) Marshal() ([]byte, error) {
	return nil, nil
}

// Unmarshal is the default no-op deserialization; concrete commands override it.
func (c *Command) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
