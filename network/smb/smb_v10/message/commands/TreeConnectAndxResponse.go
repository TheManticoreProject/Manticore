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

// TreeConnectAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3286744b-5b58-4ad5-b62e-c4f29a2492f1
type TreeConnectAndxResponse struct {
	command_interface.Command

	// Parameters

	// OptionalSupport (2 bytes): A 16-bit field. The following OptionalSupport field
	// flags are defined. Any combination of the following flags MUST be supported. All
	// undefined values are considered reserved. The server SHOULD set them to 0, and
	// the client MUST ignore them.
	OptionalSupport types.USHORT

	// Data

	// Service (variable): The type of the shared resource to which the TID is connected.
	// The Service field MUST be encoded as a null-terminated array of OEM characters, even
	// if the client and server have negotiated to use Unicode strings. The valid values for
	// this field are as follows.
	Service types.OEM_STRING

	// NativeFileSystem (variable): The name of the file system on the local resource to
	// which the returned TID is connected. If SMB_FLAGS2_UNICODE is set in the Flags2 field
	// of the SMB Header of the response, this value MUST be a null-terminated string of Unicode
	// characters. Otherwise, this field MUST be a null-terminated string of OEM characters.
	// For resources that are not backed by a file system, such as the IPC$ share used for
	// named pipes, this field MUST be set to the empty string.
	NativeFileSystem types.SMB_STRING
}

// NewTreeConnectAndxResponse creates a new TreeConnectAndxResponse structure
//
// Returns:
// - A pointer to the new TreeConnectAndxResponse structure
func NewTreeConnectAndxResponse() *TreeConnectAndxResponse {
	c := &TreeConnectAndxResponse{
		// Parameters
		OptionalSupport: types.USHORT(0),

		// Data
		Service:          types.OEM_STRING{},
		NativeFileSystem: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TREE_CONNECT_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *TreeConnectAndxResponse) IsAndX() bool {
	return true
}

// Marshal marshals the TreeConnectAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the TreeConnectAndxResponse structure
// - An error if the marshaling fails
func (c *TreeConnectAndxResponse) Marshal() ([]byte, error) {
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

	// Marshalling data Service
	bytesStream, err := c.Service.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data NativeFileSystem
	bytesStream, err = c.NativeFileSystem.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter OptionalSupport
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.OptionalSupport))
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
func (c *TreeConnectAndxResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter OptionalSupport
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for OptionalSupport")
	}
	c.OptionalSupport = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Service
	bytesRead, err = c.Service.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data NativeFileSystem
	bytesRead, err = c.NativeFileSystem.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
