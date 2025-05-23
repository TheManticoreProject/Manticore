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

// CreateNewRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/2e4852f0-8672-4d62-9848-42f931b91533
type CreateNewRequest struct {
	command_interface.Command

	// Parameters

	// A 16-bit field of 1-bit flags that represent the file attributes to assign to the file if it is created successfully.
	FileAttributes types.SMB_FILE_ATTRIBUTES

	// The time that the file was created on the client, represented as the number of seconds since Jan 1, 1970, 00:00:00.0.
	CreationTime types.FILETIME

	// Data

	// A null-terminated string that contains the fully qualified name of the file, relative to the supplied TID, to create on the server.
	FileName types.SMB_STRING
}

// NewCreateNewRequest creates a new CreateNewRequest structure
//
// Returns:
// - A pointer to the new CreateNewRequest structure
func NewCreateNewRequest() *CreateNewRequest {
	c := &CreateNewRequest{
		// Parameters
		FileAttributes: types.SMB_FILE_ATTRIBUTES{},
		CreationTime:   types.FILETIME{},

		// Data
		FileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_CREATE_NEW)

	return c
}

// Marshal marshals the CreateNewRequest structure into a byte array
//
// Returns:
// - A byte array representing the CreateNewRequest structure
// - An error if the marshaling fails
func (c *CreateNewRequest) Marshal() ([]byte, error) {
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

	// Marshalling data FileName
	c.FileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
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
func (c *CreateNewRequest) Unmarshal(data []byte) (int, error) {
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
