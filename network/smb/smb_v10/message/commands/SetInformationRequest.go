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

// SetInformationRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/76577ee1-eb2d-4db7-9bed-65c74a952741
type SetInformationRequest struct {
	command_interface.Command

	// Parameters

	// FileAttributes (2 bytes): This field is a 16-bit unsigned bit field encoded as
	// SMB_FILE_ATTRIBUTES (section 2.2.4.10.1)
	FileAttributes types.SMB_FILE_ATTRIBUTES

	// LastWriteTime (4 bytes): The time of the last write to the file, encoded as a
	// 4-byte UTIME (number of seconds since January 1, 1970 00:00:00.0).
	LastWriteTime types.ULONG

	// Reserved (10 bytes): This field is reserved, and all bytes MUST be set to 0x00
	// (USHORT Reserved[5]).
	Reserved [10]types.UCHAR

	// Data

	// BufferFormat (1 byte): This field MUST be 0x04.
	// FileName (variable): A null-terminated string that represents the fully
	// qualified name of the file relative to the supplied TID. This is the file for
	// which attributes are set.
	FileName types.SMB_STRING
}

// NewSetInformationRequest creates a new SetInformationRequest structure
//
// Returns:
// - A pointer to the new SetInformationRequest structure
func NewSetInformationRequest() *SetInformationRequest {
	c := &SetInformationRequest{
		// Parameters
		FileAttributes: types.SMB_FILE_ATTRIBUTES{},
		LastWriteTime:  types.ULONG(0),
		Reserved:       [10]types.UCHAR{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},

		// Data
		FileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_SET_INFORMATION)

	return c
}

// Marshal marshals the SetInformationRequest structure into a byte array
//
// Returns:
// - A byte array representing the SetInformationRequest structure
// - An error if the marshaling fails
func (c *SetInformationRequest) Marshal() ([]byte, error) {
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

	// Marshalling data FileName
	c.FileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FileAttributes
	bytesStream, err = c.FileAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter LastWriteTime (4-byte UTIME, little-endian)
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(c.LastWriteTime))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Reserved (10 bytes)
	rawParametersContent = append(rawParametersContent, c.Reserved[:]...)

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
func (c *SetInformationRequest) Unmarshal(rawData []byte) (int, error) {
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
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter FileAttributes
	bytesRead, err = c.FileAttributes.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastWriteTime (4-byte UTIME, little-endian)
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for LastWriteTime")
	}
	c.LastWriteTime = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+10 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	copy(c.Reserved[:], rawParametersContent[offset:offset+10])
	offset += 10

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data FileName
	bytesRead, err = c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
