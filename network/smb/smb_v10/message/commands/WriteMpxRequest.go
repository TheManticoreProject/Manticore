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

// WriteMpxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c7fa0e9f-343b-47df-8157-719a3ca9035c
type WriteMpxRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the
	// file to which the data is to be written.
	FID types.USHORT

	// TotalByteCount (2 bytes): The requested total number of bytes to write to the
	// file. The value MAY exceed the negotiated buffer size.
	TotalByteCount types.USHORT

	// Reserved (2 bytes): The server MUST ignore this value.
	Reserved types.USHORT

	// ByteOffsetToBeginWrite (4 bytes): The offset, in bytes, from the start of the file
	// at which the write is to begin. This value indicates the offset at which to write the
	// data contained in the SMB_Data.Bytes.Buffer field of the same message.
	ByteOffsetToBeginWrite types.ULONG

	// Timeout (4 bytes): This field MUST be ignored by the server.<36>
	Timeout types.ULONG

	// WriteMode (2 bytes): A 16-bit field containing flags defined as follows.
	WriteMode types.USHORT

	// RequestMask (4 bytes): This field is a bit mask indicating this SMB request's
	// identity to the server. The server's response MUST contain the logical OR of all
	// of the RequestMask values received. This response MUST be generated.
	RequestMask types.ULONG

	// DataLength (2 bytes): This field value is the number of data bytes included in
	// this request.
	DataLength types.USHORT

	// DataOffset (2 bytes): This field value is the offset, in bytes, from the start
	// of the SMB Header (section 2.2.3.1) to the start of the data buffer.
	DataOffset types.USHORT

	// Data

	// Pad (variable): Null padding bytes to align Buffer to a 16- or 32-bit boundary.
	Pad []types.UCHAR

	// Buffer (variable): The raw data, in bytes, that is to be written to the file.
	Buffer []types.UCHAR
}

// NewWriteMpxRequest creates a new WriteMpxRequest structure
//
// Returns:
// - A pointer to the new WriteMpxRequest structure
func NewWriteMpxRequest() *WriteMpxRequest {
	c := &WriteMpxRequest{
		// Parameters

		FID:                    types.USHORT(0),
		TotalByteCount:         types.USHORT(0),
		Reserved:               types.USHORT(0),
		ByteOffsetToBeginWrite: types.ULONG(0),
		Timeout:                types.ULONG(0),
		WriteMode:              types.USHORT(0),
		RequestMask:            types.ULONG(0),
		DataLength:             types.USHORT(0),
		DataOffset:             types.USHORT(0),

		// Data
		Pad:    []types.UCHAR{},
		Buffer: []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_MPX)

	return c
}

// Marshal marshals the WriteMpxRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteMpxRequest structure
// - An error if the marshaling fails
func (c *WriteMpxRequest) Marshal() ([]byte, error) {
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
	rawDataContent = append(rawDataContent, c.Pad...)

	// Marshalling data Buffer
	rawDataContent = append(rawDataContent, c.Buffer...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalByteCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalByteCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ByteOffsetToBeginWrite
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ByteOffsetToBeginWrite))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Timeout
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter WriteMode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.WriteMode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter RequestMask
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.RequestMask))
	rawParametersContent = append(rawParametersContent, buf4...)

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
func (c *WriteMpxRequest) Unmarshal(data []byte) (int, error) {
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
	rawDataContent := c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TotalByteCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalByteCount")
	}
	c.TotalByteCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ByteOffsetToBeginWrite
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ByteOffsetToBeginWrite")
	}
	c.ByteOffsetToBeginWrite = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
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

	// Unmarshalling parameter RequestMask
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for RequestMask")
	}
	c.RequestMask = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

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

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Pad
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = rawDataContent[offset : offset+int(c.DataOffset)]
	offset += int(c.DataOffset)

	// Unmarshalling data Buffer
	if len(rawDataContent) < offset+int(c.DataLength) {
		return offset, fmt.Errorf("rawParametersContent too short for Buffer")
	}
	c.Buffer = rawDataContent[offset : offset+int(c.DataLength)]

	return offset, nil
}
