package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// CloseResponseStructureSize is the fixed StructureSize value for an SMB2 CLOSE Response.
const CloseResponseStructureSize = 60

// CloseResponse is the SMB2 CLOSE Response body, sent by the server to confirm
// that an SMB2 CLOSE Request was processed. The timestamp, size, and attribute
// fields are populated only if SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB was set in the
// request; otherwise they are zero.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/c0c15c57-3f3e-452b-b51c-9cc650a13f7b
type CloseResponse struct {
	command_interface.Command

	// Flags (2 bytes): Indicates how to process the operation.
	Flags types.USHORT

	// Reserved (4 bytes): The server MUST set this to 0.
	Reserved types.ULONG

	// CreationTime (8 bytes): When the file was created, as a 64-bit FILETIME.
	CreationTime types.UINT64

	// LastAccessTime (8 bytes): When the file was last accessed, as a 64-bit FILETIME.
	LastAccessTime types.UINT64

	// LastWriteTime (8 bytes): When data was last written, as a 64-bit FILETIME.
	LastWriteTime types.UINT64

	// ChangeTime (8 bytes): When the file was last changed, as a 64-bit FILETIME.
	ChangeTime types.UINT64

	// AllocationSize (8 bytes): The allocation size of the file, in bytes.
	AllocationSize types.UINT64

	// EndOfFile (8 bytes): The end-of-file offset, in bytes.
	EndOfFile types.UINT64

	// FileAttributes (4 bytes): The attributes of the file.
	FileAttributes types.ULONG
}

// NewCloseResponse creates a new SMB2 CLOSE Response.
func NewCloseResponse() *CloseResponse {
	c := &CloseResponse{}
	c.SetCommandCode(codes.SMB2_CLOSE)
	c.StructureSize = CloseResponseStructureSize
	return c
}

// Marshal serializes the CLOSE Response body.
func (c *CloseResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 60)
	binary.LittleEndian.PutUint16(buf[0:2], CloseResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], c.Reserved)
	binary.LittleEndian.PutUint64(buf[8:16], c.CreationTime)
	binary.LittleEndian.PutUint64(buf[16:24], c.LastAccessTime)
	binary.LittleEndian.PutUint64(buf[24:32], c.LastWriteTime)
	binary.LittleEndian.PutUint64(buf[32:40], c.ChangeTime)
	binary.LittleEndian.PutUint64(buf[40:48], c.AllocationSize)
	binary.LittleEndian.PutUint64(buf[48:56], c.EndOfFile)
	binary.LittleEndian.PutUint32(buf[56:60], c.FileAttributes)
	return buf, nil
}

// Unmarshal deserializes the CLOSE Response body.
func (c *CloseResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 60 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CLOSE Response: have %d bytes, need 60", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Flags = binary.LittleEndian.Uint16(data[2:4])
	c.Reserved = binary.LittleEndian.Uint32(data[4:8])
	c.CreationTime = binary.LittleEndian.Uint64(data[8:16])
	c.LastAccessTime = binary.LittleEndian.Uint64(data[16:24])
	c.LastWriteTime = binary.LittleEndian.Uint64(data[24:32])
	c.ChangeTime = binary.LittleEndian.Uint64(data[32:40])
	c.AllocationSize = binary.LittleEndian.Uint64(data[40:48])
	c.EndOfFile = binary.LittleEndian.Uint64(data[48:56])
	c.FileAttributes = binary.LittleEndian.Uint32(data[56:60])
	return 60, nil
}
