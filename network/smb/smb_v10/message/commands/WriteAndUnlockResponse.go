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

// WriteAndUnlockResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/6d09be23-58ea-4966-b1be-86f27c8ea45a
type WriteAndUnlockResponse struct {
	command_interface.Command

	// Parameters

	// CountOfBytesWritten (2 bytes): Indicates the actual number of bytes written
	// to the file. For successful writes, this MUST equal the CountOfBytesToWrite
	// in the client Request. If the number of bytes written differs from the number
	// requested and no error is indicated, then the server has no resources available
	// to satisfy the complete write.
	CountOfBytesWritten types.USHORT
}

// NewWriteAndUnlockResponse creates a new WriteAndUnlockResponse structure
//
// Returns:
// - A pointer to the new WriteAndUnlockResponse structure
func NewWriteAndUnlockResponse() *WriteAndUnlockResponse {
	c := &WriteAndUnlockResponse{
		// Parameters
		CountOfBytesWritten: types.USHORT(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_AND_UNLOCK)

	return c
}

// Marshal marshals the WriteAndUnlockResponse structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndUnlockResponse structure
// - An error if the marshaling fails
func (c *WriteAndUnlockResponse) Marshal() ([]byte, error) {
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
			c.GetAndX().AndXCommand = c.GetCommandCode()
		}

		for _, parameter := range c.GetAndX().GetParameters() {
			c.GetParameters().AddWord(parameter)
		}
	}

	// First marshal the data and then the parameters
	// This is because some parameters are dependent on the data, for example the size of some fields within
	// the data will be stored in the parameters
	rawDataContent := []byte{}

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter CountOfBytesWritten
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.CountOfBytesWritten))
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
func (c *WriteAndUnlockResponse) Unmarshal(data []byte) (int, error) {
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
	_ = c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter CountOfBytesWritten
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesWritten")
	}
	c.CountOfBytesWritten = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
