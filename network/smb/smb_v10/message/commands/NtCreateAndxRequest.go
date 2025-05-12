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

	// Reserved (1 byte): An unused value that SHOULD be set to 0x00 when sent and MUST be ignored on receipt.
	Reserved types.UCHAR

	// NameLength (2 bytes): This field MUST be the length of the FileName field (see following) in bytes.
	NameLength types.USHORT

	// Flags (4 bytes): A 32-bit field containing a set of flags that modify the client request. Unused bit
	// fields SHOULD be set to 0 when sent and MUST be ignored on receipt.
	Flags types.ULONG

	// RootDirectoryFID (4 bytes): If nonzero, this value is the File ID of an opened root directory, and the FileName
	// field MUST be handled as relative to the directory specified by this RootDirectoryFID. If this value is 0x00000000,
	// the FileName field MUST be handled as relative to the root of the share (the TID). The RootDirectoryFID MUST have
	// been acquired in a previous message exchange.
	RootDirectoryFID types.ULONG

	// DesiredAccess (4 bytes): A 32-bit field of flags that indicate standard, specific, and generic access rights.
	// These rights are used in access-control entries (ACEs) and are the primary means of specifying the requested or
	// granted access to an object. If this value is 0x00000000, it represents a request to query the attributes without
	// accessing the file.
	DesiredAccess types.ULONG

	// AllocationSize (8 bytes): The client MUST set this value to the initial allocation size of the file in bytes.
	// The server MUST ignore this field if this request is to open an existing file. This field MUST be used only if the
	// file is created or overwritten. The value MUST be set to 0x0000000000000000 in all other cases. This does not apply
	// to directory-related requests. This is the number of bytes to be allocated, represented as a 64-bit integer value.
	AllocationSize types.LARGE_INTEGER

	// ExtFileAttributes (4 bytes): This field contains the extended file attributes of the file being requested,
	// encoded as an SMB_EXT_FILE_ATTR (section 2.2.1.2.3) data type.
	ExtFileAttributes types.SMB_EXT_FILE_ATTR

	// ShareAccess (4 bytes): A 32-bit field that specifies how the file SHOULD be shared with other processes. The names
	// in the table below are provided for reference use only. If ShareAccess values of FILE_SHARE_READ, FILE_SHARE_WRITE, or
	// FILE_SHARE_DELETE are set for a printer file or a named pipe, the server SHOULD ignore these values. The value MUST be
	// FILE_SHARE_NONE or some combination of the other values:
	ShareAccess types.ULONG

	// CreateDisposition (4 bytes): A 32-bit value that represents the action to take if the file already exists or if the
	// file is a new file and does not already exist.
	CreateDisposition types.ULONG

	// CreateOptions (4 bytes): A 32-bit field containing flag options to use if creating the file or directory.
	// This field MUST be set to 0x00000000 or a combination of the following possible values. Unused bit fields SHOULD be
	// set to 0 when sent and MUST be ignored on receipt. The following is a list of the valid values and their associated behaviors.
	// Server implementations SHOULD reserve all bits not specified in the following definitions.
	CreateOptions types.ULONG

	// ImpersonationLevel (4 bytes): This field specifies the impersonation level requested by the application that is issuing
	// the create request, and MUST contain one of the following values.
	ImpersonationLevel types.ULONG

	// SecurityFlags (1 byte): An 8-bit field containing a set of options that specify the security tracking mode. These options
	// specify whether the server is to be given a snapshot of the client's security context (called static tracking) or is to be
	// continually updated to track changes to the client's security context (called dynamic tracking). When bit 0 of the SecurityFlags
	// field is clear, static tracking is requested. When bit 0 of the SecurityFlags field is set, dynamic tracking is requested.
	// Unused bit fields SHOULD be set to 0 when sent and MUST be ignored on receipt. This field MUST be set to 0x00 or a combination
	// of the following possible values. Value names are provided for convenience only. Supported values are:
	SecurityFlags types.UCHAR

	// Data

	// FileName (variable): A string that represents the fully qualified name of the file relative to the supplied TID to create or
	// truncate on the server. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB Header of the request, the FileName string
	// MUST be a null-terminated array of 16-bit Unicode characters. Otherwise, the FileName string MUST be a null-terminated array of
	// extended ASCII (OEM) characters. If the FileName string consists of Unicode characters, this field MUST be aligned to start on
	// a 2-byte boundary from the start of the SMB Header. When opening a named pipe, the FileName field MUST contain only the relative
	// name of the pipe, that is, the "\PIPE\" prefix MUST NOT be present. This is in contrast with other commands, such as
	// SMB_COM_OPEN_ANDX and TRANS2_OPEN2, which require that the "\PIPE" prefix be present in the pathname.
	FileName types.SMB_STRING
}

// NewNtCreateAndxRequest creates a new NtCreateAndxRequest structure
//
// Returns:
// - A pointer to the new NtCreateAndxRequest structure
func NewNtCreateAndxRequest() *NtCreateAndxRequest {
	c := &NtCreateAndxRequest{
		// Parameters
		Reserved:           types.UCHAR(0),
		NameLength:         types.USHORT(0),
		Flags:              types.ULONG(0),
		RootDirectoryFID:   types.ULONG(0),
		DesiredAccess:      types.ULONG(0),
		AllocationSize:     types.LARGE_INTEGER{},
		ExtFileAttributes:  types.SMB_EXT_FILE_ATTR(0),
		ShareAccess:        types.ULONG(0),
		CreateDisposition:  types.ULONG(0),
		CreateOptions:      types.ULONG(0),
		ImpersonationLevel: types.ULONG(0),
		SecurityFlags:      types.UCHAR(0),

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
	c.FileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Reserved
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.Reserved))

	// Marshalling parameter NameLength
	buf2 := make([]byte, 2)
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
	buf8 := make([]byte, 8)
	binary.BigEndian.PutUint64(buf8, uint64(c.AllocationSize.QuadPart))
	rawParametersContent = append(rawParametersContent, buf8...)

	// Marshalling parameter ExtFileAttributes
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.ExtFileAttributes))
	rawParametersContent = append(rawParametersContent, buf4...)

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
	c.NameLength = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Flags
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Flags")
	}
	c.Flags = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter RootDirectoryFID
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for RootDirectoryFID")
	}
	c.RootDirectoryFID = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter DesiredAccess
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for DesiredAccess")
	}
	c.DesiredAccess = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter AllocationSize
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for AllocationSize")
	}
	c.AllocationSize.QuadPart = uint64(binary.BigEndian.Uint64(rawParametersContent[offset : offset+8]))
	offset += 8

	// Unmarshalling parameter ExtFileAttributes
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ExtFileAttributes")
	}
	c.ExtFileAttributes = types.SMB_EXT_FILE_ATTR(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter ShareAccess
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ShareAccess")
	}
	c.ShareAccess = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter CreateDisposition
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for CreateDisposition")
	}
	c.CreateDisposition = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter CreateOptions
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for CreateOptions")
	}
	c.CreateOptions = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter ImpersonationLevel
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for ImpersonationLevel")
	}
	c.ImpersonationLevel = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
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
	bytesRead, err = c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
