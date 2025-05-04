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

// TreeConnectRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/0036eb81-7466-4e1c-afb6-ea8bc9dd19dc
type TreeConnectRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR

	// Data
	BufferFormat1 types.UCHAR
	Path types.OEM_STRING
	BufferFormat2 types.UCHAR
	Password types.OEM_STRING
	BufferFormat3 types.UCHAR
	Service types.OEM_STRING

}

// NewTreeConnectRequest creates a new TreeConnectRequest structure
//
// Returns:
// - A pointer to the new TreeConnectRequest structure
func NewTreeConnectRequest() *TreeConnectRequest {
	c := &TreeConnectRequest{
		// Parameters
		WordCount: types.UCHAR(0),

		// Data
		BufferFormat1: types.UCHAR(0),
		Path: types.OEM_STRING{},
		BufferFormat2: types.UCHAR(0),
		Password: types.OEM_STRING{},
		BufferFormat3: types.UCHAR(0),
		Service: types.OEM_STRING{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_TREE_CONNECT)

	return c
}



// Marshal marshals the TreeConnectRequest structure into a byte array
//
// Returns:
// - A byte array representing the TreeConnectRequest structure
// - An error if the marshaling fails
func (c *TreeConnectRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data Path
	
	// Marshalling data BufferFormat2
	rawDataContent = append(rawDataContent, types.UCHAR(c.BufferFormat2))
	
	// Marshalling data Password
	
	// Marshalling data BufferFormat3
	rawDataContent = append(rawDataContent, types.UCHAR(c.BufferFormat3))
	
	// Marshalling data Service
	
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
func (c *TreeConnectRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling data BufferFormat1
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat1")
	}
	c.BufferFormat1 = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data Path
	
	// Unmarshalling data BufferFormat2
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat2")
	}
	c.BufferFormat2 = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data Password
	
	// Unmarshalling data BufferFormat3
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for BufferFormat3")
	}
	c.BufferFormat3 = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data Service

	return offset, nil
}
