package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// WriteResponseStructureSize is the fixed StructureSize value the server MUST
	// set for an SMB2 WRITE Response (16-byte fixed part + 1).
	WriteResponseStructureSize = 17

	// writeResponseFixedSize is the size, in bytes, of the fixed portion of the body.
	writeResponseFixedSize = 16
)

// WriteResponse is the SMB2 WRITE Response body, sent by the server to report how
// many bytes of an SMB2 WRITE Request were written.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/7b80a339-f4d3-4575-8ce2-70a06f24f133
type WriteResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0.
	Reserved types.USHORT

	// Count (4 bytes): The number of bytes written.
	Count types.ULONG

	// Remaining (4 bytes): Reserved; the server MUST set this to 0.
	Remaining types.ULONG

	// WriteChannelInfoOffset (2 bytes): Reserved; the server MUST set this to 0.
	WriteChannelInfoOffset types.USHORT

	// WriteChannelInfoLength (2 bytes): Reserved; the server MUST set this to 0.
	WriteChannelInfoLength types.USHORT
}

// NewWriteResponse creates a new SMB2 WRITE Response.
func NewWriteResponse() *WriteResponse {
	c := &WriteResponse{}
	c.SetCommandCode(codes.SMB2_WRITE)
	c.StructureSize = WriteResponseStructureSize
	return c
}

// Marshal serializes the WRITE Response body. A single trailing byte follows the
// 16-byte fixed part to honor the StructureSize off-by-one convention.
func (c *WriteResponse) Marshal() ([]byte, error) {
	buf := make([]byte, writeResponseFixedSize+1)
	binary.LittleEndian.PutUint16(buf[0:2], WriteResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], c.Count)
	binary.LittleEndian.PutUint32(buf[8:12], c.Remaining)
	binary.LittleEndian.PutUint16(buf[12:14], c.WriteChannelInfoOffset)
	binary.LittleEndian.PutUint16(buf[14:16], c.WriteChannelInfoLength)
	return buf, nil
}

// Unmarshal deserializes the WRITE Response body.
func (c *WriteResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < writeResponseFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 WRITE Response: have %d bytes, need at least %d", len(data), writeResponseFixedSize)
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	c.Count = binary.LittleEndian.Uint32(data[4:8])
	c.Remaining = binary.LittleEndian.Uint32(data[8:12])
	c.WriteChannelInfoOffset = binary.LittleEndian.Uint16(data[12:14])
	c.WriteChannelInfoLength = binary.LittleEndian.Uint16(data[14:16])
	return writeResponseFixedSize, nil
}
