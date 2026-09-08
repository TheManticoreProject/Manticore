package server

import (
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// lockingRange names the wire range type the message layer decodes into, so the
// widening below reads as what it is.
type lockingRange = types.LOCKING_ANDX_RANGE64

// lockWaitInterval is how often a blocked lock request retries.
//
// The server answers one request per connection at a time, so a wait holds this
// client and nothing else. Polling is used rather than waking on release because
// the thing being waited for may be released by another connection, another share
// or a closing handle, and a poll is correct for all of them without a
// notification path that would have to reach each one.
const lockWaitInterval = 10 * time.Millisecond

// handleLockingAndx answers SMB_COM_LOCKING_ANDX: it releases and acquires
// byte-range locks on an open handle.
//
// [MS-CIFS] section 3.3.5.30 processes the request in three parts, all of which
// are executed: the unlocks, then the locks, then an OpLock release if the
// OPLOCK_RELEASE bit is set. The first two are one atomic step — a failure in
// either leaves the file's lock state untouched.
//
// Wire format: [MS-CIFS] section 2.2.4.32.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleLockingAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.LockingAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	typeOfLock := uint8(request.TypeOfLock)

	// A change of lock type has to be atomic or refused, and nothing here can
	// swap a lock's type without releasing it first. [MS-CIFS] section 2.2.4.32.1:
	// "If the server cannot do this in an atomic fashion, the server MUST reject
	// this request".
	if typeOfLock&commands.LockingAndxChangeLockType != 0 {
		logger.Debugf("SMB1 server: %s asked to change a lock's type on FID 0x%04X, which is refused",
			conn.Remote, uint16(request.FID))
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	// Cancelling an outstanding lock request has nothing to cancel: a request
	// that is waiting is waiting on this connection's own goroutine, so no other
	// request from this client can arrive while it does. Answering success is
	// therefore accurate rather than a stub — there is provably no outstanding
	// request to cancel.
	if typeOfLock&commands.LockingAndxCancelLock != 0 {
		logger.Debugf("SMB1 server: %s cancelled lock requests on FID 0x%04X, of which none were outstanding",
			conn.Remote, uint16(request.FID))
		return answerLocking(conn, w)
	}

	unlocks := lockRangesOf(request.Unlocks)
	locks := lockRangesOf(request.Locks)
	shared := typeOfLock&commands.LockingAndxSharedLock != 0

	// An OpLock Break Request carries no ranges, and [MS-CIFS] section 3.3.5.30
	// is explicit that it is answered with silence: "If NumberOfRequestedUnlocks
	// and NumberOfRequestedLocks are both zero (0x0000) [...] the server MUST NOT
	// send an SMB_COM_LOCKING_ANDX Response". No OpLock is ever granted here, so
	// there is none to release, and the same section says that is not an error.
	if len(unlocks) == 0 && len(locks) == 0 {
		logger.Debugf("SMB1 server: %s sent a lock request with no ranges on FID 0x%04X, which is answered with silence",
			conn.Remote, uint16(request.FID))
		return nt_status.NT_STATUS_SUCCESS
	}

	locksTable := open.Tree.Share.locks
	if locksTable == nil {
		// A share registered without a table cannot track a lock, and granting
		// one it could not enforce would be worse than refusing.
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	deadline, waits := lockDeadline(conn, uint32(request.Timeout))
	for {
		switch outcome := locksTable.Apply(open, unlocks, locks, shared); outcome {
		case lockOutcomeGranted:
			logger.Debugf("SMB1 server: %s locked %d and unlocked %d ranges of %q",
				conn.Remote, len(locks), len(unlocks), open.Path)
			return answerLocking(conn, w)

		case lockOutcomeNotLocked:
			// Waiting cannot make a range that is not held become held, so this
			// fails immediately however long the client offered to wait.
			logger.Debugf("SMB1 server: %s asked to unlock a range of %q it does not hold",
				conn.Remote, open.Path)
			return nt_status.NT_STATUS_RANGE_NOT_LOCKED

		default:
			if !waits || !time.Now().Before(deadline) {
				logger.Debugf("SMB1 server: %s could not lock a range of %q, which is already locked",
					conn.Remote, open.Path)
				return nt_status.NT_STATUS_LOCK_NOT_GRANTED
			}
		}

		time.Sleep(lockWaitInterval)
	}
}

// answerLocking sends the response, which carries nothing but success.
func answerLocking(conn *Connection, w ResponseWriter) nt_status.NT_STATUS {
	if err := w.WriteResponse(commands.NewLockingAndxResponse()); err != nil {
		logger.Debugf("SMB1 server: failed to answer the lock request for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// lockDeadline turns a request's Timeout into a deadline.
//
// [MS-CIFS] section 2.2.4.32.1: zero fails immediately, 0xFFFFFFFF waits forever,
// and anything else is a number of milliseconds. Waiting forever is bounded by
// Config.MaxLockWait, because the connection serves nothing else while it waits —
// an unbounded wait would be a client's own denial of service, and the client
// cannot cancel it either, since the cancel would arrive behind the request it
// wanted to cancel.
//
// Parameters:
//   - conn: the connection, for its configured ceiling
//   - timeout: the request's Timeout field
//
// Returns:
//   - The deadline, and whether the request waits at all
func lockDeadline(conn *Connection, timeout uint32) (time.Time, bool) {
	if timeout == 0 {
		return time.Time{}, false
	}

	ceiling := conn.Server.config.MaxLockWait
	if ceiling <= 0 {
		ceiling = DefaultMaxLockWait
	}

	wait := ceiling
	if timeout != lockTimeoutForever {
		if requested := time.Duration(timeout) * time.Millisecond; requested < ceiling {
			wait = requested
		}
	}
	return time.Now().Add(wait), true
}

// lockTimeoutForever is the Timeout value that asks the server to wait as long as
// it takes ([MS-CIFS] section 2.2.4.32.1).
const lockTimeoutForever = 0xFFFFFFFF

// lockRangesOf widens the message layer's range entries into the form the lock
// table works with.
//
// The message layer hands over LOCKING_ANDX_RANGE64 whichever format was on the
// wire, so the two halves of the 64-bit offset and length are recombined here and
// the 32-bit case arrives with its high words already zero.
func lockRangesOf(entries []lockingRange) []lockRange {
	ranges := make([]lockRange, 0, len(entries))
	for _, entry := range entries {
		ranges = append(ranges, lockRange{
			Offset: uint64(entry.ByteOffsetHigh)<<32 | uint64(entry.ByteOffsetLow),
			Length: uint64(entry.LengthInBytesHigh)<<32 | uint64(entry.LengthInBytesLow),
			PID:    uint32(entry.PID),
		})
	}
	return ranges
}
