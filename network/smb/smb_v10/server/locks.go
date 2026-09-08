package server

import (
	"math"
	"sync"
)

// byteRangeLock is one byte-range lock granted on a file.
//
// The owning handle is part of the identity, not just bookkeeping: [MS-CIFS]
// section 3.3.5.30 holds a lock "based upon the FID used to create the lock", so
// which handle took it decides who may read and write the bytes underneath it.
type byteRangeLock struct {
	// Offset and Length describe the range. A range may extend past the end of
	// the file, and locking one does not extend the file.
	Offset uint64
	Length uint64

	// PID is the process identifier from the range entry. Only this PID may
	// release the lock.
	PID uint32

	// Shared records a shared read-only lock rather than an exclusive one, which
	// changes what other handles are refused rather than whether the lock
	// overlaps.
	Shared bool

	// owner is the handle the lock was taken on.
	owner *Open
}

// end returns the first offset past the lock, saturating rather than wrapping so
// a range described as running to the end of the address space still compares
// correctly.
func (l byteRangeLock) end() uint64 {
	end := l.Offset + l.Length
	if end < l.Offset {
		return math.MaxUint64
	}
	return end
}

// overlaps reports whether this lock covers any byte of the given range.
//
// A zero-length range covers nothing and so overlaps nothing, which is what makes
// a zero-length lock legal and inert.
func (l byteRangeLock) overlaps(offset, length uint64) bool {
	if l.Length == 0 || length == 0 {
		return false
	}
	end := offset + length
	if end < offset {
		end = math.MaxUint64
	}
	return l.Offset < end && offset < l.end()
}

// lockTable holds the byte-range locks granted on the files of one share.
//
// It belongs to the share rather than to a connection because a lock is a
// statement about a file, and the handles it has to exclude are on other
// connections as much as this one. That makes it shared state, so every path
// through it takes the mutex.
type lockTable struct {
	mutex sync.Mutex

	// granted maps a share-relative path to the locks held on that file. A path
	// with no locks carries no entry, so an unlocked file costs nothing.
	granted map[string][]byteRangeLock
}

// newLockTable builds an empty table.
func newLockTable() *lockTable {
	return &lockTable{granted: make(map[string][]byteRangeLock)}
}

// lockRange is one entry of a request's Locks or Unlocks array, in the widened
// form the message layer hands over.
type lockRange struct {
	Offset uint64
	Length uint64
	PID    uint32
}

// Apply releases and then acquires ranges on one file, as one atomic step.
//
// [MS-CIFS] section 3.3.5.30: "This client request is atomic. If any of the lock
// ranges times out because the area to be locked is already locked, or the
// lock/unlock request otherwise fails, the lock state of the file MUST NOT be
// changed." So the whole result is computed against a copy and committed only if
// every part of it succeeded.
//
// The order is the specification's: unlocks first, then locks, both of which are
// executed.
//
// Parameters:
//   - open: the handle the request arrived on
//   - unlocks: ranges to release
//   - locks: ranges to acquire
//   - shared: whether the acquired locks are shared read-only locks
//
// Returns:
//   - lockOutcome saying whether it succeeded and, if not, why
func (t *lockTable) Apply(open *Open, unlocks, locks []lockRange, shared bool) lockOutcome {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	path := open.Path

	// A copy, so a failure half-way leaves the table as it was.
	working := make([]byteRangeLock, len(t.granted[path]))
	copy(working, t.granted[path])

	for _, unlock := range unlocks {
		index := findLock(working, open, unlock)
		if index < 0 {
			// [MS-CIFS] section 3.3.5.30: a range the requesting PID did not lock
			// is STATUS_RANGE_NOT_LOCKED, and that includes a range locked by
			// another PID.
			return lockOutcomeNotLocked
		}
		working = append(working[:index], working[index+1:]...)
	}

	for _, lock := range locks {
		// "Overlapping locks are not allowed" and locking "SHOULD fail with
		// STATUS_LOCK_NOT_GRANTED if any subranges or overlapping ranges are
		// locked, even if they are currently locked by the PID requesting the new
		// lock". So there is no same-owner exemption here: any overlap refuses.
		for _, held := range working {
			if held.overlaps(lock.Offset, lock.Length) {
				return lockOutcomeConflict
			}
		}
		working = append(working, byteRangeLock{
			Offset: lock.Offset,
			Length: lock.Length,
			PID:    lock.PID,
			Shared: shared,
			owner:  open,
		})
	}

	if len(working) == 0 {
		delete(t.granted, path)
	} else {
		t.granted[path] = working
	}
	return lockOutcomeGranted
}

