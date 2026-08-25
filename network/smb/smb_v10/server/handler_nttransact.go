package server

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// ntTransactResponseWordCount is the fixed word count of an NT_TRANSACT response
// carrying no setup words, which none of the subcommands here return.
const ntTransactResponseWordCount = 18

// ntTransactResponseOverhead is the space an NT_TRANSACT response needs beyond its
// blocks. Its fixed part is larger than TRANSACTION2's, because its counts are
// 32-bit.
const ntTransactResponseOverhead = 128

// handleNtTransact answers SMB_COM_NT_TRANSACT: the primary message of an
// NT_TRANSACT.
//
// The family is the same shape as TRANSACTION2 with wider fields, so the
// reassembly is shared and only the extraction differs. The subcommand travels in
// its own Function field rather than in a setup word.
func handleNtTransact(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.NtTransactRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	totalParameters := int(request.TotalParameterCount)
	totalData := int(request.TotalDataCount)
	// The counts are 32-bit here, so the bound matters more than it did for
	// TRANSACTION2: a claimed total is an allocation.
	if totalParameters < 0 || totalData < 0 ||
		totalParameters > maxNtTransactPayload || totalData > maxNtTransactPayload {
		logger.Debugf("SMB1 server: %s declared an NT_TRANSACT of %d+%d bytes, over the limit",
			conn.Remote, totalParameters, totalData)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	reassembly := &transactionReassembly{
		family:            "NT_TRANSACT",
		subcommand:        uint16(request.Function),
		parameters:        make([]byte, totalParameters),
		data:              make([]byte, totalData),
		maxParameterCount: int(request.MaxParameterCount),
		maxDataCount:      int(request.MaxDataCount),
		setup:             request.Setup,
		started:           time.Now(),
	}

	if err := reassembly.place(
		[]byte(request.NT_Trans_Parameters), 0,
		[]byte(request.NT_Trans_Data), 0,
	); err != nil {
		logger.Debugf("SMB1 server: %s sent an inconsistent NT_TRANSACT: %v", conn.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	if reassembly.complete() {
		return conn.runNtTransact(w, req, reassembly)
	}

	conn.transaction = reassembly
	logger.Debugf("SMB1 server: %s began a fragmented NT_TRANSACT (function 0x%04X, %d+%d bytes)",
		conn.Remote, reassembly.subcommand, totalParameters, totalData)
	return nt_status.NT_STATUS_SUCCESS
}

// maxNtTransactPayload bounds a reassembled NT_TRANSACT. Its counts are 32-bit, so
// unlike TRANSACTION2 there is no natural ceiling and one has to be imposed:
// without it a client could declare four gigabytes and have it allocated.
const maxNtTransactPayload = 1 << 20

// handleNtTransactSecondary answers SMB_COM_NT_TRANSACT_SECONDARY.
func handleNtTransactSecondary(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.NtTransactSecondaryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	reassembly := conn.transaction
	if reassembly == nil || reassembly.family != "NT_TRANSACT" {
		logger.Debugf("SMB1 server: %s sent an NT_TRANSACT_SECONDARY with no matching transaction in progress",
			conn.Remote)
		return nt_status.NT_STATUS_INVALID_SMB
	}
	if time.Since(reassembly.started) > transactionReassemblyTimeout {
		conn.transaction = nil
		return nt_status.NT_STATUS_IO_TIMEOUT
	}

	if err := reassembly.place(
		[]byte(request.NT_Trans_Parameters), int(request.ParameterDisplacement),
		[]byte(request.NT_Trans_Data), int(request.DataDisplacement),
	); err != nil {
		conn.transaction = nil
		logger.Debugf("SMB1 server: %s sent an inconsistent NT_TRANSACT_SECONDARY: %v", conn.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	if !reassembly.complete() {
		return nt_status.NT_STATUS_SUCCESS
	}

	conn.transaction = nil
	return conn.runNtTransact(w, req, reassembly)
}

// runNtTransact dispatches an assembled NT_TRANSACT to its subcommand.
func (c *Connection) runNtTransact(w ResponseWriter, req *message.Message, reassembly *transactionReassembly) nt_status.NT_STATUS {
	handler, ok := ntTransactHandlers[subcommands.NtTransactSubcommand(reassembly.subcommand)]
	if !ok {
		logger.Debugf("SMB1 server: %s sent unimplemented NT_TRANSACT function 0x%04X",
			c.Remote, reassembly.subcommand)
		return nt_status.NT_STATUS_NOT_IMPLEMENTED
	}

	parameters, data, status := handler(c, req, reassembly)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	if err := c.sendNtTransactResponse(w, reassembly, parameters, data); err != nil {
		logger.Debugf("SMB1 server: failed to answer the NT_TRANSACT for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// ntTransactHandler answers one NT_TRANSACT subcommand.
type ntTransactHandler func(*Connection, *message.Message, *transactionReassembly) (parameters, data []byte, status nt_status.NT_STATUS)

// ntTransactHandlers maps a function code to its handler.
//
// NT_TRANSACT_CREATE, RENAME and the quota subcommands are absent deliberately.
// Create and rename duplicate commands that already exist in their own right, and
// nothing here tracks a quota, so answering them would mean inventing a number a
// client would then believe.
var ntTransactHandlers = map[subcommands.NtTransactSubcommand]ntTransactHandler{
	subcommands.NT_TRANSACT_QUERY_SECURITY_DESC: handleQuerySecurityDescriptor,
	subcommands.NT_TRANSACT_SET_SECURITY_DESC:   handleSetSecurityDescriptor,
	subcommands.NT_TRANSACT_IOCTL:               handleNtTransactIoctl,
}

// sendNtTransactResponse sends an NT_TRANSACT result, splitting it across as many
// messages as it needs.
func (c *Connection) sendNtTransactResponse(
	w ResponseWriter,
	reassembly *transactionReassembly,
	parameters, data []byte,
) error {
	budget := int(c.Server.config.MaxBufferSize) - ntTransactResponseOverhead
	if budget <= 0 {
		return fmt.Errorf("the negotiated buffer of %d bytes leaves no room for an NT_TRANSACT response",
			c.Server.config.MaxBufferSize)
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

		response := commands.NewNtTransactResponse()
		response.TotalParameterCount = types.ULONG(len(parameters))
		response.TotalDataCount = types.ULONG(len(data))
		response.SetupCount = types.UCHAR(0)
		response.Setup = []types.USHORT{}

		response.ParameterCount = types.ULONG(len(parameterChunk))
		response.ParameterDisplacement = types.ULONG(0)
		response.DataCount = types.ULONG(chunk)
		response.DataDisplacement = types.ULONG(sentData)

		preParameters := header.SMB_HEADER_SIZE + 1 + 2*ntTransactResponseWordCount + 2
		pad1 := (4 - (preParameters % 4)) % 4
		parameterOffset := preParameters + pad1
		response.Pad1 = make([]types.UCHAR, pad1)
		response.ParameterOffset = types.ULONG(parameterOffset)
		response.Parameters = []types.UCHAR(parameterChunk)

		afterParameters := parameterOffset + len(parameterChunk)
		pad2 := (4 - (afterParameters % 4)) % 4
		response.Pad2 = make([]types.UCHAR, pad2)
		response.DataOffset = types.ULONG(afterParameters + pad2)
		response.Data = []types.UCHAR(data[sentData : sentData+chunk])

		if err := w.WriteResponse(response); err != nil {
			return err
		}

		sentData += chunk
		if chunk == 0 && !first {
			break
		}
	}

	return nil
}

// handleQuerySecurityDescriptor answers NT_TRANSACT_QUERY_SECURITY_DESC.
//
// Request parameters: FID(2) Reserved(2) SecurityInformation(4).
func handleQuerySecurityDescriptor(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 8 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	fid := binary.LittleEndian.Uint16(parameters[0:2])
	information := SecurityInformation(binary.LittleEndian.Uint32(parameters[4:8]))

	open, status := conn.openFor(req, fid)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	provider := open.Tree.Share.Security
	if provider == nil {
		// Saying so is better than inventing a descriptor: a client that receives
		// one believes it describes the access the server enforces.
		logger.Debugf("SMB1 server: %s asked for a security descriptor on share %q, which has no security model",
			conn.Remote, open.Tree.Share.Name)
		return nil, nil, nt_status.NT_STATUS_NOT_SUPPORTED
	}

	descriptor, err := provider.SecurityDescriptor(open.Path, information)
	if err != nil {
		return nil, nil, statusForFSError(err)
	}

	// Response parameters: the length of the descriptor. A client that asked for
	// less than it takes uses this to size a second request.
	responseParameters := make([]byte, 4)
	binary.LittleEndian.PutUint32(responseParameters, uint32(len(descriptor)))

	// A client whose buffer is too small is told the size and given no data,
	// rather than a truncated descriptor it would parse as corrupt.
	if reassembly.maxDataCount > 0 && len(descriptor) > reassembly.maxDataCount {
		return responseParameters, nil, nt_status.NT_STATUS_BUFFER_TOO_SMALL
	}

	return responseParameters, descriptor, nt_status.NT_STATUS_SUCCESS
}

// handleSetSecurityDescriptor answers NT_TRANSACT_SET_SECURITY_DESC.
//
// Request parameters: FID(2) Reserved(2) SecurityInformation(4).
func handleSetSecurityDescriptor(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	parameters := reassembly.parameters
	if len(parameters) < 8 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	fid := binary.LittleEndian.Uint16(parameters[0:2])
	information := SecurityInformation(binary.LittleEndian.Uint32(parameters[4:8]))

	open, status := conn.openFor(req, fid)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}
	if open.Tree.Share.ReadOnly {
		return nil, nil, nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED
	}

	provider := open.Tree.Share.Security
	if provider == nil {
		return nil, nil, nt_status.NT_STATUS_NOT_SUPPORTED
	}

	if err := provider.SetSecurityDescriptor(open.Path, information, reassembly.data); err != nil {
		return nil, nil, statusForFSError(err)
	}

	return nil, nil, nt_status.NT_STATUS_SUCCESS
}

// handleNtTransactIoctl answers NT_TRANSACT_IOCTL.
//
// The setup words carry FunctionCode(4) FID(2) IsFsctl(1) IsFlags(1).
func handleNtTransactIoctl(conn *Connection, req *message.Message, reassembly *transactionReassembly) ([]byte, []byte, nt_status.NT_STATUS) {
	if len(reassembly.setup) < 4 {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	// The function code spans two setup words, low half first.
	functionCode := uint32(reassembly.setup[0]) | uint32(reassembly.setup[1])<<16
	fid := uint16(reassembly.setup[2])

	open, status := conn.openFor(req, fid)
	if status != nt_status.NT_STATUS_SUCCESS {
		return nil, nil, status
	}

	handler, ok := fsctlHandlers[functionCode]
	if !ok {
		logger.Debugf("SMB1 server: %s asked for FSCTL 0x%08X, which is not served", conn.Remote, functionCode)
		// The status a client expects for a control code the device does not
		// implement, which it treats as "this feature is absent" rather than as a
		// failure of the request.
		return nil, nil, nt_status.NT_STATUS_INVALID_DEVICE_REQUEST
	}

	data, status := handler(conn, open, reassembly.data)
	return nil, data, status
}

// fsctlHandler answers one file-system control code.
type fsctlHandler func(*Connection, *Open, []byte) ([]byte, nt_status.NT_STATUS)

// File-system control codes. Only codes whose answer the server can give
// truthfully are served: a control that reported a capability the storage does not
// have would be worse than one that is absent, because a client would use it.
const (
	// fsctlGetCompression reports a file's compression state.
	fsctlGetCompression = 0x0009003C
	// fsctlIsPathnameValid reports whether a name is usable on the volume.
	fsctlIsPathnameValid = 0x0009002C
)

var fsctlHandlers = map[uint32]fsctlHandler{
	fsctlGetCompression:  fsctlHandleGetCompression,
	fsctlIsPathnameValid: fsctlHandleIsPathnameValid,
}

// fsctlHandleGetCompression reports that a file is not compressed, which is true
// of everything either backend stores.
func fsctlHandleGetCompression(conn *Connection, open *Open, input []byte) ([]byte, nt_status.NT_STATUS) {
	// COMPRESSION_FORMAT_NONE.
	return make([]byte, 2), nt_status.NT_STATUS_SUCCESS
}

// fsctlHandleIsPathnameValid reports whether a name would be accepted, which is
// exactly what the path resolver decides — so the answer is the resolver's, rather
// than a second opinion that could disagree with it.
func fsctlHandleIsPathnameValid(conn *Connection, open *Open, input []byte) ([]byte, nt_status.NT_STATUS) {
	if _, err := resolvePath(trimTerminator(string(input))); err != nil {
		return nil, nt_status.NT_STATUS_OBJECT_NAME_INVALID
	}
	return nil, nt_status.NT_STATUS_SUCCESS
}

// handleNtCancel answers SMB_COM_NT_CANCEL, which asks the server to abandon a
// request still outstanding.
//
// Nothing here leaves a request outstanding: every command is answered before the
// next is read from the connection. So there is never anything to cancel, and the
// command is accepted silently — which is what it is defined to do, since a cancel
// carries no response of its own.
func handleNtCancel(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	logger.Debugf("SMB1 server: %s cancelled PID 0x%08X MID 0x%04X, which has nothing outstanding",
		conn.Remote, req.Header.GetPID(), uint16(req.Header.MID))
	return nt_status.NT_STATUS_SUCCESS
}
