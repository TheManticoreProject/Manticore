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

	// FID (2 bytes): This field MUST be a valid FID indicating the file to which the
	// data SHOULD be written.
	FID types.USHORT

	// Offset (4 bytes): If WordCount is 0x0C, this field represents a 32-bit offset,
	// measured in bytes, of where the write SHOULD start relative to the beginning of
	// the file. If WordCount is 0xE, this field represents the lower 32 bits of a
	// 64-bit offset.
	Offset types.ULONG

	// Timeout (4 bytes): This field is the time-out, in milliseconds, to wait for the
	// write to complete. This field is used only when writing to a named pipe or an
	// I/O device. It does not apply and MUST be 0x00000000 when writing to a regular
	// file.
	Timeout types.ULONG

	// WriteMode (2 bytes): A 16-bit field containing flags defined as follows:
	WriteMode types.USHORT

	// Remaining (2 bytes): This field is an advisory field telling the server
	// approximately how many bytes are to be written to this file before the next
	// non-write operation. It SHOULD include the number of bytes to be written by this
	// request. The server MAY either ignore this field or use it to perform
	// optimizations. If a pipe write spans multiple requests, the client SHOULD set
	// this field to the number of bytes remaining to be written.<61>
	Remaining types.USHORT

	// Reserved (2 bytes): This field MUST be 0x0000.
	Reserved types.USHORT

	// DataLength (2 bytes): This field is the number of bytes included in the SMB_Data
	// that are to be written to the file.
	DataLength types.USHORT

	// The DataOffset field can be used to relocate the SMB_Data.Bytes.Data block to
	// the end of the message, even if the message is a multi-part AndX chain. If the
	// SMB_Data.Bytes.Data block is relocated, the contents of SMB_Data.Bytes will not
	// be contiguous.
	DataOffset types.USHORT

	// OffsetHigh (4 bytes): This field is optional. If WordCount is 0x0C, this field
	// is not included in the request. If WordCount is 0x0E, this field represents the
	// upper 32 bits of a 64-bit offset, measured in bytes, of where the write SHOULD
	// start relative to the beginning of the file.
	OffsetHigh types.ULONG

	// Data

	// Pad (1 byte): Padding byte that MUST be ignored.
	Pad types.UCHAR

	// Data (variable): The raw bytes to be written to the file.
	Data []types.UCHAR
}

// NewWriteAndxRequest creates a new WriteAndxRequest structure
//
// Returns:
// - A pointer to the new WriteAndxRequest structure
func NewWriteAndxRequest() *WriteAndxRequest {
	c := &WriteAndxRequest{
		// Parameters
		FID:        types.USHORT(0),
		Offset:     types.ULONG(0),
		Timeout:    types.ULONG(0),
		WriteMode:  types.USHORT(0),
		Remaining:  types.USHORT(0),
		Reserved:   types.USHORT(0),
		DataLength: types.USHORT(0),
		DataOffset: types.USHORT(0),
		OffsetHigh: types.ULONG(0),

		// Data
		Pad:  types.UCHAR(0),
		Data: []types.UCHAR{},
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

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)

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

	// Marshalling parameter OffsetHigh
	if c.OffsetHigh != 0 {
		buf4 = make([]byte, 4)
		binary.BigEndian.PutUint32(buf4, uint32(c.OffsetHigh))
		rawParametersContent = append(rawParametersContent, buf4...)
	}

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

	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter WriteMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for WriteMode")
	}
	c.WriteMode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Remaining
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Remaining")
	}
	c.Remaining = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataLength
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataLength")
	}
	c.DataLength = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter OffsetHigh
	if c.GetParameters().WordCount == 0x0E {
		if len(rawParametersContent) < offset+4 {
			return offset, fmt.Errorf("rawParametersContent too short for OffsetHigh")
		}
		c.OffsetHigh = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
		offset += 4
	}

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Pad
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = types.UCHAR(rawDataContent[offset])
	offset++

	// Unmarshalling data Data
	c.Data = rawDataContent[offset : offset+int(c.DataLength)]

	return offset, nil
}
