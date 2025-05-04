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

// WriteAndCloseRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/99526875-4b5a-49cc-9686-68ab49825a65
type WriteAndCloseRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	FID types.USHORT
	CountOfBytesToWrite types.USHORT
	WriteOffsetInBytes types.ULONG
	LastWriteTime types.FILETIME

	// Data
	Pad types.UCHAR

}

// NewWriteAndCloseRequest creates a new WriteAndCloseRequest structure
//
// Returns:
// - A pointer to the new WriteAndCloseRequest structure
func NewWriteAndCloseRequest() *WriteAndCloseRequest {
	c := &WriteAndCloseRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		FID: types.USHORT(0),
		CountOfBytesToWrite: types.USHORT(0),
		WriteOffsetInBytes: types.ULONG(0),
		LastWriteTime: types.FILETIME{},

		// Data
		Pad: types.UCHAR(0),

	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_AND_CLOSE)

	return c
}



// Marshal marshals the WriteAndCloseRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndCloseRequest structure
// - An error if the marshaling fails
func (c *WriteAndCloseRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter CountOfBytesToWrite
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.CountOfBytesToWrite))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter WriteOffsetInBytes
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.WriteOffsetInBytes))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter LastWriteTime
	bytesStream, err := c.LastWriteTime.Marshal()
	if err != nil {
			return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)
	
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
func (c *WriteAndCloseRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter CountOfBytesToWrite
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesToWrite")
	}
	c.CountOfBytesToWrite = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter WriteOffsetInBytes
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for WriteOffsetInBytes")
	}
	c.WriteOffsetInBytes = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter LastWriteTime
	if len(rawParametersContent) < offset+8 {
	    return offset, fmt.Errorf("rawParametersContent too short for LastWriteTime")
	}
	bytesRead, err := c.LastWriteTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
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
