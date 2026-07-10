package nbns

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// Default UDP port for NetBIOS name service
	DefaultNBNSUDPPort = 137

	// UDP timeouts and buffer sizes
	UDPReadTimeout  = 5 * time.Second
	UDPWriteTimeout = 5 * time.Second
	MaxUDPSize      = 576 // Minimum reassembly buffer size per RFC 1001
)

// UDPServer represents a NetBIOS Name Server UDP component
type UDPServer struct {
	nbns     *NetBIOSNameServer
	conn     *net.UDPConn
	addr     *net.UDPAddr
	wg       sync.WaitGroup
	quit     chan struct{}
	handlers *PacketHandler

	// spoofHandler, when non-nil, replaces the authoritative name-table lookup
	// for the NAME QUERY opcode: name queries are answered by the poisoner
	// instead of the local name table, and unmatched queries are left
	// unanswered so legitimate resolution can still proceed.
	spoofHandler *SpoofHandler
}

// SetSpoofHandler installs an NBNS poisoning handler on the UDP server. Once
// set, NAME QUERY REQUESTs are answered by the SpoofHandler (which claims names
// the host does not own with an attacker-chosen address) rather than by the
// authoritative name table; every other opcode is handled as before. Passing
// nil restores the authoritative behaviour.
func (s *UDPServer) SetSpoofHandler(h *SpoofHandler) {
	s.spoofHandler = h
}

// EnableNodeStatus turns on the NODE STATUS responder so an NBSTAT (0x0021)
// query is answered from the local name table (RFC 1002 4.2.18). mac is reported
// as the STATISTICS UNIT_ID; a nil or non-6-byte mac reports a zeroed UNIT_ID.
// Node status is off by default.
func (s *UDPServer) EnableNodeStatus(mac net.HardwareAddr) {
	s.handlers.EnableNodeStatus(mac)
}

// SetRedirectManager installs a redirect manager so a NAME QUERY for a
// configured scope is answered with a REDIRECT NAME QUERY RESPONSE (RFC 1002
// 4.2.14). Passing nil disables redirection. Off by default.
func (s *UDPServer) SetRedirectManager(r *RedirectManager) {
	s.handlers.SetRedirectManager(r)
}

// EnableNameDefense wires the name-defence path into the UDP server: a
// NameChallenger so a conflicting registration of an owned name triggers an
// END-NODE CHALLENGE (preceded by a WACK telling the requestor to wait), and a
// conflict-demand sender so MarkNameConflict emits a NAME CONFLICT DEMAND to the
// offending owner. Both are off until this is called, leaving default behaviour
// unchanged.
func (s *UDPServer) EnableNameDefense() {
	s.handlers.SetChallenger(NewNameChallenger(s.nbns, s.handlers))
	s.nbns.SetConflictDemandSender(sendConflictDemandUDP)
}

// NewUDPServer creates a new NBNS UDP server instance
func NewUDPServer(addr string, nbns *NetBIOSNameServer) (*UDPServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	return &UDPServer{
		nbns:     nbns,
		addr:     udpAddr,
		quit:     make(chan struct{}),
		handlers: NewPacketHandler(nbns),
	}, nil
}

// Start begins listening for UDP packets
func (s *UDPServer) Start() error {
	var err error
	s.conn, err = net.ListenUDP("udp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %v", err)
	}

	s.wg.Add(1)
	go s.serve()

	log.Printf("NBNS UDP server listening on %s", s.addr)
	return nil
}

// Stop gracefully shuts down the server
func (s *UDPServer) Stop() {
	close(s.quit)
	if s.conn != nil {
		s.conn.Close()
	}
	s.wg.Wait()
}

// serve handles incoming UDP packets
func (s *UDPServer) serve() {
	defer s.wg.Done()

	buf := make([]byte, MaxUDPSize)
	for {
		select {
		case <-s.quit:
			return
		default:
			if err := s.conn.SetReadDeadline(time.Now().Add(UDPReadTimeout)); err != nil {
				log.Printf("Failed to set read deadline: %v", err)
				continue
			}

			n, remoteAddr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Failed to read UDP packet: %v", err)
				continue
			}

			// Handle the packet in a separate goroutine
			go s.handlePacket(buf[:n], remoteAddr)
		}
	}
}

// handlePacket processes a single UDP packet
func (s *UDPServer) handlePacket(data []byte, remoteAddr *net.UDPAddr) {
	var packet NBNSPacket
	bytesRead, err := packet.Unmarshal(data)
	if err != nil {
		log.Printf("Failed to unmarshal packet: %v", err)
		return
	}

	if bytesRead != len(data) {
		log.Printf("Truncated packet: expected %d bytes, got %d", len(data), bytesRead)
		return
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
		switch {
		case s.handlers.nodeStatusEnabled && s.handlers.isNodeStatusQuery(&packet):
			// A NODE STATUS REQUEST shares the query opcode and is distinguished
			// only by its NBSTAT question type; answer it from the name table.
			s.handlers.handleNodeStatus(&packet, response)
		case s.spoofHandler != nil:
			// Poisoning mode: answer only the names the operator elected to
			// spoof and stay silent otherwise, so legitimate resolution still
			// works alongside the poisoner.
			spoofed, ok := s.spoofHandler.HandleNameQuery(remoteAddr, &packet)
			if !ok {
				return
			}
			response = spoofed
		default:
			s.handlers.handleNameQueryWithRedirect(&packet, response)
		}
	case OpRegistration:
		if s.handlers.challenger != nil {
			// Defend owned names: emit an intermediate WACK to the requestor
			// before running the challenge and returning the final response.
			s.handlers.handleRegistrationWithChallenge(&packet, response, func(w *NBNSPacket) {
				s.writeResponse(w, remoteAddr)
			})
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

	// Send response
	responseData, err := response.Marshal()
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	// Check if response needs to be sent via TCP due to size
	if len(responseData) > MaxUDPSize {
		response.Header.Flags |= FlagTruncated
		responseData = responseData[:MaxUDPSize]
	}

	if err := s.conn.SetWriteDeadline(time.Now().Add(UDPWriteTimeout)); err != nil {
		log.Printf("Failed to set write deadline: %v", err)
		return
	}

	if _, err := s.conn.WriteToUDP(responseData, remoteAddr); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

// writeResponse marshals and sends a single packet to remoteAddr on the server's
// UDP socket. It is used to emit the intermediate WAIT FOR ACKNOWLEDGEMENT
// (WACK) datagram ahead of a deferred registration's final response.
func (s *UDPServer) writeResponse(packet *NBNSPacket, remoteAddr *net.UDPAddr) {
	data, err := packet.Marshal()
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}
	if err := s.conn.SetWriteDeadline(time.Now().Add(UDPWriteTimeout)); err != nil {
		log.Printf("Failed to set write deadline: %v", err)
		return
	}
	if _, err := s.conn.WriteToUDP(data, remoteAddr); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

// sendConflictDemandUDP unicasts a NAME CONFLICT DEMAND to the offending owner's
// NBNS port (137/udp). It opens a short-lived socket per demand, matching the
// fire-and-forget nature of the demand, and is the default conflict-demand
// sender wired in by EnableNameDefense.
func sendConflictDemandUDP(packet *NBNSPacket, owner net.IP) error {
	data, err := packet.Marshal()
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: owner, Port: DefaultNBNSUDPPort})
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(data)
	return err
}
