package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// The AccessMode bits of SMB_COM_OPEN ([MS-CIFS] section 2.2.4.3.1).
const (
	openAccessModeMask = 0x0007

	openAccessRead      = 0
	openAccessWrite     = 1
	openAccessReadWrite = 2
	openAccessExecute   = 3
)

// The OpenMode bits of SMB_COM_OPEN_ANDX ([MS-CIFS] section 2.2.4.41.1).
const (
	openModeFileExistsMask = 0x0003

	openModeFailIfExists  = 0
	openModeAppendIfExsts = 1
	openModeTruncate      = 2

	openModeCreateIfMissing = 0x0010
)

// The Mode values of SMB_COM_SEEK ([MS-CIFS] section 2.2.4.19.1).
const (
	seekFromStart   = 0x0000
	seekFromCurrent = 0x0001
	seekFromEnd     = 0x0002
)

// openResultsCreated is the OpenResults bit an SMB_COM_OPEN_ANDX response sets
// when the open created the file rather than opening an existing one.
const openResultsCreated = 0x0001

// coreOpenFlagsFor translates an SMB_COM_OPEN AccessMode into backend flags.
//
// The access field is a value rather than a bitmask, so an unrecognised value is
// refused rather than defaulted: [MS-CIFS] section 2.2.4.3.1 requires values 4 to
// 7 to be answered with STATUS_OS2_INVALID_ACCESS, and treating one of them as
// "read" would grant an access the client did not ask for and did not expect.
//
// Parameters:
//   - accessMode: the request's AccessMode field
//
// Returns:
//   - The backend flags, and the status to report if the mode is not usable
func coreOpenFlagsFor(accessMode uint16) (OpenFlags, nt_status.NT_STATUS) {
	switch accessMode & openAccessModeMask {
	case openAccessRead, openAccessExecute:
		// Execution is read access as far as the storage is concerned: nothing
		// here runs a file, and refusing the mode would fail an open a client is
		// entitled to make.
		return OpenFlags{Read: true, NonDirectory: true}, nt_status.NT_STATUS_SUCCESS
	case openAccessWrite:
		return OpenFlags{Write: true, NonDirectory: true}, nt_status.NT_STATUS_SUCCESS
	case openAccessReadWrite:
		return OpenFlags{Read: true, Write: true, NonDirectory: true}, nt_status.NT_STATUS_SUCCESS
	}
	return OpenFlags{}, nt_status.NT_STATUS_OS2_INVALID_ACCESS
}

// applyOpenMode folds an OpenMode field into backend flags.
//
// OpenMode is the create-disposition of the NT commands expressed in two bits
// ([MS-CIFS] section 2.2.4.41.1): what to do about a file that exists, and
// whether to create one that does not. It is shared by SMB_COM_OPEN_ANDX and
// TRANS2_OPEN2, which carry the same field.
//
// Parameters:
//   - flags: the flags to fold into, already carrying the access mode
//   - openMode: the request's OpenMode field
func applyOpenMode(flags *OpenFlags, openMode uint16) {
	if openMode&openModeCreateIfMissing != 0 {
		flags.Create = true
	}

	switch openMode & openModeFileExistsMask {
	case openModeTruncate:
		flags.Truncate = true
		// Truncating is a write, whatever access the client named.
		flags.Write = true
	case openModeFailIfExists:
		// Fail if it exists and create if it does not is a create-new; without
		// the create bit it is a plain open that must find the file, which the
		// backend already reports.
		if flags.Create {
			// Both bits, which is the convention openFlagsFor established: the
			// backend reads CreateNew as "refuse an existing file" and Create as
			// "make a missing one", and a create-new needs to say both.
			flags.CreateNew = true
		}
	case openModeAppendIfExsts:
		// Appending needs write access; the offset is the client's business,
		// since every core-set write carries one.
		flags.Write = true
	}
}

