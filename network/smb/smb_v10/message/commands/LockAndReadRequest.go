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

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the file from which the data MUST be read.
	FID types.USHORT

	// CountOfBytesToRead (2 bytes): This field is a 16-bit unsigned integer indicating the number of bytes to be read from the file
	// The client MUST ensure that the amount of data requested will fit in the negotiated maximum buffer size.
	CountOfBytesToRead types.USHORT

	// ReadOffsetInBytes (4 bytes): This field is a 32-bit unsigned integer indicating the offset in number of bytes from which to
	// begin reading from the file. The client MUST ensure that the amount of data requested fits in the negotiated maximum buffer size.
	// Because this field is limited to 32 bits, this command is inappropriate for files that have 64-bit offsets.
	ReadOffsetInBytes types.ULONG

	// EstimateOfRemainingBytesToBeRead (2 bytes): This field is a 16-bit unsigned integer indicating the remaining number of bytes that
	// the client has designated to be read from the file. This is an advisory field and can be zero.
	EstimateOfRemainingBytesToBeRead types.USHORT
}

// NewLockAndReadRequest creates a new LockAndReadRequest structure
//
// Returns:
// - A pointer to the new LockAndReadRequest structure
func NewLockAndReadRequest() *LockAndReadRequest {
	c := &LockAndReadRequest{
		// Parameters
		FID:                              types.USHORT(0),
		CountOfBytesToRead:               types.USHORT(0),
		ReadOffsetInBytes:                types.ULONG(0),
		EstimateOfRemainingBytesToBeRead: types.USHORT(0),
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
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
	_ = c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter CountOfBytesToRead
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesToRead")
	}
	c.CountOfBytesToRead = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ReadOffsetInBytes
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ReadOffsetInBytes")
	}
	c.ReadOffsetInBytes = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter EstimateOfRemainingBytesToBeRead
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for EstimateOfRemainingBytesToBeRead")
	}
	c.EstimateOfRemainingBytesToBeRead = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	return offset, nil
}
