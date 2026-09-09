package server

import (
	"encoding/binary"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// lockableServer serves one file and returns a client with two handles open on
// it, which is what the cross-handle rules need: a lock is held on a FID, and what
// it excludes is every other FID onto the same file.
func lockableServer(t *testing.T, contents string) (*smb1client.Client, smb1client.FID, smb1client.FID) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("locked.txt", []byte(contents)); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	open := func(what string) smb1client.FID {
		fid, err := client.OpenFile("locked.txt",
			fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
			fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
			fileflags.FILE_OPEN,
			fileflags.FILE_NON_DIRECTORY_FILE)
		if err != nil {
			t.Fatalf("opening the %s handle failed: %v", what, err)
		}
		return fid
	}

	return client, open("first"), open("second")
}

// sendLockingRequest sends a hand-built SMB_COM_LOCKING_ANDX and returns the raw
// reply, so a test can send more ranges than the client API does and can look at
// what came back rather than only whether it failed.
func sendLockingRequest(
	t *testing.T,
	client *smb1client.Client,
	fid smb1client.FID,
	typeOfLock uint8,
	timeout uint32,
	unlocks, locks []types.LOCKING_ANDX_RANGE64,
) []byte {
	t.Helper()

	request := newRequest(codes.SMB_COM_LOCKING_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	cmd := commands.NewLockingAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.TypeOfLock = types.UCHAR(typeOfLock)
	cmd.Timeout = types.ULONG(timeout)
	cmd.Unlocks = unlocks
	cmd.NumberOfRequestedUnlocks = types.USHORT(len(unlocks))
	cmd.Locks = locks
	cmd.NumberOfRequestedLocks = types.USHORT(len(locks))
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the lock request: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	return raw
}

// wideRange builds a 64-bit lock range.
func wideRange(pid uint16, offset, length uint64) types.LOCKING_ANDX_RANGE64 {
	return types.LOCKING_ANDX_RANGE64{
		PID:               types.USHORT(pid),
		ByteOffsetHigh:    types.ULONG(offset >> 32),
		ByteOffsetLow:     types.ULONG(offset),
		LengthInBytesHigh: types.ULONG(length >> 32),
		LengthInBytesLow:  types.ULONG(length),
	}
}

// TestByteRangeLockIsGrantedAndReleased asserts a lock is granted, that the range
// is then unavailable to another handle, and that releasing it makes it available
// again.
func TestByteRangeLockIsGrantedAndReleased(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 0, 8, true); err != nil {
		t.Fatalf("locking [0,8) failed: %v", err)
	}

	// Held, so the other handle cannot have it.
	if err := client.LockFile(second, 0, 8, true); err == nil {
		t.Fatal("a second handle locked a range the first holds")
	}

	if err := client.UnlockFile(first, 0, 8); err != nil {
		t.Fatalf("unlocking [0,8) failed: %v", err)
	}

	// Released, so it is available.
	if err := client.LockFile(second, 0, 8, true); err != nil {
		t.Fatalf("locking [0,8) after it was released failed: %v", err)
	}
}

// TestOverlappingLockIsRefusedAndAdjacentOneIsNot asserts the overlap rule, and
// that it really is an overlap rule rather than a whole-file one.
//
// [MS-CIFS] section 3.3.5.30: "Overlapping locks are not allowed."
func TestOverlappingLockIsRefusedAndAdjacentOneIsNot(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 4, 8, true); err != nil {
		t.Fatalf("locking [4,12) failed: %v", err)
	}

	for _, overlapping := range []struct {
		name           string
		offset, length uint64
	}{
		{name: "same range", offset: 4, length: 8},
		{name: "straddles the front", offset: 0, length: 8},
		{name: "straddles the back", offset: 8, length: 8},
		{name: "inside", offset: 6, length: 2},
		{name: "encloses", offset: 0, length: 16},
	} {
		t.Run(overlapping.name, func(t *testing.T) {
			if err := client.LockFile(second, overlapping.offset, overlapping.length, true); err == nil {
				t.Errorf("locking [%d,%d) succeeded while [4,12) is held",
					overlapping.offset, overlapping.offset+overlapping.length)
			}
		})
	}

	// Adjacent ranges touch but do not overlap, so both are available.
	if err := client.LockFile(second, 0, 4, true); err != nil {
		t.Errorf("locking [0,4), which ends where the held range begins, failed: %v", err)
	}
	if err := client.LockFile(second, 12, 4, true); err != nil {
		t.Errorf("locking [12,16), which begins where the held range ends, failed: %v", err)
	}
}