// coreOpen opens a path for one of the core-set commands and registers the handle.
//
// It is the shared body of SMB_COM_OPEN, SMB_COM_CREATE, SMB_COM_CREATE_NEW,
// SMB_COM_CREATE_TEMPORARY and SMB_COM_OPEN_ANDX, which differ in what they say
// rather than in what they do.
//
// Parameters:
//   - req: the request, for its TID, PID and encoding
//   - name: the path the client named, before resolution
//   - flags: what the open needs from the backend
//
// Returns:
//   - The handle, whether the target already existed, and a status
func (c *Connection) coreOpen(
	req *message.Message,
	name string,
	flags OpenFlags,
) (*Open, bool, nt_status.NT_STATUS) {
	tree, status := c.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, false, status
	}
	if tree.Share.FS == nil {
		return nil, false, nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	path, err := resolvePath(name)
	if err != nil {
		logger.Debugf("SMB1 server: %s asked to open %q, which is refused: %v", c.Remote, name, err)
		return nil, false, nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
	}

	// A read-only share refuses a modifying open whatever the client asked for,
	// checked here rather than left to the backend so a backend cannot forget it.
	if tree.Share.ReadOnly && (flags.Write || flags.Create || flags.CreateNew || flags.Truncate) {
		return nil, false, nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}

	_, statErr := tree.Share.FS.Stat(path)
	existed := statErr == nil

	file, err := tree.Share.FS.Open(path, flags)
	if err != nil {
		logger.Debugf("SMB1 server: %s could not open %q on %q: %v", c.Remote, path, tree.Share.Name, err)
		return nil, existed, statusForFSError(err)
	}

	attr, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, existed, statusForFSError(err)
	}

	fid, err := c.fids.Allocate()
	if err != nil {
		file.Close()
		logger.Warnf("SMB1 server: refusing an open from %s: %v", c.Remote, err)
		return nil, existed, nt_status.NT_STATUS_TOO_MANY_OPENED_FILES
	}

	open := &Open{
		FID:         fid,
		Tree:        tree,
		Path:        path,
		File:        file,
		IsDirectory: attr.IsDir,
		Readable:    flags.Read,
		Writable:    flags.Write && !tree.Share.ReadOnly,
		// The PID that opened the handle, so SMB_COM_PROCESS_EXIT can close what
		// a failed process left behind.
		PID:     req.Header.GetPID(),
		Created: time.Now().UTC(),
	}
	c.addOpen(open)

	logger.Debugf("SMB1 server: %s opened %q on %q as FID 0x%04X (core set)",
		c.Remote, path, tree.Share.Name, fid)
	return open, existed, nt_status.NT_STATUS_SUCCESS
}

// handleOpen answers SMB_COM_OPEN: the core-set open of an existing file.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleOpen(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.OpenRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	flags, status := coreOpenFlagsFor(uint16(request.AccessMode))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	name := decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode())
	open, _, status := conn.coreOpen(req, name, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	attr, err := open.File.Stat()
	if err != nil {
		return statusForFSError(err)
	}

	response := commands.NewOpenResponse()
	response.FID = types.USHORT(open.FID)
	response.FileAttrs.SetAttributes(legacyAttributesFor(attr))
	response.LastModified = types.ULONG(utimeOf(attr.Modified))
	response.FileSize = types.ULONG(uint32(attr.Size))
	// The granted access is echoed, which is what the client asked for: nothing
	// here narrows an open the backend accepted.
	response.AccessMode = request.AccessMode

	return conn.answer(w, response)
}

// handleOpenAndx answers SMB_COM_OPEN_ANDX, the extended core-set open.
//
// Its OpenMode says what to do about existence, which is the create-disposition
// of the NT commands in two bits ([MS-CIFS] section 2.2.4.41.1): whether to fail,
// append to or truncate an existing file, and whether to create a missing one.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleOpenAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.OpenAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	flags, status := coreOpenFlagsFor(uint16(request.AccessMode))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	applyOpenMode(&flags, uint16(request.OpenMode))

	name := decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode())
	open, existed, status := conn.coreOpen(req, name, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	attr, err := open.File.Stat()
	if err != nil {
		return statusForFSError(err)
	}

	response := commands.NewOpenAndxResponse()
	response.FID = types.USHORT(open.FID)
	response.FileAttrs = types.SMB_FILE_ATTRIBUTES{Attributes: legacyAttributesFor(attr)}
	response.LastWriteTime = types.ULONG(utimeOf(attr.Modified))
	response.FileDataSize = types.ULONG(uint32(attr.Size))
	response.AccessRights = types.USHORT(request.AccessMode)
	response.ResourceType = types.USHORT(resourceTypeDisk)
	if !existed {
		response.OpenResults = types.USHORT(openResultsCreated)
	}

	return conn.answer(w, response)
}

