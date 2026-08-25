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

	// Readable and Writable are the access the open was granted, enforced on
	// every use so a handle opened for reading cannot later be written through.
	Readable bool
	Writable bool

	// DeleteOnClose removes the file when the handle closes, which is how a
	// client deletes something it holds open.
	DeleteOnClose bool

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
func (c *Connection) addTree(tree *Tree) {
	c.trees[tree.TID] = tree
}

// removeTree drops a tree, closes every handle opened on it, and releases its
// identifier.
//
// Closing the handles matters: a client that disconnects a tree without closing
// its files expects them released, and a handle left behind would hold the
// backend's resources with nothing able to reach it.
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

	var firstErr error
	if open.File != nil {
		if err := open.File.Close(); err != nil {
			firstErr = err
		}
	}

	// A pipe handle has no backend file; what it holds is whatever the handler
	// prepared when the pipe was opened, so closing it is the handler's business.
	if open.IsPipe && open.Tree != nil && open.Tree.Share.Pipes != nil {
		if err := open.Tree.Share.Pipes.ClosePipe(open.Path); err != nil && firstErr == nil {
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
