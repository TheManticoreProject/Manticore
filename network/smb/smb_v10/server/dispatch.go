package server

import (
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// commandHandler answers one request.
//
// A handler writes its own response through w, because a command may produce no
// response at all or several (SMB_COM_ECHO produces one per EchoCount). It
// returns NT_STATUS_SUCCESS when it has answered, and any other status to have
// the dispatcher send that as an error response instead.
type commandHandler func(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS

// dispatchTable maps a command code to its handler. A parseable command with no
// entry here is answered with STATUS_NOT_IMPLEMENTED, which is a legitimate
// response rather than a placeholder: a good number of SMB_COM_* codes were
// reserved and never implemented, and the specification requires exactly that
// answer for them.
var dispatchTable = map[codes.CommandCode]commandHandler{
	// Connection and session.
	codes.SMB_COM_NEGOTIATE:          handleNegotiate,
	codes.SMB_COM_SESSION_SETUP_ANDX: handleSessionSetupAndx,
	codes.SMB_COM_LOGOFF_ANDX:        handleLogoffAndx,
	codes.SMB_COM_ECHO:               handleEcho,

	// Trees.
	codes.SMB_COM_TREE_CONNECT_ANDX: handleTreeConnectAndx,
	codes.SMB_COM_TREE_DISCONNECT:   handleTreeDisconnect,

	// File handles.
	codes.SMB_COM_NT_CREATE_ANDX: handleNtCreateAndx,
	codes.SMB_COM_CLOSE:          handleClose,
	codes.SMB_COM_READ_ANDX:      handleReadAndx,
	codes.SMB_COM_WRITE_ANDX:     handleWriteAndx,
	codes.SMB_COM_FLUSH:          handleFlush,

	// File management.
	codes.SMB_COM_DELETE:           handleDelete,
	codes.SMB_COM_RENAME:           handleRename,
	codes.SMB_COM_CREATE_DIRECTORY: handleCreateDirectory,
	codes.SMB_COM_DELETE_DIRECTORY: handleDeleteDirectory,
	codes.SMB_COM_CHECK_DIRECTORY:  handleCheckDirectory,

	// Transactions, which carry the directory-enumeration and information
	// subcommands.
	codes.SMB_COM_TRANSACTION2:           handleTransaction2,
	codes.SMB_COM_TRANSACTION2_SECONDARY: handleTransaction2Secondary,
	codes.SMB_COM_FIND_CLOSE2:            handleFindClose2,
}

// sessionlessCommands are the commands a client may send before it holds a
// session. Everything else requires one: NEGOTIATE and SESSION_SETUP_ANDX are how
// a session comes into being, and ECHO is defined to work without one.
var sessionlessCommands = map[codes.CommandCode]bool{
	codes.SMB_COM_NEGOTIATE:          true,
	codes.SMB_COM_SESSION_SETUP_ANDX: true,
	codes.SMB_COM_ECHO:               true,
}

// echoedRequestFlags are the SMB_FLAGS bits a response mirrors from its request.
// SMB_FLAGS_REPLY is set separately; every other bit is either request-only or
// reserved.
const echoedRequestFlags = flags.Flags(flags.FLAGS_CASE_INSENSITIVE | flags.FLAGS_CANONICALIZED_PATHS)

// echoedRequestFlags2 are the SMB_FLAGS2 bits a response mirrors from its
// request: they record what the client and server agreed to speak, so the reply
// has to keep speaking it.
const echoedRequestFlags2 = flags2.Flags2(flags2.FLAGS2_UNICODE |
	flags2.FLAGS2_NT_STATUS_ERROR_CODES |
	flags2.FLAGS2_LONG_NAMES_ALLOWED |
	flags2.FLAGS2_EXTENDED_SECURITY)

// replyHeader builds the response header for a request, per [MS-CIFS] 2.2.3.1.
// The identifiers are copied verbatim so the client can correlate the reply, and
// the status is encoded in whichever form the request selected.
//
// Parameters:
//   - request: the request header being answered
//   - status: the status to report
//   - signed: whether the response will carry a signature
//
// Returns:
//   - The response header
func replyHeader(request *header.Header, status nt_status.NT_STATUS, signed bool) *header.Header {
	h := header.NewHeader()

	h.Command = request.Command
	h.Status = EncodeStatus(status, request.Flags2&flags2.FLAGS2_NT_STATUS_ERROR_CODES != 0)

	h.Flags = flags.Flags(flags.FLAGS_REPLY) | (request.Flags & echoedRequestFlags)
	h.Flags2 = request.Flags2 & echoedRequestFlags2
	if signed {
		// A signed message announces the fact, so the receiver knows to check
		// the SecurityFeatures field rather than ignore it.
		h.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	}

	// The identifiers are the client's, echoed so it can match the response to
	// the request it sent.
	h.PIDHigh = request.PIDHigh
	h.PIDLow = request.PIDLow
	h.TID = request.TID
	h.UID = request.UID
	h.MID = request.MID

	h.Reserved = 0

	return h
}

// errorResponse is the body of an SMB error response: WordCount 0 followed by
// ByteCount 0, with the error itself carried in the header's Status field. The
// client side recognizes exactly this shape — an empty parameter and data block
// means "read the status".
type errorResponse struct {
	command_interface.Command
}

// newErrorResponse builds the payload-less body of an error response for a
// command code.
func newErrorResponse(command codes.CommandCode) *errorResponse {
	r := &errorResponse{}
	r.Init()
	r.SetCommandCode(command)
	return r
}

// Marshal emits WordCount 0 and ByteCount 0.
func (r *errorResponse) Marshal() ([]byte, error) {
	return []byte{0x00, 0x00, 0x00}, nil
}

// isKnownCommand reports whether a command code names a command the message
// layer can decode. It distinguishes an unrecognized code, which the
// specification answers with ERRSRV/ERRbadcmd, from a recognized code this
// server has not implemented, which is STATUS_NOT_IMPLEMENTED.
func isKnownCommand(command codes.CommandCode) bool {
	_, err := commands.CreateRequestCommand(command)
	return err == nil
}

// dispatch runs the built-in handler for a decoded request, or answers with the
// status that says why it could not.
func (s *Server) dispatch(conn *Connection, w ResponseWriter, req *message.Message) {
	// A command that acts within a session must name one that exists. Checking
	// here rather than in each handler means a command added later cannot forget
	// to.
	if !sessionlessCommands[req.Header.Command] {
		if conn.Session(uint16(req.Header.UID)) == nil {
			logger.Debugf("SMB1 server: %s sent command 0x%02X on UID 0x%04X, which names no session",
				conn.Remote, uint8(req.Header.Command), uint16(req.Header.UID))
			s.writeError(conn, w, nt_status.NT_STATUS_SMB_BAD_UID)
			return
		}
	}

	handler, ok := dispatchTable[req.Header.Command]
	if !ok {
		logger.Debugf("SMB1 server: %s sent unimplemented command 0x%02X (%s)",
			conn.Remote, uint8(req.Header.Command), req.Header.Command)
		s.writeError(conn, w, nt_status.NT_STATUS_NOT_IMPLEMENTED)
		return
	}

	if status := handler(conn, w, req); status != nt_status.NT_STATUS_SUCCESS {
		logger.Debugf("SMB1 server: command 0x%02X from %s failed with %s",
			uint8(req.Header.Command), conn.Remote, statusName(status))
		s.writeError(conn, w, status)
	}
}

// writeError sends an error response and logs a write failure, which is only
// ever a connection that has already gone away.
func (s *Server) writeError(conn *Connection, w ResponseWriter, status nt_status.NT_STATUS) {
	if err := w.WriteError(status); err != nil {
		logger.Debugf("SMB1 server: failed to send %s to %s: %v", statusName(status), conn.Remote, err)
	}
}
