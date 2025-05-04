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

// WriteAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a66126d2-a1db-446b-8736-b9f5559c49bd
type WriteAndxRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	AndXCommand types.UCHAR
	AndXReserved types.UCHAR
	AndXOffset types.USHORT
	FID types.USHORT
	Offset types.ULONG
	Timeout types.ULONG
	WriteMode types.USHORT
	Remaining types.USHORT
	Reserved types.USHORT
	DataLength types.USHORT
	DataOffset types.USHORT

	// Data
	Pad types.UCHAR

}

// NewWriteAndxRequest creates a new WriteAndxRequest structure
//
// Returns:
// - A pointer to the new WriteAndxRequest structure
func NewWriteAndxRequest() *WriteAndxRequest {
	c := &WriteAndxRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		AndXCommand: types.UCHAR(0),
		AndXReserved: types.UCHAR(0),
		AndXOffset: types.USHORT(0),
		FID: types.USHORT(0),
		Offset: types.ULONG(0),
		Timeout: types.ULONG(0),
		WriteMode: types.USHORT(0),
		Remaining: types.USHORT(0),
		Reserved: types.USHORT(0),
		DataLength: types.USHORT(0),
		DataOffset: types.USHORT(0),

		// Data
		Pad: types.UCHAR(0),

	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_ANDX)

	return c
}


// IsAndX returns true if the command is an AndX
func (c *WriteAndxRequest) IsAndX() bool {
	return true
}



// Marshal marshals the WriteAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndxRequest structure
// - An error if the marshaling fails
func (c *WriteAndxRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data Pad
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad))
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter AndXCommand
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXCommand))
	
	// Marshalling parameter AndXReserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXReserved))
	
	// Marshalling parameter AndXOffset
	buf2 := make([]byte, 2)
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
	
	// Marshalling parameter Timeout
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter WriteMode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.WriteMode))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Remaining
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Remaining))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter DataLength
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataLength))
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
func (c *WriteAndxRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter WriteMode
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for WriteMode")
	}
	c.WriteMode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Remaining
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Remaining")
	}
	c.Remaining = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter DataLength
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataLength")
	}
	c.DataLength = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data Pad
	if len(rawDataContent) < offset+1 {
	    return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = types.UCHAR(rawDataContent[offset])
	offset++

	return offset, nil
}
