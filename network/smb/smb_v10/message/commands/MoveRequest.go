package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// MoveRequest is the request for SMB_COM_MOVE (0x2A).
//
// Introduced in the LAN Manager 1.0 dialect and rendered obsolete in NT LAN Manager, this command was used to move files on the server ([MS-CIFS] §2.2.4.38). MS-CIFS does not define a message format for this obsolete command (its legacy definition appears only in [SMB-LM1X] and [XOPEN-SMB]), so it carries no parameters and no data; the
// struct exists so the command code is represented explicitly rather than falling
// through to a generic "command code not supported" error. Clients SHOULD NOT send it, and servers receiving it SHOULD return
// STATUS_NOT_IMPLEMENTED.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/817ee280-ffc9-443d-b9f3-475c4c02a4f1
type MoveRequest struct {
	command_interface.Command
}

// NewMoveRequest creates a new MoveRequest structure
//
// Returns:
// - A pointer to the new MoveRequest structure
func NewMoveRequest() *MoveRequest {
	c := &MoveRequest{}

	c.Command.SetCommandCode(codes.SMB_COM_MOVE)

	return c
}

// Marshal marshals the MoveRequest structure into a byte array
//
// Returns:
// - A byte array representing the MoveRequest structure
// - An error if the marshaling fails
func (c *MoveRequest) Marshal() ([]byte, error) {
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
func (c *MoveRequest) Unmarshal(rawData []byte) (int, error) {
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
