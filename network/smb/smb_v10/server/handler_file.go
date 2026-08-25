package server

import (
	"errors"
	"io"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// Resource types reported in a create response ([MS-CIFS] 2.2.4.64.2).
const (
	resourceTypeDisk = 0x0000
)

// treeFor resolves the tree a request acts on, and the disk share behind it.
//
// Every file command needs the same three things established — a tree, a disk
// share, and a file system — so they are checked once here rather than at the top
// of each handler where one could be forgotten.
func (c *Connection) treeFor(req *message.Message) (*Tree, nt_status.NT_STATUS) {
	tree, status := c.anyTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, status
	}
	if tree.Share.Type != ShareTypeDisk || tree.Share.FS == nil {
		return nil, nt_status.NT_STATUS_BAD_DEVICE_TYPE
	}
	return tree, nt_status.NT_STATUS_SUCCESS
}

// anyTreeFor resolves the tree a request names, whatever kind of share it is.
//
// Only the commands that serve more than one kind of share use this — an open,
// which reaches a file on a disk share and a pipe on a pipe share. Everything
// else goes through treeFor and gets the disk check for free, which is what keeps
// a file-system handler from being reached with a nil FS.
func (c *Connection) anyTreeFor(req *message.Message) (*Tree, nt_status.NT_STATUS) {
	tree := c.Tree(uint16(req.Header.TID))
	if tree == nil {
		return nil, nt_status.NT_STATUS_SMB_BAD_TID
	}
	// A tree belongs to the session that opened it: another session's identifier
	// must not reach it, even on the same connection.
	if tree.SessionUID != uint16(req.Header.UID) {
		logger.Debugf("SMB1 server: %s used TID 0x%04X from UID 0x%04X, which does not own it",
			c.Remote, tree.TID, uint16(req.Header.UID))
		return nil, nt_status.NT_STATUS_SMB_BAD_TID
	}
	return tree, nt_status.NT_STATUS_SUCCESS
}

// openFor resolves the handle a request acts on, checking it belongs to the tree
// the request named.
func (c *Connection) openFor(req *message.Message, fid uint16) (*Open, nt_status.NT_STATUS) {
	open := c.Open(fid)
	if open == nil {
		return nil, nt_status.NT_STATUS_SMB_BAD_FID
	}
	if open.Tree == nil || open.Tree.TID != uint16(req.Header.TID) {
		logger.Debugf("SMB1 server: %s used FID 0x%04X on TID 0x%04X, which does not hold it",
			c.Remote, fid, uint16(req.Header.TID))
		return nil, nt_status.NT_STATUS_SMB_BAD_FID
	}
	return open, nt_status.NT_STATUS_SUCCESS
}

