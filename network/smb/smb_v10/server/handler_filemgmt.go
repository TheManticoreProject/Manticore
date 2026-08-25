package server

import (
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// writableTreeFor resolves the tree a modifying command acts on, refusing a
// read-only share. Every command below modifies something, so the check belongs
// in one place rather than at the top of each.
func (c *Connection) writableTreeFor(req *message.Message) (*Tree, nt_status.NT_STATUS) {
	tree, status := c.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, status
	}
	if tree.Share.ReadOnly {
		logger.Debugf("SMB1 server: %s attempted to modify read-only share %q", c.Remote, tree.Share.Name)
		return nil, nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}
	return tree, nt_status.NT_STATUS_SUCCESS
}

// handleDelete answers SMB_COM_DELETE. The name may be a wildcard, in which case
// every matching file in the directory is deleted.
//
// A directory is never deleted by this command, whatever the pattern matches:
// SMB_COM_DELETE_DIRECTORY exists for that, and a client that asked to delete
// "*" does not mean to take the directories with it.
func handleDelete(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.DeleteRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	tree, status := conn.writableTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	requested := decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode())
	directory, pattern, err := resolvePathPattern(requested)
	if err != nil {
		logger.Debugf("SMB1 server: %s asked to delete %q, which is refused: %v", conn.Remote, requested, err)
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// No wildcard: one named file.
	if pattern == "" {
		if err := tree.Share.FS.Remove(directory); err != nil {
			logger.Debugf("SMB1 server: deleting %q for %s failed: %v", directory, conn.Remote, err)
			return statusForFSError(err)
		}
		return conn.answer(w, commands.NewDeleteResponse())
	}

	entries, err := tree.Share.FS.ReadDir(directory, pattern)
	if err != nil {
		return statusForFSError(err)
	}

	deleted := 0
	for _, entry := range entries {
		if entry.Attr.IsDir {
			continue
		}
		if err := tree.Share.FS.Remove(joinPath(directory, entry.Attr.Name)); err != nil {
			logger.Debugf("SMB1 server: deleting %q for %s failed: %v",
				joinPath(directory, entry.Attr.Name), conn.Remote, err)
			return statusForFSError(err)
		}
		deleted++
	}
	// A pattern that matched nothing is a failure, not a no-op: the client asked
	// for something to be deleted and nothing was.
	if deleted == 0 {
		return nt_status.NT_STATUS_NO_SUCH_FILE
	}

	return conn.answer(w, commands.NewDeleteResponse())
}

// handleRename answers SMB_COM_RENAME, which moves as well as renames since the
// destination is a full path.
func handleRename(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.RenameRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	tree, status := conn.writableTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	unicode := req.Header.Flags2.IsUnicode()
	oldPath, err := resolvePath(decodeWireString(request.OldFileName.Buffer, unicode))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}
	newPath, err := resolvePath(decodeWireString(request.NewFileName.Buffer, unicode))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}
	if oldPath == "" || newPath == "" {
		// Renaming the share root is not a thing.
		return nt_status.NT_STATUS_ACCESS_DENIED
	}

	// SMB_COM_RENAME does not replace: a destination that exists is a collision.
	// The NT_RENAME variant is what a client uses when it means to overwrite.
	if err := tree.Share.FS.Rename(oldPath, newPath, false); err != nil {
		logger.Debugf("SMB1 server: renaming %q to %q for %s failed: %v", oldPath, newPath, conn.Remote, err)
		return statusForFSError(err)
	}

	return conn.answer(w, commands.NewRenameResponse())
}

// handleCreateDirectory answers SMB_COM_CREATE_DIRECTORY.
func handleCreateDirectory(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.CreateDirectoryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	tree, status := conn.writableTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	path, err := resolvePath(decodeWireString(request.DirectoryName.Buffer, req.Header.Flags2.IsUnicode()))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}
	if path == "" {
		return nt_status.NT_STATUS_OBJECT_NAME_COLLISION
	}

	if err := tree.Share.FS.Mkdir(path); err != nil {
		logger.Debugf("SMB1 server: creating directory %q for %s failed: %v", path, conn.Remote, err)
		return statusForFSError(err)
	}

	return conn.answer(w, commands.NewCreateDirectoryResponse())
}

// handleDeleteDirectory answers SMB_COM_DELETE_DIRECTORY, which removes only an
// empty directory.
func handleDeleteDirectory(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.DeleteDirectoryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	tree, status := conn.writableTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	path, err := resolvePath(decodeWireString(request.DirectoryName.Buffer, req.Header.Flags2.IsUnicode()))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}
	if path == "" {
		// Removing the share root is refused rather than attempted.
		return nt_status.NT_STATUS_ACCESS_DENIED
	}

	if err := tree.Share.FS.Rmdir(path); err != nil {
		logger.Debugf("SMB1 server: removing directory %q for %s failed: %v", path, conn.Remote, err)
		return statusForFSError(err)
	}

	return conn.answer(w, commands.NewDeleteDirectoryResponse())
}

// handleCheckDirectory answers SMB_COM_CHECK_DIRECTORY, which reports whether a
// path names a directory. A client uses it to walk a path before opening
// anything, so it must distinguish "not there" from "there but not a directory".
func handleCheckDirectory(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.CheckDirectoryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	path, err := resolvePath(decodeWireString(request.DirectoryName.Buffer, req.Header.Flags2.IsUnicode()))
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// The share root is always a directory and needs no lookup.
	if path != "" {
		attr, err := tree.Share.FS.Stat(path)
		if err != nil {
			return statusForFSError(err)
		}
		if !attr.IsDir {
			return nt_status.NT_STATUS_NOT_A_DIRECTORY
		}
	}

	return conn.answer(w, commands.NewCheckDirectoryResponse())
}

// answer sends a response that carries nothing beyond its success, which several
// of the commands above do.
func (c *Connection) answer(w ResponseWriter, response command_interface.CommandInterface) nt_status.NT_STATUS {
	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}
