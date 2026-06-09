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
// and returns the server-assigned FileId.
//
// The path is relative to the share root; a leading backslash is stripped (an
// SMB2 CREATE name is share-relative). Wire: SMB2 CREATE.
func (c *Client) CreateFile(path string, desiredAccess, shareAccess, createDisposition, createOptions uint32) (types.SMB2_FILEID, error) {
	var fileId types.SMB2_FILEID
	if c.Session == nil || c.Session.TreeId == 0 {
		return fileId, fmt.Errorf("no tree connect established")
	}

	req := commands.NewCreateRequest()
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
		return fileId, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fileId, fmt.Errorf("create %q failed: %s", path, formatNTStatus(status))
	}

	createResponse, ok := response.Command.(*commands.CreateResponse)
	if !ok {
		return fileId, fmt.Errorf("unexpected create response command: %T", response.Command)
	}
	return createResponse.FileId, nil
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

	req := commands.NewReadRequest()
	req.FileId = fileId
	req.Offset = offset
	req.Length = length
	req.MinimumCount = 1

	response, err := c.sendReceive(c.newRequest(req), "Read")
	if err != nil {
		return nil, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		if status == ntStatusEndOfFile {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("read failed: %s", formatNTStatus(status))
	}

	readResponse, ok := response.Command.(*commands.ReadResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected read response command: %T", response.Command)
	}
	return readResponse.Data, nil
}

// WriteFile writes data to the open at the given offset and returns the number of
// bytes written. Wire: SMB2 WRITE.
func (c *Client) WriteFile(fileId types.SMB2_FILEID, offset uint64, data []byte) (uint32, error) {
	if c.Session == nil {
		return 0, fmt.Errorf("no session established")
	}

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
