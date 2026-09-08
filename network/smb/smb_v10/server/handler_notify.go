package server

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
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
			conn.Remote, open.Path, open.Tree.Share.Name)
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	changes, stop, err := watcher.Watch(open.Path, watchTree)
	if err != nil {
		return statusForFSError(err)
	}

	// Everything the answer needs is captured now: the responder is the only part
	// of the connection safe to touch from the goroutine below, and the handle
	// and share tables belong to the receive loop.
	responder := w.Defer()
	budget := reassembly.maxParameterCount

	go func() {
		defer stop()
		defer responder.Done()

		waitForChange(conn, responder, changes, completionFilter, budget)
	}()

	// The handler has answered by deferring; the response comes from above.
	return nt_status.NT_STATUS_SUCCESS
}

// waitForChange blocks until something the filter admits happens, the request is
// cancelled, or the watch ends, and answers accordingly.
//
// Parameters:
//   - conn: the connection, for logging
//   - responder: the deferred answer to complete
//   - changes: the watch's notifications
//   - completionFilter: the changes the client asked about
//   - budget: how many bytes of notification the client will take
func waitForChange(
	conn *Connection,
	responder AsyncResponder,
	changes <-chan ChangeNotification,
	completionFilter uint32,
	budget int,
) {
	for {
		select {
		case <-responder.Cancelled():
			// The client withdrew the request, or the connection went away.
			// Either way there is nobody to answer.
			logger.Debugf("SMB1 server: a directory watch for %s was cancelled", conn.Remote)
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
