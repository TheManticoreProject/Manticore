package server

import (
	"strings"
	"time"
)

// Tree is a connection to a share, established by a tree connect and named by a
// TID on every subsequent request.
type Tree struct {
	// TID is the identifier the client sends to act on this tree.
	TID uint16

	// Share is what the tree is connected to.
	Share *Share

	// Session is the session the tree belongs to. A tree is scoped to the session
	// that opened it, so another session's TID cannot be borrowed.
	SessionUID uint16

	// Created is when the tree was connected.
	Created time.Time
}

// Open is a file handle, established by a create or open and named by a FID.
type Open struct {
	// FID is the identifier the client sends to act on this handle.
	FID uint16

	// Tree is the tree the handle was opened on, and Path the share-relative path
	// it names.
	Tree *Tree
	Path string

	// File is the backend handle. It is nil for a handle onto a directory that
	// the backend declined to open, which is legitimate: a directory handle is
	// only ever used to query.
	File File

	// IsDirectory records what the handle names.
	IsDirectory bool

	// IsPipe records that the handle names a named pipe rather than a file, so
	// Path is a pipe name and there is no backend file behind it. A pipe handle
	// is what a transaction acts on: [MS-CIFS] section 3.3.5.57.7 identifies the
	// pipe by the FID in the request's setup words, not by the name it carries.
	IsPipe bool

	// Pipe is the handler's session for this open, and is what a transaction on
	// the handle runs against. It is the pipe counterpart of File.
	//
	// The session rather than the pipe's name is what the handle carries, because
	// state a client builds up on a pipe belongs to its open of it: an RPC bind
	// establishes the presentation contexts and fragment size the requests after
	// it are read against, and two clients of one pipe bind separately.
	Pipe PipeSession

	// Readable and Writable are the access the open was granted, enforced on
	// every use so a handle opened for reading cannot later be written through.
	Readable bool
	Writable bool

	// pipeOutput is the part of a pipe answer that has not been delivered yet.
	//
	// A transaction returns only as much as the client's buffer takes, and the
	// rest has to live somewhere until the client reads it: the client is told
	// more remains, and reading again is how it collects it. The handle is the
	// right owner because [MS-CIFS] section 3.3.5.57.9 identifies the pipe a read
	// acts on by its FID, and because closing the handle is what makes an
	// undrained answer collectable.
	pipeOutput []byte

	// DeleteOnClose removes the file when the handle closes, which is how a
	// client deletes something it holds open.
	DeleteOnClose bool

	// PID is the process identifier that opened the handle.
	//
	// It is recorded so SMB_COM_PROCESS_EXIT can release what a client process
	// still held when it died: [MS-CIFS] section 2.2.4.18 has the server close
	// "any resources owned by the Process ID (PID) listed in the request header",
	// and without this the command would have nothing to select on.
	PID uint32

	// Position is the handle's file pointer, which SMB_COM_SEEK moves.
	//
	// Nothing else consults it, because every read and write this server serves
	// carries its own offset. It is kept because seeking from the current
	// position is only meaningful if the position persists between calls.
	Position int64

	// Created is when the handle was opened.
	Created time.Time
}

// Tree returns the tree a TID names on this connection, or nil when it names
// none.
func (c *Connection) Tree(tid uint16) *Tree {
	return c.trees[tid]
}

// Open returns the handle a FID names on this connection, or nil when it names
// none.
func (c *Connection) Open(fid uint16) *Open {
	return c.opens[fid]
}

// addTree records a connected tree.
// OplockLevel reports the oplock this handle holds, as an OplockLevel* constant.
//
// It asks the share's table rather than caching the answer on the handle. An
// oplock is broken by whichever goroutine changes the file, which is not this
// handle's connection, so a copy kept here would be written from two goroutines
// at once — and the table is guarded and is the thing that decides who holds one.
//
// Returns:
//   - The level held, or OplockLevelNone
func (o *Open) OplockLevel() uint8 {
	if o == nil || o.Tree == nil || o.Tree.Share == nil || o.Tree.Share.oplocks == nil {
		return OplockLevelNone
	}
	if o.Tree.Share.oplocks.Holds(o) {
		return OplockLevelLevelII
	}
	return OplockLevelNone
}

// pipeSession returns the handler session this handle transacts on, which is nil
// for a handle that does not name a pipe.
func (o *Open) pipeSession() PipeSession {
	if o == nil {
		return nil
	}
	return o.Pipe
}

// drainPipeOutput removes up to limit bytes of the answer buffered on a pipe
// handle and reports whether any is left after it.
//
// The bytes are returned with the capacity clipped, so a caller that appends to
// the returned slice cannot write into what is still buffered.
//
// Parameters:
//   - limit: the most bytes to remove
//
// Returns:
//   - The bytes removed, and whether more remain buffered
func (o *Open) drainPipeOutput(limit int) ([]byte, bool) {
	if limit < 0 {
		limit = 0
	}
	if limit > len(o.pipeOutput) {
		limit = len(o.pipeOutput)
	}

	chunk := o.pipeOutput[:limit:limit]
	o.pipeOutput = o.pipeOutput[limit:]
	return chunk, len(o.pipeOutput) > 0
}