// handleCreate answers SMB_COM_CREATE: create or truncate, then open for writing.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleCreate(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.CreateRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	name := decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode())
	flags := OpenFlags{Read: true, Write: true, Create: true, Truncate: true, NonDirectory: true}

	open, _, status := conn.coreOpen(req, name, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	response := commands.NewCreateResponse()
	response.FID = types.USHORT(open.FID)
	return conn.answer(w, response)
}

// handleCreateNew answers SMB_COM_CREATE_NEW, which differs from SMB_COM_CREATE in
// refusing a file that already exists rather than truncating it.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleCreateNew(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.CreateNewRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	name := decodeWireString(request.FileName.Buffer, req.Header.Flags2.IsUnicode())
	flags := OpenFlags{Read: true, Write: true, Create: true, CreateNew: true, NonDirectory: true}

	open, _, status := conn.coreOpen(req, name, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// SMB_COM_CREATE_NEW shares the SMB_COM_CREATE response shape.
	response := commands.NewCreateResponse()
	response.FID = types.USHORT(open.FID)
	return conn.answer(w, response)
}

// handleCreateTemporary answers SMB_COM_CREATE_TEMPORARY: it creates a file with a
// server-chosen name in the directory the client names, and reports the name it
// chose.
//
// The name has to be one the client can then use, so it is generated inside the
// named directory and returned share-relative — a name the client could not
// resolve would make the handle the only way to reach the file and leave it
// unreachable after a close.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleCreateTemporary(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.CreateTemporaryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	directory := decodeWireString(request.DirectoryName.Buffer, req.Header.Flags2.IsUnicode())

	tree, status := conn.treeFor(req)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if tree.Share.FS == nil {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	// A name that does not already exist. The loop is bounded: a directory that
	// somehow holds every candidate is a refusal rather than an endless search.
	var chosen string
	for attempt := 0; attempt < temporaryNameAttempts; attempt++ {
		candidate, err := temporaryName(directory)
		if err != nil {
			return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
		}
		resolved, err := resolvePath(candidate)
		if err != nil {
			return nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD
		}
		if _, err := tree.Share.FS.Stat(resolved); err != nil {
			chosen = candidate
			break
		}
	}
	if chosen == "" {
		logger.Warnf("SMB1 server: could not find an unused temporary name in %q for %s", directory, conn.Remote)
		return nt_status.NT_STATUS_OBJECT_NAME_COLLISION
	}

	flags := OpenFlags{Read: true, Write: true, Create: true, CreateNew: true, NonDirectory: true}
	open, _, status := conn.coreOpen(req, chosen, flags)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	response := commands.NewCreateTemporaryResponse()
	response.FID = types.USHORT(open.FID)
	if err := response.TemporaryFileName.SetString(open.Path); err != nil {
		return nt_status.NT_STATUS_UNSUCCESSFUL
	}
	return conn.answer(w, response)
}

// handleRead answers SMB_COM_READ: the core-set read, whose count and offset are
// both 16- and 32-bit respectively, so it cannot address beyond 4 GiB.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleRead(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.ReadRequest)
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

	length := int(request.CountOfBytesToRead)
	if limit := int(conn.Server.config.MaxBufferSize) - readResponseOverhead; length > limit {
		length = limit
	}
	if length < 0 {
		length = 0
	}

	offset := int64(uint32(request.ReadOffsetInBytes))

	file, status := fileFor(open)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	buffer := make([]byte, length)
	read, err := file.ReadAt(buffer, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return statusForFSError(err)
	}

	response := commands.NewReadResponse()
	response.CountOfBytesReturned = types.USHORT(read)
	// The data block is BufferFormat(0x01) CountOfBytesRead(2) Bytes, per
	// [MS-CIFS] section 2.2.4.11.2, which is what an SMB_STRING in the
	// 16-bit variable-block form emits.
	response.Bytes.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT)
	response.Bytes.Buffer = []types.UCHAR(buffer[:read])
	return conn.answer(w, response)
}

