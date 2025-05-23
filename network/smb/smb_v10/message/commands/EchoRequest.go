package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// EchoRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/5dd916b2-9c38-4f0c-ba55-a2c86dd5de10
type EchoRequest struct {
	command_interface.Command

	// Parameters

	// The number of times that the server SHOULD echo the contents of the SMB_Data.Data field.
	EchoCount types.USHORT

	// Data

	// Data to echo.
	Data []types.UCHAR
}

// NewEchoRequest creates a new EchoRequest structure
//
// Returns:
// - A pointer to the new EchoRequest structure
func NewEchoRequest() *EchoRequest {
	c := &EchoRequest{
		// Parameters
		EchoCount: types.USHORT(0),

		// Data
		Data: []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_ECHO)

	return c
}

// Marshal marshals the EchoRequest structure into a byte array
//
// Returns:
// - A byte array representing the EchoRequest structure
// - An error if the marshaling fails
func (c *EchoRequest) Marshal() ([]byte, error) {
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

	// First marshal the data and then the parameters
	// This is because some parameters are dependent on the data, for example the size of some fields within
	// the data will be stored in the parameters
	rawDataContent := []byte{}

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter EchoCount
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.EchoCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameters
	c.GetParameters().AddWordsFromBytesStream(rawParametersContent)
	marshalledParameters, err := c.GetParameters().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledParameters...)

	// Marshalling data
	c.GetData().Add(rawDataContent)
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
func (c *EchoRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter EchoCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for EchoCount")
	}
	c.EchoCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Data
	c.Data = rawDataContent[offset:]
	offset += len(c.Data)

	return offset, nil
}
