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

	// FID (2 bytes): This field MUST be a valid FID indicating the file from which the
	// data MUST be read.
	FID types.USHORT

	// Offset (4 bytes): If WordCount is 0x0A, this field represents a 32-bit offset,
	// measured in bytes, of where the read MUST start relative to the beginning of the
	// file. If WordCount is 0x0C, this field represents the lower 32 bits of a 64-bit
	// offset.
	Offset types.ULONG

	// MaxCountOfBytesToReturn (2 bytes): This field represents the maximum number of
	// bytes that the client is willing to receive.
	MaxCountOfBytesToReturn types.USHORT

	// MinCountOfBytesToReturn (2 bytes): This field represents the minimum number of
	// bytes that the client is willing to receive.
	MinCountOfBytesToReturn types.USHORT

	// Timeout (4 bytes): This field represents the amount of time, in milliseconds,
	// that a server MUST wait before sending a response. It is used only when reading
	// from a named pipe or I/O device and does not apply when reading from a regular
	// file.
	Timeout types.ULONG

	// Remaining (2 bytes): Count of bytes remaining to satisfy client's read request.
	// This field is not used in the NT LAN Manager dialect. Clients MUST set this
	// field to 0x0000, and servers MUST ignore it.
	Remaining types.USHORT
}

// NewReadAndxRequest creates a new ReadAndxRequest structure
//
// Returns:
// - A pointer to the new ReadAndxRequest structure
func NewReadAndxRequest() *ReadAndxRequest {
	c := &ReadAndxRequest{
		// Parameters
		FID:                     types.USHORT(0),
		Offset:                  types.ULONG(0),
		MaxCountOfBytesToReturn: types.USHORT(0),
		MinCountOfBytesToReturn: types.USHORT(0),
		Timeout:                 types.ULONG(0),
		Remaining:               types.USHORT(0),
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
			c.GetAndX().AndXCommand = codes.SMB_COM_NO_ANDX_COMMAND
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
	_, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	_ = c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Offset
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Offset")
	}
	c.Offset = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter MaxCountOfBytesToReturn
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxCountOfBytesToReturn")
	}
	c.MaxCountOfBytesToReturn = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MinCountOfBytesToReturn
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MinCountOfBytesToReturn")
	}
	c.MinCountOfBytesToReturn = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Remaining
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Remaining")
	}
	c.Remaining = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
