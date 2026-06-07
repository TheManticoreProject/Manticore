package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// NewFileSizeResponse is the response for SMB_COM_NEW_FILE_SIZE (0x30).
//
// This command was reserved but never implemented, and its message format was
// never defined ([MS-CIFS] §2.2.4.44). It carries no parameters and no data; the
// struct exists so the command code is represented explicitly rather than falling
// through to a generic "command code not supported" error. A server receiving the
// request SHOULD return STATUS_NOT_IMPLEMENTED.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/e3b0e8ec-a0f3-48d7-92b9-25715e5ec6c8
type NewFileSizeResponse struct {
	command_interface.Command
}

// NewNewFileSizeResponse creates a new NewFileSizeResponse structure
//
// Returns:
// - A pointer to the new NewFileSizeResponse structure
func NewNewFileSizeResponse() *NewFileSizeResponse {
	c := &NewFileSizeResponse{}

	c.Command.SetCommandCode(codes.SMB_COM_NEW_FILE_SIZE)

	return c
}

// Marshal marshals the NewFileSizeResponse structure into a byte array
//
// Returns:
// - A byte array representing the NewFileSizeResponse structure
// - An error if the marshaling fails
func (c *NewFileSizeResponse) Marshal() ([]byte, error) {
	marshalledCommand := []byte{}

	// Create the Parameters structure if it is nil
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Create the Data structure if it is nil
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// In case of AndX, we need to add the parameters to the Parameters structure first
	if c.IsAndX() {
		if c.GetAndX() == nil {
			c.SetAndX(andx.NewAndX())
			c.GetAndX().AndXCommand = codes.SMB_COM_NO_ANDX_COMMAND
		}

		for _, parameter := range c.GetAndX().GetParameters() {
			c.GetParameters().AddWord(parameter)
		}
	}

	// Marshalling parameters
	// This command was never defined: no parameters are sent in this message.
	marshalledParameters, err := c.GetParameters().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledParameters...)

	// Marshalling data
	// This command was never defined: no data is sent in this message.
	marshalledData, err := c.GetData().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledData...)

	return marshalledCommand, nil
}

// Unmarshal unmarshals a byte array into the command structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshalling fails
func (c *NewFileSizeResponse) Unmarshal(rawData []byte) (int, error) {
	// Initialize the Parameters structure if it is nil to avoid a nil
	// pointer dereference when Unmarshal is called on a freshly constructed value.
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Initialize the Data structure if it is nil for the same reason.
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(rawData)
	if err != nil {
		return 0, err
	}
	_ = c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(rawData[bytesRead:])
	if err != nil {
		return 0, err
	}
	_ = c.GetData().GetBytes()

	// This command was never defined: no parameters and no data are sent in this message.
	return 0, nil
}
