package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SessionSetupAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/81e15dee-8fb6-4102-8644-7eaa7ded63f7
type SessionSetupAndxRequest struct {
	command_interface.Command

	// Parameters

	// MaxBufferSize (2 bytes): The maximum size, in bytes, of the largest SMB
	// message that the client can receive. This is the size of the largest SMB message
	// that the server can send to the client. SMB message size includes the size of
	// the SMB header, parameter, and data blocks. This size MUST NOT include any
	// transport-layer framing or other transport-layer data.
	MaxBufferSize types.USHORT

	// MaxMpxCount (2 bytes): The maximum number of pending requests supported by the
	// client. This value MUST be less than or equal to the MaxMpxCount field value
	// provided by the server in the SMB_COM_NEGOTIATE Response.
	MaxMpxCount types.USHORT

	// VcNumber (2 bytes): The number of this VC (virtual circuit) between the client
	// and the server. This field SHOULD be set to a value of 0x0000 for the first
	// virtual circuit between the client and the server and it SHOULD be set to a
	// unique nonzero value for each additional virtual circuit.
	VcNumber types.USHORT

	// SessionKey (4 bytes): The client MUST set this field to be equal to the
	// SessionKey field in the SMB_COM_NEGOTIATE Response for this SMB connection.
	SessionKey types.ULONG

	// If SMB_FLAGS2_UNICODE is set (1), the value of OEMPasswordLen MUST be 0x0000 and
	// the password MUST be encoded using UTF-16LE Unicode. Padding MUST NOT be added
	// to align this plaintext Unicode string to a word boundary.
	OEMPasswordLen types.USHORT

	// If SMB_FLAGS2_UNICODE is clear (0), the value of UnicodePasswordLen MUST be
	// 0x0000, and the password MUST be encoded using the 8-bit OEM character set
	// (extended ASCII).
	UnicodePasswordLen types.USHORT

	// Reserved (4 bytes): Reserved. This field MUST be 0x00000000. The server MUST
	// ignore the contents of this field.
	Reserved types.ULONG

	// Capabilities (4 bytes): A 32-bit field providing a set of client capability
	// indicators. The client uses this field to report its own set of capabilities to
	// the server. The client capabilities are a subset of the server capabilities.
	Capabilities capabilities.Capabilities

	// Data

	// The OEMPassword value is an array of bytes, not a null-terminated string.
	OEMPassword []types.UCHAR

	// UnicodePassword (variable): The contents of this field depends upon the
	// authentication methods in use (See section 3.2.4.2.4 for a description of
	// authentication mechanisms used with CIFS.):
	//
	//   - If Unicode has been negotiated and the client sends a plaintext password,
	// this field MUST contain the password represented in UTF-16LE Unicode.
	//
	//   - If the client uses challenge/response authentication, this field can contain
	// a cryptographic response.
	//
	//   - This field MAY be empty.
	//
	//   - If the client sends a plaintext password, then the password MUST be encoded
	// in either OEM or Unicode characters, but not both. The value of the SMB_FLAGS2_UNICODE
	// bit of the SMB_Header.Flags2 indicates the character encoding of the password.
	//
	//   - If a plaintext password is sent, then:
	//       + If SMB_FLAGS2_UNICODE is clear (0), the value of UnicodePasswordLen MUST be 0x0000,
	//     and the password MUST be encoded using the 8-bit OEM character set (extended ASCII).
	//       + If SMB_FLAGS2_UNICODE is set (1), the value of OEMPasswordLen MUST be 0x0000 and the
	// password MUST be encoded using UTF-16LE Unicode. Padding MUST NOT be added to align this
	// plaintext Unicode string to a word boundary.
	UnicodePassword []types.UCHAR

	// Pad (variable): Padding bytes. If Unicode support has been enabled and
	// SMB_FLAGS2_UNICODE is set in SMB_Header.Flags2, this field MUST contain zero
	// (0x00) or one null padding byte as needed to ensure that the AccountName string
	// is aligned on a 16-bit boundary. This also forces alignment of subsequent
	// strings without additional padding.
	Pad []types.UCHAR

	// AccountName (variable): The name of the account (username) with which the user
	// authenticates.
	AccountName types.SMB_STRING

	// PrimaryDomain (variable): A string representing the desired authentication domain.
	// This MAY be the empty string. If SMB_FLAGS2_UNICODE is set in the Flags2 field of
	// the SMB header of the request, this string MUST be a null-terminated array of
	// 16-bit Unicode characters. Otherwise, this string MUST be a null-terminated array
	// of OEM characters. If this string consists of Unicode characters, this field
	// MUST be aligned to start on a 2-byte boundary from the start of the SMB header.
	PrimaryDomain types.SMB_STRING

	// NativeOS (variable): A string representing the native operating system of the
	// CIFS client. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header
	// of the request, this string MUST be a null-terminated array of 16-bit Unicode
	// characters. Otherwise, this string MUST be a null-terminated array of OEM
	// characters. If this string consists of Unicode characters, this field MUST be
	// aligned to start on a 2-byte boundary from the start of the SMB header.
	NativeOS types.SMB_STRING

	// NativeLanMan (variable): A string that represents the native LAN manager type
	// of the client. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header
	// of the request, this string MUST be a null-terminated array of 16-bit Unicode
	// characters. Otherwise, this string MUST be a null-terminated array of OEM
	// characters. If this string consists of Unicode characters, this field MUST be
	// aligned to start on a 2-byte boundary from the start of the SMB header.
	NativeLanMan types.SMB_STRING
}

