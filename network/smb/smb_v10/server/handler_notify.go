package server

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// handleNotifyChange answers NT_TRANSACT_NOTIFY_CHANGE: it tells the client when
// the directory a handle names is modified.
//
// The command is answered when the change happens rather than when it is asked
// for, which is what it needed the deferred response path for. It is single-shot:
// [MS-CIFS] section 2.2.7.4 has the client reissue it to watch for more.
//
// Setup words: CompletionFilter(4) FID(2) WatchTree(1) Reserved(1).
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func handleNotifyChange(
	conn *Connection,
	w ResponseWriter,
	req *message.Message,
	reassembly *transactionReassembly,
) nt_status.NT_STATUS {
	// The setup arrives as words; the filter spans the first two.
	if len(reassembly.setup) < 4 {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	completionFilter := uint32(reassembly.setup[0]) | uint32(reassembly.setup[1])<<16
	fid := uint16(reassembly.setup[2])
	watchTree := uint16(reassembly.setup[3])&0x00FF != 0

	open, status := conn.openFor(req, fid)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// "A directory file MUST be opened before this command can be used." A handle
	// on a file has no entries to report changes to.
	if !open.IsDirectory {
		logger.Debugf("SMB1 server: %s watched FID 0x%04X, which is not a directory", conn.Remote, fid)
		return nt_status.NT_STATUS_NOT_A_DIRECTORY
	}
	if open.Tree == nil || open.Tree.Share.FS == nil {
		return nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	watcher, ok := open.Tree.Share.FS.(Watcher)
	if !ok {
		// A backend that cannot be watched says so. Answering success and never
		// notifying would leave the client waiting for a change it would never
		// hear about, which is worse than a refusal it can act on.
		logger.Debugf("SMB1 server: %s watched %q on share %q, which cannot be watched",
			conn.Remote, open.currentPath(), open.Tree.Share.Name)
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	// Every pending watch holds a backend watch and a goroutine until it is
	// answered, so the number a connection may have pending is bounded. The check
	// comes before the watch starts, so a refused request costs nothing.
	if conn.pendingNotifies() >= conn.Server.config.MaxNotifiesPerConnection {
		logger.Debugf("SMB1 server: %s already has %d directory watches pending",
			conn.Remote, conn.Server.config.MaxNotifiesPerConnection)
		return nt_status.NT_STATUS_INSUFFICIENT_RESOURCES
	}

	changes, stop, err := watcher.Watch(open.currentPath(), watchTree)
	if err != nil {
		return statusForFSError(err)
	}

	// Everything the answer needs is captured now: the responder is the only part
	// of the connection safe to touch from the goroutine below, and the handle
	// and share tables belong to the receive loop.
	pending := &pendingNotify{
		responder: w.Defer(),
		closed:    make(chan struct{}),
		done:      make(chan struct{}),
	}
	open.notifies = append(open.notifies, pending)
	budget := reassembly.maxParameterCount

	go func() {
		defer close(pending.done)
		defer stop()
		defer pending.responder.Done()

		waitForChange(conn, pending.responder, pending.closed, changes, completionFilter, budget)
	}()

	// The handler has answered by deferring; the response comes from above.
	return nt_status.NT_STATUS_SUCCESS
}

// pendingNotify is one NT_TRANSACT_NOTIFY_CHANGE request waiting for a change,
// recorded on the handle it watches.
//
// It is recorded there because the handle bounds its life: [MS-CIFS] section
// 3.3.5.59.4 completes the request when "the FID is closed, either by an explicit
// Close operation or another cause, such as an SMB_COM_PROCESS_EXIT [...] or
// SMB_COM_TREE_DISCONNECT". A watch that outlived its handle would hold a backend
// watch and a goroutine that nothing could reach until the connection ended.
type pendingNotify struct {
	// responder is the deferred answer.
	responder AsyncResponder

	// closed is closed by the receive loop when the handle goes, after it has
	// answered the request, and tells the goroutine to stop watching.
	closed chan struct{}

	// done is closed by the goroutine when it has finished, whichever way the
	// request ended.
	done chan struct{}
}

// finished reports whether the request is no longer pending.
func (p *pendingNotify) finished() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

// pendingNotifies counts the NT_TRANSACT_NOTIFY_CHANGE requests still pending on
// the connection, forgetting the ones that have finished.
//
// Only the open handles are consulted: closing a handle completes every request
// pending on it, so a closed handle has none left to count.
func (c *Connection) pendingNotifies() int {
	count := 0
	for _, open := range c.opens {
		kept := open.notifies[:0]
		for _, pending := range open.notifies {
			if !pending.finished() {
				kept = append(kept, pending)
			}
		}
		clear(open.notifies[len(kept):])
		open.notifies = kept
		count += len(kept)
	}
	return count
}

// cleanupNotifies completes every NT_TRANSACT_NOTIFY_CHANGE request pending on a
// handle that is being closed, and stops their watches.
//
// The status is STATUS_NOTIFY_CLEANUP, which [MS-ERREF] section 2.3.1 defines as
// "a notify change request has been completed due to closing the handle that
// made the notify change request", and which Windows sends in this case
// ([MS-SMB2] section 3.3.5.10 requires it of the SMB2 CLOSE). [MS-CIFS] section
// 3.3.5.59.4 is what requires the request to complete at all.
//
// The answer is sent here, on the receive loop, rather than left to the watch's
// goroutine, so it reaches the client before the response to whatever closed the
// handle. A request the goroutine has already answered reports ErrResponderDone,
// which is the expected outcome of that race and not worth logging.
func (c *Connection) cleanupNotifies(open *Open) {
	for _, pending := range open.notifies {
		if err := pending.responder.RespondWithError(nt_status.NT_STATUS_NOTIFY_CLEANUP); err != nil &&
			err != ErrResponderDone {
			logger.Debugf("SMB1 server: failed to end a watch for %s: %v", c.Remote, err)
		}
		close(pending.closed)
	}
	open.notifies = nil
}

// waitForChange blocks until something the filter admits happens, the request is
// cancelled, its handle is closed, or the watch ends, and answers accordingly.
//
// Parameters:
//   - conn: the connection, for logging
//   - responder: the deferred answer to complete
//   - closed: closed when the handle goes, by which time the request is answered
//   - changes: the watch's notifications
//   - completionFilter: the changes the client asked about
//   - budget: how many bytes of notification the client will take
func waitForChange(
	conn *Connection,
	responder AsyncResponder,
	closed <-chan struct{},
	changes <-chan ChangeNotification,
	completionFilter uint32,
	budget int,
) {
	for {
		select {
		case <-closed:
			// The handle was closed, and the receive loop has already answered.
			logger.Debugf("SMB1 server: a directory watch for %s ended with its handle", conn.Remote)
			return

		case <-responder.Cancelled():
			// The client withdrew the request, or the connection went away. A
			// withdrawn request is still owed its answer: [MS-CIFS] section
			// 3.3.5.52 has "The canceled commands MUST return an error result",
			// and a client that sent NT_CANCEL for a watch is waiting for that
			// result before it treats the watch as gone. On a connection that is
			// ending the write fails, which is harmless.
			logger.Debugf("SMB1 server: a directory watch for %s was cancelled", conn.Remote)
			if err := responder.RespondWithError(nt_status.NT_STATUS_CANCELLED); err != nil {
				logger.Debugf("SMB1 server: failed to answer a cancelled watch for %s: %v", conn.Remote, err)
			}
			return

		case change, open := <-changes:
			if !open {
				// The watch ended without a change — the backend stopped
				// watching. Reported rather than left hanging.
				if err := responder.RespondWithError(nt_status.NT_STATUS_NOTIFY_CLEANUP); err != nil {
					logger.Debugf("SMB1 server: failed to end a watch for %s: %v", conn.Remote, err)
				}
				return
			}

			if !changeAdmittedBy(change, completionFilter) {
				continue
			}

			logger.Debugf("SMB1 server: telling %s that %q changed (action %d)",
				conn.Remote, change.Name, change.Action)
			answerChange(conn, responder, change, budget)
			return
		}
	}
}

// answerChange sends the notification, or tells the client to re-enumerate when it
// will not fit.
//
// [MS-CIFS] section 2.2.7.4: "If too many files [...] have changed since the last
// time that the command was issued, then zero bytes are returned and
// STATUS_NOTIFY_ENUM_DIR [...] is returned in the Status field". A client handles
// that by listing the directory again, so it is a complete answer rather than a
// failure — which is what makes a small client buffer safe to honour rather than
// something to work around.
func answerChange(conn *Connection, responder AsyncResponder, change ChangeNotification, budget int) {
	encoded := encodeNotifyInformation(change)

	if budget > 0 && len(encoded) > budget {
		if err := responder.RespondWithError(nt_status.NT_STATUS_NOTIFY_ENUM_DIR); err != nil {
			logger.Debugf("SMB1 server: failed to tell %s to re-enumerate: %v", conn.Remote, err)
		}
		return
	}

	response := newNotifyChangeResponse(encoded)
	if err := responder.Respond(response); err != nil {
		logger.Debugf("SMB1 server: failed to send a change notification to %s: %v", conn.Remote, err)
	}
}

// encodeNotifyInformation renders one change as a FILE_NOTIFY_INFORMATION record.
//
// NextEntryOffset(4) Action(4) FileNameLength(4) FileName(variable), per [MS-FSCC]
// section 2.7.1. One record, so NextEntryOffset is zero and the chain ends here.
//
// The name is UTF-16LE regardless of what the message declared: this is a native
// structure, and its own definition fixes the encoding.
func encodeNotifyInformation(change ChangeNotification) []byte {
	name := utf16.EncodeUTF16LE(change.Name)

	record := make([]byte, 12, 12+len(name))
	binary.LittleEndian.PutUint32(record[4:8], change.Action)
	binary.LittleEndian.PutUint32(record[8:12], uint32(len(name)))
	return append(record, name...)
}

// changeAdmittedBy reports whether the client asked about a change of this kind.
//
// A filter of zero admits everything: a client that asked to be told about nothing
// in particular is asking to be told about anything, and refusing to notify it
// would leave it waiting forever.
func changeAdmittedBy(change ChangeNotification, completionFilter uint32) bool {
	if completionFilter == 0 {
		return true
	}
	return completionFilter&filterForAction(change.Action) != 0
}

// filterForAction maps a change to the CompletionFilter bits a client would have
// set to ask about it.
//
// The mapping is deliberately generous: an addition or a removal is a name change
// for a file and for a directory both, because this server does not record which
// of the two an entry that has just disappeared was. Erring towards notifying is
// the safer direction — a client told about a change it did not ask for
// re-enumerates and finds nothing, while one not told about a change it did ask
// for waits forever.
func filterForAction(action uint32) uint32 {
	switch action {
	case changeAdded, changeRemoved,
		uint32(filesystem.FileActionRenamedOldName), uint32(filesystem.FileActionRenamedNewName):
		return filesystem.FILE_NOTIFY_CHANGE_FILE_NAME | filesystem.FILE_NOTIFY_CHANGE_DIR_NAME
	case changeModified:
		return filesystem.FILE_NOTIFY_CHANGE_SIZE |
			filesystem.FILE_NOTIFY_CHANGE_LAST_WRITE |
			filesystem.FILE_NOTIFY_CHANGE_ATTRIBUTES
	}
	return 0
}

// newNotifyChangeResponse builds the NT_TRANSACT response carrying a change.
//
// The notification travels in the parameter block: [MS-CIFS] section 2.2.7.4 says
// "The TotalParameterCount field of the server response indicates the number of
// bytes that are being returned", and MaxDataCount is required to be zero, so
// there is no data block to put it in.
func newNotifyChangeResponse(parameters []byte) *ntTransactParametersResponse {
	return &ntTransactParametersResponse{parameters: parameters}
}

// ntTransactParametersResponse is an NT_TRANSACT response whose whole payload is
// its parameter block.
//
// The deferred responder sends a command rather than a transaction, so the
// framing an immediate transaction would get from sendNtTransactResponse is done
// here instead — for one small record it is a fixed layout rather than a
// fragmentation problem.
type ntTransactParametersResponse struct {
	command_interface.Command
	parameters []byte
}

// Marshal emits the NT_TRANSACT response words with the parameter block behind
// them.
func (r *ntTransactParametersResponse) Marshal() ([]byte, error) {
	const wordCount = 18

	// WordCount(1) Reserved1(3) TotalParameterCount(4) TotalDataCount(4)
	// ParameterCount(4) ParameterOffset(4) ParameterDisplacement(4) DataCount(4)
	// DataOffset(4) DataDisplacement(4) SetupCount(1) then ByteCount(2).
	head := make([]byte, 1+2*wordCount+1)
	head[0] = wordCount

	words := head[1 : len(head)-1]
	binary.LittleEndian.PutUint32(words[3:7], uint32(len(r.parameters)))   // TotalParameterCount
	binary.LittleEndian.PutUint32(words[11:15], uint32(len(r.parameters))) // ParameterCount

	// ParameterOffset is measured from the start of the SMB header.
	parameterOffset := header.SMB_HEADER_SIZE + len(head) + 2
	binary.LittleEndian.PutUint32(words[15:19], uint32(parameterOffset))

	byteCount := make([]byte, 2)
	binary.LittleEndian.PutUint16(byteCount, uint16(len(r.parameters)))

	// SetupCount is the last word's low byte and stays zero: no setup words come back.
	out := append([]byte{}, head...)
	out = append(out, byteCount...)
	out = append(out, r.parameters...)
	return out, nil
}

// GetCommandCode reports the command this response answers.
func (r *ntTransactParametersResponse) GetCommandCode() codes.CommandCode {
	return codes.SMB_COM_NT_TRANSACT
}
