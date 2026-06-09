package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// FlushRequestStructureSize is the fixed StructureSize value for an SMB2 FLUSH Request.
const FlushRequestStructureSize = 24

// FlushRequest is the SMB2 FLUSH Request body, sent by the client to request that
// a server flush all cached file information for the open identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/e494678b-b1fc-44a0-b86e-8195acf74ad7
type FlushRequest struct {
	command_interface.Command

	// Reserved1 (2 bytes): The client MUST set this to 0.
	Reserved1 types.USHORT

	// Reserved2 (4 bytes): The client MUST set this to 0.
	Reserved2 types.ULONG

	// FileId (16 bytes): The open to be flushed.
	FileId types.SMB2_FILEID
}

// NewFlushRequest creates a new SMB2 FLUSH Request.
func NewFlushRequest() *FlushRequest {
	c := &FlushRequest{}
	c.SetCommandCode(codes.SMB2_FLUSH)
	c.StructureSize = FlushRequestStructureSize
	return c
}

// Marshal serializes the FLUSH Request body.
func (c *FlushRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint16(buf[0:2], FlushRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved1)
	binary.LittleEndian.PutUint32(buf[4:8], c.Reserved2)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	buf = append(buf, fileId...)
	return buf, nil
}

// Unmarshal deserializes the FLUSH Request body.
func (c *FlushRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 24 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 FLUSH Request: have %d bytes, need 24", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved1 = binary.LittleEndian.Uint16(data[2:4])
	c.Reserved2 = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	return 24, nil
}
