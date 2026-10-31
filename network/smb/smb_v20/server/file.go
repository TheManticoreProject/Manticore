package server

import (
	"errors"
	"io"

	"github.com/TheManticoreProject/Manticore/logger"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// maximumAllowed asks for whatever access the share grants.
const maximumAllowed uint32 = 0x02000000

// CREATE actions reported in the response.
const (
	fileOpened      = 1
	fileCreated     = 2
	fileOverwritten = 3
)

// Open is a handle on a file or directory.
type Open struct {
	id            uint64
	session       *Session
	tree          *Tree
	path          string
	file          smb1.File
	isDir         bool
	access        uint32 // granted rights
	shareAccess   uint32 // FILE_SHARE_* the open allows others
	deleteOnClose bool
	search        *search
	notifies      []*pendingNotify
}

// fs returns the backend the open lives on.
func (o *Open) fs() smb1.FileSystem { return o.tree.Share.FS }

// openFor resolves the FileId of a request to one of the tree's handles. A
// related compounded request names the previous request's handle.
func (c *Connection) openFor(r *request, id types.SMB2_FILEID) (*Open, nt_status.NT_STATUS) {
	if id == relatedFileID {
		id = r.fileID
	}
	open := c.opens[id.Volatile]
	if open == nil || open.tree != r.tree || open.session != r.session || id.Persistent != id.Volatile {
		return nil, nt_status.NT_STATUS_FILE_CLOSED
	}
	r.fileID = id
	return open, nt_status.NT_STATUS_SUCCESS
}

// create answers CREATE (MS-SMB2 3.3.5.9).
func (c *Connection) create(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.CreateRequest)
	share := r.tree.Share
	if share == nil {
		return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
	}
	if len(c.opens) >= c.server.config.MaxOpensPerConnection {
		return nt_status.NT_STATUS_INSUFFICIENT_RESOURCES
	}
	path, err := smb1.ResolvePath(req.Name)
	if err != nil {
		return nt_status.NT_STATUS_OBJECT_NAME_INVALID
	}

	access, status := grantedAccess(req.DesiredAccess, share.ReadOnly)
	if !status.IsSuccess() {
		return status
	}
	flags, status := smb1.OpenFlagsFor(access, req.CreateDisposition, req.CreateOptions)
	if !status.IsSuccess() {
		return status
	}
	deleteOnClose := req.CreateOptions&fileflags.FILE_DELETE_ON_CLOSE != 0
	if deleteOnClose && access&fileflags.DELETE == 0 || share.ReadOnly && (flags.Create || flags.Truncate) {
		return nt_status.NT_STATUS_ACCESS_DENIED
	}

	open := &Open{
		session: r.session, tree: r.tree, path: path,
		access: access, shareAccess: req.ShareAccess, deleteOnClose: deleteOnClose,
	}
	// Checked before the backend opens anything, so a refused open cannot truncate.
	if !c.server.shareModes.acquire(open) {
		return nt_status.NT_STATUS_SHARING_VIOLATION
	}
	_, statErr := share.FS.Stat(path)
	existed := statErr == nil
	file, err := share.FS.Open(path, flags)
	if err != nil {
		c.server.shareModes.release(open)
		return smb1.StatusForFSError(err)
	}
	attr, err := share.FS.Stat(path)
	if err != nil {
		file.Close()
		c.server.shareModes.release(open)
		return smb1.StatusForFSError(err)
	}
	// A file carrying the read-only attribute cannot be opened delete-on-close,
	// [MS-FSA] section 2.1.5.1.2.1.
	if deleteOnClose && !attr.IsDir && attr.ReadOnly {
		file.Close()
		c.server.shareModes.release(open)
		return smb1.StatusForFSError(smb1.ErrCannotDelete)
	}

	c.nextFileID++
	open.id, open.file, open.isDir = c.nextFileID, file, attr.IsDir
	c.opens[open.id] = open
	r.fileID = open.fileID()

	if flags.Write || flags.Truncate || deleteOnClose {
		c.server.oplocks.breakOthers(open)
	}
	answer := commands.NewCreateResponse()
	answer.OplockLevel, answer.CreateContexts = c.server.oplocks.grant(c, open, req)
	answer.CreateAction = fileOpened
	switch {
	case !existed:
		answer.CreateAction = fileCreated
	case flags.Truncate:
		answer.CreateAction = fileOverwritten
	}
	answer.CreationTime, answer.LastAccessTime = filetime(attr.Created), filetime(attr.Accessed)
	answer.LastWriteTime, answer.ChangeTime = filetime(attr.Modified), filetime(attr.Changed)
	answer.AllocationSize, answer.EndOfFile = uint64(attr.AllocationSize), uint64(attr.Size)
	answer.FileAttributes = attributesOf(attr)
	answer.FileId = open.fileID()
	resp.command = answer
	logger.Debugf("SMB2 server: %s opened %q on %q", c.remote, path, share.Name)
	return nt_status.NT_STATUS_SUCCESS
}

