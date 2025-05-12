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

// WriteAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/43a10562-269f-4536-b460-17b494405cc4
type WriteAndxResponse struct {
	command_interface.Command

	// Parameters

	// Count (2 bytes): The number of bytes written to the file.
	Count types.USHORT

	// Available (2 bytes): This field is valid when writing to named pipes or I/O
	// devices. This field indicates the number of bytes remaining to be written after
	// the requested write was completed. If the client wrote to a disk file, this
	// field MUST be set to 0xFFFF.
	Available types.USHORT

	// Reserved (4 bytes): This field MUST be 0x00000000.
	Reserved types.ULONG
}

// NewWriteAndxResponse creates a new WriteAndxResponse structure
//
// Returns:
// - A pointer to the new WriteAndxResponse structure
func NewWriteAndxResponse() *WriteAndxResponse {
	c := &WriteAndxResponse{
		// Parameters
		Count:     types.USHORT(0),
		Available: types.USHORT(0),
		Reserved:  types.ULONG(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *WriteAndxResponse) IsAndX() bool {
	return true
}

// Marshal marshals the WriteAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndxResponse structure
// - An error if the marshaling fails
func (c *WriteAndxResponse) Marshal() ([]byte, error) {
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

	// Marshalling parameter Count
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Count))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Available
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Available))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Reserved))
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
func (c *WriteAndxResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Count
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Count")
	}
	c.Count = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Available
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Available")
	}
	c.Available = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
