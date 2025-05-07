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

// IoctlRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/c8f1b5b1-9ec1-49d2-a0e1-78ee88f39e71
type IoctlRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): The FID of the device or file to which the IOCTL is to be sent.
	FID types.USHORT

	// Category (2 bytes): The implementation-dependent device category for the
	// request.
	Category types.USHORT

	// Function (2 bytes): The implementation-dependent device function for the
	// request.
	Function types.USHORT

	// TotalParameterCount (2 bytes): The total number of IOCTL parameter bytes that
	// the client sends to the server in this request. Parameter bytes for an IOCTL are
	// carried within the SMB_Data.Parameters field of the SMB_COM_IOCTL request. This
	// value MUST be the same as ParameterCount.
	TotalParameterCount types.USHORT

	// TotalDataCount (2 bytes): The total number of IOCTL data bytes that the client
	// sends to the server in this request. Data bytes for an IOCTL are carried within
	// the SMB_Data.Data field of the SMB_COM_IOCTL request. This value MUST be the
	// same as DataCount.
	TotalDataCount types.USHORT

	// MaxParameterCount (2 bytes): The maximum number of SMB_Data.Parameters bytes
	// that the client accepts in the IOCTL response. The server MUST NOT return more
	// than this number of bytes in the SMB_Data.Parameters field of the response.
	MaxParameterCount types.USHORT

	// MaxDataCount (2 bytes): The maximum number of SMB_Data.Data bytes that the
	// client accepts in the IOCTL response. The server MUST NOT return more than this
	// number of bytes in the SMB_Data.Data field.
	MaxDataCount types.USHORT

	// Timeout (4 bytes): The value of this field MUST be the maximum number of
	// milliseconds that the server SHOULD wait for completion of the transaction
	// before generating a time-out and returning a response to the client. The client
	// SHOULD set this to 0x00000000 to indicate that no time-out is expected. A value
	// of 0x00000000 indicates that the server returns an error if the resource is not
	// immediately available. If the operation does not complete within the specified
	// time, the server aborts the request and sends a failure response.
	Timeout types.ULONG

	Reserved types.USHORT

	// ParameterCount (2 bytes): The number of IOCTL parameter bytes that the client
	// sends to the server in this request. Parameter bytes for an IOCTL are carried
	// within the SMB_Data.Parameters field of the SMB_COM_IOCTL request. This value
	// MUST be the same as TotalParameterCount.
	ParameterCount types.USHORT

	// ParameterOffset (2 bytes): The client SHOULD set the value of this field to
	// 0x0000. The server MUST ignore the value of this field.
	ParameterOffset types.USHORT

	// DataCount (2 bytes): The total number of IOCTL data bytes that the client sends
	// to the server in this request. Data bytes for an IOCTL are carried within the
	// SMB_Data.Data field of the SMB_COM_IOCTL request. This value MUST be the same as
	// TotalDataCount.
	DataCount types.USHORT

	// DataOffset (2 bytes): The client SHOULD set the value of this field to 0x0000.
	// The server MUST ignore the value of this field.
	DataOffset types.USHORT

	// Data

	// Pad1 (variable): An array of padding bytes used to align the next field to a
	// 2-byte or 4-byte boundary.
	Pad1 []types.UCHAR

	// Parameters (variable): IOCTL parameter bytes. The contents are implementation-dependent.
	Parameters []types.UCHAR

	// Pad2 (variable): An array of padding bytes, used to align the next field to a
	// 2-byte or 4-byte boundary.
	Pad2 []types.UCHAR

	// Data (variable): IOCTL data bytes. The contents are implementation-dependent.
	Data []types.UCHAR
}

// NewIoctlRequest creates a new IoctlRequest structure
//
// Returns:
// - A pointer to the new IoctlRequest structure
func NewIoctlRequest() *IoctlRequest {
	c := &IoctlRequest{
		// Parameters
		FID:                 types.USHORT(0),
		Category:            types.USHORT(0),
		Function:            types.USHORT(0),
		TotalParameterCount: types.USHORT(0),
		TotalDataCount:      types.USHORT(0),
		MaxParameterCount:   types.USHORT(0),
		MaxDataCount:        types.USHORT(0),
		Timeout:             types.ULONG(0),
		Reserved:            types.USHORT(0),
		ParameterCount:      types.USHORT(0),
		ParameterOffset:     types.USHORT(0),
		DataCount:           types.USHORT(0),
		DataOffset:          types.USHORT(0),

		// Data
		Pad1:       []types.UCHAR{},
		Parameters: []types.UCHAR{},
		Pad2:       []types.UCHAR{},
		Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_IOCTL)

	return c
}

// Marshal marshals the IoctlRequest structure into a byte array
//
// Returns:
// - A byte array representing the IoctlRequest structure
// - An error if the marshaling fails
func (c *IoctlRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Pad1
	rawDataContent = append(rawDataContent, c.Pad1...)

	// Marshalling data Parameters
	rawDataContent = append(rawDataContent, c.Parameters...)

	// Marshalling data Pad2
	rawDataContent = append(rawDataContent, c.Pad2...)

	// Marshalling data Data
	rawDataContent = append(rawDataContent, c.Data...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Category
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Category))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Function
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Function))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter MaxParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter MaxDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Timeout
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataCount))
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
func (c *IoctlRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Category
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Category")
	}
	c.Category = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Function
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Function")
	}
	c.Function = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TotalParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalParameterCount")
	}
	c.TotalParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TotalDataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TotalDataCount")
	}
	c.TotalDataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MaxParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxParameterCount")
	}
	c.MaxParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MaxDataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxDataCount")
	}
	c.MaxDataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
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

	// Unmarshalling parameter ParameterCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterCount")
	}
	c.ParameterCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ParameterOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterOffset")
	}
	c.ParameterOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataCount")
	}
	c.DataCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter DataOffset
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataOffset")
	}
	c.DataOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Pad1
	if len(rawDataContent) < offset+int(c.ParameterOffset) {
		return offset, fmt.Errorf("rawDataContent too short for Pad1")
	}
	c.Pad1 = rawDataContent[offset : offset+int(c.ParameterOffset)]
	offset += int(c.ParameterOffset)

	// Unmarshalling data Parameters
	if len(rawDataContent) < offset+int(c.ParameterCount) {
		return offset, fmt.Errorf("rawDataContent too short for Parameters")
	}
	c.Parameters = rawDataContent[offset : offset+int(c.ParameterCount)]
	offset += int(c.ParameterCount)

	// Unmarshalling data Pad2
	if len(rawDataContent) < offset+int(c.DataOffset) {
		return offset, fmt.Errorf("rawDataContent too short for Pad2")
	}
	c.Pad2 = rawDataContent[offset : offset+int(c.DataOffset)]
	offset += int(c.DataOffset)

	// Unmarshalling data Data
	if len(rawDataContent) < offset+int(c.DataCount) {
		return offset, fmt.Errorf("rawDataContent too short for Data")
	}
	c.Data = rawDataContent[offset : offset+int(c.DataCount)]
	offset += int(c.DataCount)

	return offset, nil
}
