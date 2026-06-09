package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// ntStatusNoMoreFiles (STATUS_NO_MORE_FILES) is returned by QUERY_DIRECTORY when
// enumeration is exhausted.
const ntStatusNoMoreFiles = 0x80000006

// QueryDirectory enumerates the directory open identified by fileId using the
// given FILE_INFORMATION_CLASS and search pattern (empty means "*"), returning
// the raw output buffer of packed information entries. Decoding the entries into
// typed structures is the job of the MS-FSCC information-class package (#525).
//
// A STATUS_NO_MORE_FILES result returns an empty buffer and no error, signalling
// the end of enumeration. Wire: SMB2 QUERY_DIRECTORY.
func (c *Client) QueryDirectory(fileId types.SMB2_FILEID, fileInformationClass uint8, searchPattern string, flags uint8) ([]byte, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return nil, fmt.Errorf("no tree connect established")
	}

	req := commands.NewQueryDirectoryRequest()
	req.FileId = fileId
	req.FileInformationClass = fileInformationClass
	req.Flags = flags
	req.FileName = searchPattern
	req.OutputBufferLength = c.Connection.Server.MaxTransactSize

	response, err := c.sendReceive(c.newRequest(req), "QueryDirectory")
	if err != nil {
		return nil, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		if status == ntStatusNoMoreFiles {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("query directory failed: %s", formatNTStatus(status))
	}

	queryResponse, ok := response.Command.(*commands.QueryDirectoryResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected query directory response command: %T", response.Command)
	}
	return queryResponse.OutputBuffer, nil
}
