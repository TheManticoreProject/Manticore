package nbns

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// Default TCP port for NetBIOS name service
	DefaultNBNSTCPPort = 137

	// TCP timeouts
	TCPReadTimeout  = 30 * time.Second
	TCPWriteTimeout = 30 * time.Second

	// Maximum message size
	MaxTCPMessageSize = 65535
)

// TCPServer represents a NetBIOS Name Server TCP component
type TCPServer struct {
	nbns     *NetBIOSNameServer
	listener net.Listener
	addr     string
	wg       sync.WaitGroup
	quit     chan struct{}
	tcpConns sync.Map // Track active TCP connections
	handlers *PacketHandler
}

// NewTCPServer creates a new NBNS TCP server instance
func NewTCPServer(addr string, nbns *NetBIOSNameServer) (*TCPServer, error) {
	return &TCPServer{
		nbns:     nbns,
		addr:     addr,
		quit:     make(chan struct{}),
		handlers: NewPacketHandler(nbns),
	}, nil
}

// EnableNodeStatus turns on the NODE STATUS responder so an NBSTAT (0x0021)
// query is answered from the local name table (RFC 1002 4.2.18). mac is reported
// as the STATISTICS UNIT_ID; a nil or non-6-byte mac reports a zeroed UNIT_ID.
// Node status is off by default.
func (s *TCPServer) EnableNodeStatus(mac net.HardwareAddr) {
	s.handlers.EnableNodeStatus(mac)
}

// SetRedirectManager installs a redirect manager so a NAME QUERY for a
// configured scope is answered with a REDIRECT NAME QUERY RESPONSE (RFC 1002
// 4.2.14). Passing nil disables redirection. Off by default.
func (s *TCPServer) SetRedirectManager(r *RedirectManager) {
	s.handlers.SetRedirectManager(r)
}

// EnableNameDefense installs a NameChallenger so a conflicting registration of an
// owned name triggers an END-NODE CHALLENGE before the registration is refused
// or accepted (RFC 1002 4.2.10), and a conflict-demand sender so MarkNameConflict
// emits a NAME CONFLICT DEMAND to the offending owner. The TCP message path
// returns a single framed response, so no intermediate WACK is emitted here; the
// challenge still runs. Off by default.
func (s *TCPServer) EnableNameDefense() {
	s.handlers.SetChallenger(NewNameChallenger(s.nbns, s.handlers))
	s.nbns.SetConflictDemandSender(sendConflictDemandUDP)
}

// Start begins listening for TCP connections
func (s *TCPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %v", err)
	}

	s.wg.Add(1)
	go s.serve()

	log.Printf("NBNS TCP server listening on %s", s.addr)
	return nil
}

// Stop gracefully shuts down the server
func (s *TCPServer) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all active connections
	s.tcpConns.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			conn.Close()
		}
		return true
	})

	s.wg.Wait()
}

// serve handles incoming TCP connections
func (s *TCPServer) serve() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					continue
				}
				// Non-temporary error or server shutdown
				return
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection processes a single TCP connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	// Store connection
	s.tcpConns.Store(conn.RemoteAddr().String(), conn)
	defer s.tcpConns.Delete(conn.RemoteAddr().String())

	for {
		select {
		case <-s.quit:
			return
		default:
			if err := conn.SetReadDeadline(time.Now().Add(TCPReadTimeout)); err != nil {
				log.Printf("Failed to set TCP read deadline: %v", err)
				return
			}

			// Read message length (2 bytes)
			lenBuf := make([]byte, 2)
			if _, err := io.ReadFull(conn, lenBuf); err != nil {
				if err != io.EOF {
					log.Printf("Failed to read message length: %v", err)
				}
				return
			}

			msgLen := binary.BigEndian.Uint16(lenBuf)
			if msgLen > MaxTCPMessageSize {
				log.Printf("Message too large: %d bytes", msgLen)
				return
			}

			// Read the message
			msgBuf := make([]byte, msgLen)
			if _, err := io.ReadFull(conn, msgBuf); err != nil {
				log.Printf("Failed to read message: %v", err)
				return
			}

			// Process the message
			response, err := s.handleMessage(msgBuf)
			if err != nil {
				log.Printf("Failed to handle message: %v", err)
				return
			}

			// Send response length
			if err := conn.SetWriteDeadline(time.Now().Add(TCPWriteTimeout)); err != nil {
				log.Printf("Failed to set TCP write deadline: %v", err)
				return
			}

			respLen := make([]byte, 2)
			binary.BigEndian.PutUint16(respLen, uint16(len(response)))
			if _, err := conn.Write(respLen); err != nil {
				log.Printf("Failed to write response length: %v", err)
				return
			}

			// Send response
			if _, err := conn.Write(response); err != nil {
				log.Printf("Failed to write response: %v", err)
				return
			}
		}
	}
}

// handleMessage processes a single message and returns the response
func (s *TCPServer) handleMessage(data []byte) ([]byte, error) {
	var packet NBNSPacket
	bytesRead, err := packet.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal packet: %v", err)
	}

	if bytesRead != len(data) {
		return nil, fmt.Errorf("truncated packet: expected %d bytes, got %d", len(data), bytesRead)
	}

	// Create response packet
	response := &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: packet.Header.TransactionID,
			Flags:         FlagResponse | FlagAuthoritative,
			Questions:     0,
		},
	}

	// Process based on operation code
	switch packet.Header.Flags & OpcodeMask {
	case OpNameQuery:
		if s.handlers.nodeStatusEnabled && s.handlers.isNodeStatusQuery(&packet) {
			// A NODE STATUS REQUEST shares the query opcode and is distinguished
			// only by its NBSTAT question type; answer it from the name table.
			s.handlers.handleNodeStatus(&packet, response)
		} else {
			s.handlers.handleNameQueryWithRedirect(&packet, response)
		}
	case OpRegistration:
		if s.handlers.challenger != nil {
			// The TCP path returns a single framed response, so no intermediate
			// WACK is emitted (sendWACK is nil); the challenge still runs.
			s.handlers.handleRegistrationWithChallenge(&packet, response, nil)
		} else {
			s.handlers.handleRegistration(&packet, response)
		}
	case OpRelease:
		s.handlers.handleRelease(&packet, response)
	case OpRefresh:
		s.handlers.handleRefresh(&packet, response)
	default:
		response.Header.Flags |= RcodeNotImpl
	}

	return response.Marshal()
}
