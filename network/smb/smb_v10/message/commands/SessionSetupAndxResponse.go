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
	"github.com/TheManticoreProject/Manticore/utils"
)

// SessionSetupAndxResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/e7514918-a0f6-4932-9f00-ced094445537
// SMBv1.0 extension: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/e5a467bc-cd36-4afa-825e-3f2a7bfd6189
type SessionSetupAndxResponse struct {
	command_interface.Command

	// Parameters

	// Action (2 bytes): A 16-bit field. The two lowest-order bits have been defined:
	Action types.USHORT

	// SecurityBlobLength (2 bytes): This value MUST specify the length, in bytes, of
	// the variable-length SecurityBlob that is contained within the response.
	SecurityBlobLength types.USHORT

	// Data

	// SecurityBlob (variable): This value MUST contain the authentication token being
	// returned to the client, as specified in section 3.3.5.3 and [RFC2743].
	SecurityBlob []types.UCHAR

	// Pad (variable): Padding bytes. If Unicode support has been enabled, this field
	// MUST contain zero or one null padding byte as needed to ensure that the NativeOS
	// field, which follows, is aligned on a 16-bit boundary.
	Pad []types.UCHAR

	// NativeOS (variable): A string that represents the native operating system of the server.
	// If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header of the response, then
	// the string MUST be a NULL-terminated array of 16-bit Unicode characters. Otherwise, the
	// string MUST be a NULL-terminated array of OEM characters. If the name string consists of
	// Unicode characters, then this field MUST be aligned to start on a 2-byte boundary from
	// the start of the SMB header.
	NativeOS []types.UCHAR

	// NativeLanMan (variable): A string that represents the native LAN Manager type of the
	// server. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header of the response,
	// then the string MUST be a NULL-terminated array of 16-bit Unicode characters. Otherwise,
	// the string MUST be a NULL-terminated array of OEM characters. If the name string consists
	// of Unicode characters, then this field MUST be aligned to start on a 2-byte boundary from
	// the start of the SMB header.
	NativeLanMan []types.UCHAR
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
		SecurityBlob: []types.UCHAR{},
		Pad:          []types.UCHAR{},
		NativeOS:     []types.UCHAR{},
		NativeLanMan: []types.UCHAR{},
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

	// Marshalling data SecurityBlob
	rawDataContent = append(rawDataContent, c.SecurityBlob...)
	c.SecurityBlobLength = types.USHORT(len(c.SecurityBlob))

	// Marshalling data Pad
	rawDataContent = append(rawDataContent, c.Pad...)

	// Marshalling data NativeOS
	rawDataContent = append(rawDataContent, c.NativeOS...)

	// Marshalling data NativeLanMan
	rawDataContent = append(rawDataContent, c.NativeLanMan...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter Action
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.Action))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SecurityBlobLength
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.SecurityBlobLength))
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
func (c *SessionSetupAndxResponse) Unmarshal(rawData []byte) (int, error) {
	// Initialize the Parameters structure if it is nil to avoid a nil
	// pointer dereference when Unmarshal is called on a freshly constructed value.
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Initialize the Data structure if it is nil for the same reason.
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(rawData)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(rawData[bytesRead:])
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
	if c.IsAndX() {
		offset += 4
	}

	// Unmarshalling parameter Action
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for Action")
	}
	c.Action = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SecurityBlobLength
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for SecurityBlobLength")
	}
	c.SecurityBlobLength = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data SecurityBlob
	if len(rawDataContent) < offset+int(c.SecurityBlobLength) {
		return offset, fmt.Errorf("rawDataContent too short for SecurityBlob")
	}
	c.SecurityBlob = rawDataContent[offset : offset+int(c.SecurityBlobLength)]
	offset += int(c.SecurityBlobLength)

	// Unmarshalling data Pad
	// padLen := 0
	// if (len(rawParametersContent)+3)%2 == 1 {
	// 	padLen = 1
	// }
	// if len(rawDataContent) < offset+padLen {
	// 	return offset, fmt.Errorf("rawParametersContent too short for Pad")
	// }
	// c.Pad = rawDataContent[offset : offset+padLen]
	// offset += padLen

	// Decode the NativeOS/NativeLanMan strings using the character encoding the
	// enclosing message declared via the SMB_FLAGS2_UNICODE header flag, which the
	// message layer propagates to this command before Unmarshal.
	useUnicode := c.IsUnicode()

	// Unmarshalling data NativeOS
	if useUnicode {
		nativeOSData, nativeOSLength := utils.ReadUntilNullTerminatorUTF16(rawDataContent[offset:])
		c.NativeOS = nativeOSData
		offset += nativeOSLength
	} else {
		nativeOSData, nativeOSLength := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		c.NativeOS = nativeOSData
		offset += nativeOSLength
	}

	// Unmarshalling data NativeLanMan
	if useUnicode {
		nativeLanManData, nativeLanManLength := utils.ReadUntilNullTerminatorUTF16(rawDataContent[offset:])
		c.NativeLanMan = nativeLanManData
		offset += nativeLanManLength
	} else {
		nativeLanManData, nativeLanManLength := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		c.NativeLanMan = nativeLanManData
		offset += nativeLanManLength
	}

	return offset, nil
}