// peekPipeOutput returns up to limit bytes of the answer buffered on a pipe
// handle without removing any of it, along with how much is buffered in total.
//
// Parameters:
//   - limit: the most bytes to return
//
// Returns:
//   - The bytes at the front of the buffer, and the total buffered length
func (o *Open) peekPipeOutput(limit int) ([]byte, int) {
	if limit < 0 {
		limit = 0
	}
	if limit > len(o.pipeOutput) {
		limit = len(o.pipeOutput)
	}
	return o.pipeOutput[:limit:limit], len(o.pipeOutput)
}

func (c *Connection) addTree(tree *Tree) {
	c.trees[tree.TID] = tree
}

// removeTree drops a tree, closes every handle and enumeration opened on it, and
// releases its identifier.
//
// Closing them matters: a client that disconnects a tree without closing its
// files expects them released, and anything left behind would hold the backend's
// resources with nothing able to reach it.
func (c *Connection) removeTree(tid uint16) *Tree {
	tree, ok := c.trees[tid]
	if !ok {
		return nil
	}

	for fid, open := range c.opens {
		if open.Tree == tree {
			c.closeOpen(fid)
		}
	}

	// The enumerations opened on the tree go with it, for the same reason its
	// handles do: a search names a directory on this tree, so once the tree is
	// gone there is nothing for it to continue against. Leaving them behind held
	// their snapshots for the life of the connection and kept identifiers
	// allocated that no client could still use.
	for sid, search := range c.searches {
		if search.Tree == tree {
			c.closeSearch(sid)
		}
	}

	delete(c.trees, tid)
	c.tids.Release(tid)
	return tree
}

// addOpen records an open handle.
func (c *Connection) addOpen(open *Open) {
	c.opens[open.FID] = open
}

// closeOpen closes a handle, applies a pending delete-on-close, and releases the
// identifier.
//
// Returns:
//   - The error from closing the backend handle or from the pending delete, or
//     nil
func (c *Connection) closeOpen(fid uint16) error {
	open, ok := c.opens[fid]
	if !ok {
		return nil
	}
	delete(c.opens, fid)
	c.fids.Release(fid)

	// Locks go first, and unconditionally. [MS-CIFS] section 2.2.4.32.1: "Closing
	// a file with locks still in force causes the locks to be released". A lock
	// left behind here would be owned by a handle that no longer exists, so
	// nothing could ever release it and the range would stay locked for the life
	// of the share.
	if open.Tree != nil && open.Tree.Share != nil && open.Tree.Share.locks != nil {
		open.Tree.Share.locks.ReleaseAll(open)
	}

	// An oplock goes the same way, and for the same reason: it is a promise made
	// to a handle, and a handle that no longer exists cannot be told the promise
	// has been withdrawn.
	if open.Tree != nil && open.Tree.Share != nil && open.Tree.Share.oplocks != nil {
		open.Tree.Share.oplocks.Release(open)
	}

	var firstErr error
	if open.File != nil {
		if err := open.File.Close(); err != nil {
			firstErr = err
		}
	}

	// A pipe handle has no backend file; what it holds is the handler's session,
	// so closing it is the handler's business.
	if open.Pipe != nil {
		if err := open.Pipe.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if open.DeleteOnClose && open.Tree != nil && open.Tree.Share.FS != nil {
		var err error
		if open.IsDirectory {
			err = open.Tree.Share.FS.Rmdir(open.Path)
		} else {
			err = open.Tree.Share.FS.Remove(open.Path)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// closeSessionResources drops every tree, and with it every handle, belonging to
// a session. A logoff releases what the session held rather than leaving it
// reachable through a UID that no longer names anything.
func (c *Connection) closeSessionResources(uid uint16) {
	for tid, tree := range c.trees {
		if tree.SessionUID == uid {
			c.removeTree(tid)
		}
	}
}

// shareNameFromPath extracts the share name from the UNC path a tree connect
// carries.
//
// The path is "\\server\share", and the server component is ignored: a client
// reaches this server by connecting to it, so the name it used to get here says
// nothing about which share it wants.
//
// Parameters:
//   - path: the UNC path from the request
//
// Returns:
//   - The share name, or "" when the path is not a UNC path
func shareNameFromPath(path string) string {
	normalised := strings.ReplaceAll(path, "/", "\\")
	normalised = strings.TrimRight(normalised, "\x00")

	if !strings.HasPrefix(normalised, "\\\\") {
		return ""
	}
	rest := normalised[2:]

	// Skip the server component.
	slash := strings.IndexByte(rest, '\\')
	if slash < 0 {
		return ""
	}
	share := rest[slash+1:]

	// A share name is one element; anything after it is not part of the name.
	if slash := strings.IndexByte(share, '\\'); slash >= 0 {
		share = share[:slash]
	}
	return strings.TrimRight(share, "\x00")
}
