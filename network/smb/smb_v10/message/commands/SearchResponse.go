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

// SearchResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/12a609ad-8709-4492-9b95-4a402589cf56
type SearchResponse struct {
	command_interface.Command

	// Parameters

	// Count (2 bytes): The number of directory entries returned in this response
	// message. This value MUST be less than or equal to the value of MaxCount in the
	// initial request.
	Count types.USHORT

	// Data

	// BufferFormat (1 byte): This field MUST be 0x05, which indicates that a
	// variable-size block is to follow.
	// DataLength (2 bytes): The size, in bytes, of the DirectoryInformationData array,
	// which follows. This field MUST be equal to 43 times the value of
	// SMB_Parameters.Count.
	// DirectoryInformationData (variable): Array of SMB_Directory_Information An array
	// of zero or more SMB_Directory_Information records. The structure and contents of
	// these records is specified below. Note that the SMB_Directory_Information record
	// structure is a fixed 43 bytes in length.
	SMB_Directory_Information types.SMB_STRING
}

// NewSearchResponse creates a new SearchResponse structure
//
// Returns:
// - A pointer to the new SearchResponse structure
func NewSearchResponse() *SearchResponse {
	c := &SearchResponse{
		// Parameters
		Count: types.USHORT(0),

		// Data
		SMB_Directory_Information: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_SEARCH)

	return c
}

// Marshal marshals the SearchResponse structure into a byte array
//
// Returns:
// - A byte array representing the SearchResponse structure
// - An error if the marshaling fails
func (c *SearchResponse) Marshal() ([]byte, error) {
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

	// Marshalling data SMB_Directory_Information
	marshalledSMB_Directory_Information, err := c.SMB_Directory_Information.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, marshalledSMB_Directory_Information...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Count
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Count))
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
func (c *SearchResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Count
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Count")
	}
	c.Count = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data SMB_Directory_Information
	bytesRead, err = c.SMB_Directory_Information.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
