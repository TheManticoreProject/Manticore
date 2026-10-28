package server

import (
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// CaptureConfig configures a CaptureHandler. Reporting, the output file and
// deduplication are the SMB1 capture's.
type CaptureConfig struct {
	smb1.CaptureConfig

	// PassThrough lets a recorded logon proceed through the normal policy
	// instead of refusing it with Status, so the client can still be served.
	PassThrough bool
}

// CaptureHandler harvests NTLM authentication attempts, as in the SMB1 server:
// it records the AUTHENTICATE leg of a session setup, then refuses the logon or
// lets it through. Everything else falls through to the built-in dispatch.
type CaptureHandler struct {
	*smb1.CaptureHandler
	status      nt_status.NT_STATUS
	passThrough bool
}

// NewCaptureHandler creates a capture handler.
func NewCaptureHandler(config CaptureConfig) (*CaptureHandler, error) {
	inner, err := smb1.NewCaptureHandler(config.CaptureConfig)
	if err != nil {
		return nil, err
	}
	status := config.Status
	if status == 0 {
		status = nt_status.NT_STATUS_LOGON_FAILURE
	}
	return &CaptureHandler{CaptureHandler: inner, status: status, passThrough: config.PassThrough}, nil
}

// Run records the AUTHENTICATE leg of a session setup.
func (h *CaptureHandler) Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
	if req.Header.Command != codes.SMB2_SESSION_SETUP {
		return false
	}
	accept := conn.PendingNTLM(req.Header.SessionId)
	setup, ok := req.Command.(*commands.SessionSetupRequest)
	if accept == nil || !ok || accept.AcceptAuthenticateToken(setup.SecurityBuffer) != nil {
		return false
	}

	h.Record(w.RemoteAddr(), accept)
	if h.passThrough {
		return false
	}
	w.WriteError(h.status)
	return true
}

var _ Handler = (*CaptureHandler)(nil)
