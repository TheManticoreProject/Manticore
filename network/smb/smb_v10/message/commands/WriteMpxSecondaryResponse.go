package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// WriteMpxSecondaryResponse is the response for SMB_COM_WRITE_MPX_SECONDARY (0x1F).
//
// Introduced in the LAN Manager 1.0 dialect and rendered obsolete in NT LAN Manager, this command was the secondary request used with the (also obsolete) SMB_COM_WRITE_MPX multiplexed block write ([MS-CIFS] §2.2.4.27). MS-CIFS does not define a message format for this obsolete command (its legacy definition appears only in [SMB-LM1X]), so it carries no parameters and no data; the
// struct exists so the command code is represented explicitly rather than falling
// through to a generic "command code not supported" error. A server receiving the request SHOULD return STATUS_NOT_IMPLEMENTED.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/d07bc94a-9da8-43f7-8777-9e9033891ef7
type WriteMpxSecondaryResponse struct {
	command_interface.Command
}

// NewWriteMpxSecondaryResponse creates a new WriteMpxSecondaryResponse structure
//
// Returns:
// - A pointer to the new WriteMpxSecondaryResponse structure
func NewWriteMpxSecondaryResponse() *WriteMpxSecondaryResponse {
	c := &WriteMpxSecondaryResponse{}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_MPX_SECONDARY)

	return c
}

// Marshal marshals the WriteMpxSecondaryResponse structure into a byte array
//
// Returns:
// - A byte array representing the WriteMpxSecondaryResponse structure
// - An error if the marshaling fails
func (c *WriteMpxSecondaryResponse) Marshal() ([]byte, error) {
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
	// MS-CIFS defines no message format for this obsolete command: no parameters are sent.
	marshalledParameters, err := c.GetParameters().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledParameters...)

	// Marshalling data
	// MS-CIFS defines no message format for this obsolete command: no data is sent.
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
func (c *WriteMpxSecondaryResponse) Unmarshal(rawData []byte) (int, error) {
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

	// MS-CIFS defines no message format for this obsolete command: no parameters or data are read.
	return 0, nil
}
