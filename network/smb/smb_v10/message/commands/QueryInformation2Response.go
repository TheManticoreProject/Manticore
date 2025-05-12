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

// QueryInformation2Response
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/eed3d7c3-759e-470a-8731-10583931426f
type QueryInformation2Response struct {
	command_interface.Command

	// Parameters

	// CreateDate (2 bytes): This field is the date when the file was created.
	CreateDate types.SMB_DATE

	// CreateTime (2 bytes): This field is the time on CreateDate when the file was created.
	CreationTime types.SMB_TIME

	// LastAccessDate (2 bytes): This field is the date when the file was last accessed.
	LastAccessDate types.SMB_DATE

	// LastAccessTime (2 bytes): This field is the time on LastAccessDate when the file was last accessed.
	LastAccessTime types.SMB_TIME

	// LastWriteDate (2 bytes): This field is the date when data was last written to the file.
	LastWriteDate types.SMB_DATE

	// LastWriteTime (2 bytes): This field is the time on LastWriteDate when data was last written to the file.
	LastWriteTime types.SMB_TIME

	// FileDataSize (4 bytes): This field contains the number of bytes in the file, in bytes. Because this size
	// is limited to 32 bits, this command is inappropriate for files whose size is too large.
	FileDataSize types.ULONG

	// FileAllocationSize (4 bytes): This field contains the allocation size of the file, in bytes. Because
	// this size is limited to 32 bits, this command is inappropriate for files whose size is too large.
	FileAllocationSize types.ULONG

	// FileAttributes (2 bytes): This field is a 16-bit unsigned bit field encoding the attributes of the file.
	FileAttributes types.SMB_FILE_ATTRIBUTES
}

// NewQueryInformation2Response creates a new QueryInformation2Response structure
//
// Returns:
// - A pointer to the new QueryInformation2Response structure
func NewQueryInformation2Response() *QueryInformation2Response {
	c := &QueryInformation2Response{
		// Parameters

		CreateDate:         types.SMB_DATE{},
		CreationTime:       types.SMB_TIME{},
		LastAccessDate:     types.SMB_DATE{},
		LastAccessTime:     types.SMB_TIME{},
		LastWriteDate:      types.SMB_DATE{},
		LastWriteTime:      types.SMB_TIME{},
		FileDataSize:       types.ULONG(0),
		FileAllocationSize: types.ULONG(0),
		FileAttributes:     types.SMB_FILE_ATTRIBUTES{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_QUERY_INFORMATION2)

	return c
}

// Marshal marshals the QueryInformation2Response structure into a byte array
//
// Returns:
// - A byte array representing the QueryInformation2Response structure
// - An error if the marshaling fails
func (c *QueryInformation2Response) Marshal() ([]byte, error) {
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

	// Marshalling parameter CreateDate
	byteStream, err := c.CreateDate.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter CreationTime
	byteStream, err = c.CreationTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter LastAccessDate
	byteStream, err = c.LastAccessDate.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter LastAccessTime
	byteStream, err = c.LastAccessTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter LastWriteDate
	byteStream, err = c.LastWriteDate.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter LastWriteTime
	byteStream, err = c.LastWriteTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

	// Marshalling parameter FileDataSize
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileDataSize))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter FileAllocationSize
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileAllocationSize))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter FileAttributes
	byteStream, err = c.FileAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, byteStream...)

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
func (c *QueryInformation2Response) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter CreateDate
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CreateDate")
	}
	bytesRead, err = c.CreateDate.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter CreationTime
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for CreationTime")
	}
	bytesRead, err = c.CreationTime.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastAccessDate
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for LastAccessDate")
	}
	bytesRead, err = c.LastAccessDate.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastWriteDate
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for LastWriteDate")
	}
	bytesRead, err = c.LastWriteDate.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling parameter FileDataSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for FileDataSize")
	}
	c.FileDataSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter FileAllocationSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for FileAllocationSize")
	}
	c.FileAllocationSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter FileAttributes
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FileAttributes")
	}
	_, err = c.FileAttributes.Unmarshal(rawParametersContent[offset : offset+2])
	if err != nil {
		return 0, err
	}

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
