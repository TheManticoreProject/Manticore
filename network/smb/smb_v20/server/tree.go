package server

import (
	"strings"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// ipcShareName is the inter-process share clients connect to before anything
// else. It is always present; opening a pipe on it is refused.
const ipcShareName = "IPC$"

// Maximal access reported for a tree: everything, or only reading.
const (
	fileAllAccess  uint32 = 0x001F01FF
	fileReadAccess uint32 = 0x001200A9
)

// Tree is a tree connect: a session's attachment to a share. A nil Share is IPC$.
type Tree struct {
	id    uint32
	Share *Share
}

// treeConnect answers TREE_CONNECT (MS-SMB2 3.3.5.7).
func (c *Connection) treeConnect(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.TreeConnectRequest)
	if len(r.session.trees) >= c.server.config.MaxTreesPerSession {
		return nt_status.NT_STATUS_INSUFFICIENT_RESOURCES
	}

	// The path is \\server\share; only the share name matters.
	name := req.Path[strings.LastIndexByte(req.Path, '\\')+1:]
	tree := &Tree{}
	answer := commands.NewTreeConnectResponse()
	switch share := c.server.share(name); {
	case strings.EqualFold(name, ipcShareName):
		answer.ShareType = commands.SMB2_SHARE_TYPE_PIPE
		answer.MaximalAccess = fileReadAccess
	case share != nil:
		tree.Share = share
		answer.ShareType = commands.SMB2_SHARE_TYPE_DISK
		answer.MaximalAccess = fileAllAccess
		if share.ReadOnly {
			answer.MaximalAccess = fileReadAccess
		}
	default:
		return nt_status.NT_STATUS_BAD_NETWORK_NAME
	}
	if r.session.encryptData {
		answer.ShareFlags = commands.SMB2_SHAREFLAG_ENCRYPT_DATA
	}

	c.nextTreeID++
	tree.id = c.nextTreeID
	r.session.trees[tree.id] = tree
	r.header.TreeId = tree.id
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}

// treeDisconnect answers TREE_DISCONNECT.
func (c *Connection) treeDisconnect(r *request, resp *response) nt_status.NT_STATUS {
	c.disconnectTree(r.session, r.tree)
	resp.command = commands.NewTreeDisconnectResponse()
	return nt_status.NT_STATUS_SUCCESS
}

// disconnectTree closes the handles opened through a tree and forgets it.
func (c *Connection) disconnectTree(session *Session, tree *Tree) {
	for _, open := range c.opens {
		if open.tree == tree {
			c.closeOpen(open)
		}
	}
	delete(session.trees, tree.id)
}
