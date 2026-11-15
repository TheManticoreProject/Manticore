package server

import (
	"errors"
	"io"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// applyLockRanges applies one atomic unlock-then-lock step to a handle without
// waiting. The combined lock-and-read and write-and-unlock commands carry no
// timeout, so — unlike SMB_COM_LOCKING_ANDX — they never block: a range that is
// already locked fails immediately. The resource accounting and outcome mapping
// mirror handleLockingAndx.
//
// Parameters:
//   - open: the handle the ranges apply to
//   - unlocks: ranges to release
//   - locks: ranges to acquire
//
// Returns:
//   - NT_STATUS_SUCCESS on success, or the mapped failure status
func (c *Connection) applyLockRanges(open *Open, unlocks, locks []lockRange) nt_status.NT_STATUS {
	locksTable := open.Tree.Share.locks
	if locksTable == nil {
		// A share without a lock table cannot enforce a lock; granting one would be
		// worse than refusing.
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	limit := c.Server.config.MaxLocksPerConnection
	if limit <= 0 {
		limit = DefaultMaxLocksPerConnection
	}
	growth := len(locks) - len(unlocks)
	if growth > 0 && c.locksHeld+growth > limit {
		return nt_status.NT_STATUS_INSUFF_SERVER_RESOURCES
	}

	switch locksTable.Apply(open, unlocks, locks, false) {
	case lockOutcomeGranted:
		c.locksHeld += growth
		return nt_status.NT_STATUS_SUCCESS
	case lockOutcomeNotLocked:
		return nt_status.NT_STATUS_RANGE_NOT_LOCKED
	default:
		return nt_status.NT_STATUS_LOCK_NOT_GRANTED
	}
}

// handleLockAndRead answers SMB_COM_LOCK_AND_READ: it takes an exclusive lock on
// a byte range and then reads it, in one exchange.
//
// [MS-CIFS] section 3.3.5.9: the lock is taken first, and if it cannot be granted
// the command fails with no data read. The command carries no timeout, so it does
// not wait for a conflicting lock to clear. The lock is left held on return — the
// client releases it later, usually with SMB_COM_WRITE_AND_UNLOCK.
//
// Wire format: [MS-CIFS] section 2.2.4.9.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleLockAndRead(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.LockAndReadRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if !open.Readable {
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	if open.IsDirectory {
		return nt_status.NT_STATUS_FILE_IS_A_DIRECTORY
	}

	offset := int64(uint32(request.ReadOffsetInBytes))

	// Lock the range before reading it.
	lock := lockRange{
		Offset: uint64(offset),
		Length: uint64(request.CountOfBytesToRead),
		PID:    uint32(req.Header.GetPID()),
	}
	if status := conn.applyLockRanges(open, nil, []lockRange{lock}); status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	length := int(request.CountOfBytesToRead)
	if limit := int(conn.Server.config.MaxBufferSize) - readResponseOverhead; length > limit {
		length = limit
	}
	if length < 0 {
		length = 0
	}

	file, status := fileFor(open)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	buffer := make([]byte, length)
	read, err := file.ReadAt(buffer, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return statusForFSError(err)
	}

	response := commands.NewLockAndReadResponse()
	response.CountOfBytesReturned = types.USHORT(read)
	// The data block is BufferFormat(0x01) CountOfBytesRead(2) Bytes, per
	// [MS-CIFS] section 2.2.4.9.2.
	response.BytesRead.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	response.BytesRead.Buffer = []types.UCHAR(buffer[:read])
	return conn.answer(w, response)
}

// handleWriteAndUnlock answers SMB_COM_WRITE_AND_UNLOCK: it writes a byte range
// and then releases the lock on that range, in one exchange.
//
// [MS-CIFS] section 3.3.5.11: the write happens first, then the range the client
// previously locked (with SMB_COM_LOCK_AND_READ or SMB_COM_LOCKING_ANDX) is
// released. A range that is not held is reported with STATUS_RANGE_NOT_LOCKED; the
// write has already taken effect by then, which matches the command's order.
//
// Wire format: [MS-CIFS] section 2.2.4.11.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleWriteAndUnlock(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.WriteAndUnlockRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	offset := int64(uint32(request.WriteOffsetInBytes))
	written, status := conn.coreWrite(open, offset, []byte(request.Data.Buffer), int(request.CountOfBytesToWrite))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// Release the range the client locked for this write.
	lock := lockRange{
		Offset: uint64(offset),
		Length: uint64(request.CountOfBytesToWrite),
		PID:    uint32(req.Header.GetPID()),
	}
	if status := conn.applyLockRanges(open, []lockRange{lock}, nil); status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	response := commands.NewWriteAndUnlockResponse()
	response.CountOfBytesWritten = types.USHORT(written)
	return conn.answer(w, response)
}