// NewSessionSetupAndxRequest creates a new SessionSetupAndxRequest structure
//
// Returns:
// - A pointer to the new SessionSetupAndxRequest structure
func NewSessionSetupAndxRequest() *SessionSetupAndxRequest {
	c := &SessionSetupAndxRequest{
		// Parameters
		MaxBufferSize:      types.USHORT(0),
		MaxMpxCount:        types.USHORT(0),
		VcNumber:           types.USHORT(0),
		SessionKey:         types.ULONG(0),
		OEMPasswordLen:     types.USHORT(0),
		UnicodePasswordLen: types.USHORT(0),
		Reserved:           types.ULONG(0),
		Capabilities:       capabilities.Capabilities(0),

		// Data
		OEMPassword:     []types.UCHAR{},
		UnicodePassword: []types.UCHAR{},
		Pad:             []types.UCHAR{},
		AccountName:     types.SMB_STRING{},
		PrimaryDomain:   types.SMB_STRING{},
		NativeOS:        types.SMB_STRING{},
		NativeLanMan:    types.SMB_STRING{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_SESSION_SETUP_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *SessionSetupAndxRequest) IsAndX() bool {
	return true
}

// Marshal marshals the SessionSetupAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the SessionSetupAndxRequest structure
// - An error if the marshaling fails
func (c *SessionSetupAndxRequest) Marshal() ([]byte, error) {
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

	// Marshalling data OEMPassword
	rawDataContent = append(rawDataContent, c.OEMPassword...)

	// Marshalling data UnicodePassword
	rawDataContent = append(rawDataContent, c.UnicodePassword...)

	// Marshalling data Pad
	rawDataContent = append(rawDataContent, c.Pad...)

	// Marshalling data AccountName
	c.AccountName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.AccountName.Marshal()
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

	// Marshalling data NativeOS
	c.NativeOS.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	bytesStream, err = c.NativeOS.Marshal()
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter MaxBufferSize
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxBufferSize))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter MaxMpxCount
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxMpxCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter VcNumber
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.VcNumber))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SessionKey
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.SessionKey))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter OEMPasswordLen
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.OEMPasswordLen))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter UnicodePasswordLen
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.UnicodePasswordLen))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter Reserved
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Capabilities
	buf4 = make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Capabilities))
	rawParametersContent = append(rawParametersContent, buf4...)

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
func (c *SessionSetupAndxRequest) Unmarshal(data []byte) (int, error) {
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

	// Unmarshalling parameter MaxBufferSize
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxBufferSize")
	}
	c.MaxBufferSize = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MaxMpxCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxMpxCount")
	}
	c.MaxMpxCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter VcNumber
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for VcNumber")
	}
	c.VcNumber = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SessionKey
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for SessionKey")
	}
	c.SessionKey = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter OEMPasswordLen
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for OEMPasswordLen")
	}
	c.OEMPasswordLen = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter UnicodePasswordLen
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for UnicodePasswordLen")
	}
	c.UnicodePasswordLen = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Capabilities
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Capabilities")
	}
	c.Capabilities = capabilities.Capabilities(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data OEMPassword
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for OEMPassword")
	}
	c.OEMPassword = rawDataContent[offset : offset+int(c.OEMPasswordLen)]
	offset += int(c.OEMPasswordLen)

	// Unmarshalling data UnicodePassword
	if len(rawDataContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for UnicodePassword")
	}
	c.UnicodePassword = rawDataContent[offset : offset+int(c.UnicodePasswordLen)]
	offset += int(c.UnicodePasswordLen)

	// Unmarshalling data Pad
	padLen := int(c.UnicodePasswordLen)
	if padLen%2 == 1 {
		padLen++
	}
	if len(rawDataContent) < offset+padLen {
		return offset, fmt.Errorf("rawParametersContent too short for Pad")
	}
	c.Pad = rawDataContent[offset : offset+padLen]
	offset += padLen

	// Unmarshalling data AccountName
	bytesRead, err = c.AccountName.Unmarshal(rawDataContent[offset:])
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

	return offset, nil
}
