package server

import (
	"bytes"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/createcontext"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
)

// Only level II oplocks and read leases are granted: they let a client cache
// reads, so withdrawing one needs no acknowledgement and never stalls the change
// that caused it. Write and handle caching would have to park that change until
// the client flushed and acknowledged.

// oplockTable holds the oplocks and leases granted on every file, across
// connections, so a change through one handle can break the others.
type oplockTable struct {
	mutex  sync.Mutex
	grants map[string]map[*Open]*oplockGrant
}

// oplockGrant is a level II oplock, or a read lease when lease is set.
type oplockGrant struct {
	conn     *Connection
	session  *Session
	fileID   uint64
	lease    bool
	leaseKey [16]byte
	epoch    uint16
	v2       bool
}

func newOplockTable() *oplockTable {
	return &oplockTable{grants: make(map[string]map[*Open]*oplockGrant)}
}

// grant answers a CREATE's oplock or lease request, returning the level and the
// create contexts that carry it.
func (t *oplockTable) grant(c *Connection, open *Open, req *commands.CreateRequest) (uint8, []byte) {
	if open.isDir || req.RequestedOplockLevel == commands.SMB2_OPLOCK_LEVEL_NONE {
		return commands.SMB2_OPLOCK_LEVEL_NONE, nil
	}

	g := &oplockGrant{conn: c, session: open.session, fileID: open.id}
	level := uint8(commands.SMB2_OPLOCK_LEVEL_II)
	var contexts []byte
	if req.RequestedOplockLevel == commands.SMB2_OPLOCK_LEVEL_LEASE {
		lease, ok := requestedLease(c.dialect, req.CreateContexts)
		if !ok {
			return commands.SMB2_OPLOCK_LEVEL_NONE, nil
		}
		g.lease, g.leaseKey, g.v2, g.epoch = true, lease.LeaseKey, lease.v2, lease.Epoch+1
		answer := createcontext.MarshalLeaseV1(lease.LeaseKey, createcontext.SMB2_LEASE_READ_CACHING)
		if lease.v2 {
			answer = createcontext.MarshalLeaseV2(lease.LeaseKey, createcontext.SMB2_LEASE_READ_CACHING, lease.ParentLeaseKey, g.epoch)
		}
		contexts, _ = createcontext.Marshal([]createcontext.CreateContext{answer})
		level = commands.SMB2_OPLOCK_LEVEL_LEASE
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	key := fileKey(open.tree.Share, open.path)
	if t.grants[key] == nil {
		t.grants[key] = make(map[*Open]*oplockGrant)
	}
	t.grants[key][open] = g
	return level, contexts
}

// leaseRequest is a lease request asking for read caching.
type leaseRequest struct {
	createcontext.LeaseV2
	v2 bool
}

// requestedLease finds a lease request asking for read caching. Leases need 2.1,
// and the version 2 form needs 3.x.
func requestedLease(dialect dialects.Dialect, raw []byte) (leaseRequest, bool) {
	if dialectRank[dialect] < dialectRank[dialects.SMB2_DIALECT_2_1_0] {
		return leaseRequest{}, false
	}
	contexts, err := createcontext.Parse(raw)
	if err != nil {
		return leaseRequest{}, false
	}
	for _, ctx := range contexts {
		if !bytes.Equal(ctx.Name, createcontext.NameRequestLease) {
			continue
		}
		var lease leaseRequest
		if v2, err := createcontext.ParseLeaseV2(ctx.Data); err == nil && security.IsSMB3(dialect) {
			lease = leaseRequest{LeaseV2: *v2, v2: true}
		} else if v1, err := createcontext.ParseLeaseV1(ctx.Data); err == nil {
			lease.LeaseKey, lease.LeaseState = v1.LeaseKey, v1.LeaseState
		} else {
			return leaseRequest{}, false
		}
		return lease, lease.LeaseState&createcontext.SMB2_LEASE_READ_CACHING != 0
	}
	return leaseRequest{}, false
}

// breakOthers withdraws every grant on the file except those of the acting
// open and of the lease it belongs to, notifying their holders.
func (t *oplockTable) breakOthers(actor *Open) {
	t.mutex.Lock()
	key := fileKey(actor.tree.Share, actor.path)
	own := t.grants[key][actor]
	var broken []*oplockGrant
	for open, g := range t.grants[key] {
		sameLease := own != nil && own.lease && g.lease && g.leaseKey == own.leaseKey && g.conn.clientGUID == own.conn.clientGUID
		if open == actor || sameLease {
			continue
		}
		delete(t.grants[key], open)
		broken = append(broken, g)
	}
	t.mutex.Unlock()

	notified := make(map[[16]byte]bool)
	for _, g := range broken {
		if g.lease {
			if notified[g.leaseKey] {
				continue
			}
			notified[g.leaseKey] = true
		}
		// Sent apart, so a client not reading its socket cannot stall the actor.
		go g.notify()
	}
}

// release forgets an open's grant.
func (t *oplockTable) release(open *Open) {
	if open.tree.Share == nil {
		return
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	key := fileKey(open.tree.Share, open.path)
	delete(t.grants[key], open)
	if len(t.grants[key]) == 0 {
		delete(t.grants, key)
	}
}

// rename moves an open's grant to the file's new name.
func (t *oplockTable) rename(open *Open, target string) {
	t.mutex.Lock()
	g := t.grants[fileKey(open.tree.Share, open.path)][open]
	t.mutex.Unlock()
	t.release(open)
	if g == nil {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	key := fileKey(open.tree.Share, target)
	if t.grants[key] == nil {
		t.grants[key] = make(map[*Open]*oplockGrant)
	}
	t.grants[key][open] = g
}

// notify sends the server-initiated break: an oplock or lease falls to none.
func (g *oplockGrant) notify() {
	var body command_interface.CommandInterface
	if g.lease {
		notification := commands.NewLeaseBreakNotification()
		notification.LeaseKey = g.leaseKey
		notification.CurrentLeaseState = createcontext.SMB2_LEASE_READ_CACHING
		notification.NewLeaseState = createcontext.SMB2_LEASE_NONE
		if g.v2 {
			notification.NewEpoch = g.epoch + 1
		}
		body = notification
	} else {
		notification := commands.NewOplockBreakRequest()
		notification.OplockLevel = commands.SMB2_OPLOCK_LEVEL_NONE
		notification.FileId.Persistent, notification.FileId.Volatile = g.fileID, g.fileID
		body = notification
	}

	// An unsolicited message carries MessageId 0xFFFFFFFFFFFFFFFF and no
	// session, and is not signed; it is encrypted when the holder's session is.
	m := message.NewMessage()
	m.Header.MessageId = ^uint64(0)
	m.Header.Flags = flags.SMB2_FLAGS_SERVER_TO_REDIR
	m.SetCommand(body)
	frame, err := m.Marshal()
	if err == nil && g.session.encryptData {
		frame, err = security.Encrypt(g.conn.cipher, g.session.encryptionKey, g.session.id, g.session.nonce.Add(1), frame)
	}
	if err == nil {
		err = g.conn.send(frame)
	}
	if err != nil {
		logger.Debugf("SMB2 server: failed to send a break to %s: %v", g.conn.remote, err)
	}
}
