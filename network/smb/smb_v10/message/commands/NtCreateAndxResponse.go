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

// NtCreateAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/32085986-b516-486c-abbb-0abbdf9f1909
type NtCreateAndxResponse struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	AndXCommand types.UCHAR
	AndXReserved types.UCHAR
	AndXOffset types.USHORT
	OpLockLevel types.UCHAR
	FID types.USHORT
	CreateDisposition types.ULONG
	CreateTime types.FILETIME
	LastAccessTime types.FILETIME
	LastWriteTime types.FILETIME
	LastChangeTime types.FILETIME
	ExtFileAttributes types.SMB_EXT_FILE_ATTR
	AllocationSize types.LARGE_INTEGER
	EndOfFile types.LARGE_INTEGER
	ResourceType types.USHORT
	NMPipeStatus types.SMB_NMPIPE_STATUS
	Directory types.UCHAR

	// Data
	ByteCount types.USHORT

}

// NewNtCreateAndxResponse creates a new NtCreateAndxResponse structure
//
// Returns:
// - A pointer to the new NtCreateAndxResponse structure
func NewNtCreateAndxResponse() *NtCreateAndxResponse {
	c := &NtCreateAndxResponse{
		// Parameters
		WordCount: types.UCHAR(0),
		AndXCommand: types.UCHAR(0),
		AndXReserved: types.UCHAR(0),
		AndXOffset: types.USHORT(0),
		OpLockLevel: types.UCHAR(0),
		FID: types.USHORT(0),
		CreateDisposition: types.ULONG(0),
		CreateTime: types.FILETIME{},
		LastAccessTime: types.FILETIME{},
		LastWriteTime: types.FILETIME{},
		LastChangeTime: types.FILETIME{},
		ExtFileAttributes: types.SMB_EXT_FILE_ATTR{},
		AllocationSize: types.LARGE_INTEGER{},
		EndOfFile: types.LARGE_INTEGER{},
		ResourceType: types.USHORT(0),
		NMPipeStatus: types.SMB_NMPIPE_STATUS{},
		Directory: types.UCHAR(0),

		// Data
		ByteCount: types.USHORT(0),

	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_CREATE_ANDX)

	return c
}


// IsAndX returns true if the command is an AndX
func (c *NtCreateAndxResponse) IsAndX() bool {
	return true
}



// Marshal marshals the NtCreateAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the NtCreateAndxResponse structure
// - An error if the marshaling fails
func (c *NtCreateAndxResponse) Marshal() ([]byte, error) {
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
	
	// Marshalling parameter AndXCommand
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXCommand))
	
	// Marshalling parameter AndXReserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXReserved))
	
	// Marshalling parameter AndXOffset
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AndXOffset))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter OpLockLevel
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.OpLockLevel))
	
	// Marshalling parameter FID
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter CreateDisposition
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.CreateDisposition))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter CreateTime
	bytesStream, err := c.CreateTime.Marshal()
	if err != nil {
			return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)
	
	// Marshalling parameter LastAccessTime
	bytesStream, err := c.LastAccessTime.Marshal()
	if err != nil {
			return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)
	
	// Marshalling parameter LastWriteTime
	bytesStream, err := c.LastWriteTime.Marshal()
	if err != nil {
			return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)
	
	// Marshalling parameter LastChangeTime
	bytesStream, err := c.LastChangeTime.Marshal()
	if err != nil {
			return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)
	
	// Marshalling parameter ExtFileAttributes
	
	// Marshalling parameter AllocationSize
	
	// Marshalling parameter EndOfFile
	
	// Marshalling parameter ResourceType
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ResourceType))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter NMPipeStatus
	
	// Marshalling parameter Directory
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Directory))
	
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
func (c *NtCreateAndxResponse) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter AndXCommand
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for AndXCommand")
	}
	c.AndXCommand = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXReserved
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for AndXReserved")
	}
	c.AndXReserved = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter AndXOffset
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for AndXOffset")
	}
	c.AndXOffset = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter OpLockLevel
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for OpLockLevel")
	}
	c.OpLockLevel = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter CreateDisposition
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for CreateDisposition")
	}
	c.CreateDisposition = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter CreateTime
	if len(rawParametersContent) < offset+8 {
	    return offset, fmt.Errorf("rawParametersContent too short for CreateTime")
	}
	bytesRead, err := c.CreateTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling parameter LastAccessTime
	if len(rawParametersContent) < offset+8 {
	    return offset, fmt.Errorf("rawParametersContent too short for LastAccessTime")
	}
	bytesRead, err := c.LastAccessTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling parameter LastWriteTime
	if len(rawParametersContent) < offset+8 {
	    return offset, fmt.Errorf("rawParametersContent too short for LastWriteTime")
	}
	bytesRead, err := c.LastWriteTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling parameter LastChangeTime
	if len(rawParametersContent) < offset+8 {
	    return offset, fmt.Errorf("rawParametersContent too short for LastChangeTime")
	}
	bytesRead, err := c.LastChangeTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead
	
	// Unmarshalling parameter ExtFileAttributes
	
	// Unmarshalling parameter AllocationSize
	
	// Unmarshalling parameter EndOfFile
	
	// Unmarshalling parameter ResourceType
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for ResourceType")
	}
	c.ResourceType = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter NMPipeStatus
	
	// Unmarshalling parameter Directory
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for Directory")
	}
	c.Directory = types.UCHAR(rawParametersContent[offset])
	offset++
	
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
