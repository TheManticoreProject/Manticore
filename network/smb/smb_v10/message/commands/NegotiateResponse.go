package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/utils"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// NegotiateResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a4229e1a-8a4e-489a-a2eb-11b7f360e60c
type NegotiateResponse struct {
	command_interface.Command

	// Parameters

	// The index of the dialect selected by the server from the list presented in the request. Dialect entries are numbered
	// starting with 0x0000, so a DialectIndex value of 0x0000 indicates the first entry in the list. If the server does not
	// support any of the listed dialects, it MUST return a DialectIndex of 0xFFFF.
	DialectIndex types.USHORT

	// An 8-bit field indicating the security modes supported or required by the server, as follows:
	SecurityMode securitymode.SecurityMode

	// The maximum number of outstanding SMB operations that the server supports. This value includes existing OpLocks,
	// the NT_TRANSACT_NOTIFY_CHANGE subcommand, and any other commands that are pending on the server. If the negotiated
	// MaxMpxCount is 0x0001, then OpLock support MUST be disabled for this session. The MaxMpxCount MUST be greater than 0x0000.
	// This parameter has no specific relationship to the SMB_COM_READ_MPX and SMB_COM_WRITE_MPX commands.
	MaxMpxCount types.USHORT

	// The maximum number of virtual circuits that can be established between the client and the server as part of the same SMB session.
	MaxNumberVcs types.USHORT

	// The maximum size, in bytes, of the largest SMB message that the server can receive. This is the size of the largest SMB message
	// that the client can send to the server. SMB message size includes the size of the SMB header, parameter, and data blocks. This size
	// does not include any transport-layer framing or other transport-layer data. The server SHOULD provide a MaxBufferSize of 4356 bytes,
	// and MUST be a multiple of 4 bytes. If CAP_RAW_MODE is negotiated, the SMB_COM_WRITE_RAW command can bypass the MaxBufferSize limit.
	// Otherwise, SMB messages sent to the server MUST have a total size less than or equal to the MaxBufferSize value. This includes AndX
	// chained messages.
	MaxBufferSize types.ULONG

	// This value specifies the maximum message size when the client sends an SMB_COM_WRITE_RAW Request (section 2.2.4.25.1), and the maximum
	// message size that the server MUST NOT exceed when sending an SMB_COM_READ_RAW Response (section 2.2.4.22.2). This value is significant only
	// if CAP_RAW_MODE is negotiated.
	MaxRawSize types.ULONG

	// The server SHOULD set the value to a token generated for the connection, as specified in SessionKey Generation (section 2.2.1.6.6).
	SessionKey types.ULONG

	// A 32-bit field providing a set of server capability indicators. This bit field is used to indicate to the client which features are supported
	// by the server. Any value not listed in the following table is unused. The server MUST set the unused bits to 0 in a response, and the client MUST
	// ignore these bits.
	Capabilities capabilities.Capabilities

	// The number of 100-nanosecond intervals that have elapsed since January 1, 1601, in Coordinated Universal Time (UTC) format.
	SystemTime types.FILETIME

	// A signed 16-bit signed integer that represents the server's time zone, in minutes, from UTC. The time zone of the server MUST be expressed
	// in minutes, plus or minus, from UTC.
	ServerTimeZone types.SHORT

	// This field MUST be 0x00 or 0x08. The length of the random challenge used in challenge/response authentication. If the server does not support
	// challenge/response authentication, this field MUST be 0x00. This field is often referred to in older documentation as EncryptionKeyLength.
	ChallengeLength types.UCHAR

	// Data

	// An array of unsigned bytes that MUST be ChallengeLength bytes long and MUST represent the server challenge.
	// This array MUST NOT be null-terminated. This field is often referred to in older documentation as EncryptionKey.
	Challenge []types.UCHAR

	// The null-terminated name of the NT domain or workgroup to which the server belongs.
	DomainName []types.UCHAR

	// The null-terminated name of the server.
	ServerName []types.UCHAR
}

