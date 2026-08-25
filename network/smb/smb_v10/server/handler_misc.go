package server

import (
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// MaxEchoCount bounds the number of responses one SMB_COM_ECHO request can
// produce. [MS-CIFS] 2.2.4.39 puts no ceiling on EchoCount, so an unbounded
// implementation lets a client ask for 65535 responses to a single request. The
// cap keeps that from being an amplification primitive; a client asking for more
// receives this many.
const MaxEchoCount = 64

// handleEcho answers SMB_COM_ECHO. The server sends EchoCount responses, each
// numbered from 1 and each echoing the request's data back unchanged; an
// EchoCount of zero means send nothing at all.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/36242467-7c62-4041-b60f-939683cacdf2
func handleEcho(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.EchoRequest)
	if !ok {
		// The dispatch table is keyed by command code and the message layer
		// builds the matching request type, so this cannot happen without a
		// mismatch between the two.
		return nt_status.NT_STATUS_INVALID_SMB
	}

	count := int(request.EchoCount)
	if count > MaxEchoCount {
		logger.Debugf("SMB1 server: %s requested %d echo responses, capping at %d",
			conn.Remote, count, MaxEchoCount)
		count = MaxEchoCount
	}

	for sequence := 1; sequence <= count; sequence++ {
		response := commands.NewEchoResponse()
		response.SequenceNumber = types.USHORT(sequence)
		response.Data = request.Data

		if err := w.WriteResponse(response); err != nil {
			// The client has gone away mid-sequence. The receive loop will
			// observe the same failure and close the connection, so there is
			// nothing to report back to a client that is no longer listening.
			logger.Debugf("SMB1 server: failed to send echo response %d to %s: %v",
				sequence, conn.Remote, err)
			return nt_status.NT_STATUS_SUCCESS
		}
	}

	return nt_status.NT_STATUS_SUCCESS
}
