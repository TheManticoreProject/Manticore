package server

import (
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

// TRANSACTION2 carries a subcommand whose parameter and data blocks may be larger
// than one SMB message. Both directions therefore fragment: a request arrives as a
// primary message and then as many secondary messages as it needs, and a response
// goes back the same way.
//
// The two halves are separate problems. Reassembly has to tolerate fragments
// arriving with displacements it did not choose, and must not let a client claim a
// total it never sends. Fragmentation has to fit each response inside what the
// client said it could receive. What they share is the offset arithmetic, which is
// the part that is easy to get wrong and impossible to get away with: a client
// reads the blocks at the offsets the response declares.

const (
	// trans2ResponseWordCount is the fixed word count of a TRANSACTION2 response
	// carrying no setup words, which none of the subcommands here return.
	trans2ResponseWordCount = 10

	// maxTrans2Payload bounds a reassembled request. The totals are 16-bit, so a
	// client cannot legitimately exceed this, and refusing beyond it keeps a
	// claimed total from becoming an allocation.
	maxTrans2Payload = 0xFFFF

	// transactionReassemblyTimeout bounds how long a half-delivered transaction is
	// held. A client that starts one and stops would otherwise pin the memory for
	// as long as the connection lasts.
	transactionReassemblyTimeout = 30 * time.Second
)

// transactionReassembly is a transaction being received across several messages.
//
// All three transaction families — TRANSACTION, TRANSACTION2 and NT_TRANSACT —
// carry the same shape: running totals, a per-message count, and a displacement
// saying where the message's bytes belong. They differ only in the width of those
// fields and in how the subcommand is named. So the reassembly is shared and each
// family's handler extracts the fields, which keeps the arithmetic that is easy to
// get wrong in one place.
type transactionReassembly struct {
	// family names which transaction this is, for the log line that reports a
	// malformed one.
	family string

	// subcommand is the selector the family carries: a setup word for
	// TRANSACTION2, a function code for NT_TRANSACT, a pipe operation for
	// TRANSACTION.
	subcommand uint16

	// parameters and data are the blocks being filled in, sized to the totals the
	// primary message declared.
	parameters []byte
	data       []byte

	// parametersSeen and dataSeen count the bytes placed so far, so completion is
	// known without scanning for gaps.
	parametersSeen int
	dataSeen       int

	// maxParameterCount and maxDataCount are what the client said it can receive
	// back, which bounds the response.
	maxParameterCount int
	maxDataCount      int

	// setup holds the family's setup words, which some subcommands carry their
	// arguments in rather than in the parameter block.
	setup []types.USHORT

	// name is the resource a TRANSACTION names, which is how a client says which
	// pipe it is talking to. The other two families carry no name.
	name string

	started time.Time
}

// complete reports whether every declared byte has arrived.
func (r *transactionReassembly) complete() bool {
	return r.parametersSeen >= len(r.parameters) && r.dataSeen >= len(r.data)
}

// place copies a fragment into the blocks at its declared displacement.
//
// A displacement or length that would write outside the declared total is refused
// rather than clamped: it means the client's own arithmetic disagrees with itself,
// and guessing which half to believe would let a fragment land somewhere it was
// not meant to.
func (r *transactionReassembly) place(
	parameters []byte, parameterDisplacement int,
	data []byte, dataDisplacement int,
) error {
	if err := placeFragment(r.parameters, parameters, parameterDisplacement, "parameter"); err != nil {
		return err
	}
	if err := placeFragment(r.data, data, dataDisplacement, "data"); err != nil {
		return err
	}
	r.parametersSeen += len(parameters)
	r.dataSeen += len(data)
	return nil
}

// placeFragment copies a run into a block at a displacement, refusing anything
// that would fall outside it.
func placeFragment(block, run []byte, displacement int, what string) error {
	if len(run) == 0 {
		return nil
	}
	if displacement < 0 || displacement > len(block) {
		return fmt.Errorf("%s displacement %d is outside the declared %d-byte block", what, displacement, len(block))
	}
	if displacement+len(run) > len(block) {
		return fmt.Errorf("%s fragment of %d bytes at displacement %d overruns the declared %d-byte block",
			what, len(run), displacement, len(block))
	}
	copy(block[displacement:], run)
	return nil
}

// handleTransaction2 answers SMB_COM_TRANSACTION2: the primary message of a
// transaction.
func handleTransaction2(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.Transaction2Request)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	if len(request.Setup) < 1 {
		logger.Debugf("SMB1 server: %s sent a TRANSACTION2 with no subcommand", conn.Remote)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	subcommand := uint16(request.Setup[0])

	totalParameters := int(request.TotalParameterCount)
	totalData := int(request.TotalDataCount)
	if totalParameters > maxTrans2Payload || totalData > maxTrans2Payload {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	reassembly := &transactionReassembly{
		family:            "TRANSACTION2",
		subcommand:        subcommand,
		parameters:        make([]byte, totalParameters),
		data:              make([]byte, totalData),
		maxParameterCount: int(request.MaxParameterCount),
		maxDataCount:      int(request.MaxDataCount),
		started:           time.Now(),
	}

	// The primary message carries no displacement fields: it is by definition the
	// first fragment, so both blocks start at zero.
	if err := reassembly.place(
		[]byte(request.Trans2_Parameters), 0,
		[]byte(request.Trans2_Data), 0,
	); err != nil {
		logger.Debugf("SMB1 server: %s sent an inconsistent TRANSACTION2: %v", conn.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	// The common case is a transaction that fits in one message.
	if reassembly.complete() {
		return conn.runTransaction2(w, req, reassembly)
	}

	// More is coming. Only one transaction is tracked at a time: [MS-CIFS] has a
	// client complete one before starting another, and holding several would let a
	// client pin memory by opening many and finishing none.
	conn.transaction = reassembly
	logger.Debugf("SMB1 server: %s began a fragmented TRANSACTION2 (subcommand 0x%04X, %d+%d bytes)",
		conn.Remote, subcommand, totalParameters, totalData)

	// A primary message that does not complete the transaction gets no response:
	// the client sends the rest before expecting one.
	return nt_status.NT_STATUS_SUCCESS
}

// handleTransaction2Secondary answers SMB_COM_TRANSACTION2_SECONDARY: a
// continuation of a transaction already begun.
func handleTransaction2Secondary(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.Transaction2SecondaryRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	reassembly := conn.transaction
	if reassembly == nil {
		logger.Debugf("SMB1 server: %s sent a TRANSACTION2_SECONDARY with no transaction in progress", conn.Remote)
		return nt_status.NT_STATUS_INVALID_SMB
	}
	if time.Since(reassembly.started) > transactionReassemblyTimeout {
		conn.transaction = nil
		logger.Debugf("SMB1 server: %s continued a TRANSACTION2 that had timed out", conn.Remote)
		return nt_status.NT_STATUS_IO_TIMEOUT
	}

	if err := reassembly.place(
		[]byte(request.Trans2_Parameters), int(request.ParameterDisplacement),
		[]byte(request.Trans2_Data), int(request.DataDisplacement),
	); err != nil {
		conn.transaction = nil
		logger.Debugf("SMB1 server: %s sent an inconsistent TRANSACTION2_SECONDARY: %v", conn.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	if !reassembly.complete() {
		return nt_status.NT_STATUS_SUCCESS
	}

	conn.transaction = nil
	return conn.runTransaction2(w, req, reassembly)
}

// runTransaction2 dispatches a fully assembled transaction to its subcommand and
// sends the result back, fragmenting the response if it does not fit.
func (c *Connection) runTransaction2(w ResponseWriter, req *message.Message, reassembly *transactionReassembly) nt_status.NT_STATUS {
	handler, ok := trans2Handlers[subcommands.Transaction2Subcommand(reassembly.subcommand)]
	if !ok {
		logger.Debugf("SMB1 server: %s sent unimplemented TRANSACTION2 subcommand 0x%04X",
			c.Remote, reassembly.subcommand)
		return nt_status.NT_STATUS_NOT_IMPLEMENTED
	}

	parameters, data, status := handler(c, req, reassembly)
	if status != nt_status.NT_STATUS_SUCCESS {
		return status
	}

	if err := c.sendTransaction2Response(w, reassembly, parameters, data); err != nil {
		logger.Debugf("SMB1 server: failed to answer the transaction for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// trans2Handler answers one TRANSACTION2 subcommand, returning the parameter and
// data blocks to send back.
type trans2Handler func(*Connection, *message.Message, *transactionReassembly) (parameters, data []byte, status nt_status.NT_STATUS)

// trans2Handlers maps a subcommand to its handler. A subcommand with no entry is
// answered with STATUS_NOT_IMPLEMENTED, which for the DFS and OS/2-era
// subcommands is the permanent answer rather than a placeholder.
var trans2Handlers = map[subcommands.Transaction2Subcommand]trans2Handler{
	subcommands.TRANS2_FIND_FIRST2:            handleFindFirst2,
	subcommands.TRANS2_FIND_NEXT2:             handleFindNext2,
	subcommands.TRANS2_QUERY_PATH_INFORMATION: handleQueryPathInformation,
	subcommands.TRANS2_QUERY_FILE_INFORMATION: handleQueryFileInformation,
	subcommands.TRANS2_SET_PATH_INFORMATION:   handleSetPathInformation,
	subcommands.TRANS2_SET_FILE_INFORMATION:   handleSetFileInformation,
	subcommands.TRANS2_QUERY_FS_INFORMATION:   handleQueryFsInformation,
}

// sendTransaction2Response sends a transaction result, splitting it across as many
// messages as it needs.
//
// Each message declares the running totals and its own displacement, so the client
// can place the fragments however they arrive. The offsets are absolute from the
// start of the SMB header, which is what the client reads them as.
func (c *Connection) sendTransaction2Response(
	w ResponseWriter,
	reassembly *transactionReassembly,
	parameters, data []byte,
) error {
	// What the client said it can receive bounds each message, and so does the
	// buffer the connection negotiated: a response larger than either could not be
	// read at the far end.
	budget := int(c.Server.config.MaxBufferSize) - trans2ResponseOverhead
	if budget <= 0 {
		return fmt.Errorf("the negotiated buffer of %d bytes leaves no room for a transaction response",
			c.Server.config.MaxBufferSize)
	}
	if reassembly.maxDataCount > 0 && reassembly.maxDataCount < budget {
		budget = reassembly.maxDataCount
	}

	// Parameters go first and are not split away from their transaction: every
	// subcommand here returns a handful of bytes, so a parameter block that did
	// not fit would mean the buffer is too small to talk at all.
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

		remaining := len(data) - sentData
		chunk := remaining
		if chunk > chunkBudget {
			chunk = chunkBudget
		}
		if chunk < 0 {
			chunk = 0
		}

		response := commands.NewTransaction2Response()
		response.TotalParameterCount = types.USHORT(len(parameters))
		response.TotalDataCount = types.USHORT(len(data))
		response.SetupCount = types.UCHAR(0)
		response.Setup = []types.USHORT{}

		response.ParameterCount = types.USHORT(len(parameterChunk))
		response.ParameterDisplacement = types.USHORT(0)
		response.DataCount = types.USHORT(chunk)
		response.DataDisplacement = types.USHORT(sentData)

		// The offsets: the parameter block starts after the fixed part, padded to
		// a 4-byte boundary, and the data block after the parameters, padded
		// again.
		preParameters := header.SMB_HEADER_SIZE + 1 + 2*trans2ResponseWordCount + 2
		pad1 := (4 - (preParameters % 4)) % 4
		parameterOffset := preParameters + pad1
		response.Pad1 = make([]types.UCHAR, pad1)
		response.ParameterOffset = types.USHORT(parameterOffset)
		response.Trans2_Parameters = []types.UCHAR(parameterChunk)

		afterParameters := parameterOffset + len(parameterChunk)
		pad2 := (4 - (afterParameters % 4)) % 4
		response.Pad2 = make([]types.UCHAR, pad2)
		response.DataOffset = types.USHORT(afterParameters + pad2)
		response.Trans2_Data = []types.UCHAR(data[sentData : sentData+chunk])

		if err := w.WriteResponse(response); err != nil {
			return err
		}

		sentData += chunk
		// A response with nothing left to send is finished, even if the loop
		// condition would run again on a zero-length chunk.
		if chunk == 0 && !first {
			break
		}
	}

	return nil
}

// trans2ResponseOverhead is the space a transaction response needs beyond its
// blocks: the header, the parameter words, the byte count and the two pads.
const trans2ResponseOverhead = 96
