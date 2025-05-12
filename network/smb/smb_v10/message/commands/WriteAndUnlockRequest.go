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

// WriteAndUnlockRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c03fec3f-d709-4e7f-96b1-de9760766f77
type WriteAndUnlockRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the
	// file to which the data MUST be written.
	FID types.USHORT

	// CountOfBytesToWrite (2 bytes): This field is a 16-bit unsigned integer
	// indicating the number of bytes to be written to the file. The client MUST ensure
	// that the amount of data sent can fit in the negotiated maximum buffer size.
	CountOfBytesToWrite types.USHORT

	// WriteOffsetInBytes (4 bytes): This field is a 32-bit unsigned integer indicating
	// the offset, in number of bytes, from the beginning of the file at which to begin
	// writing to the file. The client MUST ensure that the amount of data sent can fit
	// in the negotiated maximum buffer size. Because this field is limited to 32 bits,
	// this command is inappropriate for files that have 64-bit offsets.
	WriteOffsetInBytes types.ULONG

	// EstimateOfRemainingBytesToBeWritten (2 bytes): This field is a 16-bit unsigned
	// integer indicating the remaining number of bytes that the client designates to
	// write to the file. This is an advisory field and MAY be zero. This information
	// can be used by the server to optimize cache behavior.
	EstimateOfRemainingBytesToBeWritten types.USHORT

	// Data

	// BufferFormat (1 byte): This field MUST be 0x01.
	// DataLength (2 bytes): This field MUST be CountOfBytesToWrite.
	// Data (variable): The raw bytes to be written to the file.
	Data types.SMB_STRING
}

// NewWriteAndUnlockRequest creates a new WriteAndUnlockRequest structure
//
// Returns:
// - A pointer to the new WriteAndUnlockRequest structure
func NewWriteAndUnlockRequest() *WriteAndUnlockRequest {
	c := &WriteAndUnlockRequest{
		// Parameters
		FID:                                 types.USHORT(0),
		CountOfBytesToWrite:                 types.USHORT(0),
		WriteOffsetInBytes:                  types.ULONG(0),
		EstimateOfRemainingBytesToBeWritten: types.USHORT(0),

		// Data
		Data: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_WRITE_AND_UNLOCK)

	return c
}

// Marshal marshals the WriteAndUnlockRequest structure into a byte array
//
// Returns:
// - A byte array representing the WriteAndUnlockRequest structure
// - An error if the marshaling fails
func (c *WriteAndUnlockRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Data
	c.Data.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	byteStream, err := c.Data.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, byteStream...)

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

	// Marshalling parameter EstimateOfRemainingBytesToBeWritten
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.EstimateOfRemainingBytesToBeWritten))
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
func (c *WriteAndUnlockRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter EstimateOfRemainingBytesToBeWritten
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for EstimateOfRemainingBytesToBeWritten")
	}
	c.EstimateOfRemainingBytesToBeWritten = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Data
	bytesRead, err = c.Data.Unmarshal(rawDataContent)
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	return offset, nil
}
