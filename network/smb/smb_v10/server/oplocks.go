package server

import (
	"strings"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// The OpLock levels an SMB_COM_NT_CREATE_ANDX response reports ([MS-CIFS]
// section 2.2.4.64.2).
const (
	// OplockLevelNone is granted when the server grants nothing.
	OplockLevelNone = 0x00

	// OplockLevelExclusive and OplockLevelBatch are the two exclusive levels. A
	// client asks for one of them; this server never grants either, because both
	// let a client cache writes and so require a break the client must
	// acknowledge before the write behind it can proceed.
	OplockLevelExclusive = 0x01
	OplockLevelBatch     = 0x02

	// OplockLevelLevelII is a level II oplock: "multiple readers of a file and no
	// writers" ([MS-SMB] section 1.1). It permits a client to cache reads and
	// nothing else, which is why it is the level this server grants.
	OplockLevelLevelII = 0x03
)

// The NewOpLockLevel values a break notification carries ([MS-CIFS] section
// 2.2.4.32.1): "If NewOpLockLevel is 0x00, the client possesses no OpLocks on the
// file at all. If NewOpLockLevel is 0x01, then the client possesses a Level II
// OpLock."
//
// Only the first is ever sent here, since level II is the only level granted and
// there is nothing below it to break to.
const (
	breakToNone    = 0x00
	breakToLevelII = 0x01
)

// The NT_CREATE_ANDX request flags that ask for an oplock ([MS-CIFS] section
// 2.2.4.64.1).
const (
	ntCreateRequestOplock  = 0x00000002
	ntCreateRequestOpBatch = 0x00000004
)

// oplockHolder is one level II oplock granted on a file.
//
// It carries the identifiers a break notification needs rather than a pointer to
// the connection's tables, because the break is sent from whichever goroutine
// caused it: the tables belong to the receive loop of the connection that owns
// them and have no locking of their own.
type oplockHolder struct {
	// conn is the connection to notify, and fid, uid and tid name the handle in
	// the terms that connection assigned.
	conn *Connection
	fid  uint16
	uid  uint16
	tid  uint16

	// owner is the handle the oplock was granted on, which is what identifies it
	// when the handle closes or the client acknowledges a break.
	owner *Open
}

// oplockTable holds the level II oplocks granted on the files of one share.
//
// It belongs to the share rather than to a connection for the same reason the
// lock table does: an oplock is a statement about a file, and the handles it has
// to notify are on other connections as much as this one. Every path through it
// takes the mutex.
type oplockTable struct {
	mutex sync.Mutex

	// granted maps an upper-cased share-relative path to the oplocks held on that
	// file. A file with none carries no entry, so a file nobody caches costs
	// nothing.
	//
	// The key is upper-cased because the file commands match names
	// case-insensitively, and two spellings of one path have to reach one entry
	// or a break would miss a holder.
	granted map[string][]*oplockHolder
}

// newOplockTable builds an empty table.
func newOplockTable() *oplockTable {
	return &oplockTable{granted: make(map[string][]*oplockHolder)}
}

// oplockKey is the table key for a share-relative path.
func oplockKey(path string) string {
	return strings.ToUpper(path)
}

// Grant records a level II oplock on a handle.
//
// Parameters:
//   - open: the handle asking, already registered on its connection
//   - conn: the connection the handle belongs to
//   - uid: the session the handle was opened under
//
// Returns:
//   - Whether the oplock was recorded
func (t *oplockTable) Grant(open *Open, conn *Connection, uid uint16) bool {
	if open == nil || conn == nil || open.Tree == nil {
		return false
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	key := oplockKey(open.Path)
	t.granted[key] = append(t.granted[key], &oplockHolder{
		conn:  conn,
		fid:   open.FID,
		uid:   uid,
		tid:   open.Tree.TID,
		owner: open,
	})
	return true
}

// Break removes every oplock held on a path by a handle other than except, and
// returns the holders it removed so the caller can notify them.
//
// The notification is deliberately not sent here. Sending it takes the target
// connection's write lock, and doing that while holding this table's mutex would
// have a slow client on one connection stall every other connection's opens.
//
// Parameters:
//   - path: the share-relative path whose holders are losing their oplock
//   - except: a handle to leave alone, or nil to break every holder
//
// Returns:
//   - The holders that lost their oplock
func (t *oplockTable) Break(path string, except *Open) []*oplockHolder {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	key := oplockKey(path)
	held := t.granted[key]
	if len(held) == 0 {
		return nil
	}

	broken := make([]*oplockHolder, 0, len(held))
	kept := held[:0]
	for _, holder := range held {
		if except != nil && holder.owner == except {
			kept = append(kept, holder)
			continue
		}
		broken = append(broken, holder)
	}

	if len(kept) == 0 {
		delete(t.granted, key)
	} else {
		t.granted[key] = kept
	}
	return broken
}

// Release drops the oplock a handle holds, and reports whether it held one.
//
// It is what both a close and a client's break acknowledgement come to: [MS-CIFS]
// section 3.3.5.30 has the acknowledgement "set Server.Open.Oplock to NONE", and
// a handle that no longer exists cannot hold one either.
func (t *oplockTable) Release(open *Open) bool {
	if open == nil {
		return false
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	key := oplockKey(open.Path)
	held := t.granted[key]
	kept := held[:0]
	released := false
	for _, holder := range held {
		if holder.owner == open {
			released = true
			continue
		}
		kept = append(kept, holder)
	}

	if len(kept) == 0 {
		delete(t.granted, key)
	} else {
		t.granted[key] = kept
	}
	return released
}

// HeldOn reports how many oplocks are held on a path.
func (t *oplockTable) HeldOn(path string) int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return len(t.granted[oplockKey(path)])
}

// Holds reports whether a handle holds an oplock.
func (t *oplockTable) Holds(open *Open) bool {
	if open == nil {
		return false
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, holder := range t.granted[oplockKey(open.Path)] {
		if holder.owner == open {
			return true
		}
	}
	return false
}

// grantOplock decides what oplock an open gets and records it.
//
// A request for an exclusive or batch oplock is downgraded and answered with
// level II, which [MS-SMB] section 3.3.5.1.2 sanctions: it is what a share with
// SHI1005_FLAGS_FORCE_LEVELII_OPLOCK does. It is also the only honest answer
// here, because an exclusive oplock lets a client cache writes, and breaking one
// means holding the operation that caused the break until the client has flushed
// and acknowledged — which needs a request this server can suspend and resume,
// and it has none.
//
// [MS-CIFS] section 3.3.5.2.7 would additionally refuse an oplock on a file
// another handle has open for writing. This grants it and withdraws it at the
// point the writer actually changes the file instead, which keeps the same
// promise to the client — it is told before any change of its cached data can be
// seen — without a snapshot of every connection's open table, which belongs to
// those connections' own goroutines and could not be read here without a race.
//
// Parameters:
//   - conn: the connection the open belongs to
//   - open: the handle just registered
//   - uid: the session the handle was opened under
//   - createFlags: the request's Flags field
//
// Returns:
//   - The level to report in the response
func (c *Connection) grantOplock(open *Open, uid uint16, createFlags uint32) uint8 {
	if createFlags&(ntCreateRequestOplock|ntCreateRequestOpBatch) == 0 {
		// Nothing was asked for. [MS-SMB] section 3.3.5.1.2 notes that Windows
		// grants level II anyway; this does not, because a client that did not
		// ask has not agreed to answer a break.
		return OplockLevelNone
	}
	if !c.oplocksSupported() {
		return OplockLevelNone
	}
	if open.IsDirectory {
		// "If the open or create is on a directory file, then an Oplock MUST NOT
		// be granted" ([MS-CIFS] section 3.3.5.2.7).
		return OplockLevelNone
	}
	if open.Tree == nil || open.Tree.Share == nil || open.Tree.Share.oplocks == nil {
		return OplockLevelNone
	}

	if !open.Tree.Share.oplocks.Grant(open, c, uid) {
		return OplockLevelNone
	}

	logger.Debugf("SMB1 server: %s granted a level II oplock on %q (FID 0x%04X)",
		c.Remote, open.Path, open.FID)
	return OplockLevelLevelII
}

// oplocksSupported reports whether this connection may be granted an oplock.
//
// [MS-CIFS] section 3.3.5.53 requires it to be false when the client's
// MaxMpxCount is below two, and gives the reason: "a client attempting to break
// its own OpLock would always time out because there would not be enough
// outstanding command slots to properly revoke the OpLock". A client that can
// have only one command in flight cannot answer a break while it is waiting for
// anything else.
func (c *Connection) oplocksSupported() bool {
	if !c.Negotiated || c.ClientMaxMpxCount < 2 {
		return false
	}
	return serverCapabilities&capabilities.CAP_LEVEL_II_OPLOCKS != 0
}

// breakOplocksOn breaks every level II oplock on a path except one handle's,
// because the file is about to change or has just been opened for writing.
//
// A level II oplock is a promise that there are no writers, so it has to be
// withdrawn before anything makes that false. The withdrawal is immediate: the
// table stops counting the holder before this returns, so nothing that follows
// can grant or keep an oplock on the strength of it.
//
// The notification is handed to a goroutine of its own rather than sent here.
// Writing to another connection blocks until that client reads, and a client
// idling with a full receive buffer would otherwise stall the unrelated client
// whose write broke the oplock — one client's slowness becoming another's. That
// is safe precisely because a level II break needs no acknowledgement: nothing
// waits on it, and the holder has no cached writes to flush before the change
// goes ahead. It does mean the break is in flight while the change is applied
// rather than provably ahead of it, which is the guarantee level II carries in
// any implementation.
//
// Parameters:
//   - share: the share the path belongs to
//   - path: the share-relative path about to change
//   - except: the handle causing the change, which keeps whatever it holds
func (c *Connection) breakOplocksOn(share *Share, path string, except *Open) {
	if share == nil || share.oplocks == nil {
		return
	}

	for _, holder := range share.oplocks.Break(path, except) {
		go func(holder *oplockHolder) {
			if err := holder.notifyBreak(); err != nil {
				// A break that cannot be delivered has still been withdrawn:
				// the holder is out of the table. Its client keeps a cache it
				// should have dropped, so this is worth a line.
				logger.Debugf("SMB1 server: could not tell %s its oplock on %q was broken: %v",
					holder.conn.Remote, path, err)
			}
		}(holder)
	}
}

// notifyBreak sends the oplock break notification for one holder.
//
// It is an SMB_COM_LOCKING_ANDX request rather than a response — the one message
// the server originates — carrying OPLOCK_RELEASE and a NewOpLockLevel of zero.
// [MS-CIFS] section 3.3.4.2 has the other fields zero with it: "the server
// SHOULD set the Timeout, NumberOfUnlocks, NumberofLocks, and ByteCount fields to
// zero".
func (h *oplockHolder) notifyBreak() error {
	notification := commands.NewLockingAndxRequest()
	notification.FID = types.USHORT(h.fid)
	notification.TypeOfLock = types.UCHAR(commands.LockingAndxOplockRelease)
	notification.NewOpLockLevel = types.UCHAR(breakToNone)
	notification.Timeout = types.ULONG(0)

	return h.conn.SendUnsolicited(notification, h.uid, h.tid, h.conn.nextUnsolicitedMID())
}

// releaseOplock applies a client's OpLock Break Request: the acknowledgement it
// sends after being told its oplock was broken.
//
// [MS-CIFS] section 3.3.5.30 requires the server to "release the OpLock on the
// Open" and to set its state to NONE, and is equally clear that a release naming
// a handle that holds none is not an error: "If there are no outstanding OpLock
// breaks, or if the FID in the request does not match the FID of an outstanding
// OpLock Break Notification, then no OpLock is released. This does not generate
// an error."
//
// Nothing is waiting on the acknowledgement here. A level II oplock is withdrawn
// the moment the break is sent, because the client has no cached writes to flush
// first, so by the time this arrives the server has already stopped counting the
// handle as a holder — and the operation that broke it has already gone ahead.
//
// Parameters:
//   - open: the handle the client is releasing
func (c *Connection) releaseOplock(open *Open) {
	if open == nil || open.Tree == nil || open.Tree.Share == nil || open.Tree.Share.oplocks == nil {
		return
	}

	released := open.Tree.Share.oplocks.Release(open)

	logger.Debugf("SMB1 server: %s released the oplock on FID 0x%04X (%q), which it held=%t",
		c.Remote, open.FID, open.Path, released)
}
