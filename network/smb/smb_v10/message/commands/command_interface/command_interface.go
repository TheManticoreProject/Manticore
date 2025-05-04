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
}

// GetCommandCode returns the command code
func (c *Command) GetCommandCode() codes.CommandCode {
	return c.CommandCode
}

// SetCommandCode sets the command code
func (c *Command) SetCommandCode(commandCode codes.CommandCode) {
	c.CommandCode = commandCode
}

// GetAndX returns the AndX of the command
func (c *Command) GetAndX() *andx.AndX {
	return c.AndX
}

// SetAndX sets the AndX of the command
func (c *Command) SetAndX(andX *andx.AndX) {
	c.AndX = andX
}

// IsAndX returns true if the command is an AndX
func (c *Command) IsAndX() bool {
	return false
}

// GetParameters returns the parameters of the command
func (c *Command) GetParameters() *parameters.Parameters {
	return c.Parameters
}

// SetParameters sets the parameters of the command
func (c *Command) SetParameters(parameters *parameters.Parameters) {
	c.Parameters = parameters
}

// GetData returns the data of the command
func (c *Command) GetData() *data.Data {
	return c.Data
}

// SetData sets the data of the command
func (c *Command) SetData(data *data.Data) {
	c.Data = data
}

// GetNextCommand returns the next command in the chain
// If the command is not an AndX, it returns nil
func (c *Command) GetNextCommand() CommandInterface {
	if c.AndX != nil && c.IsAndX() {
		return c.NextCommand
	}
	return nil
}

// SetNextCommand sets the next command in the chain
func (c *Command) SetNextCommand(nextCommand CommandInterface) {
	c.NextCommand = nextCommand
}

// AddCommandToChain adds a command to the chain
func (c *Command) AddCommandToChain(nextCommand CommandInterface) {
	if c.NextCommand == nil {
		c.NextCommand = nextCommand
	} else {
		c.NextCommand.AddCommandToChain(nextCommand)
	}
}

// GetChainLength returns the length of the chain of commands
func (c *Command) GetChainLength() uint {
	if c.NextCommand == nil {
		return 1
	}
	return 1 + c.NextCommand.GetChainLength()
}

// Marshal returns the marshalled command
func (c *Command) Marshal() ([]byte, error) {
	return nil, nil
}

// Unmarshal unmarshals the command
func (c *Command) Unmarshal(data []byte) (int, error) {
	return 0, nil
}
