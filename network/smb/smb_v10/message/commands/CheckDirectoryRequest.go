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

// CheckDirectoryRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/dc566429-904b-4bf5-8158-d68f2370ae68
type CheckDirectoryRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR

	// Data
	BufferFormat types.UCHAR
	DirectoryName types.SMB_STRING

}

// NewCheckDirectoryRequest creates a new CheckDirectoryRequest structure
//
// Returns:
// - A pointer to the new CheckDirectoryRequest structure
func NewCheckDirectoryRequest() *CheckDirectoryRequest {
	c := &CheckDirectoryRequest{
		// Parameters
		WordCount: types.UCHAR(0),

		// Data
		BufferFormat: types.UCHAR(0),
		DirectoryName: types.SMB_STRING{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_CHECK_DIRECTORY)

	return c
}



// Marshal marshals the CheckDirectoryRequest structure into a byte array
//
// Returns:
// - A byte array representing the CheckDirectoryRequest structure
// - An error if the marshaling fails
func (c *CheckDirectoryRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data BufferFormat
	rawDataContent = append(rawDataContent, types.UCHAR(c.BufferFormat))
	
	// Marshalling data DirectoryName
	bytesStream, err := c.DirectoryName.Marshal()
	if err != nil {
			return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
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
func (c *CheckDirectoryRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	bytesRead, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0
	
	// Unmarshalling parameter WordCount
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for WordCount")
	}
	c.WordCount = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data BufferFormat
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat")
	}
	c.BufferFormat = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data DirectoryName
	bytesRead, err := c.DirectoryName.Unmarshal(rawDataContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead

	return offset, nil
}
