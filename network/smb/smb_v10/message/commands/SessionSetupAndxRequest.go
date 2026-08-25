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
	"github.com/TheManticoreProject/Manticore/utils"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// SessionSetupAndxRequest
// Source for CIFS base: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/81e15dee-8fb6-4102-8644-7eaa7ded63f7
// SMBv1.0 extension: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/a00d0361-3544-4845-96ab-309b4bb7705d
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

	// When extended security is not being used, the following fields (OEMPasswordLen,
	// UnicodePasswordLen) are used to authenticate the user:

	// If SMB_FLAGS2_UNICODE is set (1), the value of OEMPasswordLen MUST be 0x0000 and
	// the password MUST be encoded using UTF-16LE Unicode. Padding MUST NOT be added
	// to align this plaintext Unicode string to a word boundary.
	OEMPasswordLen types.USHORT

	// If SMB_FLAGS2_UNICODE is clear (0), the value of UnicodePasswordLen MUST be
	// 0x0000, and the password MUST be encoded using the 8-bit OEM character set
	// (extended ASCII).
	UnicodePasswordLen types.USHORT

	// When extended security is being used, the following field (SecurityBlobLength)
	// is used to authenticate the user:

	// SecurityBlobLength (2 bytes): This value MUST specify the length in bytes
	// of the variable-length SecurityBlob field that is contained within the request.
	SecurityBlobLength types.USHORT

	// Reserved (4 bytes): Reserved. This field MUST be 0x00000000. The server MUST
	// ignore the contents of this field.
	Reserved types.ULONG

	// Capabilities (4 bytes): A 32-bit field providing a set of client capability
	// indicators. The client uses this field to report its own set of capabilities to
	// the server. The client capabilities are a subset of the server capabilities.
	Capabilities capabilities.Capabilities

	// Data

	// When extended security is not being used, the following fields (OEMPassword,
	// UnicodePassword, Pad, AccountName, PrimaryDomain) are used to authenticate the
	// user:

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

	// When extended security is being used, the following field (SecurityBlob)
	// is used to authenticate the user:

	// SecurityBlob (variable): This field MUST be the authentication token sent to the
	// server, as specified in section 3.2.4.2.4 and in [RFC2743]. This field MUST be
	// aligned to start on a 2-byte boundary from the start of the SMB header.
	SecurityBlob []types.UCHAR

	// NativeOS (variable): A string representing the native operating system of the
	// CIFS client. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header
	// of the request, this string MUST be a null-terminated array of 16-bit Unicode
	// characters. Otherwise, this string MUST be a null-terminated array of OEM
	// characters. If this string consists of Unicode characters, this field MUST be
	// aligned to start on a 2-byte boundary from the start of the SMB header.
	NativeOS string

	// NativeLanMan (variable): A string that represents the native LAN manager type
	// of the client. If SMB_FLAGS2_UNICODE is set in the Flags2 field of the SMB header
	// of the request, this string MUST be a null-terminated array of 16-bit Unicode
	// characters. Otherwise, this string MUST be a null-terminated array of OEM
	// characters. If this string consists of Unicode characters, this field MUST be
	// aligned to start on a 2-byte boundary from the start of the SMB header.
	NativeLanMan string
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
		// When extended security is not being used, the following fields (OEMPassword,
		// UnicodePassword, Pad, AccountName, PrimaryDomain) are used to authenticate the
		// user:
		OEMPassword:     []types.UCHAR{},
		UnicodePassword: []types.UCHAR{},
		Pad:             []types.UCHAR{},
		AccountName:     types.SMB_STRING{},
		PrimaryDomain:   types.SMB_STRING{},
		// When extended security is being used, the following field (SecurityBlob)
		// is used to authenticate the user:
		SecurityBlob: []types.UCHAR{},
		// Other data fields
		NativeOS:     "",
		NativeLanMan: "",
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

	var err error

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

	if c.Capabilities.HasCapability(capabilities.CAP_EXTENDED_SECURITY) {
		// Marshalling data SecurityBlob
		rawDataContent = append(rawDataContent, c.SecurityBlob...)
		c.SecurityBlobLength = types.USHORT(len(c.SecurityBlob))

		// Marshalling data NativeOS
		if c.Capabilities.HasCapability(capabilities.CAP_UNICODE) {
			rawDataContent = append(rawDataContent, utf16.EncodeUTF16LE(c.NativeOS)...)
			rawDataContent = append(rawDataContent, []byte{0, 0}...)
		} else {
			rawDataContent = append(rawDataContent, c.NativeOS...)
			rawDataContent = append(rawDataContent, []byte{0}...)
		}

		// Marshalling data NativeLanMan
		if c.Capabilities.HasCapability(capabilities.CAP_UNICODE) {
			rawDataContent = append(rawDataContent, utf16.EncodeUTF16LE(c.NativeLanMan)...)
			rawDataContent = append(rawDataContent, []byte{0, 0}...)
		} else {
			rawDataContent = append(rawDataContent, c.NativeLanMan...)
			rawDataContent = append(rawDataContent, []byte{0}...)
		}
	} else {
		// Marshalling data OEMPassword
		rawDataContent = append(rawDataContent, c.OEMPassword...)
		c.OEMPasswordLen = types.USHORT(len(c.OEMPassword))

		// Marshalling data UnicodePassword
		rawDataContent = append(rawDataContent, c.UnicodePassword...)
		c.UnicodePasswordLen = types.USHORT(len(c.UnicodePassword))

		// Marshalling data Pad
		rawDataContent = append(rawDataContent, c.Pad...)

		// Marshalling data AccountName (raw null-terminated string, no buffer-format prefix)
		rawDataContent = append(rawDataContent, c.AccountName.Buffer...)
		rawDataContent = append(rawDataContent, 0x00)

		// Marshalling data PrimaryDomain (raw null-terminated string, no buffer-format prefix)
		rawDataContent = append(rawDataContent, c.PrimaryDomain.Buffer...)
		rawDataContent = append(rawDataContent, 0x00)

		// Marshalling data NativeOS (raw null-terminated string, no buffer-format prefix)
		if c.Capabilities.HasCapability(capabilities.CAP_UNICODE) {
			rawDataContent = append(rawDataContent, utf16.EncodeUTF16LE(c.NativeOS)...)
			rawDataContent = append(rawDataContent, 0x00, 0x00)
		} else {
			rawDataContent = append(rawDataContent, c.NativeOS...)
			rawDataContent = append(rawDataContent, 0x00)
		}

		// Marshalling data NativeLanMan (raw null-terminated string, no buffer-format prefix)
		if c.Capabilities.HasCapability(capabilities.CAP_UNICODE) {
			rawDataContent = append(rawDataContent, utf16.EncodeUTF16LE(c.NativeLanMan)...)
			rawDataContent = append(rawDataContent, 0x00, 0x00)
		} else {
			rawDataContent = append(rawDataContent, c.NativeLanMan...)
			rawDataContent = append(rawDataContent, 0x00)
		}
	}

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter MaxBufferSize
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.MaxBufferSize))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter MaxMpxCount
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.MaxMpxCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter VcNumber
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.VcNumber))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SessionKey
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(c.SessionKey))
	rawParametersContent = append(rawParametersContent, buf4...)

	if c.Capabilities.HasCapability(capabilities.CAP_EXTENDED_SECURITY) {
		// Marshalling parameter SecurityBlobLength
		buf2 = make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(c.SecurityBlobLength))
		rawParametersContent = append(rawParametersContent, buf2...)
	} else {
		// Marshalling parameter OEMPasswordLen
		buf2 = make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(c.OEMPasswordLen))
		rawParametersContent = append(rawParametersContent, buf2...)

		// Marshalling parameter UnicodePasswordLen
		buf2 = make([]byte, 2)
		binary.LittleEndian.PutUint16(buf2, uint16(c.UnicodePasswordLen))
		rawParametersContent = append(rawParametersContent, buf2...)
	}

	// Marshalling parameter Reserved
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(c.Reserved))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter Capabilities
	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(c.Capabilities))
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
func (c *SessionSetupAndxRequest) Unmarshal(rawData []byte) (int, error) {
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

	// Determine which variant of the request this is from the parameter
	// WordCount. The Capabilities field that carries CAP_EXTENDED_SECURITY is
	// located at the end of the parameter block and therefore cannot be used to
	// drive parsing of the fields that precede it. Per [MS-CIFS] 2.2.4.53.1 the
	// non-extended-security request has WordCount 0x0D, while [MS-SMB] 2.2.4.5
	// defines WordCount 0x0C for the extended-security request.
	isExtendedSecurity := c.GetParameters().WordCount == 0x0C

	// First unmarshal the parameters
	offset = 0
	if c.IsAndX() {
		offset += 4
	}

	// Unmarshalling parameter MaxBufferSize
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxBufferSize")
	}
	c.MaxBufferSize = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter MaxMpxCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxMpxCount")
	}
	c.MaxMpxCount = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter VcNumber
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for VcNumber")
	}
	c.VcNumber = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SessionKey
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for SessionKey")
	}
	c.SessionKey = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	if isExtendedSecurity {
		// Unmarshalling parameter SecurityBlobLength
		if len(rawParametersContent) < offset+2 {
			return offset, fmt.Errorf("rawParametersContent too short for SecurityBlobLength")
		}
		c.SecurityBlobLength = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2
	} else {
		// Unmarshalling parameter OEMPasswordLen
		if len(rawParametersContent) < offset+2 {
			return offset, fmt.Errorf("rawParametersContent too short for OEMPasswordLen")
		}
		c.OEMPasswordLen = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2

		// Unmarshalling parameter UnicodePasswordLen
		if len(rawParametersContent) < offset+2 {
			return offset, fmt.Errorf("rawParametersContent too short for UnicodePasswordLen")
		}
		c.UnicodePasswordLen = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
		offset += 2
	}

	// Unmarshalling parameter Reserved
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Reserved")
	}
	c.Reserved = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter Capabilities
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Capabilities")
	}
	c.Capabilities = capabilities.Capabilities(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Then unmarshal the data
	offset = 0

	if isExtendedSecurity {
		// Unmarshalling data SecurityBlob
		// The length of the SecurityBlob is carried by the SecurityBlobLength
		// parameter (parsed above), not by a length field inside the data block.
		if len(rawDataContent) < offset+int(c.SecurityBlobLength) {
			return offset, fmt.Errorf("rawDataContent too short for SecurityBlob")
		}
		c.SecurityBlob = rawDataContent[offset : offset+int(c.SecurityBlobLength)]
		offset += int(c.SecurityBlobLength)

		// Unmarshalling data NativeOS
		nativeOSdata, bytesRead := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		offset += bytesRead
		c.NativeOS = string(nativeOSdata)

		// Unmarshalling data NativeLanMan
		nativeLanMandata, bytesRead := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		offset += bytesRead
		c.NativeLanMan = string(nativeLanMandata)
	} else {
		// Unmarshalling data OEMPassword
		// OEMPasswordLen is client-controlled, so the bound has to cover the
		// whole slice rather than just its first byte.
		if len(rawDataContent) < offset+int(c.OEMPasswordLen) {
			return offset, fmt.Errorf("rawDataContent too short for OEMPassword")
		}
		c.OEMPassword = rawDataContent[offset : offset+int(c.OEMPasswordLen)]
		offset += int(c.OEMPasswordLen)

		// Unmarshalling data UnicodePassword
		// UnicodePasswordLen is client-controlled, so the bound has to cover the
		// whole slice rather than just its first byte.
		if len(rawDataContent) < offset+int(c.UnicodePasswordLen) {
			return offset, fmt.Errorf("rawDataContent too short for UnicodePassword")
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

		// Unmarshalling data AccountName (raw null-terminated string, no buffer-format prefix)
		accountNameData, accountNameBytesRead := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		c.AccountName.Buffer = accountNameData
		c.AccountName.Length = types.USHORT(len(accountNameData))
		offset += accountNameBytesRead

		// Unmarshalling data PrimaryDomain (raw null-terminated string, no buffer-format prefix)
		primaryDomainData, primaryDomainBytesRead := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		c.PrimaryDomain.Buffer = primaryDomainData
		c.PrimaryDomain.Length = types.USHORT(len(primaryDomainData))
		offset += primaryDomainBytesRead

		// Unmarshalling NativeOS (raw null-terminated string, no buffer-format prefix)
		nativeOSdata, nativeOSBytesRead := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		offset += nativeOSBytesRead
		c.NativeOS = string(nativeOSdata)

		// Unmarshalling NativeLanMan (raw null-terminated string, no buffer-format prefix)
		nativeLanMandata, nativeLanManBytesRead := utils.ReadUntilNullTerminator(rawDataContent[offset:])
		offset += nativeLanManBytesRead
		c.NativeLanMan = string(nativeLanMandata)
	}

	return offset, nil
}
