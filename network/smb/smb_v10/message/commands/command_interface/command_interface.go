package command_interface

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

type CommandInterface interface {
	// GetCommandCode returns the command code
	GetCommandCode() codes.CommandCode

	// SetCommandCode sets the command code
	SetCommandCode(codes.CommandCode)

	// GetAndX returns the AndX of the command
	GetAndX() *andx.AndX

	// SetAndX sets the AndX of the command
	SetAndX(*andx.AndX)

	// IsAndX returns true if the command is an AndX
	IsAndX() bool

	// GetParameters returns the parameters of the command
	GetParameters() *parameters.Parameters

	// SetParameters sets the parameters of the command
	SetParameters(*parameters.Parameters)

	// GetData returns the data of the command
	GetData() *data.Data

	// SetData sets the data of the command
	SetData(*data.Data)

	// GetNextCommand returns the next command in the chain
	// If the command is not an AndX, it returns nil
	GetNextCommand() CommandInterface

	// SetNextCommand sets the next command in the chain
	SetNextCommand(CommandInterface)

	// AddCommandToChain adds a command to the chain
	AddCommandToChain(CommandInterface)

	// GetChainLength returns the length of the chain of commands
	GetChainLength() uint

	// Marshal returns the marshalled command
	Marshal() ([]byte, error)

	// Unmarshal unmarshals the command
	Unmarshal([]byte) (int, error)

	// Init initializes the command
	Init()

	// SetUnicode records whether the enclosing SMB message uses Unicode strings
	// (the SMB_FLAGS2_UNICODE header flag), so that Unmarshal can decode
	// string fields with the correct character encoding.
	SetUnicode(bool)

	// IsUnicode reports whether the enclosing SMB message uses Unicode strings.
	IsUnicode() bool
}

// ChainPositioned is implemented by a command whose wire format carries an offset
// measured from the start of the SMB header rather than from the command itself.
//
// Such a command is only correct at the front of a message: batched second or
// later, its own offsets have to account for everything ahead of it.
// Message.Marshal tells each command where it begins before marshalling it, so a
// command that needs the number has it, and one that does not is unaffected.
type ChainPositioned interface {
	// SetChainOffset records how far past the end of the SMB header this command
	// begins. Zero means it is the first command in the message.
	SetChainOffset(offset int)
}

// Command is a struct that implements the CommandInterface
type Command struct {
	// Command code
	CommandCode codes.CommandCode

	// AndX command
	AndX *andx.AndX

	// Parameters
	Parameters *parameters.Parameters

	// Data
	Data *data.Data

	// Next command
	NextCommand CommandInterface

	// Unicode records whether the enclosing SMB message uses Unicode strings
	// (the SMB_FLAGS2_UNICODE header flag). It is set by the message layer
	// before Unmarshal so string fields are decoded with the right encoding.
	Unicode bool
}

// Init initializes the command
//
// Parameters:
//   - commandCode: The command code to set
func (c *Command) Init() {
	c.AndX = nil

	if c.Parameters == nil {
		c.Parameters = parameters.NewParameters()
	}

	if c.Data == nil {
		c.Data = data.NewData()
	}

	c.NextCommand = nil
}

// SetUnicode records whether the enclosing SMB message uses Unicode strings.
//
// Parameters:
//   - unicode: True if the SMB_FLAGS2_UNICODE header flag is set
func (c *Command) SetUnicode(unicode bool) {
	c.Unicode = unicode
}

// IsUnicode reports whether the enclosing SMB message uses Unicode strings.
//
// Returns:
//   - bool: True if the SMB_FLAGS2_UNICODE header flag is set
func (c *Command) IsUnicode() bool {
	return c.Unicode
}

// GetCommandCode returns the command code
//
// Returns:
//   - codes.CommandCode: The command code
func (c *Command) GetCommandCode() codes.CommandCode {
	return c.CommandCode
}

// SetCommandCode sets the command code
//
// Parameters:
//   - commandCode: The command code to set
func (c *Command) SetCommandCode(commandCode codes.CommandCode) {
	c.CommandCode = commandCode
}

// GetAndX returns the AndX of the command
//
// Returns:
//   - *andx.AndX: The AndX of the command
func (c *Command) GetAndX() *andx.AndX {
	return c.AndX
}

// SetAndX sets the AndX of the command
//
// Parameters:
//   - andX: The AndX to set
func (c *Command) SetAndX(andX *andx.AndX) {
	c.AndX = andX
}

// IsAndX returns true if the command is an AndX
//
// Returns:
//   - bool: True if the command is an AndX, false otherwise
func (c *Command) IsAndX() bool {
	return false
}

// GetParameters returns the parameters of the command
//
// Returns:
//   - *parameters.Parameters: The parameters of the command
func (c *Command) GetParameters() *parameters.Parameters {
	return c.Parameters
}

// SetParameters sets the parameters of the command
//
// Parameters:
//   - parameters: The parameters to set
func (c *Command) SetParameters(parameters *parameters.Parameters) {
	c.Parameters = parameters
}

// GetData returns the data of the command
//
// Returns:
//   - *data.Data: The data of the command
func (c *Command) GetData() *data.Data {
	return c.Data
}

// SetData sets the data of the command
//
// Parameters:
//   - data: The data to set
func (c *Command) SetData(data *data.Data) {
	c.Data = data
}

// GetNextCommand returns the next command in the chain
// If the command is not an AndX, it returns nil
//
// Returns:
//   - CommandInterface: The next command in the chain
func (c *Command) GetNextCommand() CommandInterface {
	// Whatever is linked is returned, with no test for an AndX block.
	//
	// Gating this on c.AndX != nil looks like it identifies an AndX command, but
	// it does not: the block is recorded when a chain is parsed and absent when a
	// chain is built, so the gate hid a command that had been linked deliberately
	// and made every walker stop at the first one. Whether a command may be
	// followed at all is an AndX question, and Message.Marshal asks it there.
	return c.NextCommand
}

// SetNextCommand sets the next command in the chain
//
// Parameters:
//   - nextCommand: The next command to set
func (c *Command) SetNextCommand(nextCommand CommandInterface) {
	c.NextCommand = nextCommand
}

// AddCommandToChain adds a command to the chain
//
// Parameters:
//   - nextCommand: The next command to add to the chain
func (c *Command) AddCommandToChain(nextCommand CommandInterface) {
	if c.NextCommand == nil {
		c.NextCommand = nextCommand
	} else {
		c.NextCommand.AddCommandToChain(nextCommand)
	}
}

// GetChainLength returns the length of the chain of commands
//
// Returns:
//   - uint: The length of the chain of commands
func (c *Command) GetChainLength() uint {
	if c.NextCommand == nil {
		return 1
	}
	return 1 + c.NextCommand.GetChainLength()
}

// Marshal returns the marshalled command
//
// Returns:
//   - []byte: The marshalled command
//   - error: An error if the marshalling fails
func (c *Command) Marshal() ([]byte, error) {
	return nil, nil
}

// Unmarshal unmarshals the command
//
// Parameters:
//   - data: The data to unmarshal
//
// Returns:
//   - int: The number of bytes unmarshalled
//   - error: An error if the unmarshalling fails
func (c *Command) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