// handleWrite answers SMB_COM_WRITE.
//
// A write of zero bytes truncates the file to the offset, which is the one piece
// of this command that is not obvious: [MS-CIFS] section 3.3.5.14 defines a
// CountOfBytesToWrite of zero as a request to set the length rather than as a
// write of nothing, so treating it as a no-op would silently drop a truncation.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleWrite(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.WriteRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	written, status := conn.coreWrite(open, int64(uint32(request.WriteOffsetInBytes)),
		[]byte(request.Data.Buffer), int(request.CountOfBytesToWrite))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	response := commands.NewWriteResponse()
	response.CountOfBytesWritten = types.USHORT(written)
	return conn.answer(w, response)
}

// handleWriteAndClose answers SMB_COM_WRITE_AND_CLOSE: a write followed by a
// close, in one exchange.
//
// The close happens even if the write failed, because the client will not send
// another close for this handle: the command's contract is that the handle is gone
// afterwards, and leaving it open would leak it.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleWriteAndClose(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.WriteAndCloseRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	written, writeStatus := conn.coreWrite(open, int64(uint32(request.WriteOffsetInBytes)),
		[]byte(request.Data), int(request.CountOfBytesToWrite))

	// The handle goes whatever the write did.
	if err := conn.closeOpen(uint16(request.FID)); err != nil {
		logger.Debugf("SMB1 server: closing FID 0x%04X for %s reported %v",
			uint16(request.FID), conn.Remote, err)
	}

	if writeStatus != nt_status.NT_STATUS_SUCCESS {
		return writeStatus
	}

	response := commands.NewWriteAndCloseResponse()
	response.CountOfBytesWritten = types.USHORT(written)
	return conn.answer(w, response)
}

