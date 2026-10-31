package server

import (
	"encoding/binary"
	"sync"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// pendingNotify is a CHANGE_NOTIFY answered STATUS_PENDING whose final response
// is still owed. Whoever completes it first sends that response: a change, a
// CANCEL, or the handle closing.
type pendingNotify struct {
	asyncID   uint64
	messageID uint64
	session   *Session
	once      sync.Once
	stopped   chan struct{}
	reply     func(nt_status.NT_STATUS, command_interface.CommandInterface)
}

// finish sends the final response unless the request was already completed.
func (p *pendingNotify) finish(status nt_status.NT_STATUS, body command_interface.CommandInterface) {
	p.once.Do(func() {
		close(p.stopped)
		p.reply(status, body)
	})
}

// finished reports whether the final response has been sent.
func (p *pendingNotify) finished() bool {
	select {
	case <-p.stopped:
		return true
	default:
		return false
	}
}

// changeNotify answers CHANGE_NOTIFY (MS-SMB2 3.3.5.19) asynchronously: the
// interim response goes now, the final one when the directory changes.
func (c *Connection) changeNotify(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.ChangeNotifyRequest)
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if !open.isDir {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	if status := open.require(fileflags.FILE_LIST_DIRECTORY); !status.IsSuccess() {
		return status
	}
	watcher, ok := open.fs().(smb1.Watcher)
	if !ok {
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}
	if c.pendingNotifies() >= c.server.config.MaxNotifiesPerConnection {
		return nt_status.NT_STATUS_INSUFFICIENT_RESOURCES
	}
	changes, stopWatch, err := watcher.Watch(open.currentPath(), req.Flags&commands.SMB2_WATCH_TREE != 0)
	if err != nil {
		return smb1.StatusForFSError(err)
	}

	c.nextAsyncID++
	p := &pendingNotify{
		asyncID:   c.nextAsyncID,
		messageID: r.header.MessageId,
		session:   r.session,
		stopped:   make(chan struct{}),
		reply:     c.asyncReplier(r, c.nextAsyncID),
	}
	c.trackNotify(open, p)

	// The final response must not overtake the interim one (MS-SMB2 3.3.4.2), so
	// the watch only starts reporting once the interim response is on the wire.
	interimSent := make(chan struct{})
	resp.onSent = func([]byte) { close(interimSent) }

	go func() {
		defer stopWatch()
		select {
		case <-interimSent:
		case <-p.stopped:
			return
		}
		for {
			select {
			case <-p.stopped:
				return
			case change, ok := <-changes:
				if !ok {
					p.finish(nt_status.NT_STATUS_NOTIFY_CLEANUP, commands.NewChangeNotifyResponse())
					return
				}
				if admits(req.CompletionFilter, change) {
					status, body := notification(change, changes, req.CompletionFilter, int(req.OutputBufferLength))
					p.finish(status, body)
					return
				}
			}
		}
	}()

	logger.Debugf("SMB2 server: %s watching %q (filter 0x%x) as CHANGE_NOTIFY %d", c.remote, open.currentPath(), req.CompletionFilter, p.asyncID)
	resp.asyncID = p.asyncID
	return nt_status.NT_STATUS_PENDING
}

// trackNotify records a pending request on its open and forgets the ones that
// already completed.
func (c *Connection) trackNotify(open *Open, p *pendingNotify) {
	kept := open.notifies[:0]
	for _, other := range open.notifies {
		if other.finished() {
			delete(c.notifies, other.asyncID)
		} else {
			kept = append(kept, other)
		}
	}
	open.notifies = append(kept, p)
	c.notifies[p.asyncID] = p
}

// pendingNotifies counts the CHANGE_NOTIFY requests still pending on the
// connection, forgetting the ones that already completed.
func (c *Connection) pendingNotifies() int {
	for id, p := range c.notifies {
		if p.finished() {
			delete(c.notifies, id)
		}
	}
	return len(c.notifies)
}

// cancel answers a CANCEL by completing the request it names, by AsyncId for an
// asynchronous one and by MessageId otherwise, among those of session.
func (c *Connection) cancel(h *header.Header, session *Session) {
	for id, p := range c.notifies {
		if p.session != session {
			continue
		}
		if h.Flags.IsAsync() && id == h.AsyncId || !h.Flags.IsAsync() && p.messageID == h.MessageId {
			delete(c.notifies, id)
			p.finish(nt_status.NT_STATUS_CANCELLED, nil)
		}
	}
}

// cleanupNotifies completes the requests pending on a handle being closed.
func (c *Connection) cleanupNotifies(open *Open) {
	for _, p := range open.notifies {
		delete(c.notifies, p.asyncID)
		p.finish(nt_status.NT_STATUS_NOTIFY_CLEANUP, commands.NewChangeNotifyResponse())
	}
	open.notifies = nil
}

// asyncReplier returns the function that sends the final response of an
// asynchronous request. It captures everything it needs, so it is safe to call
// from any goroutine.
func (c *Connection) asyncReplier(r *request, asyncID uint64) func(nt_status.NT_STATUS, command_interface.CommandInterface) {
	session, messageID, charge := r.session, r.header.MessageId, r.header.CreditCharge
	sign := r.header.Flags.IsSigned() || session.signingRequired
	var encryptedBy *Session
	if r.encrypted {
		encryptedBy = session
	}

	return func(status nt_status.NT_STATUS, body command_interface.CommandInterface) {
		logger.Debugf("SMB2 server: %s CHANGE_NOTIFY %d completed with %s", c.remote, asyncID, status)
		if body == nil {
			body = commands.NewErrorResponse(codes.SMB2_CHANGE_NOTIFY)
		}
		h := header.NewHeader()
		h.Command = codes.SMB2_CHANGE_NOTIFY
		h.Status = uint32(status)
		h.Flags = flags.SMB2_FLAGS_SERVER_TO_REDIR | flags.SMB2_FLAGS_ASYNC_COMMAND
		h.MessageId = messageID
		h.AsyncId = asyncID
		h.SessionId = session.id
		h.CreditCharge = charge

		resp := &response{header: h, command: body, session: session, sign: sign}
		if err := c.sendResponses([]*response{resp}, encryptedBy); err != nil {
			logger.Debugf("SMB2 server: failed to complete a CHANGE_NOTIFY for %s: %v", c.remote, err)
		}
	}
}

// notification renders a change, and any others already waiting, as chained
// FILE_NOTIFY_INFORMATION records. When they do not fit in the client's buffer
// it is told to list the directory again instead.
func notification(first smb1.ChangeNotification, more <-chan smb1.ChangeNotification, filter uint32, budget int) (nt_status.NT_STATUS, command_interface.CommandInterface) {
	var output []byte
	last := 0
	for change, ok := first, true; ok; {
		if admits(filter, change) {
			start := (len(output) + 3) &^ 3
			if len(output) > 0 {
				binary.LittleEndian.PutUint32(output[last:], uint32(start-last))
			}
			name := utf16.EncodeUTF16LE(change.Name)
			output = append(output, make([]byte, start-len(output))...)
			output = binary.LittleEndian.AppendUint32(output, 0)
			output = binary.LittleEndian.AppendUint32(output, change.Action)
			output = binary.LittleEndian.AppendUint32(output, uint32(len(name)))
			output = append(output, name...)
			last = start
		}
		select {
		case change, ok = <-more:
		default:
			ok = false
		}
	}

	answer := commands.NewChangeNotifyResponse()
	if len(output) > budget {
		return nt_status.NT_STATUS_NOTIFY_ENUM_DIR, answer
	}
	answer.OutputBuffer = output
	return nt_status.NT_STATUS_SUCCESS, answer
}

// admits reports whether a CompletionFilter asks about a change. A filter of
// zero asks about everything, as in the SMB1 server. A name change is reported
// for files and directories alike, since a removed entry's kind is no longer known.
func admits(filter uint32, change smb1.ChangeNotification) bool {
	if filter == 0 {
		return true
	}
	switch filesystem.FileAction(change.Action) {
	case filesystem.FileActionAdded, filesystem.FileActionRemoved,
		filesystem.FileActionRenamedOldName, filesystem.FileActionRenamedNewName:
		return filter&(filesystem.FILE_NOTIFY_CHANGE_FILE_NAME|filesystem.FILE_NOTIFY_CHANGE_DIR_NAME) != 0
	case filesystem.FileActionModified:
		return filter&(filesystem.FILE_NOTIFY_CHANGE_SIZE|filesystem.FILE_NOTIFY_CHANGE_LAST_WRITE|filesystem.FILE_NOTIFY_CHANGE_ATTRIBUTES) != 0
	}
	return false
}
