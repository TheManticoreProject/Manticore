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

// UnlockByteRangeRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/7d3d2faf-8421-4acc-885c-805162028764
type UnlockByteRangeRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit signed integer indicating the
	// file from which the data MUST be read.
	FID types.SHORT

	// CountOfBytesToUnlock (4 bytes): This field is a 32-bit unsigned integer
	// indicating the number of contiguous bytes to be unlocked.
	CountOfBytesToUnlock types.ULONG

	// UnlockOffsetInBytes (4 bytes): This field is a 32-bit unsigned integer
	// indicating the offset, in number of bytes, from which to begin the unlock.
	// Because this field is limited to 32-bits, this command is inappropriate for
	// files that have 64-bit offsets.
	UnlockOffsetInBytes types.ULONG
}

// NewUnlockByteRangeRequest creates a new UnlockByteRangeRequest structure
//
// Returns:
// - A pointer to the new UnlockByteRangeRequest structure
func NewUnlockByteRangeRequest() *UnlockByteRangeRequest {
	c := &UnlockByteRangeRequest{
		// Parameters
		FID:                  types.SHORT(0),
		CountOfBytesToUnlock: types.ULONG(0),
		UnlockOffsetInBytes:  types.ULONG(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_UNLOCK_BYTE_RANGE)

	return c
}

// Marshal marshals the UnlockByteRangeRequest structure into a byte array
//
// Returns:
// - A byte array representing the UnlockByteRangeRequest structure
// - An error if the marshaling fails
func (c *UnlockByteRangeRequest) Marshal() ([]byte, error) {
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

	// Marshalling parameter CountOfBytesToUnlock
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.CountOfBytesToUnlock))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter UnlockOffsetInBytes
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.UnlockOffsetInBytes))
	rawParametersContent = append(rawParametersContent, buf4...)

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
func (c *UnlockByteRangeRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(data[bytesRead:])
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
	c.FID = types.SHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter CountOfBytesToUnlock
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesToUnlock")
	}
	c.CountOfBytesToUnlock = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter UnlockOffsetInBytes
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for UnlockOffsetInBytes")
	}
	c.UnlockOffsetInBytes = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
