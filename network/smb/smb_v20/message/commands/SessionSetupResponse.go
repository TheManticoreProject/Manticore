package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// SessionSetupResponseStructureSize is the fixed StructureSize value the server
	// MUST set for an SMB2 SESSION_SETUP Response (8-byte fixed part + 1).
	SessionSetupResponseStructureSize = 9

	// sessionSetupResponseFixedSize is the size, in bytes, of the fixed portion
	// of the body that precedes the variable security Buffer.
	sessionSetupResponseFixedSize = 8

	// SMB2_SESSION_FLAG_IS_GUEST indicates the client was authenticated as a guest.
	SMB2_SESSION_FLAG_IS_GUEST = 0x0001
	// SMB2_SESSION_FLAG_IS_NULL indicates the client was authenticated anonymously.
	SMB2_SESSION_FLAG_IS_NULL = 0x0002
	// SMB2_SESSION_FLAG_ENCRYPT_DATA indicates the server requires encryption (SMB 3.x only).
	SMB2_SESSION_FLAG_ENCRYPT_DATA = 0x0004
)

// SessionSetupResponse is the SMB2 SESSION_SETUP Response body, sent by the
// server in response to an SMB2 SESSION_SETUP Request.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/0324190f-a31b-4666-9fa9-5c624273a694
type SessionSetupResponse struct {
	command_interface.Command

	// SessionFlags (2 bytes): Additional information about the session (guest/null/encrypt).
	SessionFlags types.USHORT

	// SecurityBuffer (variable): The GSS authentication token. SecurityBufferOffset
	// and SecurityBufferLength are computed from this on marshal.
	SecurityBuffer []byte
}

// NewSessionSetupResponse creates a new SMB2 SESSION_SETUP Response.
func NewSessionSetupResponse() *SessionSetupResponse {
	c := &SessionSetupResponse{SecurityBuffer: []byte{}}
	c.SetCommandCode(codes.SMB2_SESSION_SETUP)
	c.StructureSize = SessionSetupResponseStructureSize
	return c
}

// Marshal serializes the SESSION_SETUP Response body.
func (c *SessionSetupResponse) Marshal() ([]byte, error) {
	buf := make([]byte, sessionSetupResponseFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], SessionSetupResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.SessionFlags)

	// SecurityBufferOffset is measured from the start of the SMB2 header.
	if len(c.SecurityBuffer) > 0 {
		binary.LittleEndian.PutUint16(buf[4:6], uint16(header.SMB2_HEADER_SIZE+sessionSetupResponseFixedSize))
	}
	binary.LittleEndian.PutUint16(buf[6:8], uint16(len(c.SecurityBuffer)))

	buf = append(buf, c.SecurityBuffer...)
	return buf, nil
}

// Unmarshal deserializes the SESSION_SETUP Response body.
func (c *SessionSetupResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < sessionSetupResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 SESSION_SETUP Response: have %d bytes, need at least %d", len(data), sessionSetupResponseFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.SessionFlags = binary.LittleEndian.Uint16(data[2:4])
	securityBufferOffset := int(binary.LittleEndian.Uint16(data[4:6]))
	securityBufferLength := int(binary.LittleEndian.Uint16(data[6:8]))

	consumed := sessionSetupResponseFixedSize
	if securityBufferLength > 0 {
		start := securityBufferOffset - header.SMB2_HEADER_SIZE
		if start < sessionSetupResponseFixedSize || start+securityBufferLength > len(data) {
			return 0, fmt.Errorf("SMB2 SESSION_SETUP Response security buffer out of bounds: offset %d length %d", securityBufferOffset, securityBufferLength)
		}
		c.SecurityBuffer = make([]byte, securityBufferLength)
		copy(c.SecurityBuffer, data[start:start+securityBufferLength])
		consumed = start + securityBufferLength
	} else {
		c.SecurityBuffer = []byte{}
	}

	return consumed, nil
}
