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

// SessionSetupAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/e7514918-a0f6-4932-9f00-ced094445537
type SessionSetupAndxResponse struct {
	command_interface.Command

	// Parameters

	// Action (2 bytes): A 16-bit field. The two lowest-order bits have been defined:
	Action types.USHORT

	// Data

	// Pad (variable): Padding bytes. If Unicode support has been enabled, this field
	// MUST contain zero or one null padding byte as needed to ensure that the NativeOS
	// field, which follows, is aligned on a 16-bit boundary.
	Pad []types.UCHAR

	// NativeOS (variable): A string that represents the native operating system of the
	// server. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header of
	// the response, the string MUST be a null-terminated array of 16-bit Unicode
	// characters. Otherwise, the string MUST be a null-terminated array of OEM
	// characters. If the string consists of Unicode characters, this field MUST be
	// aligned to start on a 2-byte boundary from the start of the SMB header.<102>
	NativeOS types.SMB_STRING

	// NativeLanMan (variable): A string that represents the native LAN Manager type
	// of the server. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header
	// of the response, the string MUST be a null-terminated array of 16-bit Unicode
	// characters. Otherwise, the string MUST be a null-terminated array of OEM characters.
	// If the string consists of Unicode characters, this field MUST be aligned to start
	// on a 2-byte boundary from the start of the SMB header.
	NativeLanMan types.SMB_STRING

	// PrimaryDomain (variable): A string representing the primary domain or workgroup
	// name of the server. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header
	// of the response, the string MUST be a null-terminated array of 16-bit Unicode characters.
	// Otherwise, the string MUST be a null-terminated array of OEM characters. If the string
	// consists of Unicode characters, this field MUST be aligned to start on a 2-byte boundary
	// from the start of the SMB header.
	PrimaryDomain types.SMB_STRING
}

// NewSessionSetupAndxResponse creates a new SessionSetupAndxResponse structure
//
// Returns:
// - A pointer to the new SessionSetupAndxResponse structure
func NewSessionSetupAndxResponse() *SessionSetupAndxResponse {
	c := &SessionSetupAndxResponse{
		// Parameters
		Action: types.USHORT(0),

		// Data
		Pad:           []types.UCHAR{},
		NativeOS:      types.SMB_STRING{},
		NativeLanMan:  types.SMB_STRING{},
		PrimaryDomain: types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_SESSION_SETUP_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *SessionSetupAndxResponse) IsAndX() bool {
	return true
}

// Marshal marshals the SessionSetupAndxResponse structure into a byte array
//
// Returns:
// - A byte array representing the SessionSetupAndxResponse structure
// - An error if the marshaling fails
func (c *SessionSetupAndxResponse) Marshal() ([]byte, error) {
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

	// Marshalling data Pad
	rawDataContent = append(rawDataContent, c.Pad...)

	// Marshalling data NativeOS
	c.NativeOS.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	bytesStream, err := c.NativeOS.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data NativeLanMan
	c.NativeLanMan.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	bytesStream, err = c.NativeLanMan.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data PrimaryDomain
	c.PrimaryDomain.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	bytesStream, err = c.PrimaryDomain.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Action
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.Action))
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
func (c *SessionSetupAndxResponse) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter Action
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Action")
	}
	c.Action = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data Pad
	padLen := 0
	if (len(rawParametersContent)+3)%2 == 1 {
		padLen = 1
	}
	if len(rawDataContent) < offset+padLen {
		return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = rawDataContent[offset : offset+padLen]
	offset += padLen

	// Unmarshalling data NativeOS
	bytesRead, err = c.NativeOS.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data NativeLanMan
	bytesRead, err = c.NativeLanMan.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data PrimaryDomain
	bytesRead, err = c.PrimaryDomain.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
