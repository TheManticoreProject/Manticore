package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// LogoffResponseStructureSize is the fixed StructureSize value for an SMB2 LOGOFF Response.
const LogoffResponseStructureSize = 4

// LogoffResponse is the SMB2 LOGOFF Response body, sent by the server in
// response to an SMB2 LOGOFF Request.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/7539feb4-6fbb-4996-81ac-06863bb1a89e
type LogoffResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0, and the client MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewLogoffResponse creates a new SMB2 LOGOFF Response.
func NewLogoffResponse() *LogoffResponse {
	c := &LogoffResponse{}
	c.SetCommandCode(codes.SMB2_LOGOFF)
	c.StructureSize = LogoffResponseStructureSize
	return c
}

// Marshal serializes the LOGOFF Response body.
func (c *LogoffResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], LogoffResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the LOGOFF Response body.
func (c *LogoffResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 LOGOFF Response: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
