package llmnr

import (
	"fmt"
	"net"
)

// ResponseWriter interface is used by an LLMNR handler to construct a response
type ResponseWriter interface {
	WriteMessage(*Message) error
	GetRemoteAddr() net.Addr
}

type responseWriter struct {
	Server     *Server
	RemoteAddr net.Addr
}

func (w *responseWriter) GetRemoteAddr() net.Addr {
	return w.RemoteAddr
}

// NewResponseWriter creates a new ResponseWriter instance.
//
// Parameters:
// - server: The Server instance that received the query.
// - remoteAddr: The address of the client that sent the query.
//
// Returns:
// - A new ResponseWriter instance.
func NewResponseWriter(server *Server, remoteAddr net.Addr) ResponseWriter {
	return &responseWriter{
		Server:     server,
		RemoteAddr: remoteAddr,
	}
}

// WriteMessage sends a response message to the remote address associated with the responseWriter.
//
// Parameters:
// - msg: The message to be sent. It must not be nil.
//
// Returns:
// - An error if the message is nil, if encoding the message fails, or if sending the message fails.
//
// The function sets the message as a response, encodes it, and sends it to the remote address using the server's UDP connection.
func (w *responseWriter) WriteMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	msg.SetResponse()

	encoded, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	_, err = w.Server.Conn.WriteToUDP(encoded, w.RemoteAddr.(*net.UDPAddr))

	return err
}
