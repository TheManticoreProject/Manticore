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

// WriteAndCloseRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/99526875-4b5a-49cc-9686-68ab49825a65
type WriteAndCloseRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the
	// file to which the data SHOULD be written.
	FID types.USHORT

	// CountOfBytesToWrite (2 bytes): This field is a 16-bit unsigned integer
	// indicating the number of bytes to be written to the file. The client MUST ensure
	// that the amount of data sent can fit in the negotiated maximum buffer size. If
	// the value of this field is zero (0x0000), the server MUST truncate or extend the
	// file to match the WriteOffsetInBytes.
	CountOfBytesToWrite types.USHORT

	// WriteOffsetInBytes  (4 bytes): This field is a 32-bit unsigned integer indicating
	// the offset, in number of bytes, from the beginning of the file at which to begin
	// writing to the file. The client MUST ensure that the amount of data sent can
	// it in the negotiated maximum buffer size. Because this field is limited to 32-bits,
	// this command is inappropriate for files that have 64-bit offsets.
	WriteOffsetInBytes types.ULONG

	// LastWriteTime (4 bytes): This field is a 32-bit unsigned integer indicating the
	// number of seconds since Jan 1, 1970, 00:00:00.0. The server SHOULD set the last
	// write time of the file represented by the FID to this value. If the value is
	// zero (0x00000000), the server SHOULD use the current local time of the server to
	// set the value. Failure to set the time MUST NOT result in an error response from
	// the server.
	LastWriteTime types.FILETIME

	// Reserved (12 bytes): This field is optional. This field is reserved, and all
	// entries MUST be zero (0x00000000). This field is used only in the 12-word version
	// of the request.
	Reserved [3]types.ULONG

	// Data

	// Pad (1 byte): The value of this field SHOULD be ignored. This is padding to
	// force the byte alignment to a double word boundary.
	Pad types.UCHAR

	// Data (variable): The raw bytes to be written to the file.
	Data []types.UCHAR
}

// NewWriteAndCloseRequest creates a new WriteAndCloseRequest structure
//
// Returns:
// - A pointer to the new WriteAndCloseRequest structure
func NewWriteAndCloseRequest() *WriteAndCloseRequest {
	c := &WriteAndCloseRequest{
		// Parameters

		FID:                 types.USHORT(0),
		CountOfBytesToWrite: types.USHORT(0),
		WriteOffsetInBytes:  types.ULONG(0),
		LastWriteTime:       types.FILETIME{},
		Reserved:            [3]types.ULONG{0, 0, 0},

		// Data
		Pad: types.UCHAR(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_AND_CLOSE)

	return c
}

// Marshal marshals the WriteAndCloseRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndCloseRequest structure
// - An error if the marshaling fails
func (c *WriteAndCloseRequest) Marshal() ([]byte, error) {
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
	rawDataContent = append(rawDataContent, types.UCHAR(c.Pad))

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)
	c.CountOfBytesToWrite = types.USHORT(len(c.Data))

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter CountOfBytesToWrite
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.CountOfBytesToWrite))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter WriteOffsetInBytes
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.WriteOffsetInBytes))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter LastWriteTime
	bytesStream, err := c.LastWriteTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter Reserved
	if c.Reserved != [3]types.ULONG{0, 0, 0} {
		for _, reserved := range c.Reserved {
			buf4 = make([]byte, 4)
			binary.BigEndian.PutUint32(buf4, uint32(reserved))
			rawParametersContent = append(rawParametersContent, buf4...)
		}
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
func (c *WriteAndCloseRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter CountOfBytesToWrite
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CountOfBytesToWrite")
	}
	c.CountOfBytesToWrite = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter WriteOffsetInBytes
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for WriteOffsetInBytes")
	}
	c.WriteOffsetInBytes = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter LastWriteTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for LastWriteTime")
	}
	bytesRead, err = c.LastWriteTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter Reserved
	if c.GetParameters().WordCount == 12 {
		if len(rawParametersContent) < offset+12 {
			return offset, fmt.Errorf("rawParametersContent too short for Reserved")
		}
		c.Reserved = [3]types.ULONG{
			types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4])),
			types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset+4 : offset+8])),
			types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset+8 : offset+12])),
		}
		offset += 12
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
	c.Data = rawDataContent[offset : offset+int(c.CountOfBytesToWrite)]

	return offset, nil
}
