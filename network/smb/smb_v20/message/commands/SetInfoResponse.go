package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// SetInfoResponseStructureSize is the fixed StructureSize value for an SMB2
// SET_INFO Response. The response is exactly this 2-byte field.
const SetInfoResponseStructureSize = 2

// SetInfoResponse is the SMB2 SET_INFO Response body, sent by the server to
// confirm that an SMB2 SET_INFO Request was processed. It carries no fields
// beyond StructureSize.
//
// Source: [MS-SMB2] section 2.2.40 SMB2 SET_INFO Response.
type SetInfoResponse struct {
	command_interface.Command
}

// NewSetInfoResponse creates a new SMB2 SET_INFO Response.
func NewSetInfoResponse() *SetInfoResponse {
	c := &SetInfoResponse{}
	c.SetCommandCode(codes.SMB2_SET_INFO)
	c.StructureSize = SetInfoResponseStructureSize
	return c
}

// Marshal serializes the SET_INFO Response body.
func (c *SetInfoResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf[0:2], SetInfoResponseStructureSize)
	return buf, nil
}

// Unmarshal deserializes the SET_INFO Response body.
func (c *SetInfoResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 SET_INFO Response: have %d bytes, need 2", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	return 2, nil
}
