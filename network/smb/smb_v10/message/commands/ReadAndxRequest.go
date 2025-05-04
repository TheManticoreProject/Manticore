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

// ReadAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/7e6c7cc2-c3f1-4335-8263-d7412f77140e
type ReadAndxRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	AndXCommand types.UCHAR
	AndXReserved types.UCHAR
	AndXOffset types.USHORT
	FID types.USHORT
	Offset types.ULONG
	MaxCountOfBytesToReturn types.USHORT
	MinCountOfBytesToReturn types.USHORT
	Timeout types.ULONG
	Remaining types.USHORT

	// Data
	ByteCount types.USHORT

}

// NewReadAndxRequest creates a new ReadAndxRequest structure
//
// Returns:
// - A pointer to the new ReadAndxRequest structure
func NewReadAndxRequest() *ReadAndxRequest {
	c := &ReadAndxRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		AndXCommand: types.UCHAR(0),
		AndXReserved: types.UCHAR(0),
		AndXOffset: types.USHORT(0),
		FID: types.USHORT(0),
		Offset: types.ULONG(0),
		MaxCountOfBytesToReturn: types.USHORT(0),
		MinCountOfBytesToReturn: types.USHORT(0),
		Timeout: types.ULONG(0),
		Remaining: types.USHORT(0),

		// Data
		ByteCount: types.USHORT(0),

	}

	c.Command.SetCommandCode(codes.SMB_COM_READ_ANDX)

	return c
}


// IsAndX returns true if the command is an AndX
func (c *ReadAndxRequest) IsAndX() bool {
	return true
}



// Marshal marshals the ReadAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the ReadAndxRequest structure
// - An error if the marshaling fails
func (c *ReadAndxRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data ByteCount
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ByteCount))
	rawDataContent = append(rawDataContent, buf2...)
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter AndXCommand
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXCommand))
	
	// Marshalling parameter AndXReserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXReserved))
	
	// Marshalling parameter AndXOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AndXOffset))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter FID
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Offset
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Offset))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter MaxCountOfBytesToReturn
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxCountOfBytesToReturn))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter MinCountOfBytesToReturn
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MinCountOfBytesToReturn))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Timeout
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter Remaining
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Remaining))
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
func (c *ReadAndxRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter AndXCommand
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for AndXCommand")
	}
	c.AndXCommand = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXReserved
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for AndXReserved")
	}
	c.AndXReserved = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXOffset
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for AndXOffset")
	}
	c.AndXOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Offset
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Offset")
	}
	c.Offset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter MaxCountOfBytesToReturn
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MaxCountOfBytesToReturn")
	}
	c.MaxCountOfBytesToReturn = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter MinCountOfBytesToReturn
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for MinCountOfBytesToReturn")
	}
	c.MinCountOfBytesToReturn = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter Remaining
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Remaining")
	}
	c.Remaining = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data ByteCount
	if len(rawDataContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for ByteCount")
	}
	c.ByteCount = types.USHORT(binary.BigEndian.Uint16(rawDataContent[offset:offset+2]))
	offset += 2

	return offset, nil
}
