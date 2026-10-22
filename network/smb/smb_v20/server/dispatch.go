package server

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// handler answers one request by setting resp.command, and returns its status.
type handler func(c *Connection, r *request, resp *response) nt_status.NT_STATUS

// handlers are the commands served. LOCK and OPLOCK_BREAK acknowledgements are
// answered STATUS_NOT_SUPPORTED.
var handlers = map[codes.CommandCode]handler{
	codes.SMB2_NEGOTIATE:       (*Connection).negotiate,
	codes.SMB2_SESSION_SETUP:   (*Connection).sessionSetup,
	codes.SMB2_LOGOFF:          (*Connection).logoff,
	codes.SMB2_TREE_CONNECT:    (*Connection).treeConnect,
	codes.SMB2_TREE_DISCONNECT: (*Connection).treeDisconnect,
	codes.SMB2_CREATE:          (*Connection).create,
	codes.SMB2_CLOSE:           (*Connection).close,
	codes.SMB2_FLUSH:           (*Connection).flush,
	codes.SMB2_READ:            (*Connection).read,
	codes.SMB2_WRITE:           (*Connection).write,
	codes.SMB2_IOCTL:           (*Connection).ioctl,
	codes.SMB2_ECHO:            (*Connection).echo,
	codes.SMB2_QUERY_DIRECTORY: (*Connection).queryDirectory,
	codes.SMB2_QUERY_INFO:      (*Connection).queryInfo,
	codes.SMB2_SET_INFO:        (*Connection).setInfo,
	codes.SMB2_CHANGE_NOTIFY:   (*Connection).changeNotify,
}

// dispatch runs the handler of a request.
func (c *Connection) dispatch(r *request, resp *response) nt_status.NT_STATUS {
	h, ok := handlers[r.header.Command]
	if !ok {
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}
	return h(c, r, resp)
}

// echo answers ECHO.
func (c *Connection) echo(r *request, resp *response) nt_status.NT_STATUS {
	resp.command = commands.NewEchoResponse()
	return nt_status.NT_STATUS_SUCCESS
}
