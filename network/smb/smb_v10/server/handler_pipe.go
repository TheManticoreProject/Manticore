package server

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// PipeHandler serves the named pipes on an IPC share.
//
// A pipe is request-response: a client writes a message and reads the answer,
// which is how MS-RPC travels over SMB. Transact is the operation that does both
// in one exchange and the one every RPC client uses, so a handler that implements
// only that is a complete one for the purpose.
//
// A handler is called on the goroutine serving the connection, so an
// implementation that blocks holds that client up. It may be called concurrently
// for different connections.
type PipeHandler interface {
	// OpenPipe reports whether a pipe exists under this handler, and prepares
	// whatever per-open state it needs. The name has no leading separator and no
	// "PIPE" prefix: "srvsvc", not "\\PIPE\\srvsvc".
	OpenPipe(name string) error

	// Transact writes a message to a pipe and returns the answer. maxOutput is
	// the largest answer the server will take in one exchange, which is
	// maxPipeAnswerSize rather than the client's buffer: fitting the answer to
	// the client is the server's job, because only the server knows how the
	// client will read the rest. A handler whose answer would exceed maxOutput
	// should return what fits and report that more remains.
	Transact(name string, input []byte, maxOutput int) (output []byte, moreRemains bool, err error)

	// ClosePipe releases whatever OpenPipe prepared.
	ClosePipe(name string) error
}

// pipeNameOf extracts a pipe's name from the path a client opened.
//
// A client reaches a pipe by opening "\PIPE\srvsvc" on IPC$, or sometimes
// "\srvsvc", and the name is matched case-insensitively as Windows does. The
// leading separator and the "PIPE" element are stripped so a handler sees the name
// alone.
func pipeNameOf(path string) string {
	trimmed := strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
	if upper := strings.ToUpper(trimmed); strings.HasPrefix(upper, "PIPE/") {
		trimmed = trimmed[len("PIPE/"):]
	} else if upper == "PIPE" {
		return ""
	}
	return strings.Trim(trimmed, "/")
}