// TestExclusiveLockRefusesAnotherHandlesAccess asserts an exclusive lock is
// enforced against reads and writes through another handle, and not against the
// handle that holds it.
//
// A lock nothing honours is decoration; this is the assertion that makes the
// feature real.
func TestExclusiveLockRefusesAnotherHandlesAccess(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 0, 8, true); err != nil {
		t.Fatalf("locking [0,8) failed: %v", err)
	}

	if _, err := client.ReadFile(second, 0, 8); err == nil {
		t.Error("another handle read a range under an exclusive lock")
	}
	if _, err := client.WriteFile(second, 0, []byte("XXXXXXXX")); err == nil {
		t.Error("another handle wrote a range under an exclusive lock")
	}

	// Beyond the lock, the other handle is unaffected.
	if _, err := client.ReadFile(second, 8, 8); err != nil {
		t.Errorf("another handle could not read outside the locked range: %v", err)
	}

	// The owning handle is never refused by its own lock.
	if _, err := client.ReadFile(first, 0, 8); err != nil {
		t.Errorf("the owning handle could not read its own locked range: %v", err)
	}
	if _, err := client.WriteFile(first, 0, []byte("ABCDEFGH")); err != nil {
		t.Errorf("the owning handle could not write its own locked range: %v", err)
	}
}

// TestSharedLockRefusesWritesOnly asserts a shared read-only lock lets another
// handle read the range and refuses only its writes.
//
// [MS-CIFS] section 3.3.5.30: "If the lock is a shared read lock, other FIDs
// indicating a separate Open of the same file MUST be denied write access to the
// same bytes."
func TestSharedLockRefusesWritesOnly(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 0, 8, false); err != nil {
		t.Fatalf("taking a shared lock on [0,8) failed: %v", err)
	}

	if _, err := client.ReadFile(second, 0, 8); err != nil {
		t.Errorf("another handle could not read a range under a shared lock: %v", err)
	}
	if _, err := client.WriteFile(second, 0, []byte("XXXXXXXX")); err == nil {
		t.Error("another handle wrote a range under a shared lock")
	}
}

// TestUnlockingARangeNotHeldIsRefused asserts a release of something not held is
// reported rather than ignored.
//
// [MS-CIFS] section 3.3.5.30 names the status: "If the PID in the unlock request
// does not match Server.Open.Locks in the Open, the server MUST send an error
// response message with status set to STATUS_RANGE_NOT_LOCKED".
func TestUnlockingARangeNotHeldIsRefused(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	// Never locked at all.
	raw := sendLockingRequest(t, client, first, commands.LockingAndxLargeFiles, 0,
		[]types.LOCKING_ANDX_RANGE64{wideRange(0x1234, 0, 8)}, nil)
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != uint32(nt_status.NT_STATUS_RANGE_NOT_LOCKED) {
		t.Errorf("unlocking a range never held reported 0x%08X, want STATUS_RANGE_NOT_LOCKED (0x%08X)",
			status, uint32(nt_status.NT_STATUS_RANGE_NOT_LOCKED))
	}

	// Held, but by the other handle.
	if err := client.LockFile(first, 0, 8, true); err != nil {
		t.Fatalf("locking [0,8) failed: %v", err)
	}
	if err := client.UnlockFile(second, 0, 8); err == nil {
		t.Error("a handle released a lock held by another handle")
	}
}

// TestClosingAHandleReleasesItsLocks asserts a closed handle leaves nothing
// locked.
//
// A lock owned by a handle that no longer exists could never be released, so the
// range would stay locked for as long as the share did.
func TestClosingAHandleReleasesItsLocks(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 0, 16, true); err != nil {
		t.Fatalf("locking [0,16) failed: %v", err)
	}
	if err := client.CloseFile(first); err != nil {
		t.Fatalf("closing the holding handle failed: %v", err)
	}

	if err := client.LockFile(second, 0, 16, true); err != nil {
		t.Fatalf("the range was still locked after its holder closed: %v", err)
	}
	if _, err := client.ReadFile(second, 0, 16); err != nil {
		t.Errorf("reading after the holder closed failed: %v", err)
	}
}

