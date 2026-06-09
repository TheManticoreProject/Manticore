package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// SessionSetupRequestStructureSize is the fixed StructureSize value the client
	// MUST set for an SMB2 SESSION_SETUP Request (24-byte fixed part + 1).
	SessionSetupRequestStructureSize = 25

	// sessionSetupRequestFixedSize is the size, in bytes, of the fixed portion of
	// the body that precedes the variable security Buffer.
	sessionSetupRequestFixedSize = 24

	// SMB2_SESSION_FLAG_BINDING indicates the request binds an existing session to a new connection.
	SMB2_SESSION_FLAG_BINDING = 0x01
)

// SessionSetupRequest is the SMB2 SESSION_SETUP Request body, sent by the client
// to request a new authenticated session.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/5a3c2c28-d6b0-48ed-b917-a86b2ca4575f
type SessionSetupRequest struct {
	command_interface.Command

	// Flags (1 byte): Session binding flags (SMB 3.x only; else 0).
	Flags types.UCHAR

	// SecurityMode (1 byte): Whether SMB signing is enabled or required at the client.
	SecurityMode securitymode.SecurityMode

	// Capabilities (4 bytes): Client capabilities.
	Capabilities capabilities.Capabilities

	// Channel (4 bytes): MUST NOT be used; the client sets this to 0.
	Channel types.ULONG

	// PreviousSessionId (8 bytes): A previously established session identifier, or 0.
	PreviousSessionId types.UINT64

	// SecurityBuffer (variable): The GSS authentication token. SecurityBufferOffset
	// and SecurityBufferLength are computed from this on marshal.
	SecurityBuffer []byte
}

// NewSessionSetupRequest creates a new SMB2 SESSION_SETUP Request.
func NewSessionSetupRequest() *SessionSetupRequest {
	c := &SessionSetupRequest{SecurityBuffer: []byte{}}
	c.SetCommandCode(codes.SMB2_SESSION_SETUP)
	c.StructureSize = SessionSetupRequestStructureSize
	return c
}

// Marshal serializes the SESSION_SETUP Request body.
func (c *SessionSetupRequest) Marshal() ([]byte, error) {
	buf := make([]byte, sessionSetupRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], SessionSetupRequestStructureSize)
	buf[2] = byte(c.Flags)
	buf[3] = byte(c.SecurityMode)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(c.Capabilities))
	binary.LittleEndian.PutUint32(buf[8:12], c.Channel)

	// SecurityBufferOffset is measured from the start of the SMB2 header.
	if len(c.SecurityBuffer) > 0 {
		binary.LittleEndian.PutUint16(buf[12:14], uint16(header.SMB2_HEADER_SIZE+sessionSetupRequestFixedSize))
	}
	binary.LittleEndian.PutUint16(buf[14:16], uint16(len(c.SecurityBuffer)))
	binary.LittleEndian.PutUint64(buf[16:24], c.PreviousSessionId)

	buf = append(buf, c.SecurityBuffer...)
	return buf, nil
}

// Unmarshal deserializes the SESSION_SETUP Request body.
func (c *SessionSetupRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < sessionSetupRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 SESSION_SETUP Request: have %d bytes, need at least %d", len(data), sessionSetupRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Flags = data[2]
	c.SecurityMode = securitymode.SecurityMode(data[3])
	c.Capabilities = capabilities.Capabilities(binary.LittleEndian.Uint32(data[4:8]))
	c.Channel = binary.LittleEndian.Uint32(data[8:12])
	securityBufferOffset := int(binary.LittleEndian.Uint16(data[12:14]))
	securityBufferLength := int(binary.LittleEndian.Uint16(data[14:16]))
	c.PreviousSessionId = binary.LittleEndian.Uint64(data[16:24])

	consumed := sessionSetupRequestFixedSize
	if securityBufferLength > 0 {
		start := securityBufferOffset - header.SMB2_HEADER_SIZE
		if start < sessionSetupRequestFixedSize || start+securityBufferLength > len(data) {
			return 0, fmt.Errorf("SMB2 SESSION_SETUP Request security buffer out of bounds: offset %d length %d", securityBufferOffset, securityBufferLength)
		}
		c.SecurityBuffer = make([]byte, securityBufferLength)
		copy(c.SecurityBuffer, data[start:start+securityBufferLength])
		consumed = start + securityBufferLength
	} else {
		c.SecurityBuffer = []byte{}
	}

	return consumed, nil
}
