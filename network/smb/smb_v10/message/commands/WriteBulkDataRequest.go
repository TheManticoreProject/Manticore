package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// WriteBulkDataRequest is the request for SMB_COM_WRITE_BULK_DATA (0xDA).
//
// This command was reserved but never implemented; no formal definition was ever provided (the related SMB_COM_READ_BULK and SMB_COM_WRITE_BULK were likewise never implemented) ([MS-CIFS] §2.2.4.73). It carries no parameters and no data; the struct
// exists so the command code is represented explicitly rather than falling through to
// a generic "command code not supported" error. Clients SHOULD NOT send it, and servers receiving it SHOULD return
// STATUS_NOT_IMPLEMENTED.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/0cc41665-80d5-49aa-af4e-6fff0ed1820f
type WriteBulkDataRequest struct {
	command_interface.Command
}

// NewWriteBulkDataRequest creates a new WriteBulkDataRequest structure
//
// Returns:
// - A pointer to the new WriteBulkDataRequest structure
func NewWriteBulkDataRequest() *WriteBulkDataRequest {
	c := &WriteBulkDataRequest{}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_BULK_DATA)

	return c
}

// Marshal marshals the WriteBulkDataRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteBulkDataRequest structure
// - An error if the marshaling fails
func (c *WriteBulkDataRequest) Marshal() ([]byte, error) {
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
func (c *WriteBulkDataRequest) Unmarshal(rawData []byte) (int, error) {
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