// statusForFSError maps a backend failure onto the protocol status that describes
// it, so a client sees a meaningful refusal rather than a generic one.
func statusForFSError(err error) nt_status.NT_STATUS {
	switch {
	case err == nil:
		return nt_status.NT_STATUS_SUCCESS
	case errors.Is(err, ErrNotFound):
		return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
	case errors.Is(err, ErrExists):
		return nt_status.NT_STATUS_OBJECT_NAME_COLLISION
	case errors.Is(err, ErrNotDirectory):
		return nt_status.NT_STATUS_NOT_A_DIRECTORY
	case errors.Is(err, ErrIsDirectory):
		return nt_status.NT_STATUS_FILE_IS_A_DIRECTORY
	case errors.Is(err, ErrNotEmpty):
		return nt_status.NT_STATUS_DIRECTORY_NOT_EMPTY
	case errors.Is(err, ErrReadOnly):
		return nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	case errors.Is(err, ErrAccessDenied):
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	return nt_status.NT_STATUS_UNSUCCESSFUL
}

// openFlagsFor translates a create request's access and options into what a
// backend needs.
//
// The disposition decides what happens about existence, and the options decide
// what the target must be. A request whose options contradict each other is
// refused rather than resolved arbitrarily.
func openFlagsFor(desiredAccess, createDisposition, createOptions uint32) (OpenFlags, nt_status.NT_STATUS) {
	flags := OpenFlags{
		Read:  desiredAccess&(fileflags.GENERIC_READ|fileflags.GENERIC_ALL|fileflags.FILE_READ_DATA) != 0,
		Write: desiredAccess&(fileflags.GENERIC_WRITE|fileflags.GENERIC_ALL|fileflags.FILE_WRITE_DATA|fileflags.FILE_APPEND_DATA) != 0,
	}
	// An open that asked for nothing in particular still reads: a client commonly
	// opens purely to query, and refusing that would break it.
	if !flags.Read && !flags.Write {
		flags.Read = true
	}

	switch createDisposition {
	case fileflags.FILE_OPEN:
		// Must exist.
	case fileflags.FILE_CREATE:
		flags.Create, flags.CreateNew = true, true
	case fileflags.FILE_OPEN_IF:
		flags.Create = true
	case fileflags.FILE_OVERWRITE:
		flags.Truncate = true
	case fileflags.FILE_OVERWRITE_IF, fileflags.FILE_SUPERSEDE:
		flags.Create, flags.Truncate = true, true
	default:
		return OpenFlags{}, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	flags.Directory = createOptions&fileflags.FILE_DIRECTORY_FILE != 0
	flags.NonDirectory = createOptions&fileflags.FILE_NON_DIRECTORY_FILE != 0
	if flags.Directory && flags.NonDirectory {
		return OpenFlags{}, nt_status.NT_STATUS_INVALID_PARAMETER
	}
	// Creating or truncating a directory is not a thing.
	if flags.Directory && flags.Truncate {
		return OpenFlags{}, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	return flags, nt_status.NT_STATUS_SUCCESS
}

// handleNtCreateAndx answers SMB_COM_NT_CREATE_ANDX: the open path, which creates
// or opens a file or directory and returns the handle a client acts through.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/e1d7f4a5-c9d8-4a4e-9b0e-0e8f3f2a4c5d
func handleNtCreateAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.NtCreateAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	// An open is the one command that serves both kinds of share, so it resolves
	// the tree without the disk check and applies its own.
	tree, status := conn.anyTreeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// Unicode is a per-request property, not a per-connection one: this
	// repository's client sends file-I/O messages in OEM even when the connection
	// negotiated Unicode, so the request's own flag is what governs.
	requested := decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode())

	// A pipe share has no file system behind it, so it takes a different path
	// entirely: what a client opens there is a pipe, and the handle it gets back
	// is what a later transaction names.
	if tree.Share.Type == ShareTypeNamedPipe {
		return conn.openPipe(w, tree, requested)
	}
	if tree.Share.Type != ShareTypeDisk || tree.Share.FS == nil {
		return nt_status.NT_STATUS_BAD_DEVICE_TYPE
	}

	path, err := resolvePath(requested)
	if err != nil {
		logger.Debugf("SMB1 server: %s asked to open %q, which is refused: %v",
			conn.Remote, requested, err)
		return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	flags, status := openFlagsFor(uint32(request.DesiredAccess), uint32(request.CreateDisposition), uint32(request.CreateOptions))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// A read-only share refuses a modifying open whatever access the client asked
	// for. Enforced here rather than in the backend so a backend cannot forget it.
	if tree.Share.ReadOnly && (flags.Write || flags.Create || flags.CreateNew || flags.Truncate) {
		logger.Debugf("SMB1 server: %s asked to open %q for writing on read-only share %q",
			conn.Remote, path, tree.Share.Name)
		return nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}

	// Whether the target existed decides what the response reports it did.
	_, statErr := tree.Share.FS.Stat(path)
	existed := statErr == nil

	file, err := tree.Share.FS.Open(path, flags)
	if err != nil {
		logger.Debugf("SMB1 server: %s could not open %q on %q: %v", conn.Remote, path, tree.Share.Name, err)
		return statusForFSError(err)
	}

	attr, err := file.Stat()
	if err != nil {
		file.Close()
		return statusForFSError(err)
	}

	fid, err := conn.fids.Allocate()
	if err != nil {
		file.Close()
		logger.Warnf("SMB1 server: refusing an open from %s: %v", conn.Remote, err)
		return nt_status.NT_STATUS_TOO_MANY_OPENED_FILES
	}

	open := &Open{
		FID:           fid,
		Tree:          tree,
		Path:          path,
		File:          file,
		IsDirectory:   attr.IsDir,
		Readable:      flags.Read,
		Writable:      flags.Write && !tree.Share.ReadOnly,
		DeleteOnClose: uint32(request.CreateOptions)&fileflags.FILE_DELETE_ON_CLOSE != 0,
		Created:       time.Now().UTC(),
	}
	conn.addOpen(open)

	logger.Debugf("SMB1 server: %s opened %q on %q as FID 0x%04X", conn.Remote, path, tree.Share.Name, fid)

	response := commands.NewNtCreateAndxResponse()
	response.FID = types.USHORT(fid)
	response.CreateDisposition = types.ULONG(createActionFor(existed, flags))
	response.CreateTime = *msdtyp.NewFILETIMEFromTime(attr.Created)
	response.LastAccessTime = *msdtyp.NewFILETIMEFromTime(attr.Accessed)
	response.LastWriteTime = *msdtyp.NewFILETIMEFromTime(attr.Modified)
	response.LastChangeTime = *msdtyp.NewFILETIMEFromTime(attr.Changed)
	response.ExtFileAttributes = types.SMB_EXT_FILE_ATTR(attributesFor(attr))
	response.AllocationSize = types.LARGE_INTEGER{QuadPart: uint64(attr.AllocationSize)}
	response.EndOfFile = types.LARGE_INTEGER{QuadPart: uint64(attr.Size)}
	response.ResourceType = types.USHORT(resourceTypeDisk)
	if attr.IsDir {
		response.Directory = types.UCHAR(1)
	}

	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer the open for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// createActionFor reports what an open did, which a client uses to tell a create
// from an open of something that was already there.
func createActionFor(existed bool, flags OpenFlags) uint32 {
	switch {
	case !existed:
		return fileflags.FILE_CREATE
	case flags.Truncate:
		return fileflags.FILE_OVERWRITE
	}
	return fileflags.FILE_OPEN
}

// attributesFor renders a backend's view of an entry as the attribute bits a
// client expects. FILE_ATTRIBUTE_NORMAL is only meaningful alone, so it is used
// when nothing else applies.
func attributesFor(attr FileAttr) uint32 {
	attributes := uint32(0)
	if attr.IsDir {
		attributes |= fileflags.FILE_ATTRIBUTE_DIRECTORY
	}
	if attr.ReadOnly {
		attributes |= fileflags.FILE_ATTRIBUTE_READONLY
	}
	if attributes == 0 {
		attributes = fileflags.FILE_ATTRIBUTE_NORMAL
	}
	return attributes
}

// handleClose answers SMB_COM_CLOSE: it releases a handle, applying any pending
// delete-on-close.
func handleClose(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.CloseRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// A client may set the modification time as it closes. 0 and -1 both mean
	// "leave it alone", which is why the field cannot simply be applied.
	if request.LastTimeModified != 0 && request.LastTimeModified != 0xFFFFFFFF && open.Writable &&
		open.Tree != nil && open.Tree.Share.FS != nil {
		modified := time.Unix(int64(request.LastTimeModified), 0).UTC()
		if err := open.Tree.Share.FS.SetAttr(open.Path, FileAttr{Modified: modified}, AttrMask{Modified: true}); err != nil {
			logger.Debugf("SMB1 server: could not set the modification time of %q: %v", open.Path, err)
		}
	}

	if err := conn.closeOpen(uint16(request.FID)); err != nil {
		logger.Debugf("SMB1 server: closing FID 0x%04X for %s reported %v", uint16(request.FID), conn.Remote, err)
	}

	if err := w.WriteResponse(commands.NewCloseResponse()); err != nil {
		logger.Debugf("SMB1 server: failed to answer the close for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// handleReadAndx answers SMB_COM_READ_ANDX.
//
// A read past the end of the file returns nothing rather than failing, and a read
// that reaches the end returns what there was: SMB reports a short count, so
// truncation is the normal answer rather than an error.
func handleReadAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.ReadAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if !open.Readable {
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	if open.IsDirectory {
		return nt_status.NT_STATUS_FILE_IS_A_DIRECTORY
	}

	// The offset is 64-bit when the request carried the high word, which it does
	// only in the twelve-word form.
	offset := int64(uint64(request.Offset) | uint64(request.OffsetHigh)<<32)
	if offset < 0 {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	// Bounded by what the client asked for and by what the connection agreed a
	// message may carry, so a client cannot ask for a response it could not
	// receive.
	length := int(request.MaxCountOfBytesToReturn)
	if limit := int(conn.Server.config.MaxBufferSize) - readResponseOverhead; length > limit {
		length = limit
	}
	if length < 0 {
		length = 0
	}

	buffer := make([]byte, length)
	read, err := open.File.ReadAt(buffer, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		logger.Debugf("SMB1 server: reading %q for %s failed: %v", open.Path, conn.Remote, err)
		return statusForFSError(err)
	}

	response := commands.NewReadAndxResponse()
	response.Data = buffer[:read]
	response.DataLength = types.USHORT(read)

	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer the read for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// readResponseOverhead is the space a read response needs beyond its data: the
// 32-byte header, the parameter words and the byte count. Reserving it keeps a
// full-size read inside the negotiated buffer.
const readResponseOverhead = 64

// handleWriteAndx answers SMB_COM_WRITE_ANDX.
func handleWriteAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.WriteAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if !open.Writable {
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	if open.IsDirectory {
		return nt_status.NT_STATUS_FILE_IS_A_DIRECTORY
	}

	offset := int64(uint64(request.Offset) | uint64(request.OffsetHigh)<<32)
	if offset < 0 {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	// The declared length governs, but never past what actually arrived: the two
	// disagreeing is a malformed request, not licence to read past the buffer.
	data := []byte(request.Data)
	if declared := int(request.DataLength); declared < len(data) {
		data = data[:declared]
	}

	written, err := open.File.WriteAt(data, offset)
	if err != nil {
		logger.Debugf("SMB1 server: writing %q for %s failed: %v", open.Path, conn.Remote, err)
		return statusForFSError(err)
	}

	// WritethroughMode asks for the write to be committed before the response.
	if uint16(request.WriteMode)&writeThroughMode != 0 {
		if err := open.File.Sync(); err != nil {
			logger.Debugf("SMB1 server: syncing %q for %s failed: %v", open.Path, conn.Remote, err)
			return statusForFSError(err)
		}
	}

	response := commands.NewWriteAndxResponse()
	response.Count = types.USHORT(written)

	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer the write for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// writeThroughMode is the WriteMode bit asking for the write to reach storage
// before the response ([MS-CIFS] 2.2.4.43.1).
const writeThroughMode = 0x0001

// handleFlush answers SMB_COM_FLUSH: it commits one handle, or every handle on
// the tree when the FID is the 0xFFFF wildcard.
func handleFlush(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.FlushRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	if uint16(request.FID) == 0xFFFF {
		tree, status := conn.treeFor(req)
		if status != nt_status.NT_STATUS_SUCCESS {
			return status
		}
		for _, open := range conn.opens {
			if open.Tree == tree && open.File != nil {
				if err := open.File.Sync(); err != nil {
					logger.Debugf("SMB1 server: syncing %q failed: %v", open.Path, err)
				}
			}
		}
	} else {
		open, status := conn.openFor(req, uint16(request.FID))
		if status != nt_status.NT_STATUS_SUCCESS {
			return status
		}
		if open.File != nil {
			if err := open.File.Sync(); err != nil {
				return statusForFSError(err)
			}
		}
	}

	if err := w.WriteResponse(commands.NewFlushResponse()); err != nil {
		logger.Debugf("SMB1 server: failed to answer the flush for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}
