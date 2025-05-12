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

// OpenAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/dbce00e7-68a1-41c6-982d-9483c902ad9b
type OpenAndxResponse struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): A valid FID representing the open instance of the file.
	FID types.USHORT

	// FileAttrs (2 bytes): The actual file system attributes of the file. If none of
	// the attribute bytes is set, the file attributes refer to a regular file.
	FileAttrs types.SMB_FILE_ATTRIBUTES

	// LastWriteTime (4 bytes): A 32-bit integer time value of the last modification to
	// the file.
	LastWriteTime types.FILETIME

	// FileDataSize (4 bytes): The number of bytes in the file. This field is advisory
	// and MAY be used.
	FileDataSize types.ULONG

	// AccessRights (2 bytes): A 16-bit field that shows granted access rights to the
	// file.
	AccessRights types.USHORT

	// ResourceType (2 bytes): A 16-bit field that shows the resource type opened.
	ResourceType types.USHORT

	// NMPipeStatus (2 bytes): A 16-bit field that contains the status of the named
	// pipe if the resource type opened is a named pipe.
	NMPipeStatus types.SMB_NMPIPE_STATUS

	// OpenResults (2 bytes): A 16-bit field that shows the results of the open operation.
	OpenResults types.USHORT

	// Reserved (6 bytes): Reserved and MUST be set to 0x00000000000000.
	Reserved [3]types.USHORT
}

// NewOpenAndxResponse creates a new OpenAndxResponse structure
//
// Returns:
// - A pointer to the new OpenAndxResponse structure
func NewOpenAndxResponse() *OpenAndxResponse {
	c := &OpenAndxResponse{
		// Parameters
		FID:           types.USHORT(0),
		FileAttrs:     types.SMB_FILE_ATTRIBUTES{},
		LastWriteTime: types.FILETIME{},
		FileDataSize:  types.ULONG(0),
		AccessRights:  types.USHORT(0),
		ResourceType:  types.USHORT(0),
		NMPipeStatus:  types.SMB_NMPIPE_STATUS{},
		OpenResults:   types.USHORT(0),
		Reserved:      [3]types.USHORT{0, 0, 0},
	}

	c.Command.SetCommandCode(codes.SMB_COM_OPEN_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *OpenAndxResponse) IsAndX() bool {
	return true
}

// Marshal marshals the OpenAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the OpenAndxResponse structure
// - An error if the marshaling fails
func (c *OpenAndxResponse) Marshal() ([]byte, error) {
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

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter FileAttrs
	bytesStream, err := c.FileAttrs.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter LastWriteTime
	bytesStream, err = c.LastWriteTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter FileDataSize
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileDataSize))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter AccessRights
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AccessRights))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter ResourceType
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ResourceType))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter NMPipeStatus

	// Marshalling parameter OpenResults
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.OpenResults))
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
func (c *OpenAndxResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter FileAttrs
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FileAttrs")
	}
	bytesRead, err = c.FileAttrs.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastWriteTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for LastWriteTime")
	}
	bytesRead, err = c.LastWriteTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter FileDataSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for FileDataSize")
	}
	c.FileDataSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter AccessRights
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for AccessRights")
	}
	c.AccessRights = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter ResourceType
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ResourceType")
	}
	c.ResourceType = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter NMPipeStatus

	// Unmarshalling parameter OpenResults
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for OpenResults")
	}
	c.OpenResults = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