// fileID is the SMB2_FILEID naming the open on the wire.
func (o *Open) fileID() types.SMB2_FILEID {
	return types.SMB2_FILEID{Persistent: o.id, Volatile: o.id}
}

// close answers CLOSE.
func (c *Connection) close(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.CloseRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}

	answer := commands.NewCloseResponse()
	if req.Flags&commands.SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB != 0 {
		if attr, err := open.fs().Stat(open.path); err == nil {
			answer.Flags = commands.SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB
			answer.CreationTime, answer.LastAccessTime = filetime(attr.Created), filetime(attr.Accessed)
			answer.LastWriteTime, answer.ChangeTime = filetime(attr.Modified), filetime(attr.Changed)
			answer.AllocationSize, answer.EndOfFile = uint64(attr.AllocationSize), uint64(attr.Size)
			answer.FileAttributes = attributesOf(attr)
		}
	}
	if err := c.closeOpen(open); err != nil {
		return smb1.StatusForFSError(err)
	}
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}

// closeOpen releases a handle, applying a pending delete.
func (c *Connection) closeOpen(open *Open) error {
	delete(c.opens, open.id)
	c.cleanupNotifies(open)
	c.server.oplocks.release(open)
	c.server.shareModes.release(open)
	open.file.Close()
	if !open.deleteOnClose {
		return nil
	}
	if open.isDir {
		return open.fs().Rmdir(open.path)
	}
	return open.fs().Remove(open.path)
}

// read answers READ.
func (c *Connection) read(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.ReadRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if open.isDir {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if status := open.require(readDataRights); !status.IsSuccess() {
		return status
	}
	// A multi-credit connection pays one credit per 64 KiB read (MS-SMB2
	// 3.3.5.2.5); one without LARGE_MTU was offered 64 KiB at most.
	limit := c.server.config.MaxReadSize
	if c.capabilities.Has(capabilities.SMB2_GLOBAL_CAP_LARGE_MTU) {
		if uint64(max(r.header.CreditCharge, 1))<<16 < uint64(req.Length) {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
	} else {
		limit = min(limit, 1<<16)
	}
	if req.Length > limit {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	data := make([]byte, req.Length)
	n, err := open.file.ReadAt(data, int64(req.Offset))
	if err != nil && !errors.Is(err, io.EOF) {
		return smb1.StatusForFSError(err)
	}
	if n == 0 && req.Length > 0 || uint32(n) < req.MinimumCount {
		return nt_status.NT_STATUS_END_OF_FILE
	}

	answer := commands.NewReadResponse()
	answer.Data = data[:n]
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}

// write answers WRITE.
func (c *Connection) write(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.WriteRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if status := open.require(writeDataRights); !status.IsSuccess() {
		return status
	}
	if open.isDir {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	if uint32(len(req.Data)) > c.server.config.MaxWriteSize {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	c.server.oplocks.breakOthers(open)
	n, err := open.file.WriteAt(req.Data, int64(req.Offset))
	if err != nil {
		return smb1.StatusForFSError(err)
	}
	answer := commands.NewWriteResponse()
	answer.Count = uint32(n)
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}

// flush answers FLUSH.
func (c *Connection) flush(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.FlushRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if err := open.file.Sync(); err != nil {
		return smb1.StatusForFSError(err)
	}
	resp.command = commands.NewFlushResponse()
	return nt_status.NT_STATUS_SUCCESS
}

// ioctl answers IOCTL. Only the secure negotiate validation is served.
func (c *Connection) ioctl(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.IoctlRequest)
	switch req.CtlCode {
	case filesystem.FSCTL_VALIDATE_NEGOTIATE_INFO:
	case filesystem.FSCTL_DFS_GET_REFERRALS, filesystem.FSCTL_DFS_GET_REFERRALS_EX:
		// MS-SMB2 3.3.5.15.2: what a server without DFS answers.
		return nt_status.NT_STATUS_FS_DRIVER_REQUIRED
	default:
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}
	output, ok := c.validateNegotiateInfo(req.Input)
	if !ok {
		// MS-SMB2 3.3.5.15.12: the connection is terminated.
		c.terminate = true
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	answer := commands.NewIoctlResponse()
	answer.CtlCode = req.CtlCode
	answer.FileId = req.FileId
	answer.Output = output
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}
