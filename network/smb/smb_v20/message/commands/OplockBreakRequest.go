package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// OplockBreakRequestStructureSize is the fixed StructureSize value for an SMB2
// OPLOCK_BREAK Acknowledgment.
const OplockBreakRequestStructureSize = 24

// OplockBreakRequest is the SMB2 OPLOCK_BREAK Acknowledgment, sent by the client
// in response to an oplock break notification to acknowledge the reduced oplock
// level. (The lease-break acknowledgment, StructureSize 36, is an SMB 2.1+
// feature and is not modeled here.)
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/8b2f9f49-93de-479c-81c2-795b1059deea
type OplockBreakRequest struct {
	command_interface.Command

	// OplockLevel (1 byte): The lowered oplock level the client accepts.
	OplockLevel types.UCHAR

	// Reserved (1 byte): The client MUST set this to 0.
	Reserved types.UCHAR

	// Reserved2 (4 bytes): The client MUST set this to 0.
	Reserved2 types.ULONG

	// FileId (16 bytes): The open whose oplock is being acknowledged.
	FileId types.SMB2_FILEID
}

// NewOplockBreakRequest creates a new SMB2 OPLOCK_BREAK Acknowledgment.
func NewOplockBreakRequest() *OplockBreakRequest {
	c := &OplockBreakRequest{}
	c.SetCommandCode(codes.SMB2_OPLOCK_BREAK)
	c.StructureSize = OplockBreakRequestStructureSize
	return c
}

// Marshal serializes the OPLOCK_BREAK Acknowledgment body.
func (c *OplockBreakRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint16(buf[0:2], OplockBreakRequestStructureSize)
	buf[2] = byte(c.OplockLevel)
	buf[3] = byte(c.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], c.Reserved2)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	buf = append(buf, fileId...)
	return buf, nil
}

// Unmarshal deserializes the OPLOCK_BREAK Acknowledgment body.
func (c *OplockBreakRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 24 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 OPLOCK_BREAK Acknowledgment: have %d bytes, need 24", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.OplockLevel = data[2]
	c.Reserved = data[3]
	c.Reserved2 = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}
	return 24, nil
}
