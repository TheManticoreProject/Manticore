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

// NtCreateAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/f2a0f032-7545-41c9-9ceb-aab39852c11a
type NtCreateAndxRequest struct {
	command_interface.Command

	// Parameters
	WordCount types.UCHAR
	AndXCommand types.UCHAR
	AndXReserved types.UCHAR
	AndXOffset types.USHORT
	Reserved types.UCHAR
	NameLength types.USHORT
	Flags types.ULONG
	RootDirectoryFID types.ULONG
	DesiredAccess types.ULONG
	AllocationSize types.LARGE_INTEGER
	ExtFileAttributes types.SMB_EXT_FILE_ATTR
	ShareAccess types.ULONG
	CreateDisposition types.ULONG
	CreateOptions types.ULONG
	ImpersonationLevel types.ULONG
	SecurityFlags types.UCHAR

	// Data
	FileName types.SMB_STRING

}

// NewNtCreateAndxRequest creates a new NtCreateAndxRequest structure
//
// Returns:
// - A pointer to the new NtCreateAndxRequest structure
func NewNtCreateAndxRequest() *NtCreateAndxRequest {
	c := &NtCreateAndxRequest{
		// Parameters
		WordCount: types.UCHAR(0),
		AndXCommand: types.UCHAR(0),
		AndXReserved: types.UCHAR(0),
		AndXOffset: types.USHORT(0),
		Reserved: types.UCHAR(0),
		NameLength: types.USHORT(0),
		Flags: types.ULONG(0),
		RootDirectoryFID: types.ULONG(0),
		DesiredAccess: types.ULONG(0),
		AllocationSize: types.LARGE_INTEGER{},
		ExtFileAttributes: types.SMB_EXT_FILE_ATTR{},
		ShareAccess: types.ULONG(0),
		CreateDisposition: types.ULONG(0),
		CreateOptions: types.ULONG(0),
		ImpersonationLevel: types.ULONG(0),
		SecurityFlags: types.UCHAR(0),

		// Data
		FileName: types.SMB_STRING{},

	}

	c.Command.SetCommandCode(codes.SMB_COM_NT_CREATE_ANDX)

	return c
}


// IsAndX returns true if the command is an AndX
func (c *NtCreateAndxRequest) IsAndX() bool {
	return true
}



// Marshal marshals the NtCreateAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the NtCreateAndxRequest structure
// - An error if the marshaling fails
func (c *NtCreateAndxRequest) Marshal() ([]byte, error) {
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
	
	// Marshalling data FileName
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
			return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)
	
	// Then marshal the parameters
	rawParametersContent := []byte{}
	
	// Marshalling parameter WordCount
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.WordCount))
	
	// Marshalling parameter AndXCommand
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXCommand))
	
	// Marshalling parameter AndXReserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.AndXReserved))
	
	// Marshalling parameter AndXOffset
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AndXOffset))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Reserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved))
	
	// Marshalling parameter NameLength
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.NameLength))
	rawParametersContent = append(rawParametersContent, buf2...)
	
	// Marshalling parameter Flags
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Flags))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter RootDirectoryFID
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.RootDirectoryFID))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter DesiredAccess
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.DesiredAccess))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter AllocationSize
	
	// Marshalling parameter ExtFileAttributes
	
	// Marshalling parameter ShareAccess
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ShareAccess))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter CreateDisposition
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.CreateDisposition))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter CreateOptions
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.CreateOptions))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter ImpersonationLevel
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ImpersonationLevel))
	rawParametersContent = append(rawParametersContent, buf4...)
	
	// Marshalling parameter SecurityFlags
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.SecurityFlags))
	
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
func (c *NtCreateAndxRequest) Unmarshal(data []byte) (int, error) {
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
	
	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for Reserved")
	}
	c.Reserved = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Unmarshalling parameter NameLength
	if len(rawParametersContent) < offset+2 {
	    return offset, fmt.Errorf("rawParametersContent too short for NameLength")
	}
	c.NameLength = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset:offset+2]))
	offset += 2
	
	// Unmarshalling parameter Flags
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for Flags")
	}
	c.Flags = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter RootDirectoryFID
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for RootDirectoryFID")
	}
	c.RootDirectoryFID = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter DesiredAccess
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for DesiredAccess")
	}
	c.DesiredAccess = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter AllocationSize
	
	// Unmarshalling parameter ExtFileAttributes
	
	// Unmarshalling parameter ShareAccess
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for ShareAccess")
	}
	c.ShareAccess = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter CreateDisposition
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for CreateDisposition")
	}
	c.CreateDisposition = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter CreateOptions
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for CreateOptions")
	}
	c.CreateOptions = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter ImpersonationLevel
	if len(rawParametersContent) < offset+4 {
	    return offset, fmt.Errorf("rawParametersContent too short for ImpersonationLevel")
	}
	c.ImpersonationLevel = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset:offset+4]))
	offset += 4
	
	// Unmarshalling parameter SecurityFlags
	if len(rawParametersContent) < offset+1 {
	    return offset, fmt.Errorf("data too short for SecurityFlags")
	}
	c.SecurityFlags = types.UCHAR(rawParametersContent[offset])
	offset++
	
	// Then unmarshal the data
	offset = 0
	
	// Unmarshalling data FileName
	bytesRead, err := c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
	    return offset, err
	}
	offset += bytesRead

	return offset, nil
}
