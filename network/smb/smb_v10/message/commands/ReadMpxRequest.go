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

// ReadMpxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3e066ba0-9fce-43c7-85fd-756e4721f7ee
type ReadMpxRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the
	// file from which the data MUST be read.
	FID types.USHORT

	// Offset (4 bytes): The offset, in bytes, from the start of the file at which the
	// read begins.
	Offset types.ULONG

	// MaxCountOfBytesToReturn (2 bytes): This field represents the maximum number of
	// bytes that the client is willing to receive.
	MaxCountOfBytesToReturn types.USHORT

	// MinCountOfBytesToReturn (2 bytes): This field represents the minimum number of
	// bytes that the client is willing to receive.
	MinCountOfBytesToReturn types.USHORT

	// Timeout (4 bytes): The number of milliseconds that the server is requested to
	// wait while processing this command. This field is optionally honored only when
	// reading from a named pipe or I/O device. It does not apply when reading from a
	// regular file.
	Timeout types.ULONG

	// Reserved (2 bytes): This field MUST be set to 0x0000.
	Reserved types.USHORT
}

// NewReadMpxRequest creates a new ReadMpxRequest structure
//
// Returns:
// - A pointer to the new ReadMpxRequest structure
func NewReadMpxRequest() *ReadMpxRequest {
	c := &ReadMpxRequest{
		// Parameters
		FID:                     types.USHORT(0),
		Offset:                  types.ULONG(0),
		MaxCountOfBytesToReturn: types.USHORT(0),
		MinCountOfBytesToReturn: types.USHORT(0),
		Timeout:                 types.ULONG(0),
		Reserved:                types.USHORT(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_READ_MPX)

	return c
}

// Marshal marshals the ReadMpxRequest structure into a byte array
//
// Returns:
// - A byte array representing the ReadMpxRequest structure
// - An error if the marshaling fails
func (c *ReadMpxRequest) Marshal() ([]byte, error) {
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

	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved))
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
func (c *ReadMpxRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
