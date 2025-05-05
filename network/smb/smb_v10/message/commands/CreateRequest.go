package commands

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// CreateRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/24462104-6b35-4fe4-aeaa-7c30cd727bc9
type CreateRequest struct {
	command_interface.Command

	// Parameters

	// A 16-bit field of 1-bit flags that represent the file attributes to assign to the file if it is created successfully.
	FileAttributes types.SMB_FILE_ATTRIBUTES

	// The time that the file was created on the client, represented as the number of seconds since Jan 1, 1970, 00:00:00.0.
	CreationTime types.FILETIME

	// Data

	// A null-terminated string that represents the fully qualified name of the file relative to the supplied TID to create or truncate on the server.
	FileName types.SMB_STRING
}

// NewCreateRequest creates a new CreateRequest structure
//
// Returns:
// - A pointer to the new CreateRequest structure
func NewCreateRequest() *CreateRequest {
	c := &CreateRequest{
		// Parameters
		FileAttributes: types.SMB_FILE_ATTRIBUTES{},
		CreationTime:   types.FILETIME{},

		// Data
		FileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_CREATE)

	return c
}

// Marshal marshals the CreateRequest structure into a byte array
//
// Returns:
// - A byte array representing the CreateRequest structure
// - An error if the marshaling fails
func (c *CreateRequest) Marshal() ([]byte, error) {
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

	// Marshalling data FileName
	c.FileName.SetBufferFormat(0x04)
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FileAttributes
	bytesStream, err = c.FileAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter CreationTime
	bytesStream, err = c.CreationTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

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
func (c *CreateRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter FileAttributes
	bytesRead, err = c.FileAttributes.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter CreationTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for CreationTime")
	}
	bytesRead, err = c.CreationTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data FileName
	bytesRead, err = c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