// coreWrite is the shared body of the core-set writes.
//
// Parameters:
//   - open: the handle to write through
//   - offset: where to write
//   - data: the bytes that arrived
//   - declared: how many the request said it carried
//
// Returns:
//   - How many bytes were written, and a status
func (c *Connection) coreWrite(open *Open, offset int64, data []byte, declared int) (int, nt_status.NT_STATUS) {
	if !open.Writable {
		return 0, nt_status.NT_STATUS_ACCESS_DENIED
	}
	if open.IsDirectory {
		return 0, nt_status.NT_STATUS_FILE_IS_A_DIRECTORY
	}
	if offset < 0 {
		return 0, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	file, status := fileFor(open)
	if status != nt_status.NT_STATUS_SUCCESS {
		return 0, status
	}

	// The break goes out before the file changes, for both of the changes this
	// path makes: a client caching reads on the promise that nothing is writing
	// has to be told before the data under its cache moves.
	c.breakOplocksOn(open.Tree.Share, open.Path, open)

	// A count of zero sets the length rather than writing nothing.
	if declared == 0 {
		if err := file.Truncate(offset); err != nil {
			return 0, statusForFSError(err)
		}
		return 0, nt_status.NT_STATUS_SUCCESS
	}

	// Never past what actually arrived: the two disagreeing is a malformed
	// request, not licence to read past the buffer.
	if declared < len(data) {
		data = data[:declared]
	}

	written, err := file.WriteAt(data, offset)
	if err != nil {
		return written, statusForFSError(err)
	}
	return written, nt_status.NT_STATUS_SUCCESS
}

// handleSeek answers SMB_COM_SEEK: it moves the handle's file pointer and reports
// where it ended up.
//
// The pointer is per-handle state that nothing else here consults, because every
// read and write this server serves carries its own offset. It is kept anyway,
// because seeking from the current position is only meaningful if the position is
// remembered between calls — a server that reported an answer without storing it
// would give a different result for the same sequence of requests.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleSeek(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.SeekRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	open, status := conn.openFor(req, uint16(request.FID))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}
	if open.IsPipe {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	// Offset is signed: seeking backwards from the end or the current position is
	// the point of those modes.
	delta := int64(int32(uint32(request.Offset)))

	var base int64
	switch uint16(request.Mode) {
	case seekFromStart:
		base = 0
	case seekFromCurrent:
		base = open.Position
	case seekFromEnd:
		file, fileStatus := fileFor(open)
		if fileStatus != nt_status.NT_STATUS_SUCCESS {
			return fileStatus
		}
		attr, err := file.Stat()
		if err != nil {
			return statusForFSError(err)
		}
		base = attr.Size
	default:
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	position := base + delta
	if position < 0 {
		// Seeking before the start of a file is refused rather than clamped: a
		// clamped answer would tell the client the seek succeeded somewhere it
		// did not ask for.
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	open.Position = position

	response := commands.NewSeekResponse()
	response.Offset = types.ULONG(uint32(position))
	return conn.answer(w, response)
}

// handleTreeConnect answers SMB_COM_TREE_CONNECT, the deprecated form of the tree
// connect.
//
// It carries the share path and returns a TID, exactly as the AndX form does, so
// it shares that handler's share lookup and tree allocation.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleTreeConnect(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.TreeConnectRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	unicode := req.Header.Flags2.IsUnicode()
	tree, status := conn.connectTree(req,
		decodeWireString(request.Path.Buffer, unicode),
		decodeOEMString(request.Service.Buffer))
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	response := commands.NewTreeConnectResponse()
	response.MaxBufferSize = types.USHORT(conn.Server.config.MaxBufferSize)
	response.TID = types.USHORT(tree.TID)

	w.SetResponseTID(tree.TID)
	return conn.answer(w, response)
}

// handleProcessExit answers SMB_COM_PROCESS_EXIT: it releases what the request's
// process identifier still holds.
//
// [MS-CIFS] section 2.2.4.18: "Upon receiving an SMB_COM_PROCESS_EXIT request, the
// server MUST close any resources owned by the Process ID (PID) listed in the
// request header." The command exists to clean up after a client process that died
// without closing its handles, so the PID recorded on each open is what makes it
// mean anything.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleProcessExit(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	if _, ok := req.Command.(*commands.ProcessExitRequest); !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	pid := req.Header.GetPID()

	// Collected first: closing a handle mutates the table being walked.
	doomed := []uint16{}
	for fid, open := range conn.opens {
		if open.PID == pid {
			doomed = append(doomed, fid)
		}
	}

	for _, fid := range doomed {
		if err := conn.closeOpen(fid); err != nil {
			logger.Debugf("SMB1 server: closing FID 0x%04X for exited PID 0x%08X reported %v", fid, pid, err)
		}
	}

	if len(doomed) > 0 {
		logger.Debugf("SMB1 server: released %d handles held by exited PID 0x%08X for %s",
			len(doomed), pid, conn.Remote)
	}

	return conn.answer(w, commands.NewProcessExitResponse())
}

// temporaryNameAttempts bounds the search for an unused temporary name.
const temporaryNameAttempts = 32

// temporaryName invents a name inside a directory for SMB_COM_CREATE_TEMPORARY.
//
// The name is 8.3-shaped so that it can also be seen through the core-set search
// commands, which cannot carry a longer one. It is drawn from the system's
// cryptographic random source rather than from a counter, because a counter shared
// across connections would need locking and would still collide after a restart.
//
// Parameters:
//   - directory: the share-relative directory the client named
//
// Returns:
//   - A share-relative candidate name, or an error if the directory is not usable
func temporaryName(directory string) (string, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}

	name := fmt.Sprintf("TMP%s.TMP", strings.ToUpper(hex.EncodeToString(suffix)))
	trimmed := strings.Trim(strings.ReplaceAll(directory, "\\", "/"), "/")
	if trimmed == "" {
		return name, nil
	}
	return trimmed + "/" + name, nil
}
