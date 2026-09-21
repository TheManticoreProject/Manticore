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

// SMB2 share flags (TREE_CONNECT Response ShareFlags field, MS-SMB2 2.2.10).
const (
	SMB2_SHAREFLAG_MANUAL_CACHING              uint32 = 0x00000000
	SMB2_SHAREFLAG_AUTO_CACHING                uint32 = 0x00000010
	SMB2_SHAREFLAG_VDO_CACHING                 uint32 = 0x00000020
	SMB2_SHAREFLAG_NO_CACHING                  uint32 = 0x00000030
	SMB2_SHAREFLAG_DFS                         uint32 = 0x00000001
	SMB2_SHAREFLAG_DFS_ROOT                    uint32 = 0x00000002
	SMB2_SHAREFLAG_RESTRICT_EXCLUSIVE_OPENS    uint32 = 0x00000100
	SMB2_SHAREFLAG_FORCE_SHARED_DELETE         uint32 = 0x00000200
	SMB2_SHAREFLAG_ALLOW_NAMESPACE_CACHING     uint32 = 0x00000400
	SMB2_SHAREFLAG_ACCESS_BASED_DIRECTORY_ENUM uint32 = 0x00000800
	SMB2_SHAREFLAG_FORCE_LEVELII_OPLOCK        uint32 = 0x00001000
	SMB2_SHAREFLAG_ENABLE_HASH_V1              uint32 = 0x00002000
	SMB2_SHAREFLAG_ENABLE_HASH_V2              uint32 = 0x00004000
	SMB2_SHAREFLAG_ENCRYPT_DATA                uint32 = 0x00008000
	SMB2_SHAREFLAG_IDENTITY_REMOTING           uint32 = 0x00040000
	SMB2_SHAREFLAG_COMPRESS_DATA               uint32 = 0x00100000
	SMB2_SHAREFLAG_ISOLATE_TRANSPORT           uint32 = 0x00200000
)

// SMB2 share capability flags (TREE_CONNECT Response Capabilities field, MS-SMB2 2.2.10).
const (
	SMB2_SHARE_CAP_DFS                     uint32 = 0x00000008
	SMB2_SHARE_CAP_CONTINUOUS_AVAILABILITY uint32 = 0x00000010
	SMB2_SHARE_CAP_SCALEOUT               uint32 = 0x00000020
	SMB2_SHARE_CAP_CLUSTER                uint32 = 0x00000040
	SMB2_SHARE_CAP_ASYMMETRIC             uint32 = 0x00000080
	SMB2_SHARE_CAP_REDIRECT_TO_OWNER      uint32 = 0x00000100
)

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
