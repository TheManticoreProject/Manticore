package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TreeConnectRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/0036eb81-7466-4e1c-afb6-ea8bc9dd19dc
type TreeConnectRequest struct {
	command_interface.Command

	// Data

	// BufferFormat1 (1 byte): A buffer format identifier. The value of this field MUST
	// be 0x04.
	// Path (variable): A null-terminated string that represents the server and share
	// name of the resource to which the client is attempting to connect. This field
	// MUST be encoded using Universal Naming Convention (UNC) syntax. The string MUST
	// be a null-terminated array of OEM characters, even if the client and server have
	// negotiated to use Unicode strings.
	Path types.OEM_STRING

	// BufferFormat2 (1 byte): A buffer format identifier. The value of this field MUST
	// be 0x04.
	// Password (variable): A null-terminated string that represents a share password
	// in plaintext form. The string MUST be a null-terminated array of OEM characters,
	// even if the client and server have negotiated to use Unicode strings.
	Password types.OEM_STRING

	// BufferFormat3 (1 byte): A buffer format identifier. The value of this field MUST
	// be 0x04.
	// Service (variable): A null-terminated string representing the type of resource
	// that the client intends to access. This field MUST be a null-terminated array of
	// OEM characters, even if the client and server have negotiated to use Unicode
	// strings. The valid values for this field are as follows:
	Service types.OEM_STRING
}

// NewTreeConnectRequest creates a new TreeConnectRequest structure
//
// Returns:
// - A pointer to the new TreeConnectRequest structure
func NewTreeConnectRequest() *TreeConnectRequest {
	c := &TreeConnectRequest{

		// Data
		Path:     types.OEM_STRING{},
		Password: types.OEM_STRING{},
		Service:  types.OEM_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_TREE_CONNECT)

	return c
}

// Marshal marshals the TreeConnectRequest structure into a byte array
//
// Returns:
// - A byte array representing the TreeConnectRequest structure
// - An error if the marshaling fails
func (c *TreeConnectRequest) Marshal() ([]byte, error) {
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

	// Marshalling data Path
	byteStream, err := c.Path.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, byteStream...)

	// Marshalling data Password
	byteStream, err = c.Password.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, byteStream...)

	// Marshalling data Service
	byteStream, err = c.Service.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, byteStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

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
func (c *TreeConnectRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	_ = c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0
	// No parameters are sent in this message

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Path
	bytesRead, err = c.Path.Unmarshal(rawDataContent)
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling data Password
	bytesRead, err = c.Password.Unmarshal(rawDataContent)
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshalling data Service
	bytesRead, err = c.Service.Unmarshal(rawDataContent)
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	return offset, nil
}
