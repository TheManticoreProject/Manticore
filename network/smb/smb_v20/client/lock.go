package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// Lock acquires a byte-range lock on the open file: a shared (read) lock when
// exclusive is false, or an exclusive (write) lock when true. By default the
// server blocks until the range is available; when failImmediately is set, a
// conflicting range returns STATUS_LOCK_NOT_GRANTED instead of waiting.
// Wire: SMB2 LOCK.
func (c *Client) Lock(fileId types.SMB2_FILEID, offset, length uint64, exclusive, failImmediately bool) error {
	flags := uint32(commands.SMB2_LOCKFLAG_SHARED_LOCK)
	if exclusive {
		flags = commands.SMB2_LOCKFLAG_EXCLUSIVE_LOCK
	}
	if failImmediately {
		flags |= commands.SMB2_LOCKFLAG_FAIL_IMMEDIATELY
	}
	return c.lock(fileId, offset, length, flags)
}

// Unlock releases a byte-range lock previously acquired with Lock. The offset and
// length MUST exactly match a locked range. Wire: SMB2 LOCK (unlock flag).
func (c *Client) Unlock(fileId types.SMB2_FILEID, offset, length uint64) error {
	return c.lock(fileId, offset, length, commands.SMB2_LOCKFLAG_UNLOCK)
}

// lock issues an SMB2 LOCK with a single lock element carrying the given flags.
func (c *Client) lock(fileId types.SMB2_FILEID, offset, length uint64, flags uint32) error {
	if c.Session == nil || c.Session.TreeId == 0 {
		return fmt.Errorf("no tree connect established")
	}

	req := commands.NewLockRequest()
	req.FileId = fileId
	req.Locks = []commands.LockElement{{
		Offset: types.UINT64(offset),
		Length: types.UINT64(length),
		Flags:  types.ULONG(flags),
	}}

	response, err := c.sendReceive(c.newRequest(req), "Lock")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("lock failed: %s", formatNTStatus(status))
	}
	return nil
}
