package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// TreeConnectResponseStructureSize is the fixed StructureSize value for an SMB2 TREE_CONNECT Response.
const TreeConnectResponseStructureSize = 16

// SMB2 share types (TREE_CONNECT Response ShareType field).
const (
	SMB2_SHARE_TYPE_DISK  = 0x01
	SMB2_SHARE_TYPE_PIPE  = 0x02
	SMB2_SHARE_TYPE_PRINT = 0x03
)

// SMB2_SHAREFLAG_ENCRYPT_DATA indicates the server requires messages on this
// share to be encrypted (SMB 3.x; TREE_CONNECT Response ShareFlags field).
const SMB2_SHAREFLAG_ENCRYPT_DATA = 0x00008000

// TreeConnectResponse is the SMB2 TREE_CONNECT Response body, sent by the server
// when a TREE_CONNECT request succeeds.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/dd34e26c-a75e-47fa-aab2-6efc27502e96
type TreeConnectResponse struct {
	command_interface.Command

	// ShareType (1 byte): The type of share being accessed (disk/pipe/print).
	ShareType types.UCHAR

	// Reserved (1 byte): The server MUST set this to 0.
	Reserved types.UCHAR

	// ShareFlags (4 bytes): Properties for this share (caching, DFS, encryption, ...).
	ShareFlags types.ULONG

	// Capabilities (4 bytes): Capabilities for this share.
	Capabilities types.ULONG

	// MaximalAccess (4 bytes): The maximal access for the user on this share.
	MaximalAccess types.ULONG
}

// NewTreeConnectResponse creates a new SMB2 TREE_CONNECT Response.
func NewTreeConnectResponse() *TreeConnectResponse {
	c := &TreeConnectResponse{}
	c.SetCommandCode(codes.SMB2_TREE_CONNECT)
	c.StructureSize = TreeConnectResponseStructureSize
	return c
}

// Marshal serializes the TREE_CONNECT Response body.
func (c *TreeConnectResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint16(buf[0:2], TreeConnectResponseStructureSize)
	buf[2] = byte(c.ShareType)
	buf[3] = byte(c.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], c.ShareFlags)
	binary.LittleEndian.PutUint32(buf[8:12], c.Capabilities)
	binary.LittleEndian.PutUint32(buf[12:16], c.MaximalAccess)
	return buf, nil
}

// Unmarshal deserializes the TREE_CONNECT Response body.
func (c *TreeConnectResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 TREE_CONNECT Response: have %d bytes, need 16", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.ShareType = data[2]
	c.Reserved = data[3]
	c.ShareFlags = binary.LittleEndian.Uint32(data[4:8])
	c.Capabilities = binary.LittleEndian.Uint32(data[8:12])
	c.MaximalAccess = binary.LittleEndian.Uint32(data[12:16])
	return 16, nil
}
