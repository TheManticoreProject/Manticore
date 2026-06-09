package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// OplockBreakResponseStructureSize is the fixed StructureSize value for an SMB2
// OPLOCK_BREAK Notification and Response.
const OplockBreakResponseStructureSize = 24

// OplockBreakResponse is the server-sent SMB2 OPLOCK_BREAK message: both the
// unsolicited Notification (informing the client its oplock is being broken) and
// the Response to an Oplock Break Acknowledgment share this layout, which is
// identical to the acknowledgment. (Lease-break forms are SMB 2.1+ and are not
// modeled here.)
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/90d23bb5-cbda-410e-a5c2-ca53674656c9
type OplockBreakResponse struct {
	command_interface.Command

	// OplockLevel (1 byte): The oplock level the server is breaking to / granting.
	OplockLevel types.UCHAR

	// Reserved (1 byte): The server MUST set this to 0.
	Reserved types.UCHAR

	// Reserved2 (4 bytes): The server MUST set this to 0.
	Reserved2 types.ULONG

	// FileId (16 bytes): The open whose oplock is being broken.
	FileId types.SMB2_FILEID
}

// NewOplockBreakResponse creates a new SMB2 OPLOCK_BREAK Notification/Response.
func NewOplockBreakResponse() *OplockBreakResponse {
	c := &OplockBreakResponse{}
	c.SetCommandCode(codes.SMB2_OPLOCK_BREAK)
	c.StructureSize = OplockBreakResponseStructureSize
	return c
}

// Marshal serializes the OPLOCK_BREAK Notification/Response body.
func (c *OplockBreakResponse) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint16(buf[0:2], OplockBreakResponseStructureSize)
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

// Unmarshal deserializes the OPLOCK_BREAK Notification/Response body.
func (c *OplockBreakResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < 24 {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 OPLOCK_BREAK Notification/Response: have %d bytes, need 24", len(data))
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
