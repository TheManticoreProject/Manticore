package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// DeleteDirectoryRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/f944f5bb-0668-4cdf-b6ff-bec3f6ea8667
type DeleteDirectoryRequest struct {
	command_interface.Command

	// Data

	// A null-terminated string that contains the full pathname, relative to the supplied TID, of the directory to be deleted.
	DirectoryName types.SMB_STRING
}

// NewDeleteDirectoryRequest creates a new DeleteDirectoryRequest structure
//
// Returns:
// - A pointer to the new DeleteDirectoryRequest structure
func NewDeleteDirectoryRequest() *DeleteDirectoryRequest {
	c := &DeleteDirectoryRequest{
		// Data
		DirectoryName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_DELETE_DIRECTORY)

	return c
}

// Marshal marshals the DeleteDirectoryRequest structure into a byte array
//
// Returns:
// - A byte array representing the DeleteDirectoryRequest structure
// - An error if the marshaling fails
func (c *DeleteDirectoryRequest) Marshal() ([]byte, error) {
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

	// Marshalling data DirectoryName
	c.DirectoryName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.DirectoryName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

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
func (c *DeleteDirectoryRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	_ = c.GetParameters().GetBytes()
	bytesRead, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data DirectoryName
	bytesRead, err = c.DirectoryName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
