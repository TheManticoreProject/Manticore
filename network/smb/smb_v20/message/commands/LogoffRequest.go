package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// LogoffRequestStructureSize is the fixed StructureSize value for an SMB2 LOGOFF Request.
const LogoffRequestStructureSize = 4

// LogoffRequest is the SMB2 LOGOFF Request body, used by the client to request
// termination of a particular session.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/abdc4ea9-52df-480e-9a36-34f104797d2c
type LogoffRequest struct {
	command_interface.Command

	// Reserved (2 bytes): The client MUST set this to 0, and the server MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewLogoffRequest creates a new SMB2 LOGOFF Request.
func NewLogoffRequest() *LogoffRequest {
	c := &LogoffRequest{}
	c.SetCommandCode(codes.SMB2_LOGOFF)
	c.StructureSize = LogoffRequestStructureSize
	return c
}

// Marshal serializes the LOGOFF Request body.
func (c *LogoffRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], LogoffRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the LOGOFF Request body.
func (c *LogoffRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 LOGOFF Request: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