// lockOutcome is why an Apply succeeded or failed, kept separate from a status so
// the caller can decide whether waiting could change the answer.
type lockOutcome int

const (
	// lockOutcomeGranted means the whole request was applied.
	lockOutcomeGranted lockOutcome = iota

	// lockOutcomeConflict means a range asked for is already locked. Waiting may
	// change this, so it is the outcome a timeout retries on.
	lockOutcomeConflict

	// lockOutcomeNotLocked means a range asked to be released is not held by the
	// requesting handle and PID. Waiting cannot change this.
	lockOutcomeNotLocked
)

// findLock locates a lock matching a range exactly, held on the given handle by
// the range's PID, and returns its index or -1.
//
// The match is exact rather than by containment: a lock is released as a unit, and
// splitting one to satisfy a partial unlock would leave the client holding
// something it never asked for.
func findLock(held []byteRangeLock, open *Open, wanted lockRange) int {
	for index, lock := range held {
		if lock.owner == open && lock.PID == wanted.PID &&
			lock.Offset == wanted.Offset && lock.Length == wanted.Length {
			return index
		}
	}
	return -1
}

// Blocks reports whether a lock held by some other handle stops this one reading
// or writing a range.
//
// [MS-CIFS] section 3.3.5.30: "any process (PID) using the FID specified in the
// creation of the lock has access to the locked bytes. If the lock is an exclusive
// lock, other FIDs indicating a separate Open of the same file MUST be denied
// access to the same bytes. If the lock is a shared read lock, other FIDs
// indicating a separate Open of the same file MUST be denied write access to the
// same bytes."
//
// So the owning handle is never blocked by its own lock, an exclusive lock blocks
// both directions for everyone else, and a shared lock blocks only writes.
//
// Parameters:
//   - path: the share-relative path being accessed
//   - open: the handle doing the access
//   - offset, length: the range being accessed
//   - writing: whether the access is a write
//
// Returns:
//   - true when the access must be refused
func (t *lockTable) Blocks(path string, open *Open, offset, length uint64, writing bool) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, held := range t.granted[path] {
		if held.owner == open {
			continue
		}
		if !held.overlaps(offset, length) {
			continue
		}
		if writing || !held.Shared {
			return true
		}
	}
	return false
}

// ReleaseAll drops every lock held on a handle, which is what closing it does.
//
// [MS-CIFS] section 2.2.4.32.1: "Closing a file with locks still in force causes
// the locks to be released in a nondeterministic order." Releasing them all at
// once satisfies that and leaves no lock owned by a handle that no longer exists —
// which would otherwise be a lock nothing could ever release.
func (t *lockTable) ReleaseAll(open *Open) {
	if open == nil {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	held, present := t.granted[open.Path]
	if !present {
		return
	}

	kept := held[:0]
	for _, lock := range held {
		if lock.owner != open {
			kept = append(kept, lock)
		}
	}
	if len(kept) == 0 {
		delete(t.granted, open.Path)
		return
	}
	t.granted[open.Path] = kept
}

// HeldOn returns how many locks are held on a path, for tests and for a caller
// that wants to report on a share.
func (t *lockTable) HeldOn(path string) int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return len(t.granted[path])
}