// TestLockRequestIsAtomic asserts a request whose second range conflicts grants
// neither range.
//
// [MS-CIFS] section 3.3.5.30: "This client request is atomic. If any of the lock
// ranges times out because the area to be locked is already locked, or the
// lock/unlock request otherwise fails, the lock state of the file MUST NOT be
// changed."
func TestLockRequestIsAtomic(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	// The second handle holds [8,12), so the pair below cannot be granted whole.
	if err := client.LockFile(second, 8, 4, true); err != nil {
		t.Fatalf("locking [8,12) failed: %v", err)
	}

	raw := sendLockingRequest(t, client, first, commands.LockingAndxLargeFiles, 0, nil,
		[]types.LOCKING_ANDX_RANGE64{
			wideRange(0x1234, 0, 4),
			wideRange(0x1234, 8, 4),
		})
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != uint32(nt_status.NT_STATUS_LOCK_NOT_GRANTED) {
		t.Fatalf("the request reported 0x%08X, want STATUS_LOCK_NOT_GRANTED (0x%08X)",
			status, uint32(nt_status.NT_STATUS_LOCK_NOT_GRANTED))
	}

	// [0,4) must not have been granted on the way to the failure, so the other
	// handle can still take it.
	if err := client.LockFile(second, 0, 4, true); err != nil {
		t.Errorf("[0,4) was left locked by a request that failed: %v", err)
	}
}

// TestLockRequestWithNoRangesIsAnsweredWithSilence asserts a request carrying no
// ranges gets no response at all.
//
// That is an OpLock Break Request, and [MS-CIFS] section 3.3.5.30 is explicit:
// "If NumberOfRequestedUnlocks and NumberOfRequestedLocks are both zero (0x0000)
// [...] the server MUST NOT send an SMB_COM_LOCKING_ANDX Response". A server that
// answered anyway would leave one unread reply on the connection and every
// subsequent response would be read against the wrong request.
func TestLockRequestWithNoRangesIsAnsweredWithSilence(t *testing.T) {
	client, first, _ := lockableServer(t, "0123456789abcdef")

	request := newRequest(codes.SMB_COM_LOCKING_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	cmd := commands.NewLockingAndxRequest()
	cmd.FID = types.USHORT(first)
	cmd.TypeOfLock = types.UCHAR(commands.LockingAndxLargeFiles | commands.LockingAndxOplockRelease)
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the lock request: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Nothing may be read for that request, so the next thing on the connection
	// has to be the answer to what is sent next. An echo is used because its
	// response is unmistakable.
	echoed, err := client.Echo([]byte("after the break acknowledgement"))
	if err != nil {
		t.Fatalf("the connection did not survive a break acknowledgement: %v", err)
	}
	if string(echoed) != "after the break acknowledgement" {
		t.Fatalf("the echo returned %q, so the reply stream is out of step — the lock request was answered",
			echoed)
	}
}

// TestChangingALockTypeIsRefused asserts the atomic lock-type change is refused
// rather than approximated.
//
// [MS-CIFS] section 2.2.4.32.1: "If the server cannot do this in an atomic
// fashion, the server MUST reject this request". Releasing and re-taking the lock
// would open a window in which the range is unlocked, which is exactly what the
// client asked to avoid.
func TestChangingALockTypeIsRefused(t *testing.T) {
	client, first, _ := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 0, 8, false); err != nil {
		t.Fatalf("taking a shared lock failed: %v", err)
	}

	raw := sendLockingRequest(t, client, first,
		commands.LockingAndxLargeFiles|commands.LockingAndxChangeLockType, 0, nil,
		[]types.LOCKING_ANDX_RANGE64{wideRange(0x1234, 0, 8)})
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != uint32(nt_status.NT_STATUS_NOT_SUPPORTED) {
		t.Errorf("a lock-type change reported 0x%08X, want STATUS_NOT_SUPPORTED (0x%08X)",
			status, uint32(nt_status.NT_STATUS_NOT_SUPPORTED))
	}
}

// TestBlockedLockWaitsAndThenFails asserts a request that offers to wait is
// refused once its timeout runs out, and that the wait is bounded by the timeout
// rather than by the server's ceiling.
func TestBlockedLockWaitsAndThenFails(t *testing.T) {
	client, first, second := lockableServer(t, "0123456789abcdef")

	if err := client.LockFile(first, 0, 8, true); err != nil {
		t.Fatalf("locking [0,8) failed: %v", err)
	}

	// 50 ms, so the wait is observable but the test is not slow.
	raw := sendLockingRequest(t, client, second, commands.LockingAndxLargeFiles, 50, nil,
		[]types.LOCKING_ANDX_RANGE64{wideRange(0x4321, 0, 8)})
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != uint32(nt_status.NT_STATUS_LOCK_NOT_GRANTED) {
		t.Errorf("a lock that waited reported 0x%08X, want STATUS_LOCK_NOT_GRANTED (0x%08X)",
			status, uint32(nt_status.NT_STATUS_LOCK_NOT_GRANTED))
	}

	// The connection is still usable after the wait.
	if _, err := client.Echo([]byte("still here")); err != nil {
		t.Errorf("the connection did not survive a lock wait: %v", err)
	}

}
