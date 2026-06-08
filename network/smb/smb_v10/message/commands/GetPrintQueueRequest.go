package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// GetPrintQueueRequest is the request for SMB_COM_GET_PRINT_QUEUE (0xC3).
//
// An original Core Protocol command rendered obsolete in NT LAN Manager, this command was used to list the items in a print queue associated with a TID ([MS-CIFS] §2.2.4.70). MS-CIFS does not define a message format for this obsolete command (its legacy definition appears only in [SMB-CORE]), so it carries no parameters and no data; the
// struct exists so the command code is represented explicitly rather than falling
// through to a generic "command code not supported" error. Clients SHOULD NOT send it, and servers receiving it MUST return
// STATUS_NOT_IMPLEMENTED.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/8aaa6b27-b144-4cd6-9171-102217b1406d
type GetPrintQueueRequest struct {
	command_interface.Command
}

// NewGetPrintQueueRequest creates a new GetPrintQueueRequest structure
//
// Returns:
// - A pointer to the new GetPrintQueueRequest structure
func NewGetPrintQueueRequest() *GetPrintQueueRequest {
	c := &GetPrintQueueRequest{}

	c.Command.SetCommandCode(codes.SMB_COM_GET_PRINT_QUEUE)

	return c
}

// Marshal marshals the GetPrintQueueRequest structure into a byte array
//
// Returns:
// - A byte array representing the GetPrintQueueRequest structure
// - An error if the marshaling fails
func (c *GetPrintQueueRequest) Marshal() ([]byte, error) {
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
func (c *GetPrintQueueRequest) Unmarshal(rawData []byte) (int, error) {
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
