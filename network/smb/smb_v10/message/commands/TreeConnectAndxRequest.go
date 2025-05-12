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

// TreeConnectAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/90bf689a-8536-4f03-9f1b-683ee4bdd67c
type TreeConnectAndxRequest struct {
	command_interface.Command

	// Parameters

	// Flags (2 bytes): A 16-bit field used to modify the SMB_COM_TREE_CONNECT_ANDX
	// Request (section 2.2.4.55.1). The client MUST set reserved values to 0, and the
	// server MUST ignore them.
	Flags types.USHORT

	// PasswordLength (2 bytes): This field MUST be the length, in bytes, of the
	// SMB_Data.Bytes.Password field.
	PasswordLength types.USHORT

	// Data

	// Password (variable): An array of bytes.
	//
	// - If the server is operating in share level access control mode and plaintext
	// passwords have been negotiated, then the Password MUST be an OEM_STRING
	// representing the user's password in plaintext.
	//
	// - If the server is operating in share level access control mode and
	// challenge/response authentication has been negotiated, then the Password
	// MUST be an authentication response.
	//
	// - If authentication is not used, then the Password SHOULD be a single null
	// padding byte (which takes the place of the Pad[] byte).
	//
	// The SMB_Parameters.Bytes.PasswordLength MUST be the full length of the
	// Password field. If the Password is the null padding byte, the password
	// length is 1.
	Password []types.UCHAR

	// Pad (variable): Padding bytes. If Unicode support has been enabled and
	// SMB_FLAGS2_UNICODE is set in SMB_Header.Flags2, this field MUST contain zero or
	// one null padding bytes as needed to ensure that the Path string is aligned on a
	// 16-bit boundary.
	Pad []types.UCHAR

	// Path (variable): A null-terminated string that represents the server and share
	// name of the resource to which the client attempts to connect. This field MUST be
	// encoded using Universal Naming Convention (UNC) syntax. If SMB_FLAGS2_UNICODE is
	// set in the Flags2 field of the SMB Header of the request, the string MUST be a
	// null-terminated array of 16-bit Unicode characters. Otherwise, the string MUST
	// be a null-terminated array of OEM characters. If the string consists of Unicode
	// characters, this field MUST be aligned to start on a 2-byte boundary from the
	// start of the SMB Header. A path in UNC syntax would be represented by a string
	// in the following form:
	Path types.SMB_STRING

	// Service (variable): The type of resource that the client attempts to access.
	// This field MUST be a null-terminated array of OEM characters even if the client
	// and server have negotiated to use Unicode strings. The valid values for this
	// field are as follows:
	Service types.OEM_STRING
}

// NewTreeConnectAndxRequest creates a new TreeConnectAndxRequest structure
//
// Returns:
// - A pointer to the new TreeConnectAndxRequest structure
func NewTreeConnectAndxRequest() *TreeConnectAndxRequest {
	c := &TreeConnectAndxRequest{
		// Parameters
		Flags:          types.USHORT(0),
		PasswordLength: types.USHORT(0),

		// Data
		Pad:     []types.UCHAR{},
		Path:    types.SMB_STRING{},
		Service: types.OEM_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TREE_CONNECT_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *TreeConnectAndxRequest) IsAndX() bool {
	return true
}

// Marshal marshals the TreeConnectAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the TreeConnectAndxRequest structure
// - An error if the marshaling fails
func (c *TreeConnectAndxRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Password
	rawDataContent = append(rawDataContent, c.Password...)
	c.PasswordLength = types.USHORT(len(c.Password))

	// Marshalling data Pad
	rawDataContent = append(rawDataContent, c.Pad...)

	// Marshalling data Path
	bytesStream, err := c.Path.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data Service
	bytesStream, err = c.Service.Marshal()
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

	// Marshalling parameter PasswordLength
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.PasswordLength))
	rawParametersContent = append(rawParametersContent, buf2...)

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
func (c *TreeConnectAndxRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter PasswordLength
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for PasswordLength")
	}
	c.PasswordLength = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Password
	if len(rawDataContent) < offset+int(c.PasswordLength) {
		return offset, fmt.Errorf("rawParametersContent too short for Password")
	}
	c.Password = rawDataContent[offset : offset+int(c.PasswordLength)]
	offset += int(c.PasswordLength)

	// Unmarshalling data Pad
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawDataContent too short for Pad")
	}
	c.Pad = rawDataContent[offset : offset+1]
	offset++

	// Unmarshalling data Path
	bytesRead, err = c.Path.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data Service
	bytesRead, err = c.Service.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
