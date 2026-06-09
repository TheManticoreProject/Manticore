package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// TreeDisconnectRequestStructureSize is the fixed StructureSize value for an SMB2 TREE_DISCONNECT Request.
const TreeDisconnectRequestStructureSize = 4

// TreeDisconnectRequest is the SMB2 TREE_DISCONNECT Request body, used by the
// client to request that the tree connect identified by the TreeId in the SMB2
// header be disconnected.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/8efb5d72-d3da-499b-bf57-846b9da99f3b
type TreeDisconnectRequest struct {
	command_interface.Command

	// Reserved (2 bytes): The client MUST set this to 0, and the server MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewTreeDisconnectRequest creates a new SMB2 TREE_DISCONNECT Request.
func NewTreeDisconnectRequest() *TreeDisconnectRequest {
	c := &TreeDisconnectRequest{}
	c.SetCommandCode(codes.SMB2_TREE_DISCONNECT)
	c.StructureSize = TreeDisconnectRequestStructureSize
	return c
}

// Marshal serializes the TREE_DISCONNECT Request body.
func (c *TreeDisconnectRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], TreeDisconnectRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the TREE_DISCONNECT Request body.
func (c *TreeDisconnectRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 TREE_DISCONNECT Request: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
