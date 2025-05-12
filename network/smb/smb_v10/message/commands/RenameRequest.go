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

// RenameRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c970f3bf-806e-4309-8ea9-6515605f450d
type RenameRequest struct {
	command_interface.Command

	// Parameters

	// SearchAttributes (2 bytes): Indicates the file attributes that the file(s) to be
	// renamed MUST have. If the value of this field is 0x0000, then only normal files
	// MUST be matched to be renamed. If the System or Hidden attributes are specified,
	// then entries with those attributes MAY be matched in addition to the normal
	// files. Read-only files MUST NOT be renamed. The read-only attribute of the file
	// MUST be cleared before it can be renamed.
	SearchAttributes types.SMB_FILE_ATTRIBUTES

	// Data

	// BufferFormat1 (1 byte): This field MUST be 0x04.
	// OldFileName (variable): A null-terminated string that contains the name of the
	// file or files to be renamed. Wildcards MAY be used in the filename component of
	// the path.
	OldFileName types.SMB_STRING

	// BufferFormat2 (1 byte): This field MUST be 0x04.
	// NewFileName (variable): A null-terminated string containing the new name(s) to
	// be given to the file(s) that matches OldFileName or the name of the destination
	// directory into which the files matching OldFileName MUST be moved.
	NewFileName types.SMB_STRING
}

// NewRenameRequest creates a new RenameRequest structure
//
// Returns:
// - A pointer to the new RenameRequest structure
func NewRenameRequest() *RenameRequest {
	c := &RenameRequest{
		// Parameters
		SearchAttributes: types.SMB_FILE_ATTRIBUTES{},

		// Data
		OldFileName: types.SMB_STRING{},
		NewFileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_RENAME)

	return c
}

// Marshal marshals the RenameRequest structure into a byte array
//
// Returns:
// - A byte array representing the RenameRequest structure
// - An error if the marshaling fails
func (c *RenameRequest) Marshal() ([]byte, error) {
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

	// Marshalling data OldFileName
	c.OldFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.OldFileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data NewFileName
	c.NewFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err = c.NewFileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter SearchAttributes
	marshalledSearchAttributes, err := c.SearchAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledSearchAttributes...)

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
func (c *RenameRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter SearchAttributes
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for SearchAttributes")
	}
	c.SearchAttributes.Unmarshal(rawParametersContent[offset : offset+2])
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data OldFileName
	bytesRead, err = c.OldFileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data NewFileName
	bytesRead, err = c.NewFileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
