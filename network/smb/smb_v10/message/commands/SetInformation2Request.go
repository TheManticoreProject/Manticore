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

// SetInformation2Request
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/de521278-f800-4a57-b524-4811fe9edd8f
type SetInformation2Request struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This is the FID representing the file for which attributes are to
	// be set.
	FID types.USHORT

	// CreateDate (2 bytes): This is the date when the file was created.
	CreateDate types.SMB_DATE

	// CreateTime (2 bytes): This is the time on CreateDate when the file was created.
	CreationTime types.SMB_TIME

	// LastAccessDate (2 bytes): This is the date when the file was last accessed.
	LastAccessDate types.SMB_DATE

	// LastAccessTime (2 bytes): This is the time on LastAccessDate when the file was
	// last accessed.
	LastAccessTime types.SMB_TIME

	// LastWriteDate (2 bytes): This is the date when data was last written to the
	// file.
	LastWriteDate types.SMB_DATE

	// LastWriteTime (2 bytes): This is the time on LastWriteDate when data was last
	// written to the file.
	LastWriteTime types.SMB_TIME
}

// NewSetInformation2Request creates a new SetInformation2Request structure
//
// Returns:
// - A pointer to the new SetInformation2Request structure
func NewSetInformation2Request() *SetInformation2Request {
	c := &SetInformation2Request{
		// Parameters
		FID:            types.USHORT(0),
		CreateDate:     types.SMB_DATE{},
		CreationTime:   types.SMB_TIME{},
		LastAccessDate: types.SMB_DATE{},
		LastAccessTime: types.SMB_TIME{},
		LastWriteDate:  types.SMB_DATE{},
		LastWriteTime:  types.SMB_TIME{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_SET_INFORMATION2)

	return c
}

// Marshal marshals the SetInformation2Request structure into a byte array
//
// Returns:
// - A byte array representing the SetInformation2Request structure
// - An error if the marshaling fails
func (c *SetInformation2Request) Marshal() ([]byte, error) {
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter CreateDate
	marshalledCreateDate, err := c.CreateDate.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledCreateDate...)

	// Marshalling parameter CreationTime
	marshalledCreationTime, err := c.CreationTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledCreationTime...)

	// Marshalling parameter LastAccessDate
	marshalledLastAccessDate, err := c.LastAccessDate.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledLastAccessDate...)

	// Marshalling parameter LastAccessTime
	marshalledLastAccessTime, err := c.LastAccessTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledLastAccessTime...)

	// Marshalling parameter LastWriteDate
	marshalledLastWriteDate, err := c.LastWriteDate.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledLastWriteDate...)

	// Marshalling parameter LastWriteTime
	marshalledLastWriteTime, err := c.LastWriteTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, marshalledLastWriteTime...)

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
func (c *SetInformation2Request) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter CreateDate
	bytesRead, err = c.CreateDate.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter CreationTime
	bytesRead, err = c.CreationTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastAccessDate
	bytesRead, err = c.LastAccessDate.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastAccessTime
	bytesRead, err = c.LastAccessTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastWriteDate
	bytesRead, err = c.LastWriteDate.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastWriteTime
	bytesRead, err = c.LastWriteTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
