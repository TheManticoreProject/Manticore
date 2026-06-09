package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// FlushResponseStructureSize is the fixed StructureSize value for an SMB2 FLUSH Response.
const FlushResponseStructureSize = 4

// FlushResponse is the SMB2 FLUSH Response body, sent by the server to confirm
// that an SMB2 FLUSH Request was processed.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/42f78e6a-e25f-48f5-8f08-b4f1bb4c4fa4
type FlushResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0, and the client MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewFlushResponse creates a new SMB2 FLUSH Response.
func NewFlushResponse() *FlushResponse {
	c := &FlushResponse{}
	c.SetCommandCode(codes.SMB2_FLUSH)
	c.StructureSize = FlushResponseStructureSize
	return c
}

// Marshal serializes the FLUSH Response body.
func (c *FlushResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], FlushResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the FLUSH Response body.
func (c *FlushResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 FLUSH Response: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
