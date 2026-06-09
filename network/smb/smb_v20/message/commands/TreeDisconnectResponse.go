package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// TreeDisconnectResponseStructureSize is the fixed StructureSize value for an SMB2 TREE_DISCONNECT Response.
const TreeDisconnectResponseStructureSize = 4

// TreeDisconnectResponse is the SMB2 TREE_DISCONNECT Response body, sent by the
// server in response to an SMB2 TREE_DISCONNECT Request.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/dd34e26c-a75e-47fa-aab2-6efc27502e96
type TreeDisconnectResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0, and the client MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewTreeDisconnectResponse creates a new SMB2 TREE_DISCONNECT Response.
func NewTreeDisconnectResponse() *TreeDisconnectResponse {
	c := &TreeDisconnectResponse{}
	c.SetCommandCode(codes.SMB2_TREE_DISCONNECT)
	c.StructureSize = TreeDisconnectResponseStructureSize
	return c
}

// Marshal serializes the TREE_DISCONNECT Response body.
func (c *TreeDisconnectResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], TreeDisconnectResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the TREE_DISCONNECT Response body.
func (c *TreeDisconnectResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 TREE_DISCONNECT Response: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
