package server

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"

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
	// id is the Volatile half of the FileId on the wire, and the key in
	// c.opens. persistent is the Persistent half, an independent random value:
	// Windows returns two unrelated numbers here, so equal halves counting from
	// one per connection (which is what a plain counter gives) is a fingerprint.
	id            uint64
	persistent    uint64
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

	// pipe is set on a named-pipe open on IPC$, which has no file.
	pipe *pipeOpen

	// renamed is where a rename moved the file after path was resolved. It is
	// written by whichever connection renamed the file, so it is atomic.
	renamed atomic.Pointer[string]
}

// fs returns the backend the open lives on.
func (o *Open) fs() smb1.FileSystem { return o.tree.Share.FS }

// currentPath returns the share-relative path the open's file has now, which
// a rename through any open of it, or of a directory above it, changes.
func (o *Open) currentPath() string {
	if renamed := o.renamed.Load(); renamed != nil {
		return *renamed
	}
	return o.path
}

// openFor resolves the FileId of a request to one of the tree's handles. A
// related compounded request names the previous request's handle.
func (c *Connection) openFor(r *request, id types.SMB2_FILEID) (*Open, nt_status.NT_STATUS) {
	if id == relatedFileID {
		id = r.fileID
	}
	open := c.opens[id.Volatile]
	if open == nil || open.tree != r.tree || open.session != r.session || id.Persistent != open.persistent {
		return nil, nt_status.NT_STATUS_FILE_CLOSED
	}
	// A pipe has no file behind it, so only the commands that move messages
	// through it, and CLOSE, act on a pipe handle.
	if open.pipe != nil && !pipeCommands[r.header.Command] {
		return nil, nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}
	r.fileID = id
	return open, nt_status.NT_STATUS_SUCCESS
}

// create answers CREATE (MS-SMB2 3.3.5.9).
func (c *Connection) create(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.CreateRequest)
	share := r.tree.Share
	if share == nil {
		// IPC$: the name is a named pipe.
		return c.createPipe(r, resp, req)
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
	// MAXIMUM_ALLOWED on a file carrying the read-only attribute is granted
	// everything but the right to change its data, [MS-FSA] section 2.1.5.1.2.1.
	// Only the rights MAXIMUM_ALLOWED added are dropped: asking for the right to
	// change the data explicitly is still refused by the backend.
	if req.DesiredAccess&maximumAllowed != 0 {
		if attr, err := share.FS.Stat(path); err == nil && !attr.IsDir && attr.ReadOnly {
			explicit, _ := grantedAccess(req.DesiredAccess&^maximumAllowed, false)
			access &^= writeDataRights &^ explicit
		}
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

	open.id, open.persistent = c.newFileID()
	open.file, open.isDir = file, attr.IsDir
	c.opens[open.id] = open
	c.server.openFiles.add(open)
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
	return types.SMB2_FILEID{Persistent: o.persistent, Volatile: o.id}
}

// newFileID mints the two halves of a FileId for a new open: a Volatile value
// unique among this connection's live handles (it keys c.opens) and a Persistent
// value unrelated to it. Both are random and non-zero, so a capture shows two
// independent numbers the way Windows does, not a counter whose halves are equal
// and start at one. The relatedFileID sentinel (all ones) is avoided so a real
// handle can never be mistaken for the "same as the previous request" marker.
func (c *Connection) newFileID() (volatile, persistent uint64) {
	for {
		volatile = randomFileIDHalf()
		if volatile == relatedFileID.Volatile {
			continue
		}
		if _, taken := c.opens[volatile]; !taken {
			break
		}
	}
	for {
		persistent = randomFileIDHalf()
		if persistent != relatedFileID.Persistent {
			return volatile, persistent
		}
	}
}

// randomFileIDHalf returns a non-zero random 64-bit value. crypto/rand.Read does
// not fail as of Go 1.24 (it panics if the OS entropy source does), so there is
// no error to propagate.
func randomFileIDHalf() uint64 {
	for {
		var b [8]byte
		_, _ = rand.Read(b[:])
		if v := binary.LittleEndian.Uint64(b[:]); v != 0 {
			return v
		}
	}
}

// close answers CLOSE.
func (c *Connection) close(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.CloseRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}

	answer := commands.NewCloseResponse()
	if open.pipe == nil && req.Flags&commands.SMB2_CLOSE_FLAG_POSTQUERY_ATTRIB != 0 {
		if attr, err := open.fs().Stat(open.currentPath()); err == nil {
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
	if open.pipe != nil {
		return open.pipe.session.Close()
	}
	c.server.openFiles.remove(open)
	c.cleanupNotifies(open)
	c.server.oplocks.release(open)
	c.server.shareModes.release(open)
	open.file.Close()
	if !open.deleteOnClose {
		return nil
	}
	if open.isDir {
		return open.fs().Rmdir(open.currentPath())
	}
	return open.fs().Remove(open.currentPath())
}

// read answers READ.
func (c *Connection) read(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.ReadRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if open.pipe != nil {
		return c.readPipe(open, req, resp)
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
	if open.pipe != nil {
		return c.writePipe(open, req, resp)
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

// ioctl answers IOCTL. The secure negotiate validation and the named-pipe
// transceive are served.
func (c *Connection) ioctl(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.IoctlRequest)
	switch req.CtlCode {
	case filesystem.FSCTL_VALIDATE_NEGOTIATE_INFO:
	case filesystem.FSCTL_PIPE_TRANSCEIVE:
		return c.transceivePipe(r, req, resp)
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
