package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// LeaseBreakNotificationStructureSize is the fixed StructureSize for an SMB2
// Lease Break Notification (MS-SMB2 2.2.23.2). The server sends this when it
// needs to break a lease granted to the client.
const LeaseBreakNotificationStructureSize = 44

// SMB2 lease break flags (MS-SMB2 2.2.23.2).
const SMB2_NOTIFY_BREAK_LEASE_FLAG_ACK_REQUIRED uint32 = 0x01

// LeaseBreakNotification is the server-sent SMB2 Lease Break Notification
// (MS-SMB2 2.2.23.2). It shares the OPLOCK_BREAK command code (0x0012) but is
// distinguished by its 44-byte StructureSize (vs. 24 for the oplock form).
type LeaseBreakNotification struct {
	command_interface.Command

	NewEpoch          uint16
	Flags             uint32
	LeaseKey          [16]byte
	CurrentLeaseState uint32
	NewLeaseState     uint32
}

// NewLeaseBreakNotification creates a new SMB2 Lease Break Notification.
func NewLeaseBreakNotification() *LeaseBreakNotification {
	c := &LeaseBreakNotification{}
	c.SetCommandCode(codes.SMB2_OPLOCK_BREAK)
	c.StructureSize = LeaseBreakNotificationStructureSize
	return c
}

// AckRequired reports whether the server expects a Lease Break Acknowledgment.
func (c *LeaseBreakNotification) AckRequired() bool {
	return c.Flags&SMB2_NOTIFY_BREAK_LEASE_FLAG_ACK_REQUIRED != 0
}

// Marshal serializes the Lease Break Notification body.
func (c *LeaseBreakNotification) Marshal() ([]byte, error) {
	buf := make([]byte, 44)
	binary.LittleEndian.PutUint16(buf[0:2], LeaseBreakNotificationStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], c.NewEpoch)
	binary.LittleEndian.PutUint32(buf[4:8], c.Flags)
	copy(buf[8:24], c.LeaseKey[:])
	binary.LittleEndian.PutUint32(buf[24:28], c.CurrentLeaseState)
	binary.LittleEndian.PutUint32(buf[28:32], c.NewLeaseState)
	return buf, nil
}

// Unmarshal deserializes the Lease Break Notification body.
func (c *LeaseBreakNotification) Unmarshal(data []byte) (int, error) {
	if len(data) < 44 {
		return 0, fmt.Errorf("data too short for Lease Break Notification: have %d, need 44", len(data))
	}
	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	c.NewEpoch = binary.LittleEndian.Uint16(data[2:4])
	c.Flags = binary.LittleEndian.Uint32(data[4:8])
	copy(c.LeaseKey[:], data[8:24])
	c.CurrentLeaseState = binary.LittleEndian.Uint32(data[24:28])
	c.NewLeaseState = binary.LittleEndian.Uint32(data[28:32])
	return 44, nil
}
