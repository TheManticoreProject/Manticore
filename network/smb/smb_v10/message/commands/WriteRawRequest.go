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

// WriteRawRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1ff2a25f-efe2-470c-a780-b06ef46c4089
type WriteRawRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the
	// file, named pipe, or device to which the data MUST be written.
	FID types.USHORT

	// CountOfBytes (2 bytes): The total number of bytes to be written to the file during
	// the entire dialog. The value MAY exceed the maximum buffer size (MaxBufferSize)
	// established for the session.
	CountOfBytes types.USHORT

	// Reserved1 (2 bytes): This field is reserved and MUST be ignored by the server.
	Reserved1 types.USHORT

	// Offset (4 bytes): The offset, in bytes, from the start of the file at which the
	// write SHOULD begin. If WordCount is 0x0E, this is the lower 32 bits of a 64-bit value.
	Offset types.ULONG

	// Timeout (4 bytes): This field is the time-out, in milliseconds, to wait for the write
	// to complete. This field is optionally honored only when writing to a named pipe or I/O
	// device. It does not apply and MUST be 0x00000000 when writing to a regular file.
	Timeout types.ULONG

	// WriteMode (2 bytes): A 16-bit field containing flags defined as follows. The flag
	// names below are provided for reference only.
	// If WritethroughMode is not set, this SMB is assumed to be a form of write behind
	// (cached write). The SMB transport layer guarantees delivery of raw data from the client.
	// If an error occurs at the server end, all bytes MUST be received and discarded. If an
	// error occurs while writing data to disk (such as disk full) the next access to the file
	// handle (another write, close, read, etc.) MUST result in an error, reporting this situation.
	// If WritethroughMode is set, the server MUST receive the data, write it to disk and
	// then send a Final Server Response (section 2.2.4.25.3) indicating the result of the write.
	// The total number of bytes successfully written MUST also be returned in the SMB_Parameters.Count
	// field of the response.
	WriteMode types.USHORT

	// Reserved2 (4 bytes): This field MUST be 0x00000000.
	Reserved2 types.ULONG

	// DataLength (2 bytes): This field is the number of bytes included in the SMB_Data
	// block that are to be written to the file.
	DataLength types.USHORT

	// DataOffset (2 bytes): This field is the offset, in bytes, from the start of the
	// SMB Header (section 2.2.3.1) to the start of the data to be written to the file
	// from the Data[] field. Specifying this offset allows the client to efficiently
	// align the data buffer.
	DataOffset types.USHORT

	// OffsetHigh (4 bytes): If WordCount is 0x0E, this is the upper 32 bits of the 64-bit
	// offset in bytes from the start of the file at which the write MUST start. Support
	// of this field is optional.
	OffsetHigh types.ULONG

	// Data

	// Pad (variable): Padding bytes for the client to align the data on an appropriate
	// boundary for transfer of the SMB transport. The server MUST ignore these bytes.
	Pad []types.UCHAR

	// Data (variable): The bytes to be written to the file.
	Data []types.UCHAR
}

// NewWriteRawRequest creates a new WriteRawRequest structure
//
// Returns:
// - A pointer to the new WriteRawRequest structure
func NewWriteRawRequest() *WriteRawRequest {
	c := &WriteRawRequest{
		// Parameters
		FID:          types.USHORT(0),
		CountOfBytes: types.USHORT(0),
		Reserved1:    types.USHORT(0),
		Offset:       types.ULONG(0),
		Timeout:      types.ULONG(0),
		WriteMode:    types.USHORT(0),
		Reserved2:    types.ULONG(0),
		DataLength:   types.USHORT(0),
		DataOffset:   types.USHORT(0),
		OffsetHigh:   types.ULONG(0),

		// Data

		Pad: []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_RAW)

	return c
}

// Marshal marshals the WriteRawRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteRawRequest structure
// - An error if the marshaling fails
func (c *WriteRawRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Pad
	rawDataContent = append(rawDataContent, c.Pad...)

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter CountOfBytes
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.CountOfBytes))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved1
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved1))
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

	// Marshalling parameter Reserved2
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Reserved2))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter DataLength
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataLength))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter OffsetHigh
	if c.OffsetHigh != 0x00000000 {
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
func (c *WriteRawRequest) Unmarshal(data []byte) (int, error) {
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

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
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

	// Unmarshalling parameter CountOfBytes
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CountOfBytes")
	}
	c.CountOfBytes = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved1
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved1")
	}
	c.Reserved1 = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
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

	// Unmarshalling parameter Reserved2
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved2")
	}
	c.Reserved2 = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
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
	// TODO: Compute padding length
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = rawDataContent[offset : offset+int(c.DataLength)]
	offset += int(c.DataLength)

	// Unmarshalling data Data
	if len(rawDataContent) < offset+int(c.DataLength) {
		return offset, fmt.Errorf("rawDataContent too short for Data")
	}
	c.Data = rawDataContent[offset : offset+int(c.DataLength)]
	offset += int(c.DataLength)

	return offset, nil
}
