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

// OpenRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/ab9bb872-1967-4088-b444-6a35d0af305e
type OpenRequest struct {
	command_interface.Command

	// Parameters

	// AccessMode (2 bytes): A 16-bit field for encoding the requested access mode. See
	// section 3.2.4.5.1 for a discussion on sharing modes.
	AccessMode types.USHORT

	// SearchAttributes (2 bytes): Specifies the type of file. This field is used as a
	// search mask. Both the FileName and the SearchAttributes of a file MUST match in
	// order for the file to be opened.<28>
	SearchAttributes types.SMB_FILE_ATTRIBUTES

	// Data

	// FileName (variable): A null-terminated string containing the file name of the
	// file to be opened.
	FileName types.SMB_STRING
}

// NewOpenRequest creates a new OpenRequest structure
//
// Returns:
// - A pointer to the new OpenRequest structure
func NewOpenRequest() *OpenRequest {
	c := &OpenRequest{
		// Parameters

		AccessMode:       types.USHORT(0),
		SearchAttributes: types.SMB_FILE_ATTRIBUTES{},

		// Data
		FileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_OPEN)

	return c
}

// Marshal marshals the OpenRequest structure into a byte array
//
// Returns:
// - A byte array representing the OpenRequest structure
// - An error if the marshaling fails
func (c *OpenRequest) Marshal() ([]byte, error) {
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
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter AccessMode
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AccessMode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SearchAttributes
	byteStream, err := c.SearchAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

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
func (c *OpenRequest) Unmarshal(data []byte) (int, error) {
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

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter AccessMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for AccessMode")
	}
	c.AccessMode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SearchAttributes
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for SearchAttributes")
	}
	bytesRead, err = c.SearchAttributes.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data FileName
	bytesRead, err = c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	return offset, nil
}
