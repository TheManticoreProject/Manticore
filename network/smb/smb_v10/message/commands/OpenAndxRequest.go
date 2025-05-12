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

// OpenAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3a760987-f60d-4012-930b-fe90328775cc
type OpenAndxRequest struct {
	command_interface.Command

	// Parameters

	// Flags (2 bytes): A 16-bit field of flags for requesting attribute data and
	// locking.
	Flags types.USHORT

	// AccessMode (2 bytes): A 16-bit field for encoding the requested access mode. See
	// section 3.2.4.5.1 for a discussion on sharing modes.
	AccessMode types.USHORT

	// SearchAttrs (2 bytes): The set of attributes that the file MUST have in order to
	// be found. If none of the attribute bytes is set, the file attributes MUST refer
	// to a regular file.
	SearchAttrs types.SMB_FILE_ATTRIBUTES

	// FileAttrs (2 bytes): The set of attributes that the file is to have if the file
	// needs to be created. If none of the attribute bytes is set, the file attributes
	// MUST refer to a regular file.
	FileAttrs types.SMB_FILE_ATTRIBUTES

	// CreationTime (4 bytes): A 32-bit integer time value to be assigned to the file
	// as the time of creation if the file is created.
	CreationTime types.FILETIME

	// OpenMode (2 bytes): A 16-bit field that controls the way a file SHOULD be
	// treated when it is opened for use by certain extended SMB requests.
	OpenMode types.USHORT

	// AllocationSize (4 bytes): The number of bytes to reserve on file creation or
	// truncation. This field MAY be ignored by the server.
	AllocationSize types.ULONG

	// Timeout (4 bytes): This field is a 32-bit unsigned integer value containing the
	// number of milliseconds to wait on a blocked open request before returning
	// without successfully opening the file.
	Timeout types.ULONG

	// Reserved (2 bytes): Reserved and MUST be set to 0x0000.
	Reserved [2]types.USHORT

	// Data

	// FileName (variable): A buffer containing the name of the file to be opened.
	FileName types.SMB_STRING
}

// NewOpenAndxRequest creates a new OpenAndxRequest structure
//
// Returns:
// - A pointer to the new OpenAndxRequest structure
func NewOpenAndxRequest() *OpenAndxRequest {
	c := &OpenAndxRequest{
		// Parameters
		Flags:          types.USHORT(0),
		AccessMode:     types.USHORT(0),
		SearchAttrs:    types.SMB_FILE_ATTRIBUTES{},
		FileAttrs:      types.SMB_FILE_ATTRIBUTES{},
		CreationTime:   types.FILETIME{},
		OpenMode:       types.USHORT(0),
		AllocationSize: types.ULONG(0),
		Timeout:        types.ULONG(0),
		Reserved:       [2]types.USHORT{0, 0},

		// Data
		FileName: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_OPEN_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *OpenAndxRequest) IsAndX() bool {
	return true
}

// Marshal marshals the OpenAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the OpenAndxRequest structure
// - An error if the marshaling fails
func (c *OpenAndxRequest) Marshal() ([]byte, error) {
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
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Flags
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Flags))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter AccessMode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.AccessMode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SearchAttrs
	bytesStream, err = c.SearchAttrs.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter FileAttrs
	bytesStream, err = c.FileAttrs.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter CreationTime
	bytesStream, err = c.CreationTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameter OpenMode
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.OpenMode))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter AllocationSize
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.AllocationSize))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Timeout
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Reserved
	buf2 = make([]byte, 2)
	for i := range c.Reserved {
		binary.BigEndian.PutUint16(buf2, uint16(c.Reserved[i]))
		rawParametersContent = append(rawParametersContent, buf2...)
	}

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
func (c *OpenAndxRequest) Unmarshal(data []byte) (int, error) {
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

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter Flags
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Flags")
	}
	c.Flags = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter AccessMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for AccessMode")
	}
	c.AccessMode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SearchAttrs
	bytesRead, err = c.SearchAttrs.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter FileAttrs
	bytesRead, err = c.FileAttrs.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter CreationTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for CreationTime")
	}
	bytesRead, err = c.CreationTime.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling parameter OpenMode
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for OpenMode")
	}
	c.OpenMode = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter AllocationSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for AllocationSize")
	}
	c.AllocationSize = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	for i := range c.Reserved {
		c.Reserved[i] = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2
	}

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
