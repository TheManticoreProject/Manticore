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

// TreeConnectResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/f9a8a713-1c53-4fb0-908e-625389840cf8
type TreeConnectResponse struct {
	command_interface.Command

	// Parameters

	// MaxBufferSize (2 bytes): The maximum size, in bytes, of the largest SMB message
	// that the server can receive. This is the size of the largest SMB message that
	// the client can send to the server. SMB message size includes the size of the SMB
	// Header (section 2.2.3.1), parameter, and data blocks. This size MUST NOT include
	// any transport-layer framing or other transport-layer data.
	MaxBufferSize types.USHORT

	// TID (2 bytes): The newly generated Tree ID, used in subsequent CIFS client
	// requests to refer to a resource relative to the SMB_Data.Bytes.Path specified in
	// the request. Most access to the server requires a valid TID, whether the resource
	// is password protected or not. The value 0xFFFF is reserved; the server MUST NOT
	// return a TID value of 0xFFFF.
	TID types.USHORT
}

// NewTreeConnectResponse creates a new TreeConnectResponse structure
//
// Returns:
// - A pointer to the new TreeConnectResponse structure
func NewTreeConnectResponse() *TreeConnectResponse {
	c := &TreeConnectResponse{
		// Parameters
		MaxBufferSize: types.USHORT(0),
		TID:           types.USHORT(0),
	}

	c.Command.SetCommandCode(codes.SMB_COM_TREE_CONNECT)

	return c
}

// Marshal marshals the TreeConnectResponse structure into a byte array
//
// Returns:
// - A byte array representing the TreeConnectResponse structure
// - An error if the marshaling fails
func (c *TreeConnectResponse) Marshal() ([]byte, error) {
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter MaxBufferSize
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxBufferSize))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TID
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.TID))
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
func (c *TreeConnectResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter MaxBufferSize
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxBufferSize")
	}
	c.MaxBufferSize = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for TID")
	}
	c.TID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
