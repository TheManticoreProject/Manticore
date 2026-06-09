package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// EchoRequestStructureSize is the fixed StructureSize value for an SMB2 ECHO Request.
const EchoRequestStructureSize = 4

// EchoRequest is the SMB2 ECHO Request body, sent by a client to determine
// whether a server is still processing requests.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/d939504d-57e2-4c0e-8ad5-1678b6fccca1
type EchoRequest struct {
	command_interface.Command

	// Reserved (2 bytes): The client MUST set this to 0, and the server MUST
	// ignore it on receipt.
	Reserved types.USHORT
}

// NewEchoRequest creates a new SMB2 ECHO Request.
func NewEchoRequest() *EchoRequest {
	c := &EchoRequest{}
	c.SetCommandCode(codes.SMB2_ECHO)
	c.StructureSize = EchoRequestStructureSize
	return c
}

// Marshal serializes the ECHO Request body.
func (c *EchoRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], EchoRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.Reserved)
	return buf, nil
}

// Unmarshal deserializes the ECHO Request body.
func (c *EchoRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 ECHO Request: have %d bytes, need 4", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Reserved = binary.LittleEndian.Uint16(data[2:4])
	return 4, nil
}
