package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// CancelRequestStructureSize is the fixed StructureSize value for an SMB2 CANCEL Request.
const CancelRequestStructureSize = 4

// CancelRequest is the SMB2 CANCEL Request body, sent by the client to cancel a
// previously sent request. The MessageId of the request to be canceled is set
// in the SMB2 header. CANCEL has no response.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/91913fc6-4ec9-4a83-961b-370070067e63
type CancelRequest struct {
	command_interface.Command

	// Reserved (2 bytes): The client MUST set this to 0, and the server MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewCancelRequest creates a new SMB2 CANCEL Request.
func NewCancelRequest() *CancelRequest {
	c := &CancelRequest{}
	c.SetCommandCode(codes.SMB2_CANCEL)
	c.StructureSize = CancelRequestStructureSize
	return c
}

// Marshal serializes the CANCEL Request body.
func (c *CancelRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], CancelRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the CANCEL Request body.
func (c *CancelRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 CANCEL Request: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
