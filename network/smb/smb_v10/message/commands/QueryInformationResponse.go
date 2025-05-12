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

// QueryInformationResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/847573c9-cbe6-4dcb-a0db-9b5af815759b
type QueryInformationResponse struct {
	command_interface.Command

	// Parameters

	// FileAttributes (2 bytes): This field is a 16-bit unsigned bit field encoded as
	// SMB_FILE_ATTRIBUTES (see section 2.2.1.2.4).
	FileAttributes types.SMB_FILE_ATTRIBUTES

	// LastWriteTime (4 bytes): The time of the last write to the file.
	LastWriteTime types.FILETIME

	// FileSize (4 bytes): This field contains the size of the file, in bytes. Because
	// this size is limited to 32 bits, this command is inappropriate for files whose
	// size is too large.
	FileSize types.ULONG

	// Reserved (10 bytes): This field is reserved, and all entries MUST be set to 0x00.
	Reserved [5]types.USHORT
}

// NewQueryInformationResponse creates a new QueryInformationResponse structure
//
// Returns:
// - A pointer to the new QueryInformationResponse structure
func NewQueryInformationResponse() *QueryInformationResponse {
	c := &QueryInformationResponse{
		// Parameters

		FileAttributes: types.SMB_FILE_ATTRIBUTES{},
		LastWriteTime:  types.FILETIME{},
		FileSize:       types.ULONG(0),
		Reserved:       [5]types.USHORT{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_QUERY_INFORMATION)

	return c
}

// Marshal marshals the QueryInformationResponse structure into a byte array
//
// Returns:
// - A byte array representing the QueryInformationResponse structure
// - An error if the marshaling fails
func (c *QueryInformationResponse) Marshal() ([]byte, error) {
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

	// Marshalling parameter FileAttributes
	byteStream, err := c.FileAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter LastWriteTime
	bytesStream, err := c.LastWriteTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter FileSize
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileSize))
	rawParametersContent = append(rawParametersContent, buf4...)

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
func (c *QueryInformationResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter FileAttributes
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FileAttributes")
	}
	bytesRead, err = c.FileAttributes.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastWriteTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for LastWriteTime")
	}
	bytesRead, err = c.LastWriteTime.Unmarshal(rawParametersContent[offset : offset+8])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter FileSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for FileSize")
	}
	c.FileSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
