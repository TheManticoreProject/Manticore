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
	if c.AndX != nil && c.IsAndX() {
		return c.NextCommand
	}
	return nil
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
