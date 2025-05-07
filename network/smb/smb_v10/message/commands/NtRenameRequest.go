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

// NtRenameRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/d777310e-deb1-490c-9157-26456c0b0116
type NtRenameRequest struct {
	command_interface.Command

	// Parameters

	// SearchAttributes (2 bytes): This field indicates the attributes that the target file(s) MUST have. If the attribute is 0x0000,
	// then only normal files are renamed or linked. If the system file or hidden attributes are specified, then the rename is inclusive of both special types.
	SearchAttributes types.SMB_FILE_ATTRIBUTES

	// InformationLevel (2 bytes): This field MUST be one of the three values shown in the following table.
	InformationLevel types.USHORT

	// Reserved (4 bytes): This field SHOULD be set to 0x00000000 by the client and MUST be ignored by the server.
	Reserved types.ULONG

	// Data

	// OldFileName (variable): A null-terminated string containing the full path name of the file to be manipulated.
	// Wildcards are not supported.
	OldFileName types.SMB_STRING

	// NewFileName (variable): A null-terminated string containing the new full path name to be assigned to the file
	// provided in OldFileName or the full path into which the file is to be moved.
	NewFileName types.SMB_STRING
}

// NewNtRenameRequest creates a new NtRenameRequest structure
//
// Returns:
// - A pointer to the new NtRenameRequest structure
func NewNtRenameRequest() *NtRenameRequest {
	c := &NtRenameRequest{
		// Parameters
		SearchAttributes: types.SMB_FILE_ATTRIBUTES{},
		InformationLevel: types.USHORT(0),
		Reserved:         types.ULONG(0),

		// Data
		OldFileName: types.SMB_STRING{},
		NewFileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_RENAME)

	return c
}

// Marshal marshals the NtRenameRequest structure into a byte array
//
// Returns:
// - A byte array representing the NtRenameRequest structure
// - An error if the marshaling fails
func (c *NtRenameRequest) Marshal() ([]byte, error) {
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

	// Marshalling data OldFileName
	c.OldFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.OldFileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data NewFileName
	c.NewFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err = c.NewFileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter SearchAttributes
	bytesStream, err = c.SearchAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter InformationLevel
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.InformationLevel))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Reserved))
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
func (c *NtRenameRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter SearchAttributes
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for SearchAttributes")
	}
	bytesRead, err = c.SearchAttributes.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter InformationLevel
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for InformationLevel")
	}
	c.InformationLevel = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data OldFileName
	bytesRead, err = c.OldFileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data NewFileName
	bytesRead, err = c.NewFileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
