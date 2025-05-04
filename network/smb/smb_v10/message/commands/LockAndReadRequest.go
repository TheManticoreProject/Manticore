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

// LockAndReadRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/4652d923-dc4e-4611-b17e-9215d8c66f2e
type LockAndReadRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	FID types.USHORT
	CountOfBytesToRead types.USHORT
	ReadOffsetInBytes types.ULONG
	EstimateOfRemainingBytesToBeRead types.USHORT

	// Data
	ByteCount types.USHORT

}

// NewLockAndReadRequest creates a new LockAndReadRequest structure
//
// Returns:
// - A pointer to the new LockAndReadRequest structure
func NewLockAndReadRequest() *LockAndReadRequest {
	c := &LockAndReadRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		FID: types.USHORT(0),
		CountOfBytesToRead: types.USHORT(0),
		ReadOffsetInBytes: types.ULONG(0),
		EstimateOfRemainingBytesToBeRead: types.USHORT(0),

		// Data
		ByteCount: types.USHORT(0),

	}

	c.Command.SetCommandCode(codes.SMB_COM_LOCK_AND_READ)

	return c
}



// Marshal marshals the LockAndReadRequest structure into a byte array
//
// Returns:
// - A byte array representing the LockAndReadRequest structure
// - An error if the marshaling fails
func (c *LockAndReadRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling parameter FID
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter CountOfBytesToRead
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.CountOfBytesToRead))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter ReadOffsetInBytes
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ReadOffsetInBytes))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter EstimateOfRemainingBytesToBeRead
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.EstimateOfRemainingBytesToBeRead))
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
func (c *LockAndReadRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter CountOfBytesToRead
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesToRead")
	}
	c.CountOfBytesToRead = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter ReadOffsetInBytes
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for ReadOffsetInBytes")
	}
	c.ReadOffsetInBytes = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter EstimateOfRemainingBytesToBeRead
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for EstimateOfRemainingBytesToBeRead")
	}
	c.EstimateOfRemainingBytesToBeRead = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
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
