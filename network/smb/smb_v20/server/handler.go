package server

import (
	"net"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// Handler observes or answers a request before the built-in dispatch sees it,
// as in the SMB1 server.
//
// Run reports whether it handled the request: true stops the chain and the
// built-in dispatch, false passes the request on. Handlers run before the
// session, signing and tree checks, so one that answers is responsible for its
// own; one that only observes simply returns false.
//
// Handlers are shared across connections and run on the goroutine serving the
// connection, so one that keeps state must guard it.
type Handler interface {
	Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool

// Run calls f.
func (f HandlerFunc) Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
	return f(srv, conn, w, req)
}

// ResponseWriter sets the response to the request being handled. The server
// builds the correlated header, signs or encrypts it, and frames it with the
// rest of a compounded message. A handler that returns true without writing
// sends no response at all.
type ResponseWriter interface {
	// RemoteAddr returns the address of the client being answered.
	RemoteAddr() net.Addr

	// WriteResponse answers with cmd and STATUS_SUCCESS.
	WriteResponse(cmd command_interface.CommandInterface)

	// WriteResponseWithStatus answers with cmd and a status that still carries
	// a body, such as STATUS_MORE_PROCESSING_REQUIRED.
	WriteResponseWithStatus(cmd command_interface.CommandInterface, status nt_status.NT_STATUS)

	// WriteError answers with an ERROR response carrying status.
	WriteError(status nt_status.NT_STATUS)

	// SetResponseSessionID sets the SessionId the response carries, for the
	// session setup leg that assigns one.
	SetResponseSessionID(id uint64)
}

// responseWriter records what a handler answered.
type responseWriter struct {
	remote    net.Addr
	command   command_interface.CommandInterface
	status    nt_status.NT_STATUS
	written   bool
	sessionID uint64
}

func (w *responseWriter) RemoteAddr() net.Addr { return w.remote }

func (w *responseWriter) WriteResponse(cmd command_interface.CommandInterface) {
	w.WriteResponseWithStatus(cmd, nt_status.NT_STATUS_SUCCESS)
}

func (w *responseWriter) WriteResponseWithStatus(cmd command_interface.CommandInterface, status nt_status.NT_STATUS) {
	w.command, w.status, w.written = cmd, status, true
}

func (w *responseWriter) WriteError(status nt_status.NT_STATUS) {
	w.WriteResponseWithStatus(nil, status)
}

func (w *responseWriter) SetResponseSessionID(id uint64) { w.sessionID = id }

// RegisterHandler appends a handler to the chain run before dispatch.
func (s *Server) RegisterHandler(handler Handler) {
	if handler == nil {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handlers = append(s.handlers, handler)
}

// runHandlers offers a decoded request to the handler chain and reports whether
// one of them took it, with the response it wrote.
func (c *Connection) runHandlers(r *request) (bool, *responseWriter) {
	if r.command == nil {
		return false, nil
	}
	c.server.mutex.RLock()
	handlers := c.server.handlers
	c.server.mutex.RUnlock()

	req := &message.Message{Header: r.header, Command: r.command}
	for _, h := range handlers {
		w := &responseWriter{remote: c.remote}
		if h.Run(c.server, c, w, req) {
			return true, w
		}
	}
	return false, nil
}

// RemoteAddr returns the address of the client.
func (c *Connection) RemoteAddr() net.Addr { return c.remote }

// PendingNTLM returns the NTLM exchange awaiting its AUTHENTICATE on a session,
// or nil. A handler reads the challenge the client answers from it.
func (c *Connection) PendingNTLM(sessionID uint64) *spnego.AcceptContext {
	if session := c.sessions[sessionID]; session != nil {
		return session.ntlm
	}
	return nil
}
