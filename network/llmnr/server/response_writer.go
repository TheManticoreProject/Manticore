package server

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// ResponseWriter interface is used by an LLMNR handler to construct a response
type ResponseWriter interface {
	WriteMessage(*message.Message) error
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
//
// A UDP datagram cannot carry more than constants.MaxPacketSize bytes, so the
// response is encoded with MarshalWithTruncation: if the full response would
// exceed that limit it is truncated and the TC (truncation) bit is set, which
// tells the querying host to retry the same query over TCP (RFC 4795 §2.4) where
// the complete answer is served by the TCP responder.
func (w *responseWriter) WriteMessage(msg *message.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	msg.SetResponse()

	encoded, _, err := msg.MarshalWithTruncation(constants.MaxPacketSize)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	// The UDP responder can only reply over its net.UDPConn, so the remote
	// address must be a *net.UDPAddr. Assert with the comma-ok form and return a
	// descriptive error rather than risk a panic on an unexpected address type.
	udpAddr, ok := w.RemoteAddr.(*net.UDPAddr)
	if !ok {
		return fmt.Errorf("cannot write UDP response: remote address is %T, want *net.UDPAddr", w.RemoteAddr)
	}

	_, err = w.Server.Conn.WriteToUDP(encoded, udpAddr)

	return err
}

// tcpResponseWriter is the ResponseWriter used by the TCP responder. Unlike the
// UDP writer it never truncates: TCP has no 512-byte datagram limit, so the full
// answer is written framed with the DNS-over-TCP two-byte length prefix
// (RFC 1035 §4.2.2). Serving the complete answer over TCP is the whole point of
// the TC-triggered fallback (RFC 4795 §2.4).
type tcpResponseWriter struct {
	conn net.Conn
}

// newTCPResponseWriter creates a ResponseWriter that writes responses back over
// the given TCP connection.
func newTCPResponseWriter(conn net.Conn) ResponseWriter {
	return &tcpResponseWriter{conn: conn}
}

// GetRemoteAddr returns the address of the connected TCP peer.
func (w *tcpResponseWriter) GetRemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

// WriteMessage marks msg as a response, encodes it in full (no truncation), and
// writes it to the TCP connection framed with the two-byte big-endian length
// prefix.
func (w *tcpResponseWriter) WriteMessage(msg *message.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	msg.SetResponse()

	encoded, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	return message.WriteTCPMessage(w.conn, encoded)
}
