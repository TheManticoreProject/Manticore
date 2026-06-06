package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// readAndxResponseParameterWords is the fixed number of 2-byte parameter words in
// an SMB_COM_READ_ANDX response: the 2-word AndX block plus Available,
// DataCompactionMode, Reserved1, DataLength, DataOffset, and the 5-word Reserved2.
const readAndxResponseParameterWords = 12

// readAndxResponseDataOffset is the offset, in bytes from the start of the SMB
// header, at which this response's data bytes begin when no pad is inserted:
// header + WordCount(1) + parameter words + ByteCount(2).
const readAndxResponseDataOffset = header.SMB_HEADER_SIZE + 1 + 2*readAndxResponseParameterWords + 2

// ReadAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/89d6b552-5406-445c-85d5-54c80b94a20f
type ReadAndxResponse struct {
	command_interface.Command

	// Parameters

	// Available (2 bytes): This field is valid when reading from named pipes. This
	// field indicates the number of bytes remaining to be read after the requested
	// read was completed.
	Available types.USHORT

	// DataCompactionMode (2 bytes): Reserved and MUST be 0x0000.
	DataCompactionMode types.USHORT

	// Reserved1 (2 bytes): This field MUST be 0x0000.
	Reserved1 types.USHORT

	// DataLength (2 bytes): The number of data bytes included in the response. If this
	// value is less than the value in the
	// Request.SMB_Parameters.MaxCountOfBytesToReturn field, it indicates that the read
	// operation has reached the end of the file (EOF).
	DataLength types.USHORT

	// DataOffset (2 bytes): The offset in bytes from the header of the read data.
	DataOffset types.USHORT

	// Reserved2 (10 bytes): Reserved. All entries MUST be 0x0000. The last 5 words are
	// reserved in order to make the SMB_COM_READ_ANDX Response the same size as the
	// SMB_COM_WRITE_ANDX Response.
	Reserved2 [5]types.USHORT

	// Data

	// Data (variable): The actual bytes read from the file. On the wire these are
	// located at DataOffset (measured from the start of the SMB header) and are
	// DataLength bytes long.
	Data []types.UCHAR
}

// NewReadAndxResponse creates a new ReadAndxResponse structure
//
// Returns:
// - A pointer to the new ReadAndxResponse structure
func NewReadAndxResponse() *ReadAndxResponse {
	c := &ReadAndxResponse{
		// Parameters

		Available:          types.USHORT(0),
		DataCompactionMode: types.USHORT(0),
		Reserved1:          types.USHORT(0),
		DataLength:         types.USHORT(0),
		DataOffset:         types.USHORT(0),

		// Data
		Data: []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_READ_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *ReadAndxResponse) IsAndX() bool {
	return true
}

// Marshal marshals the ReadAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the ReadAndxResponse structure
// - An error if the marshaling fails
func (c *ReadAndxResponse) Marshal() ([]byte, error) {
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

	// Place the file data immediately after the data block's ByteCount (no pad)
	// and advertise its length and its absolute offset from the start of the SMB
	// header so the parameter fields below carry the matching values.
	c.DataLength = types.USHORT(len(c.Data))
	c.DataOffset = types.USHORT(readAndxResponseDataOffset)

	rawDataContent := []byte{}
	rawDataContent = append(rawDataContent, c.Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Available
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.Available))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataCompactionMode
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataCompactionMode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved1
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.Reserved1))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataLength
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataLength))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved2
	for _, reserved := range c.Reserved2 {
		buf2 = make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(reserved))
		rawParametersContent = append(rawParametersContent, buf2...)
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
func (c *ReadAndxResponse) Unmarshal(rawData []byte) (int, error) {
	// Initialize the Parameters structure if it is nil to avoid a nil
	// pointer dereference when Unmarshal is called on a freshly constructed value.
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Initialize the Data structure if it is nil for the same reason.
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(rawData)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(rawData[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0
	if c.IsAndX() {
		offset += 4
	}

	// Unmarshalling parameter Available
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Available")
	}
	c.Available = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataCompactionMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataCompactionMode")
	}
	c.DataCompactionMode = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved1
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved1")
	}
	c.Reserved1 = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataLength
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataLength")
	}
	c.DataLength = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved2
	for i := range c.Reserved2 {
		if len(rawParametersContent) < offset+2 {
			return offset, fmt.Errorf("rawParametersContent too short for Reserved2")
		}
		c.Reserved2[i] = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2
	}

	// Then unmarshal the data. The file bytes are the trailing DataLength bytes of
	// the data block; any pad inserted to align the data to DataOffset precedes them.
	offset = 0
	if c.DataLength > 0 {
		if int(c.DataLength) > len(rawDataContent) {
			return offset, fmt.Errorf("ReadAndx DataLength %d exceeds data block size %d", c.DataLength, len(rawDataContent))
		}
		c.Data = append([]types.UCHAR{}, rawDataContent[len(rawDataContent)-int(c.DataLength):]...)
		offset = int(c.DataLength)
	} else {
		c.Data = []types.UCHAR{}
	}

	return offset, nil
}