// NewNegotiateResponse creates a new NegotiateResponse structure
//
// Returns:
// - A pointer to the new NegotiateResponse structure
func NewNegotiateResponse() *NegotiateResponse {
	c := &NegotiateResponse{
		// Parameters
		DialectIndex:    types.USHORT(0),
		SecurityMode:    securitymode.SecurityMode(0),
		MaxMpxCount:     types.USHORT(0),
		MaxNumberVcs:    types.USHORT(0),
		MaxBufferSize:   types.ULONG(0),
		MaxRawSize:      types.ULONG(0),
		SessionKey:      types.ULONG(0),
		Capabilities:    capabilities.Capabilities(0),
		SystemTime:      types.FILETIME{},
		ServerTimeZone:  types.SHORT(0),
		ChallengeLength: types.UCHAR(0),

		// Data
		Challenge:  []types.UCHAR{},
		DomainName: []types.UCHAR{},
		ServerName: []types.UCHAR{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_NEGOTIATE)

	return c
}

// GetSelectedDialect returns the selected dialect from the NegotiateResponse structure
//
// Parameters:
// - dialects: The dialects to search in
//
// Returns:
// - The selected dialect
// - An error if the selected dialect is not found
func (c *NegotiateResponse) GetSelectedDialect(dialects dialects.Dialects) (string, error) {
	if c.DialectIndex >= types.USHORT(len(dialects.Dialects)) {
		return "", fmt.Errorf("dialect index out of bounds")
	}

	return dialects.Dialects[c.DialectIndex], nil
}

// Marshal marshals the NegotiateResponse structure into a byte array
//
// Returns:
// - A byte array representing the NegotiateResponse structure
// - An error if the marshaling fails
func (c *NegotiateResponse) Marshal() ([]byte, error) {
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
	// Marshalling data Challenge
	c.ChallengeLength = types.UCHAR(len(c.Challenge))
	rawDataContent = append(rawDataContent, c.Challenge...)
	// Marshalling data DomainName
	rawDataContent = append(rawDataContent, c.DomainName...)

	// Then marshal the parameters
	rawParametersContent := []byte{}
	// Marshalling parameter DialectIndex
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.DialectIndex))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter SecurityMode
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.SecurityMode))
	// Marshalling parameter MaxMpxCount
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.MaxMpxCount))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter MaxNumberVcs
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.MaxNumberVcs))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter MaxBufferSize
	buf2 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf2, uint32(c.MaxBufferSize))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter MaxRawSize
	buf2 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf2, uint32(c.MaxRawSize))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter SessionKey
	buf2 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf2, uint32(c.SessionKey))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter Capabilities
	buf2 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf2, uint32(c.Capabilities))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter SystemTime
	bytesStream, err := c.SystemTime.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)
	// Marshalling parameter ServerTimeZone
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(c.ServerTimeZone))
	rawParametersContent = append(rawParametersContent, buf2...)
	// Marshalling parameter ChallengeLength
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.ChallengeLength))

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
func (c *NegotiateResponse) Unmarshal(marshalledData []byte) (int, error) {
	offset := 0

	// Create the Parameters structure if it is nil
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Create the Data structure if it is nil
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(marshalledData)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(marshalledData[bytesRead:])
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
	// Unmarshalling parameter DialectIndex
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for DialectIndex")
	}
	c.DialectIndex = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2
	// Unmarshalling parameter SecurityMode
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for SecurityMode")
	}
	c.SecurityMode = securitymode.SecurityMode(rawParametersContent[offset])
	offset++
	// Unmarshalling parameter MaxMpxCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxMpxCount")
	}
	c.MaxMpxCount = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2
	// Unmarshalling parameter MaxNumberVcs
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxNumberVcs")
	}
	c.MaxNumberVcs = types.USHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2
	// Unmarshalling parameter MaxBufferSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxBufferSize")
	}
	c.MaxBufferSize = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4
	// Unmarshalling parameter MaxRawSize
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxRawSize")
	}
	c.MaxRawSize = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4
	// Unmarshalling parameter SessionKey
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for SessionKey")
	}
	c.SessionKey = types.ULONG(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4
	// Unmarshalling parameter Capabilities
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Capabilities")
	}
	c.Capabilities = capabilities.Capabilities(binary.LittleEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4
	// Unmarshalling parameter SystemTime
	if len(rawParametersContent) < offset+8 {
		return offset, fmt.Errorf("rawParametersContent too short for SystemTime")
	}
	bytesRead, err = c.SystemTime.Unmarshal(rawParametersContent[offset : offset+8])
	if err != nil {
		return offset, err
	}
	offset += bytesRead
	// Unmarshalling parameter ServerTimeZone
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for ServerTimeZone")
	}
	c.ServerTimeZone = types.SHORT(binary.LittleEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2
	// Unmarshalling parameter ChallengeLength
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("rawParametersContent too short for ChallengeLength")
	}
	c.ChallengeLength = types.UCHAR(rawParametersContent[offset])
	offset++

	// Then unmarshal the data
	offset = 0
	// Unmarshalling data Challenge
	if len(rawDataContent) < offset+int(c.ChallengeLength) {
		return offset, fmt.Errorf("rawDataContent too short for Challenge")
	}
	c.Challenge = rawDataContent[offset : offset+int(c.ChallengeLength)]
	offset += int(c.ChallengeLength)
	rawDataContent = rawDataContent[offset:]

	// Unmarshalling data DomainName
	domainName, offset := utils.GetNullTerminatedUnicodeString(rawDataContent)
	c.DomainName = []types.UCHAR(domainName)
	rawDataContent = rawDataContent[offset:]

	// Unmarshalling data ServerName
	serverName, offset := utils.GetNullTerminatedUnicodeString(rawDataContent)
	c.ServerName = []types.UCHAR(serverName)
	// rawDataContent = rawDataContent[offset:]

	return offset, nil
}
