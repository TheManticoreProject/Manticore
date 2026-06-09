package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// LockResponseStructureSize is the fixed StructureSize value for an SMB2 LOCK Response.
const LockResponseStructureSize = 4

// LockResponse is the SMB2 LOCK Response body, sent by the server to confirm that
// an SMB2 LOCK Request was processed.
//
// Source: [MS-SMB2] section 2.2.27 SMB2 LOCK Response.
type LockResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0, and the client MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewLockResponse creates a new SMB2 LOCK Response.
func NewLockResponse() *LockResponse {
	c := &LockResponse{}
	c.SetCommandCode(codes.SMB2_LOCK)
	c.StructureSize = LockResponseStructureSize
	return c
}

// Marshal serializes the LOCK Response body.
func (c *LockResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], LockResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the LOCK Response body.
func (c *LockResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 LOCK Response: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
