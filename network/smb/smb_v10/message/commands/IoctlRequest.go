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

// IoctlRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c8f1b5b1-9ec1-49d2-a0e1-78ee88f39e71
type IoctlRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	FID types.USHORT
	Category types.USHORT
	Function types.USHORT
	TotalParameterCount types.USHORT
	TotalDataCount types.USHORT
	MaxParameterCount types.USHORT
	MaxDataCount types.USHORT
	Timeout types.ULONG
	Reserved types.USHORT
	ParameterCount types.USHORT
	ParameterOffset types.USHORT
	DataCount types.USHORT
	DataOffset types.USHORT

	// Data
	Pad1 []types.UCHAR
	Pad2 []types.UCHAR

}

// NewIoctlRequest creates a new IoctlRequest structure
//
// Returns:
// - A pointer to the new IoctlRequest structure
func NewIoctlRequest() *IoctlRequest {
	c := &IoctlRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		FID: types.USHORT(0),
		Category: types.USHORT(0),
		Function: types.USHORT(0),
		TotalParameterCount: types.USHORT(0),
		TotalDataCount: types.USHORT(0),
		MaxParameterCount: types.USHORT(0),
		MaxDataCount: types.USHORT(0),
		Timeout: types.ULONG(0),
		Reserved: types.USHORT(0),
		ParameterCount: types.USHORT(0),
		ParameterOffset: types.USHORT(0),
		DataCount: types.USHORT(0),
		DataOffset: types.USHORT(0),

		// Data
		Pad1: []types.UCHAR{},
		Pad2: []types.UCHAR{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_IOCTL)

	return c
}



// Marshal marshals the IoctlRequest structure into a byte array
//
// Returns:
// - A byte array representing the IoctlRequest structure
// - An error if the marshaling fails
func (c *IoctlRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Category
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Category))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Function
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Function))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter TotalParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter TotalDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter MaxParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter MaxDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Timeout
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter ParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter ParameterOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter DataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataOffset))
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
func (c *IoctlRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Category
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Category")
	}
	c.Category = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Function
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Function")
	}
	c.Function = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter TotalParameterCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for TotalParameterCount")
	}
	c.TotalParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter TotalDataCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for TotalDataCount")
	}
	c.TotalDataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter MaxParameterCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MaxParameterCount")
	}
	c.MaxParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter MaxDataCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MaxDataCount")
	}
	c.MaxDataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter ParameterCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for ParameterCount")
	}
	c.ParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter ParameterOffset
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for ParameterOffset")
	}
	c.ParameterOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter DataCount
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataCount")
	}
	c.DataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
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
