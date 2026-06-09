package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// Ioctl issues an SMB2 IOCTL on the open identified by fileId and returns the
// output buffer. ctlCode is the control code; when isFsctl is true the
// SMB2_0_IOCTL_IS_FSCTL flag is set (the usual case for FSCTL_* codes).
// maxOutputResponse bounds the bytes the server may return.
//
// STATUS_BUFFER_OVERFLOW is treated as success: it is a warning that the output
// did not fit, and the partial data that did fit is returned. Callers that need
// the remainder (e.g. a pipe transceive whose reply spans multiple reads) issue
// follow-up READs. Wire: SMB2 IOCTL.
func (c *Client) Ioctl(fileId types.SMB2_FILEID, ctlCode uint32, input []byte, isFsctl bool, maxOutputResponse uint32) ([]byte, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return nil, fmt.Errorf("no tree connect established")
	}

	req := commands.NewIoctlRequest()
	req.CtlCode = types.ULONG(ctlCode)
	req.FileId = fileId
	req.Input = input
	req.MaxOutputResponse = types.ULONG(maxOutputResponse)
	if isFsctl {
		req.Flags = commands.SMB2_0_IOCTL_IS_FSCTL
	}

	response, err := c.sendReceive(c.newRequest(req), "Ioctl")
	if err != nil {
		return nil, err
	}
	if status := statusFromResponse(response); status != 0x00000000 && status != ntStatusBufferOverflow {
		return nil, fmt.Errorf("ioctl 0x%08x failed: %s", ctlCode, formatNTStatus(status))
	}

	ioctlResponse, ok := response.Command.(*commands.IoctlResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected ioctl response command: %T", response.Command)
	}
	return ioctlResponse.Output, nil
}

// TransactNamedPipe writes a message to the named pipe identified by fileId and
// returns the pipe's reply in a single round-trip, via FSCTL_PIPE_TRANSCEIVE. It
// is the SMB2 basis for DCE/RPC over named pipes (ncacn_np). If the reply exceeds
// maxOutputResponse the returned data is the portion that fit; the caller reads
// the remainder with ReadFile.
func (c *Client) TransactNamedPipe(fileId types.SMB2_FILEID, input []byte, maxOutputResponse uint32) ([]byte, error) {
	return c.Ioctl(fileId, filesystem.FSCTL_PIPE_TRANSCEIVE, input, true, maxOutputResponse)
}
