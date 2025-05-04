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
	WordCount types.UCHAR
	CreateDate types.SMB_DATE
	CreationTime types.SMB_TIME
	LastAccessDate types.SMB_DATE
	LastAccessTime types.SMB_TIME
	LastWriteDate types.SMB_DATE
	LastWriteTime types.SMB_TIME
	FileDataSize types.ULONG
	FileAllocationSize types.ULONG
	FileAttributes types.SMB_FILE_ATTRIBUTES

	// Data
	ByteCount types.USHORT

}

// NewQueryInformation2Response creates a new QueryInformation2Response structure
//
// Returns:
// - A pointer to the new QueryInformation2Response structure
func NewQueryInformation2Response() *QueryInformation2Response {
	c := &QueryInformation2Response{
		// Parameters
		WordCount: types.UCHAR(0),
		CreateDate: types.SMB_DATE{},
		CreationTime: types.SMB_TIME{},
		LastAccessDate: types.SMB_DATE{},
		LastAccessTime: types.SMB_TIME{},
		LastWriteDate: types.SMB_DATE{},
		LastWriteTime: types.SMB_TIME{},
		FileDataSize: types.ULONG(0),
		FileAllocationSize: types.ULONG(0),
		FileAttributes: types.SMB_FILE_ATTRIBUTES{},

		// Data
		ByteCount: types.USHORT(0),

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
	
	// Marshalling data ByteCount
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ByteCount))
	rawDataContent = append(rawDataContent, buf2...)
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter CreateDate
	
	// Marshalling parameter CreationTime
	
	// Marshalling parameter LastAccessDate
	
	// Marshalling parameter LastAccessTime
	
	// Marshalling parameter LastWriteDate
	
	// Marshalling parameter LastWriteTime
	
	// Marshalling parameter FileDataSize
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileDataSize))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter FileAllocationSize
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.FileAllocationSize))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter FileAttributes
	
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
	bytesRead, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0
	
	// Unmarshalling parameter WordCount
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for WordCount")
	}
	c.WordCount = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter CreateDate
	
	// Unmarshalling parameter CreationTime
	
	// Unmarshalling parameter LastAccessDate
	
	// Unmarshalling parameter LastAccessTime
	
	// Unmarshalling parameter LastWriteDate
	
	// Unmarshalling parameter LastWriteTime
	
	// Unmarshalling parameter FileDataSize
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for FileDataSize")
	}
	c.FileDataSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter FileAllocationSize
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for FileAllocationSize")
	}
	c.FileAllocationSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter FileAttributes
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data ByteCount
	if len(rawDataContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for ByteCount")
	}
	c.ByteCount = types.USHORT(binary.BigEndian.Uint16(rawDataContent[offset:offset+2]))
	offset += 2

	return offset, nil
}
