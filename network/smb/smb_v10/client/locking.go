package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TypeOfLock bit flags for SMB_COM_LOCKING_ANDX (MS-CIFS 2.2.4.32.1).
const (
	lockTypeSharedLock = 0x01 // shared (read-only) lock; absence means exclusive
	lockTypeLargeFiles = 0x10 // ranges are in 64-bit LOCKING_ANDX_RANGE64 format
)

// range64 builds a 64-bit lock range for the given file offset and length, owned
// by the supplied PID.
func range64(pid uint16, offset, length uint64) types.LOCKING_ANDX_RANGE64 {
	return types.LOCKING_ANDX_RANGE64{
		PID:               types.USHORT(pid),
		ByteOffsetHigh:    types.ULONG(offset >> 32),
		ByteOffsetLow:     types.ULONG(offset),
		LengthInBytesHigh: types.ULONG(length >> 32),
		LengthInBytesLow:  types.ULONG(length),
	}
}

// LockFile locks a byte range [offset, offset+length) on the open file referenced
// by fid. When exclusive is false the lock is a shared (read-only) lock. The
// request fails immediately (no wait) if the range cannot be locked.
//
// Wire: SMB_COM_LOCKING_ANDX with a single 64-bit lock range.
func (c *Client) LockFile(fid FID, offset, length uint64, exclusive bool) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_LOCKING_ANDX)

	cmd := commands.NewLockingAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.TypeOfLock = types.UCHAR(lockTypeLargeFiles)
	if !exclusive {
		cmd.TypeOfLock |= lockTypeSharedLock
	}
	cmd.Timeout = types.ULONG(0)
	cmd.NumberOfRequestedLocks = types.USHORT(1)
	cmd.Locks = []types.LOCKING_ANDX_RANGE64{range64(uint16(msg.Header.GetPID()), offset, length)}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "LockingAndx(lock)")
	if err != nil {
		return err
	}
	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("LockingAndx lock failed: 0x%08x", response.Header.Status)
	}
	return nil
}

// UnlockFile releases a previously acquired byte-range lock [offset, offset+length)
// on the open file referenced by fid.
//
// Wire: SMB_COM_LOCKING_ANDX with a single 64-bit unlock range.
func (c *Client) UnlockFile(fid FID, offset, length uint64) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_LOCKING_ANDX)

	cmd := commands.NewLockingAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.TypeOfLock = types.UCHAR(lockTypeLargeFiles)
	cmd.NumberOfRequestedUnlocks = types.USHORT(1)
	cmd.Unlocks = []types.LOCKING_ANDX_RANGE64{range64(uint16(msg.Header.GetPID()), offset, length)}

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "LockingAndx(unlock)")
	if err != nil {
		return err
	}
	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("LockingAndx unlock failed: 0x%08x", response.Header.Status)
	}
	return nil
}
