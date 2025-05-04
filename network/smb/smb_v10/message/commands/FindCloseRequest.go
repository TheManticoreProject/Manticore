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

// FindCloseRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/f62c901b-c2e0-412e-a7df-c4f3889a2412
type FindCloseRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	MaxCount types.USHORT
	SearchAttributes types.USHORT

	// Data
	BufferFormat1 types.UCHAR
	FileName types.SMB_STRING
	BufferFormat2 types.UCHAR
	ResumeKeyLength types.USHORT
	ResumeKey types.SMB_Resume_Key

}

// NewFindCloseRequest creates a new FindCloseRequest structure
//
// Returns:
// - A pointer to the new FindCloseRequest structure
func NewFindCloseRequest() *FindCloseRequest {
	c := &FindCloseRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		MaxCount: types.USHORT(0),
		SearchAttributes: types.USHORT(0),

		// Data
		BufferFormat1: types.UCHAR(0),
		FileName: types.SMB_STRING{},
		BufferFormat2: types.UCHAR(0),
		ResumeKeyLength: types.USHORT(0),
		ResumeKey: types.SMB_Resume_Key{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_FIND_CLOSE)

	return c
}



// Marshal marshals the FindCloseRequest structure into a byte array
//
// Returns:
// - A byte array representing the FindCloseRequest structure
// - An error if the marshaling fails
func (c *FindCloseRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data BufferFormat1
	rawDataContent = append(rawDataContent, types.UCHAR(c.BufferFormat1))
	
	// Marshalling data FileName
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
			return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)
	
	// Marshalling data BufferFormat2
	rawDataContent = append(rawDataContent, types.UCHAR(c.BufferFormat2))
	
	// Marshalling data ResumeKeyLength
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ResumeKeyLength))
	rawDataContent = append(rawDataContent, buf2...)
	
	// Marshalling data ResumeKey
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter MaxCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter SearchAttributes
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.SearchAttributes))
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
func (c *FindCloseRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter MaxCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MaxCount")
	}
	c.MaxCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter SearchAttributes
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for SearchAttributes")
	}
	c.SearchAttributes = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data BufferFormat1
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat1")
	}
	c.BufferFormat1 = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data FileName
	bytesRead, err := c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling data BufferFormat2
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat2")
	}
	c.BufferFormat2 = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data ResumeKeyLength
	if len(rawDataContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for ResumeKeyLength")
	}
	c.ResumeKeyLength = types.USHORT(binary.BigEndian.Uint16(rawDataContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling data ResumeKey

	return offset, nil
}
