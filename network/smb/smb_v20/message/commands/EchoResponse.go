package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// EchoResponseStructureSize is the fixed StructureSize value for an SMB2 ECHO Response.
const EchoResponseStructureSize = 4

// EchoResponse is the SMB2 ECHO Response body, sent by the server to confirm
// that an SMB2 ECHO Request was processed.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/2abe9b3c-c5ab-417f-bcc3-9ab51f2fce35
type EchoResponse struct {
	command_interface.Command

	// Reserved (2 bytes): The server MUST set this to 0, and the client MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewEchoResponse creates a new SMB2 ECHO Response.
func NewEchoResponse() *EchoResponse {
	c := &EchoResponse{}
	c.SetCommandCode(codes.SMB2_ECHO)
	c.StructureSize = EchoResponseStructureSize
	return c
}

// Marshal serializes the ECHO Response body.
func (c *EchoResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], EchoResponseStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the ECHO Response body.
func (c *EchoResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 ECHO Response: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
