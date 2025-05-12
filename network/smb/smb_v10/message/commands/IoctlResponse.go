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

// IoctlResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/27cb85fe-071a-41aa-9068-317720909892
type IoctlResponse struct {
	command_interface.Command

	// Parameters

	// TotalParameterCount (2 bytes): The total number of IOCTL parameter bytes that
	// the server sends to the client in this response. Parameter bytes for an IOCTL
	// are carried within the SMB_Data.Parameters field of the SMB_COM_IOCTL request.
	// This value MUST be the same as ParameterCount, and this value MUST be less than
	// or equal to the MaxParameterCount field value in the client's request.
	TotalParameterCount types.USHORT

	// TotalDataCount (2 bytes): The total number of IOCTL data bytes that the server
	// sends to the client in this response. Data bytes for an IOCTL are carried within
	// the SMB_Data.Data field of the SMB_COM_IOCTL request. This value MUST be the
	// same as DataCount, and this value MUST be less than or equal to the MaxDataCount
	// field value in the client's request.
	TotalDataCount types.USHORT

	// ParameterCount (2 bytes): The total number of IOCTL parameter bytes that the
	// server sends to the client in this response. Parameter bytes for an IOCTL are
	// carried within the SMB_Data.Parameters field of the SMB_COM_IOCTL request. This
	// value MUST be the same as TotalParameterCount and this value MUST be less than
	// or equal to the MaxParameterCount field value in the client's request.
	ParameterCount types.USHORT

	// ParameterOffset (2 bytes): This field MUST contain the number of bytes from the
	// start of the SMB Header (section 2.2.3.1) to the start of the
	// SMB_Data.Parameters field. Client implementations MUST use this value to locate
	// the IOCTL parameter block within the response.
	ParameterOffset types.USHORT

	ParameterDisplacement types.USHORT

	// DataCount (2 bytes): The total number of IOCTL data bytes that the server sends
	// to the client in this response. Data bytes for an IOCTL are carried within the
	// SMB_Data.Data field of the SMB_COM_IOCTL request. This value MUST be the same as
	// TotalDataCount, and this value MUST be less than or equal to the MaxDataCount
	// field value of the client's request.
	DataCount types.USHORT

	// DataOffset (2 bytes): This field MUST be the number of bytes from the start of
	// the SMB Header of the response to the start of the SMB_Data.Data field. Client
	// implementations MUST use this value to locate the IOCTL data block within the
	// response.
	DataOffset types.USHORT

	// DataDisplacement (2 bytes): The server SHOULD set the value of this field to
	// 0x0000. The client MUST ignore the value of this field.
	DataDisplacement types.USHORT

	// Data

	// Pad1 (variable): An array of padding bytes used to align the next field to a 16-
	// or 32-bit boundary.
	Pad1 []types.UCHAR

	// Parameters (variable): IOCTL parameter bytes. The contents are implementation-dependent.
	Parameters []types.UCHAR

	// Pad2 (variable): An array of padding bytes used to align the next field to a 16-
	// or 32-bit boundary.
	Pad2 []types.UCHAR

	// Data (variable): IOCTL data bytes. The contents are implementation-dependent.
	Data []types.UCHAR
}

// NewIoctlResponse creates a new IoctlResponse structure
//
// Returns:
// - A pointer to the new IoctlResponse structure
func NewIoctlResponse() *IoctlResponse {
	c := &IoctlResponse{
		// Parameters
		TotalParameterCount:   types.USHORT(0),
		TotalDataCount:        types.USHORT(0),
		ParameterCount:        types.USHORT(0),
		ParameterOffset:       types.USHORT(0),
		ParameterDisplacement: types.USHORT(0),
		DataCount:             types.USHORT(0),
		DataOffset:            types.USHORT(0),
		DataDisplacement:      types.USHORT(0),

		// Data
		Pad1:       []types.UCHAR{},
		Parameters: []types.UCHAR{},
		Pad2:       []types.UCHAR{},
		Data:       []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_IOCTL)

	return c
}

// Marshal marshals the IoctlResponse structure into a byte array
//
// Returns:
// - A byte array representing the IoctlResponse structure
// - An error if the marshaling fails
func (c *IoctlResponse) Marshal() ([]byte, error) {
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

	// Marshalling parameter TotalParameterCount
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TotalDataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TotalDataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ParameterDisplacement
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ParameterDisplacement))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataOffset))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter DataDisplacement
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.DataDisplacement))
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
func (c *IoctlResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter ParameterDisplacement
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ParameterDisplacement")
	}
	c.ParameterDisplacement = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
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

	// Unmarshalling parameter DataDisplacement
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DataDisplacement")
	}
	c.DataDisplacement = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
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
