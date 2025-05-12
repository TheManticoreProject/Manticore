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

// OpenResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/20829e08-c77a-42f3-b427-2ef87d3cf212
type OpenResponse struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): The FID returned for the open file.
	FID types.USHORT

	// FileAttrs (2 bytes): The set of attributes currently assigned to the file. This
	// field is formatted in the same way as the SearchAttributes field in the request.
	FileAttrs types.SMB_FILE_ATTRIBUTES

	// LastModified (4 bytes): The time of the last modification to the opened file.
	LastModified types.FILETIME

	// FileSize (4 bytes): The current size of the opened file, in bytes.
	FileSize types.ULONG

	// AccessMode (2 bytes): A 16-bit field for encoding the granted access mode. This
	// field is formatted in the same way as the Request equivalent.
	AccessMode types.USHORT
}

// NewOpenResponse creates a new OpenResponse structure
//
// Returns:
// - A pointer to the new OpenResponse structure
func NewOpenResponse() *OpenResponse {
	c := &OpenResponse{
		// Parameters
		FID:          types.USHORT(0),
		FileAttrs:    types.SMB_FILE_ATTRIBUTES{},
		LastModified: types.FILETIME{},
		FileSize:     types.ULONG(0),
		AccessMode:   types.USHORT(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_OPEN)

	return c
}

// Marshal marshals the OpenResponse structure into a byte array
//
// Returns:
// - A byte array representing the OpenResponse structure
// - An error if the marshaling fails
func (c *OpenResponse) Marshal() ([]byte, error) {
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter FileAttrs
	byteStream, err := c.FileAttrs.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter LastModified
	bytesStream, err := c.LastModified.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter FileSize
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileSize))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter AccessMode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AccessMode))
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
func (c *OpenResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter FileAttrs
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FileAttrs")
	}
	bytesRead, err = c.FileAttrs.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastModified
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for LastModified")
	}
	bytesRead, err = c.LastModified.Unmarshal(rawParametersContent[offset : offset+8])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter FileSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for FileSize")
	}
	c.FileSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter AccessMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for AccessMode")
	}
	c.AccessMode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
