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

	// OpLockLevel (1 byte): The OpLock level granted to the client process.
	OpLockLevel types.UCHAR

	// FID (2 bytes): A FID representing the file or directory that was created or opened.
	FID types.USHORT

	// CreateDisposition (4 bytes): A 32-bit value that represents the action to take if the file already exists or if
	// the file is a new file and does not already exist.
	CreateDisposition types.ULONG

	// CreateTime (8 bytes): A 64-bit integer value representing the time that the file was created. The time value is
	// a signed 64-bit integer representing either an absolute time or a time interval. Times are specified in units of 100ns.
	// A positive value expresses an absolute time, where the base time (the 64- bit integer with value 0x0000000000000000) is
	// the beginning of the year 1601 AD in the Gregorian calendar. A negative value expresses a time interval relative to some
	// base time, usually the current time.
	CreateTime types.FILETIME

	// LastAccessTime (8 bytes): The time that the file was last accessed encoded in the same format as CreateTime.
	LastAccessTime types.FILETIME

	// LastWriteTime (8 bytes): The time that the file was last written, encoded in the same format as CreateTime.
	LastWriteTime types.FILETIME

	// LastChangeTime (8 bytes): The time that the file was last changed, encoded in the same format as CreateTime.
	LastChangeTime types.FILETIME

	// ExtFileAttributes (4 bytes): This field contains the extended file attributes that the server assigned to the file or
	// directory as a result of the command, encoded as an SMB_EXT_FILE_ATTR (section 2.2.1.2.3) data type.
	ExtFileAttributes types.SMB_EXT_FILE_ATTR

	// AllocationSize (8 bytes): The number of bytes allocated to the file by the server.
	AllocationSize types.LARGE_INTEGER

	// EndOfFile (8 bytes): The end of file offset value.
	EndOfFile types.LARGE_INTEGER

	// ResourceType (2 bytes): The file type. This field MUST be interpreted as follows.
	ResourceType types.USHORT

	// NMPipeStatus (2 bytes): A 16-bit field that shows the status of the named pipe if the resource type opened is a named pipe.
	// This field is formatted as an SMB_NMPIPE_STATUS (section 2.2.1.3).
	NMPipeStatus types.SMB_NMPIPE_STATUS

	// Directory (1 byte): If the returned FID represents a directory, the server MUST set this value to a nonzero value (0x01 is commonly used).
	// If the FID is not a directory, the server MUST set this value to 0x00 (FALSE).
	Directory types.UCHAR
}

// NewNtCreateAndxResponse creates a new NtCreateAndxResponse structure
//
// Returns:
// - A pointer to the new NtCreateAndxResponse structure
func NewNtCreateAndxResponse() *NtCreateAndxResponse {
	c := &NtCreateAndxResponse{
		// Parameters
		OpLockLevel:       types.UCHAR(0),
		FID:               types.USHORT(0),
		CreateDisposition: types.ULONG(0),
		CreateTime:        types.FILETIME{},
		LastAccessTime:    types.FILETIME{},
		LastWriteTime:     types.FILETIME{},
		LastChangeTime:    types.FILETIME{},
		ExtFileAttributes: types.SMB_EXT_FILE_ATTR(0),
		AllocationSize:    types.LARGE_INTEGER{},
		EndOfFile:         types.LARGE_INTEGER{},
		ResourceType:      types.USHORT(0),
		NMPipeStatus:      types.SMB_NMPIPE_STATUS{},
		Directory:         types.UCHAR(0),
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

	// Marshalling parameter OpLockLevel
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.OpLockLevel))

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
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
	bytesStream, err = c.LastAccessTime.Marshal()
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

	// Marshalling parameter LastChangeTime
	bytesStream, err = c.LastChangeTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter ExtFileAttributes
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ExtFileAttributes))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter AllocationSize
	buf8 := make([]byte, 8)
	binary.BigEndian.PutUint64(buf8, uint64(c.AllocationSize.QuadPart))
	rawParametersContent = append(rawParametersContent, buf8...)

	// Marshalling parameter EndOfFile
	buf8 = make([]byte, 8)
	binary.BigEndian.PutUint64(buf8, uint64(c.EndOfFile.QuadPart))
	rawParametersContent = append(rawParametersContent, buf8...)

	// Marshalling parameter ResourceType
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.ResourceType))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter NMPipeStatus
	bytesStream, err = c.NMPipeStatus.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

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
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter CreateDisposition
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for CreateDisposition")
	}
	c.CreateDisposition = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter CreateTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for CreateTime")
	}
	bytesRead, err = c.CreateTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter LastAccessTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for LastAccessTime")
	}
	bytesRead, err = c.LastAccessTime.Unmarshal(rawParametersContent[offset:])
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

	// Unmarshalling parameter LastChangeTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for LastChangeTime")
	}
	bytesRead, err = c.LastChangeTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter ExtFileAttributes
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ExtFileAttributes")
	}
	c.ExtFileAttributes = types.SMB_EXT_FILE_ATTR(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter AllocationSize
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for AllocationSize")
	}
	c.AllocationSize.QuadPart = uint64(binary.BigEndian.Uint64(rawParametersContent[offset : offset+8]))
	offset += 8

	// Unmarshalling parameter EndOfFile
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for EndOfFile")
	}
	c.EndOfFile.QuadPart = uint64(binary.BigEndian.Uint64(rawParametersContent[offset : offset+8]))
	offset += 8

	// Unmarshalling parameter ResourceType
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ResourceType")
	}
	c.ResourceType = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter NMPipeStatus
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for NMPipeStatus")
	}
	bytesRead, err = c.NMPipeStatus.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter Directory
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for Directory")
	}
	c.Directory = types.UCHAR(rawParametersContent[offset])
	offset++

	// Then unmarshal the data
	offset = 0
	// No data is sent by this message.

	return offset, nil
}
