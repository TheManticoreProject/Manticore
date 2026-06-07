package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// IoctlSecondaryRequest is the request for SMB_COM_IOCTL_SECONDARY (0x28).
//
// Introduced in the LAN Manager 1.0 dialect and rendered obsolete in NT LAN Manager, this command (a companion to the deprecated SMB_COM_IOCTL) was reserved but not implemented, so its message format was never defined ([MS-CIFS] §2.2.4.36). It carries no parameters and no data; the struct
// exists so the command code is represented explicitly rather than falling through to
// a generic "command code not supported" error. Clients SHOULD NOT send it, and servers receiving it MUST return
// STATUS_NOT_IMPLEMENTED.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3a5f8e47-16e6-484d-9346-6c4cbdc22dec
type IoctlSecondaryRequest struct {
	command_interface.Command
}

// NewIoctlSecondaryRequest creates a new IoctlSecondaryRequest structure
//
// Returns:
// - A pointer to the new IoctlSecondaryRequest structure
func NewIoctlSecondaryRequest() *IoctlSecondaryRequest {
	c := &IoctlSecondaryRequest{}

	c.Command.SetCommandCode(codes.SMB_COM_IOCTL_SECONDARY)

	return c
}

// Marshal marshals the IoctlSecondaryRequest structure into a byte array
//
// Returns:
// - A byte array representing the IoctlSecondaryRequest structure
// - An error if the marshaling fails
func (c *IoctlSecondaryRequest) Marshal() ([]byte, error) {
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
func (c *IoctlSecondaryRequest) Unmarshal(rawData []byte) (int, error) {
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