// handleTransaction answers SMB_COM_TRANSACTION, the family that carries the named
// pipe operations.
//
// Unlike the other two families the subcommand travels in a setup word AND the
// request names the pipe in its Name field, so both are needed to know what is
// being asked of what.
func handleTransaction(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.TransactionRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	if len(request.Setup) < 1 {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	totalParameters := int(request.TotalParameterCount)
	totalData := int(request.TotalDataCount)
	if totalParameters > maxTrans2Payload || totalData > maxTrans2Payload {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	reassembly := &transactionReassembly{
		family:            "TRANSACTION",
		subcommand:        uint16(request.Setup[0]),
		parameters:        make([]byte, totalParameters),
		data:              make([]byte, totalData),
		maxParameterCount: int(request.MaxParameterCount),
		maxDataCount:      int(request.MaxDataCount),
		setup:             request.Setup,
		// The Name is OEM or UTF-16 according to the request's own flag, as every
		// other name-carrying field is.
		name:    decodeWireString(request.Name.Buffer, req.Header.Flags2.IsUnicode()),
		started: time.Now(),
	}

	if err := reassembly.place(
		[]byte(request.Trans_Parameters), 0,
		[]byte(request.Trans_Data), 0,
	); err != nil {
		logger.Debugf("SMB1 server: %s sent an inconsistent TRANSACTION: %v", conn.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	if reassembly.complete() {
		return conn.runTransaction(w, req, reassembly)
	}

	conn.transaction = reassembly
	return nt_status.NT_STATUS_SUCCESS
}

// handleTransactionSecondary answers SMB_COM_TRANSACTION_SECONDARY.
func handleTransactionSecondary(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.TransactionSecondaryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	reassembly := conn.transaction
	if reassembly == nil || reassembly.family != "TRANSACTION" {
		return nt_status.NT_STATUS_INVALID_SMB
	}
	if time.Since(reassembly.started) > transactionReassemblyTimeout {
		conn.transaction = nil
		return nt_status.NT_STATUS_IO_TIMEOUT
	}

	// TransactionSecondaryRequest names its blocks Trans2_* even though it belongs
	// to the TRANSACTION family rather than to TRANSACTION2. The names are
	// misleading but the fields are the right ones.
	if err := reassembly.place(
		[]byte(request.Trans2_Parameters), int(request.ParameterDisplacement),
		[]byte(request.Trans2_Data), int(request.DataDisplacement),
	); err != nil {
		conn.transaction = nil
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	if !reassembly.complete() {
		return nt_status.NT_STATUS_SUCCESS
	}

	conn.transaction = nil
	return conn.runTransaction(w, req, reassembly)
}

// runTransaction dispatches an assembled TRANSACTION to its pipe operation.
func (c *Connection) runTransaction(w ResponseWriter, req *message.Message, reassembly *transactionReassembly) nt_status.NT_STATUS {
	tree := c.Tree(uint16(req.Header.TID))
	if tree == nil {
		return nt_status.NT_STATUS_SMB_BAD_TID
	}
	if tree.Share.Type != ShareTypeNamedPipe {
		logger.Debugf("SMB1 server: %s sent a pipe transaction on share %q, which is a %q share",
			c.Remote, tree.Share.Name, tree.Share.Type)
		return nt_status.NT_STATUS_BAD_DEVICE_TYPE
	}
	if tree.Share.Pipes == nil {
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	switch subcommands.TransactionSubcommand(reassembly.subcommand) {
	case subcommands.TRANS_TRANSACT_NMPIPE:
		return c.transactNamedPipe(w, req, tree, reassembly)

	case subcommands.TRANS_READ_NMPIPE:
		return c.readNamedPipe(w, reassembly)

	case subcommands.TRANS_PEEK_NMPIPE:
		return c.peekNamedPipe(w, reassembly)

	case subcommands.TRANS_WAIT_NMPIPE:
		// The pipe is available as soon as the handler knows the name, so waiting
		// for it succeeds immediately rather than blocking on nothing.
		name := pipeNameOf(reassembly.name)
		if err := tree.Share.Pipes.OpenPipe(name); err != nil {
			return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
		}
		return c.answerTransaction(w, reassembly, nil, nil)

	case subcommands.TRANS_QUERY_NMPIPE_STATE:
		// A message-mode, blocking, client-end pipe with unlimited instances,
		// which is what an RPC client expects to find.
		state := make([]byte, 2)
		binary.LittleEndian.PutUint16(state, pipeStateMessageMode)
		return c.answerTransaction(w, reassembly, state, nil)

	case subcommands.TRANS_SET_NMPIPE_STATE:
		// Accepted and ignored: the only state a client sets is blocking mode and
		// read mode, and both are already what this reports.
		return c.answerTransaction(w, reassembly, nil, nil)
	}

	logger.Debugf("SMB1 server: %s sent unimplemented pipe operation 0x%04X", c.Remote, reassembly.subcommand)
	return nt_status.NT_STATUS_NOT_IMPLEMENTED
}

// maxPipeAnswerSize bounds the answer the server will take from a handler in one
// exchange, and so bounds what one handle may hold undelivered.
//
// It is not the client's buffer: the client's buffer bounds one response, and an
// answer larger than that is delivered across several reads. It is a ceiling on
// buffering, because the client chooses when to read and a handler that produced
// a very large answer would otherwise pin that memory until the handle closed.
const maxPipeAnswerSize = 64 * 1024

// namedPipeStateConnectionOk is the NamedPipeState a peek reports for a pipe whose
// handler is present and answering ([MS-CIFS] section 2.2.5.5.2).
const namedPipeStateConnectionOk = 0x0003

// peekParametersSize is the size of a peek response's Trans_Parameters:
// ReadDataAvailable(2) MessageBytesLength(2) NamedPipeState(2), per [MS-CIFS]
// section 2.2.5.5.2.
const peekParametersSize = 6

// pipeReadBudget is how many bytes a pipe response may carry, from the client's
// MaxDataCount. A client that named no limit gets the largest a transaction can
// describe.
//
// Parameters:
//   - maxDataCount: the MaxDataCount the client sent
//
// Returns:
//   - The number of bytes the response may carry
func pipeReadBudget(maxDataCount int) int {
	if maxDataCount <= 0 || maxDataCount > maxTrans2Payload {
		return maxTrans2Payload
	}
	return maxDataCount
}

// pipeOpenFrom resolves the pipe handle a transaction acts on from its setup
// words.
//
// [MS-CIFS] sections 3.3.5.57.6, 3.3.5.57.7 and 3.3.5.57.9 all identify the pipe
// by the FID in Setup[1], so this is the one place that reads it.
//
// Parameters:
//   - reassembly: the assembled transaction
//
// Returns:
//   - The open handle, or the status that says why there is none
func (c *Connection) pipeOpenFrom(reassembly *transactionReassembly) (*Open, nt_status.NT_STATUS) {
	if len(reassembly.setup) < 2 {
		logger.Debugf("SMB1 server: %s sent pipe operation 0x%04X with no FID in its setup words",
			c.Remote, reassembly.subcommand)
		return nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	fid := uint16(reassembly.setup[1])
	open := c.Open(fid)
	switch {
	case open == nil:
		logger.Debugf("SMB1 server: %s sent a pipe operation on FID 0x%04X, which is not open", c.Remote, fid)
		return nil, nt_status.NT_STATUS_SMB_BAD_FID
	case !open.IsPipe:
		logger.Debugf("SMB1 server: %s sent a pipe operation on FID 0x%04X, which is a file", c.Remote, fid)
		return nil, nt_status.NT_STATUS_INVALID_HANDLE
	}
	return open, nt_status.NT_STATUS_SUCCESS
}

// readNamedPipe answers TRANS_READ_NMPIPE: it returns the next part of the answer
// buffered on a pipe handle.
//
// This is one of the two ways a client collects an answer that did not fit in the
// transaction that produced it; SMB_COM_READ_ANDX on the same handle is the other.
//
// A read with nothing buffered answers with no data rather than blocking. The
// specification has a blocking pipe wait "indefinitely for data to become
// available" ([MS-CIFS] section 3.3.5.57.9), but a handler here answers a
// transaction rather than producing data on its own, so there is nothing for a
// wait to become available from, and blocking would hold the connection forever.
//
// Wire format: [MS-CIFS] section 2.2.5.8.2. Server behaviour: section 3.3.5.57.9.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func (c *Connection) readNamedPipe(w ResponseWriter, reassembly *transactionReassembly) nt_status.NT_STATUS {
	open, status := c.pipeOpenFrom(reassembly)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	chunk, moreRemains := open.drainPipeOutput(pipeReadBudget(reassembly.maxDataCount))

	// [MS-CIFS] section 3.3.5.57.9: on either success or STATUS_BUFFER_OVERFLOW
	// the server MUST construct a TRANS_READ_NMPIPE Response, so the status
	// travels with the data.
	answer := nt_status.NT_STATUS_SUCCESS
	if moreRemains {
		answer = nt_status.NT_STATUS_BUFFER_OVERFLOW
	}

	logger.Debugf("SMB1 server: read %d bytes of pipe %q for %s, more remains=%t",
		len(chunk), open.Path, c.Remote, moreRemains)

	if err := c.sendTransactionResponse(w, reassembly, nil, chunk, answer); err != nil {
		logger.Debugf("SMB1 server: failed to answer the pipe read for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// peekNamedPipe answers TRANS_PEEK_NMPIPE: it returns the front of the answer
// buffered on a pipe handle without consuming it, plus how much is there.
//
// A client peeks to size a read before making it, so the counts matter as much as
// the data.
//
// Wire format: [MS-CIFS] section 2.2.5.5.2. Server behaviour: section 3.3.5.57.6.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/
func (c *Connection) peekNamedPipe(w ResponseWriter, reassembly *transactionReassembly) nt_status.NT_STATUS {
	open, status := c.pipeOpenFrom(reassembly)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	// The client's budget covers the parameters as well as the data, and the
	// parameters go out in the first message, so the peeked chunk has to leave
	// room for them. Peeking the full budget instead would push the tail of the
	// chunk into a second message and leave ReadDataAvailable and
	// MessageBytesLength describing bytes this response did not carry — the two
	// counts a client sizes its next read from.
	budget := pipeReadBudget(reassembly.maxDataCount) - peekParametersSize
	if budget < 0 {
		budget = 0
	}
	chunk, available := open.peekPipeOutput(budget)

	// MessageBytesLength is what is left of the peeked message after what is
	// being returned.
	parameters := make([]byte, peekParametersSize)
	binary.LittleEndian.PutUint16(parameters[0:2], clampToUint16(available))
	binary.LittleEndian.PutUint16(parameters[2:4], clampToUint16(available-len(chunk)))
	binary.LittleEndian.PutUint16(parameters[4:6], namedPipeStateConnectionOk)

	answer := nt_status.NT_STATUS_SUCCESS
	if available > len(chunk) {
		answer = nt_status.NT_STATUS_BUFFER_OVERFLOW
	}

	logger.Debugf("SMB1 server: peeked %d of %d buffered bytes of pipe %q for %s",
		len(chunk), available, open.Path, c.Remote)

	if err := c.sendTransactionResponse(w, reassembly, parameters, chunk, answer); err != nil {
		logger.Debugf("SMB1 server: failed to answer the pipe peek for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// clampToUint16 fits a length into the 16-bit field that describes it, saturating
// rather than wrapping: a count that wrapped would describe less data than is
// there and stop a client reading the rest.
func clampToUint16(value int) uint16 {
	if value < 0 {
		return 0
	}
	if value > 0xFFFF {
		return 0xFFFF
	}
	return uint16(value)
}

// pipeStateMessageMode is the NMPIPE_STATE a query reports: message-type pipe,
// message-mode reads, blocking, unlimited instances ([MS-CIFS] 2.2.7.2).
const pipeStateMessageMode = 0x0500

// transactNamedPipe answers TRANS_TRANSACT_NMPIPE: it writes the request's data to
// a pipe and returns the answer as the response data.
//
// This is the operation MS-RPC travels over, so it is the one that matters: an RPC
// bind and every call after it is a write-then-read on one pipe, which this does
// in a single exchange.
func (c *Connection) transactNamedPipe(
	w ResponseWriter,
	req *message.Message,
	tree *Tree,
	reassembly *transactionReassembly,
) nt_status.NT_STATUS {
	// [MS-CIFS] section 3.3.5.57.7 is explicit about which pipe is acted on: the
	// one "identified by the SMB_Parameters.Words.Setup.FID field of the request",
	// which is Setup[1] behind the subcommand. So the handle the client opened is
	// what names the pipe, and the request's Name field is boilerplate — the
	// subcommand sets it to "\PIPE\", which names nothing on its own.
	var open *Open
	name := ""
	if len(reassembly.setup) >= 2 {
		resolved, status := c.pipeOpenFrom(reassembly)
		if status != nt_status.NT_STATUS_SUCCESS {
			return status
		}
		open, name = resolved, resolved.Path
	}
	// A client that sent no FID may still have named the pipe outright. That is
	// not what the subcommand specifies, but answering it costs nothing and a
	// client that does it has said unambiguously what it means.
	if name == "" {
		name = pipeNameOf(reassembly.name)
	}
	if name == "" {
		logger.Debugf("SMB1 server: %s sent a pipe transaction naming no pipe", c.Remote)
		return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
	}

	// The handler is asked for the whole answer, not for the part that fits: the
	// client's buffer bounds what one response may carry, not what the pipe
	// produced, and the difference is what gets buffered for the client to read.
	output, moreRemains, err := tree.Share.Pipes.Transact(name, reassembly.data, maxPipeAnswerSize)
	if err != nil {
		logger.Debugf("SMB1 server: pipe %q refused a transaction from %s: %v", name, c.Remote, err)
		return statusForPipeError(err)
	}

	// Cut the answer to what the client said it could receive, and keep the rest
	// on the handle. Keeping it is what makes STATUS_BUFFER_OVERFLOW below an
	// instruction the client can act on: discarding it would leave the client
	// reading for data that no longer exists.
	deliver := output
	if budget := pipeReadBudget(reassembly.maxDataCount); len(deliver) > budget {
		deliver = output[:budget]
	}
	remainder := output[len(deliver):]
	if len(remainder) > 0 {
		if open == nil {
			// No FID was sent, so there is no handle to hold the rest on and no
			// handle for the client to read it from either. The answer is cut and
			// the overflow reported, which is all this case can honestly do.
			logger.Debugf("SMB1 server: %s transacted pipe %q without a FID, so %d undelivered bytes are dropped",
				c.Remote, name, len(remainder))
		} else {
			open.pipeOutput = append(open.pipeOutput, remainder...)
		}
	}

	// What the client is told about a truncated answer is the whole difference
	// between a working RPC conversation and a stuck one. [MS-CIFS] section
	// 3.3.5.57.7: on STATUS_BUFFER_OVERFLOW the server still "MUST construct a
	// TRANS_TRANSACT_NMPIPE Response" — the status travels alongside the data
	// rather than instead of it, and it is what tells the client to read again.
	status := nt_status.NT_STATUS_SUCCESS
	if moreRemains || len(remainder) > 0 {
		status = nt_status.NT_STATUS_BUFFER_OVERFLOW
		logger.Debugf("SMB1 server: pipe %q has more to return to %s", name, c.Remote)
	}

	if err := c.sendTransactionResponse(w, reassembly, nil, deliver, status); err != nil {
		logger.Debugf("SMB1 server: failed to answer the pipe transaction for %s: %v", c.Remote, err)
	}
	// The response has been sent, status included, so the dispatcher must not send
	// an error response on top of it.
	return nt_status.NT_STATUS_SUCCESS
}

// statusForPipeError maps a handler's failure onto the status a client expects.
func statusForPipeError(err error) nt_status.NT_STATUS {
	switch {
	case err == nil:
		return nt_status.NT_STATUS_SUCCESS
	case strings.Contains(err.Error(), "not found"):
		return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
	}
	return nt_status.NT_STATUS_UNSUCCESSFUL
}

// answerTransaction sends a transaction result and reports success, for the
// branches whose only failure would be a connection already going away.
func (c *Connection) answerTransaction(w ResponseWriter, reassembly *transactionReassembly, parameters, data []byte) nt_status.NT_STATUS {
	if err := c.sendTransactionResponse(w, reassembly, parameters, data, nt_status.NT_STATUS_SUCCESS); err != nil {
		logger.Debugf("SMB1 server: failed to answer the pipe transaction for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// sendTransactionResponse sends a TRANSACTION result, splitting it across as many
// messages as it needs.
// The status is carried on the response itself rather than in place of it, which
// STATUS_BUFFER_OVERFLOW requires: it means "here is what fits, ask again".
func (c *Connection) sendTransactionResponse(
	w ResponseWriter,
	reassembly *transactionReassembly,
	parameters, data []byte,
	status nt_status.NT_STATUS,
) error {
	const wordCount = 10

	budget := int(c.Server.config.MaxBufferSize) - trans2ResponseOverhead
	if budget <= 0 {
		return fmt.Errorf("the negotiated buffer leaves no room for a transaction response")
	}
	if reassembly.maxDataCount > 0 && reassembly.maxDataCount < budget {
		budget = reassembly.maxDataCount
	}
	if len(parameters) > budget {
		return fmt.Errorf("the response parameters are %d bytes, over the %d-byte budget", len(parameters), budget)
	}

	sentData := 0
	for first := true; first || sentData < len(data); first = false {
		chunkBudget := budget
		parameterChunk := []byte{}
		if first {
			parameterChunk = parameters
			chunkBudget -= len(parameters)
		}

		chunk := len(data) - sentData
		if chunk > chunkBudget {
			chunk = chunkBudget
		}
		if chunk < 0 {
			chunk = 0
		}

		response := commands.NewTransactionResponse()
		response.TotalParameterCount = types.USHORT(len(parameters))
		response.TotalDataCount = types.USHORT(len(data))
		response.SetupCount = types.UCHAR(0)
		response.Setup = []types.USHORT{}
		response.ParameterCount = types.USHORT(len(parameterChunk))
		response.ParameterDisplacement = types.USHORT(0)
		response.DataCount = types.USHORT(chunk)
		response.DataDisplacement = types.USHORT(sentData)

		preParameters := header.SMB_HEADER_SIZE + 1 + 2*wordCount + 2
		pad1 := (4 - (preParameters % 4)) % 4
		parameterOffset := preParameters + pad1
		response.Pad1 = make([]types.UCHAR, pad1)
		response.ParameterOffset = types.USHORT(parameterOffset)
		response.Trans_Parameters = []types.UCHAR(parameterChunk)

		afterParameters := parameterOffset + len(parameterChunk)
		pad2 := (4 - (afterParameters % 4)) % 4
		response.Pad2 = make([]types.UCHAR, pad2)
		response.DataOffset = types.USHORT(afterParameters + pad2)
		response.Trans_Data = []types.UCHAR(data[sentData : sentData+chunk])

		var err error
		if status == nt_status.NT_STATUS_SUCCESS {
			err = w.WriteResponse(response)
		} else {
			err = w.WriteResponseWithStatus(response, status)
		}
		if err != nil {
			return err
		}

		sentData += chunk
		if chunk == 0 && !first {
			break
		}
	}

	return nil
}

// openPipe answers an NT_CREATE_ANDX against a pipe share.
//
// The handle it returns is the point of the exchange: [MS-CIFS] section 3.3.5.57.7
// identifies the pipe a transaction acts on by the FID in the request's setup
// words, so a client has to open the pipe before it can transact on it. This is
// the first half of every MS-RPC conversation over SMB1.
func (c *Connection) openPipe(w ResponseWriter, tree *Tree, requested string) nt_status.NT_STATUS {
	if tree.Share.Pipes == nil {
		logger.Debugf("SMB1 server: %s asked to open a pipe on share %q, which serves none",
			c.Remote, tree.Share.Name)
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	name := pipeNameOf(requested)
	if name == "" {
		return nt_status.NT_STATUS_OBJECT_NAME_INVALID
	}

	if err := tree.Share.Pipes.OpenPipe(name); err != nil {
		logger.Debugf("SMB1 server: %s could not open pipe %q: %v", c.Remote, name, err)
		return statusForPipeError(err)
	}

	fid, err := c.fids.Allocate()
	if err != nil {
		// The handler prepared state for a handle that will not exist, so give it
		// back rather than leaking it for the life of the connection.
		if closeErr := tree.Share.Pipes.ClosePipe(name); closeErr != nil {
			logger.Debugf("SMB1 server: releasing pipe %q after a failed open reported %v", name, closeErr)
		}
		logger.Warnf("SMB1 server: refusing a pipe open from %s: %v", c.Remote, err)
		return nt_status.NT_STATUS_TOO_MANY_OPENED_FILES
	}

	open := &Open{
		FID:  fid,
		Tree: tree,
		// The name is stored as the handler sees it, since that is what every
		// later call passes back to the handler.
		Path:     name,
		IsPipe:   true,
		Readable: true,
		Writable: true,
		Created:  time.Now().UTC(),
	}
	c.addOpen(open)

	logger.Debugf("SMB1 server: %s opened pipe %q on %q as FID 0x%04X", c.Remote, name, tree.Share.Name, fid)

	response := commands.NewNtCreateAndxResponse()
	response.FID = types.USHORT(fid)
	response.CreateDisposition = types.ULONG(fileflags.FILE_OPEN)
	// A message-mode pipe, which is the only kind a transacted exchange may run
	// against ([MS-CIFS] section 2.2.5.6).
	response.ResourceType = types.USHORT(resourceTypeMessageModePipe)
	response.NMPipeStatus = types.SMB_NMPIPE_STATUS{
		// ICount is the instance count; Flags carries the high half of the 16-bit
		// status, which is where the read-mode and pipe-type fields live.
		ICount: uint8(pipeStateMessageMode & 0xFF),
		Flags:  uint8(pipeStateMessageMode >> 8),
	}
	response.ExtFileAttributes = types.SMB_EXT_FILE_ATTR(fileflags.FILE_ATTRIBUTE_NORMAL)

	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer the pipe open for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// resourceTypeMessageModePipe is the ResourceType an open reports for a pipe that
// carries messages rather than a byte stream ([MS-CIFS] section 2.2.4.64.2).
const resourceTypeMessageModePipe = 0x0002
