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

// SeekRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/e9dd996c-ba2b-474b-ae5d-5f65c3be1251
type SeekRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): The File ID of the open file within which to seek.
	FID types.USHORT

	// Mode (2 bytes): The seek mode. Possible values are as follows.
	Mode types.USHORT

	// Offset (4 bytes): A 32-bit signed long value indicating the file position,
	// relative to the position indicated in Mode, to which to set the updated file
	// pointer. The value of Offset ranges from -2 gigabytes to +2 gigabytes
	// ((-2**31) to (2**31 -1) bytes).
	Offset types.LONG
}

// NewSeekRequest creates a new SeekRequest structure
//
// Returns:
// - A pointer to the new SeekRequest structure
func NewSeekRequest() *SeekRequest {
	c := &SeekRequest{
		// Parameters
		FID:    types.USHORT(0),
		Mode:   types.USHORT(0),
		Offset: types.LONG(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_SEEK)

	return c
}

// Marshal marshals the SeekRequest structure into a byte array
//
// Returns:
// - A byte array representing the SeekRequest structure
// - An error if the marshaling fails
func (c *SeekRequest) Marshal() ([]byte, error) {
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

	// Marshalling parameter Mode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Mode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Offset
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Offset))
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
func (c *SeekRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Mode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Mode")
	}
	c.Mode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Offset
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Offset")
	}
	c.Offset = types.LONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
