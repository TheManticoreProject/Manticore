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

// FindCloseResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/4c9f95de-365b-4a41-b4fb-fe0233b280eb
type FindCloseResponse struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	Count types.USHORT

	// Data
	BufferFormat types.UCHAR
	DataLength types.USHORT

}

// NewFindCloseResponse creates a new FindCloseResponse structure
//
// Returns:
// - A pointer to the new FindCloseResponse structure
func NewFindCloseResponse() *FindCloseResponse {
	c := &FindCloseResponse{
		// Parameters
		WordCount: types.UCHAR(0),
		Count: types.USHORT(0),

		// Data
		BufferFormat: types.UCHAR(0),
		DataLength: types.USHORT(0),

	}

	c.Command.SetCommandCode(codes.SMB_COM_FIND_CLOSE)

	return c
}



// Marshal marshals the FindCloseResponse structure into a byte array
//
// Returns:
// - A byte array representing the FindCloseResponse structure
// - An error if the marshaling fails
func (c *FindCloseResponse) Marshal() ([]byte, error) {
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
	
	// Marshalling data DataLength
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataLength))
	rawDataContent = append(rawDataContent, buf2...)
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter Count
	buf2 = make([]byte, 2)
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
func (c *FindCloseResponse) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter Count
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Count")
	}
	c.Count = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data BufferFormat
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat")
	}
	c.BufferFormat = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data DataLength
	if len(rawDataContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataLength")
	}
	c.DataLength = types.USHORT(binary.BigEndian.Uint16(rawDataContent[offset:offset+2]))
	offset += 2

	return offset, nil
}
