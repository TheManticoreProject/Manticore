package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// errChangeNotifyCancelled is returned by ChangeNotify when the wait is
// interrupted by Cancel.
var errChangeNotifyCancelled = fmt.Errorf("change notify cancelled")

// IsChangeNotifyCancelled reports whether err is the error ChangeNotify returns
// when it was interrupted by Cancel.
func IsChangeNotifyCancelled(err error) bool {
	return err == errChangeNotifyCancelled
}

// ChangeNotify monitors the directory open identified by fileId and blocks until a
// change matching completionFilter (a combination of filesystem.FILE_NOTIFY_CHANGE_*
// bits) occurs, then returns the changes. When watchTree is set the whole subtree
// is monitored.
//
// The call blocks: the server registers the watch (replying STATUS_PENDING) and
// sends the result only when a change happens. A concurrent Cancel interrupts the
// wait, in which case ChangeNotify returns an error for which
// IsChangeNotifyCancelled reports true. A STATUS_NOTIFY_ENUM_DIR result (too many
// changes to report) returns no entries and no error; the caller should
// re-enumerate. Wire: SMB2 CHANGE_NOTIFY.
func (c *Client) ChangeNotify(fileId types.SMB2_FILEID, completionFilter uint32, watchTree bool) ([]filesystem.FileNotifyInformation, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return nil, fmt.Errorf("no tree connect established")
	}

	req := commands.NewChangeNotifyRequest()
	req.FileId = fileId
	req.CompletionFilter = types.ULONG(completionFilter)
	req.OutputBufferLength = types.ULONG(c.Connection.Server.MaxTransactSize)
	if watchTree {
		req.Flags = commands.SMB2_WATCH_TREE
	}

	response, err := c.sendReceive(c.newRequest(req), "ChangeNotify")
	if err != nil {
		return nil, err
	}
	switch status := statusFromResponse(response); status {
	case 0x00000000:
		// changes available below
	case ntStatusCancelled:
		return nil, errChangeNotifyCancelled
	case ntStatusNotifyEnumDir:
		return nil, nil
	default:
		return nil, fmt.Errorf("change notify failed: %s", formatNTStatus(status))
	}

	notifyResponse, ok := response.Command.(*commands.ChangeNotifyResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected change notify response command: %T", response.Command)
	}
	return filesystem.ParseFileNotifyInformation(notifyResponse.OutputBuffer), nil
}
