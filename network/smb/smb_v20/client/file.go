package client

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// ntStatusEndOfFile (STATUS_END_OF_FILE) is returned by a READ when the requested
// offset is at or past the end of the file.
const ntStatusEndOfFile = 0xC0000011

// CreateFile opens or creates a file (or pipe) on the currently connected tree
// and returns the server-assigned FileId. It requests no oplock.
//
// The path is relative to the share root; a leading backslash is stripped (an
// SMB2 CREATE name is share-relative). Wire: SMB2 CREATE.
func (c *Client) CreateFile(path string, desiredAccess, shareAccess, createDisposition, createOptions uint32) (types.SMB2_FILEID, error) {
	fileId, _, err := c.createFile(path, desiredAccess, shareAccess, createDisposition, createOptions, commands.SMB2_OPLOCK_LEVEL_NONE)
	return fileId, err
}

// CreateFileWithOplock opens or creates a file like CreateFile but also requests
// an oplock (one of commands.SMB2_OPLOCK_LEVEL_*) and returns the oplock level the
// server granted. When the server later needs to break that oplock — because
// another client opens the same file — it sends an OPLOCK_BREAK notification; the
// caller observes it with WaitOplockBreak and replies with AcknowledgeOplockBreak.
func (c *Client) CreateFileWithOplock(path string, desiredAccess, shareAccess, createDisposition, createOptions uint32, requestedOplock uint8) (types.SMB2_FILEID, uint8, error) {
	return c.createFile(path, desiredAccess, shareAccess, createDisposition, createOptions, requestedOplock)
}

// createFile performs an SMB2 CREATE and returns the FileId and the granted oplock
// level.
func (c *Client) createFile(path string, desiredAccess, shareAccess, createDisposition, createOptions uint32, requestedOplock uint8) (types.SMB2_FILEID, uint8, error) {
	var fileId types.SMB2_FILEID
	if c.Session == nil || c.Session.TreeId == 0 {
		return fileId, 0, fmt.Errorf("no tree connect established")
	}

	req := commands.NewCreateRequest()
	req.RequestedOplockLevel = types.UCHAR(requestedOplock)
	req.DesiredAccess = desiredAccess
	req.ShareAccess = shareAccess
	req.CreateDisposition = createDisposition
	req.CreateOptions = createOptions
	// Impersonation (2): give the server a static snapshot of the client's context.
	req.ImpersonationLevel = 0x00000002
	// FILE_ATTRIBUTE_NORMAL.
	req.FileAttributes = 0x00000080
	req.Name = strings.TrimPrefix(path, "\\")

	response, err := c.sendReceive(c.newRequest(req), "Create")
	if err != nil {
		return fileId, 0, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fileId, 0, fmt.Errorf("create %q failed: %s", path, formatNTStatus(status))
	}

	createResponse, ok := response.Command.(*commands.CreateResponse)
	if !ok {
		return fileId, 0, fmt.Errorf("unexpected create response command: %T", response.Command)
	}
	return createResponse.FileId, uint8(createResponse.OplockLevel), nil
}

// CloseFile closes an open identified by FileId. Wire: SMB2 CLOSE.
func (c *Client) CloseFile(fileId types.SMB2_FILEID) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	req := commands.NewCloseRequest()
	req.FileId = fileId

	response, err := c.sendReceive(c.newRequest(req), "Close")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("close failed: %s", formatNTStatus(status))
	}
	return nil
}

// ReadFile reads up to length bytes from the open at the given offset. A read at
// or past end-of-file returns an empty slice and no error. Wire: SMB2 READ.
func (c *Client) ReadFile(fileId types.SMB2_FILEID, offset uint64, length uint32) ([]byte, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	// A single READ is bounded by the negotiated MaxReadSize; satisfy a larger
	// request by issuing successive reads until length bytes are read or the file
	// ends. A request within MaxReadSize is a single read, as before.
	maxRead := c.Connection.Server.MaxReadSize
	if maxRead == 0 {
		maxRead = 65536
	}

	out := make([]byte, 0, length)
	remaining := length
	pos := offset
	for remaining > 0 {
		chunk := remaining
		if chunk > maxRead {
			chunk = maxRead
		}
		data, eof, err := c.readChunk(fileId, pos, chunk)
		if err != nil {
			return nil, err
		}
		out = append(out, data...)
		// Stop at end-of-file or a short read (the server returned less than was
		// asked for, which for a pipe is one message and for a file is the tail).
		if eof || len(data) == 0 || uint32(len(data)) < chunk {
			break
		}
		pos += uint64(len(data))
		remaining -= uint32(len(data))
	}
	return out, nil
}

// readChunk performs a single SMB2 READ of up to length bytes at offset. eof is
// true when the server reports end-of-file (STATUS_END_OF_FILE).
func (c *Client) readChunk(fileId types.SMB2_FILEID, offset uint64, length uint32) (data []byte, eof bool, err error) {
	req := commands.NewReadRequest()
	req.FileId = fileId
	req.Offset = offset
	req.Length = length
	req.MinimumCount = 1

	response, err := c.sendReceive(c.newRequest(req), "Read")
	if err != nil {
		return nil, false, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		if status == ntStatusEndOfFile {
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("read failed: %s", formatNTStatus(status))
	}

	readResponse, ok := response.Command.(*commands.ReadResponse)
	if !ok {
		return nil, false, fmt.Errorf("unexpected read response command: %T", response.Command)
	}
	return readResponse.Data, false, nil
}

// WriteFile writes data to the open at the given offset and returns the number of
// bytes written. Wire: SMB2 WRITE.
func (c *Client) WriteFile(fileId types.SMB2_FILEID, offset uint64, data []byte) (uint32, error) {
	if c.Session == nil {
		return 0, fmt.Errorf("no session established")
	}

	// A single WRITE is bounded by the negotiated MaxWriteSize; split a larger
	// payload across successive writes. A payload within MaxWriteSize is a single
	// write, as before.
	maxWrite := c.Connection.Server.MaxWriteSize
	if maxWrite == 0 {
		maxWrite = 65536
	}

	var total uint32
	pos := offset
	for len(data) > 0 {
		n := uint32(len(data))
		if n > maxWrite {
			n = maxWrite
		}
		written, err := c.writeChunk(fileId, pos, data[:n])
		if err != nil {
			return total, err
		}
		// The server-reported count is untrusted: a value larger than the chunk
		// submitted would slice data out of range below. Clamp it to the bytes
		// actually sent and reject the response as malformed.
		if written > n {
			return total, fmt.Errorf("write reported %d bytes written, more than the %d submitted", written, n)
		}
		total += written
		pos += uint64(written)
		data = data[written:]
		if written == 0 {
			break
		}
	}
	return total, nil
}

// writeChunk performs a single SMB2 WRITE of data at offset and returns the
// number of bytes the server accepted.
func (c *Client) writeChunk(fileId types.SMB2_FILEID, offset uint64, data []byte) (uint32, error) {
	req := commands.NewWriteRequest()
	req.FileId = fileId
	req.Offset = offset
	req.Data = data

	response, err := c.sendReceive(c.newRequest(req), "Write")
	if err != nil {
		return 0, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return 0, fmt.Errorf("write failed: %s", formatNTStatus(status))
	}

	writeResponse, ok := response.Command.(*commands.WriteResponse)
	if !ok {
		return 0, fmt.Errorf("unexpected write response command: %T", response.Command)
	}
	return writeResponse.Count, nil
}
