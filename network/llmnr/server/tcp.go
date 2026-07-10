package server

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// ListenAndServeTCP starts the optional LLMNR TCP responder and blocks serving
// it until the server is closed. RFC 4795 §2.4 requires a responder to listen on
// TCP port 5355 on the unicast address(es) it answers from, so that a querying
// host whose UDP response was truncated (TC bit set) can retry the query over
// TCP and receive the complete answer.
//
// It binds s.TCPListenAddr (defaulting to ":5355" when empty), then accepts
// connections and serves each with the same handler chain as the UDP path. It is
// independent of and non-breaking to ListenAndServe: a caller that wants both
// transports starts ListenAndServe and ListenAndServeTCP in separate goroutines,
// and Close shuts down both. Readiness can be observed with ListeningTCP.
//
// Returns:
//   - An error if no handlers are registered or the TCP socket cannot be bound.
//     A nil error is returned after a clean Close.
func (s *Server) ListenAndServeTCP() error {
	if len(s.snapshotHandlers()) == 0 {
		return fmt.Errorf("no handlers registered")
	}

	addr := s.TCPListenAddr
	if addr == "" {
		addr = fmt.Sprintf(":%d", constants.ListenPort)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP: %w", err)
	}
	s.tcpListener = listener

	s.listeningMu.Lock()
	s.listeningTCP = true
	s.listeningMu.Unlock()

	return s.serveTCP()
}

// ListeningTCP reports whether the server has bound its TCP responder socket and
// is ready to accept connections. Like Listening, it is safe to call
// concurrently with ListenAndServeTCP so a caller that started the TCP responder
// in a background goroutine can poll for readiness without racing on the
// tcpListener field.
func (s *Server) ListeningTCP() bool {
	s.listeningMu.RLock()
	defer s.listeningMu.RUnlock()
	return s.listeningTCP
}

// TCPAddr returns the address the TCP responder is bound to, or nil when the TCP
// responder has not been started. It is primarily useful when TCPListenAddr
// requested an ephemeral port (e.g. "127.0.0.1:0" in tests) and the caller needs
// the concrete port that was assigned.
func (s *Server) TCPAddr() net.Addr {
	if s.tcpListener == nil {
		return nil
	}
	return s.tcpListener.Addr()
}

// serveTCP accepts and dispatches TCP connections until the server is closed.
// Each accepted connection is handled in its own goroutine so a slow or stalled
// client cannot block others.
func (s *Server) serveTCP() error {
	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			// A closed listener (from Close) surfaces here as an accept error;
			// treat that as a clean shutdown rather than a failure.
			select {
			case <-s.Closed:
				return nil
			default:
			}
			if s.Debug {
				logger.Debugf("Error accepting TCP connection: %s\n", err.Error())
			}
			continue
		}
		go s.handleTCPConn(conn)
	}
}

// handleTCPConn serves a single LLMNR-over-TCP connection: it reads one
// length-prefixed query (RFC 1035 §4.2.2 framing), runs the handler chain if the
// message is a query, and lets the handlers write a length-prefixed response
// through the TCP ResponseWriter. The connection is closed when handling
// completes.
func (s *Server) handleTCPConn(conn net.Conn) {
	defer conn.Close()

	payload, err := message.ReadTCPMessage(conn)
	if err != nil {
		if s.Debug {
			logger.Debugf("Error reading TCP query: %s\n", err.Error())
		}
		return
	}

	msg := message.Message{}
	if _, err := msg.Unmarshal(payload); err != nil {
		if s.Debug {
			logger.Debugf("Error decoding TCP message: %s\n", err.Error())
		}
		return
	}

	if !msg.IsQuery() {
		if s.Debug {
			logger.Debugf("Received non-query TCP message from %s\n", conn.RemoteAddr().String())
		}
		return
	}
	if s.Debug {
		logger.Debugf("Received query TCP message from %s\n", conn.RemoteAddr().String())
	}

	writer := newTCPResponseWriter(conn)
	s.processHandlers(s, conn.RemoteAddr(), writer, &msg)
}
