package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// LeaseBreakAckStructureSize is the fixed StructureSize for an SMB2 Lease
// Break Acknowledgment (MS-SMB2 2.2.24.2).
const LeaseBreakAckStructureSize = 36

// LeaseBreakAckRequest is the client-sent SMB2 Lease Break Acknowledgment
// (MS-SMB2 2.2.24.2). It shares the OPLOCK_BREAK command code (0x0012) but is
// distinguished by its 36-byte StructureSize (vs. 24 for the oplock form).
type LeaseBreakAckRequest struct {
	command_interface.Command

	Flags      uint32
	LeaseKey   [16]byte
	LeaseState uint32
}

// NewLeaseBreakAckRequest creates a new SMB2 Lease Break Acknowledgment.
func NewLeaseBreakAckRequest() *LeaseBreakAckRequest {
	c := &LeaseBreakAckRequest{}
	c.SetCommandCode(codes.SMB2_OPLOCK_BREAK)
	c.StructureSize = LeaseBreakAckStructureSize
	return c
}

// Marshal serializes the Lease Break Acknowledgment body.
func (c *LeaseBreakAckRequest) Marshal() ([]byte, error) {
	buf := make([]byte, 36)
	binary.LittleEndian.PutUint16(buf[0:2], LeaseBreakAckStructureSize)
	// buf[2:4] Reserved
	binary.LittleEndian.PutUint32(buf[4:8], c.Flags)
	copy(buf[8:24], c.LeaseKey[:])
	binary.LittleEndian.PutUint32(buf[24:28], c.LeaseState)
	// buf[28:36] LeaseDuration (reserved, must be 0)
	return buf, nil
}

// Unmarshal deserializes the Lease Break Acknowledgment body.
func (c *LeaseBreakAckRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < 36 {
		return 0, fmt.Errorf("data too short for Lease Break Acknowledgment: have %d, need 36", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.Flags = binary.LittleEndian.Uint32(data[4:8])
	copy(c.LeaseKey[:], data[8:24])
	c.LeaseState = binary.LittleEndian.Uint32(data[24:28])
	return 36, nil
}
