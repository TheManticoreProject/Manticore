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

// NtTransactSecondaryRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/4173c449-a6e1-4fa9-b980-708a229fdb3a
type NtTransactSecondaryRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	TotalParameterCount types.ULONG
	TotalDataCount types.ULONG
	ParameterCount types.ULONG
	ParameterOffset types.ULONG
	ParameterDisplacement types.ULONG
	DataCount types.ULONG
	DataOffset types.ULONG
	DataDisplacement types.ULONG
	Reserved2 types.UCHAR

	// Data
	Pad1 []types.UCHAR
	Pad2 []types.UCHAR

}

// NewNtTransactSecondaryRequest creates a new NtTransactSecondaryRequest structure
//
// Returns:
// - A pointer to the new NtTransactSecondaryRequest structure
func NewNtTransactSecondaryRequest() *NtTransactSecondaryRequest {
	c := &NtTransactSecondaryRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		TotalParameterCount: types.ULONG(0),
		TotalDataCount: types.ULONG(0),
		ParameterCount: types.ULONG(0),
		ParameterOffset: types.ULONG(0),
		ParameterDisplacement: types.ULONG(0),
		DataCount: types.ULONG(0),
		DataOffset: types.ULONG(0),
		DataDisplacement: types.ULONG(0),
		Reserved2: types.UCHAR(0),

		// Data
		Pad1: []types.UCHAR{},
		Pad2: []types.UCHAR{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_TRANSACT_SECONDARY)

	return c
}



// Marshal marshals the NtTransactSecondaryRequest structure into a byte array
//
// Returns:
// - A byte array representing the NtTransactSecondaryRequest structure
// - An error if the marshaling fails
func (c *NtTransactSecondaryRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data Pad1
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad1))
	
	// Marshalling data Pad2
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad2))
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter TotalParameterCount
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter TotalDataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter ParameterCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter ParameterOffset
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter ParameterDisplacement
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ParameterDisplacement))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter DataCount
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter DataOffset
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter DataDisplacement
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DataDisplacement))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter Reserved2
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved2))
	
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
func (c *NtTransactSecondaryRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter TotalParameterCount
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for TotalParameterCount")
	}
	c.TotalParameterCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter TotalDataCount
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for TotalDataCount")
	}
	c.TotalDataCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter ParameterCount
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for ParameterCount")
	}
	c.ParameterCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter ParameterOffset
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for ParameterOffset")
	}
	c.ParameterOffset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter ParameterDisplacement
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for ParameterDisplacement")
	}
	c.ParameterDisplacement = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter DataCount
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataCount")
	}
	c.DataCount = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter DataDisplacement
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataDisplacement")
	}
	c.DataDisplacement = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter Reserved2
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for Reserved2")
	}
	c.Reserved2 = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data Pad1
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for Pad1")
	}
	c.Pad1 = types.UCHAR(rawDataContent[offset])
	offset++
	
	// Unmarshalling data Pad2
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for Pad2")
	}
	c.Pad2 = types.UCHAR(rawDataContent[offset])
	offset++

	return offset, nil
}
