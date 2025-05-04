package nbtns

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// Default UDP port for NetBIOS name service
	DefaultNBTNSUDPPort = 137

	// UDP timeouts and buffer sizes
	UDPReadTimeout  = 5 * time.Second
	UDPWriteTimeout = 5 * time.Second
	MaxUDPSize      = 576 // Minimum reassembly buffer size per RFC 1001
)

// UDPServer represents a NetBIOS Name Server UDP component
type UDPServer struct {
	nbtns    *NetBIOSNameServer
	conn     *net.UDPConn
	addr     *net.UDPAddr
	wg       sync.WaitGroup
	quit     chan struct{}
	handlers *PacketHandler
}

// NewUDPServer creates a new NBTNS UDP server instance
func NewUDPServer(addr string, nbtns *NetBIOSNameServer) (*UDPServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	return &UDPServer{
		nbtns:    nbtns,
		addr:     udpAddr,
		quit:     make(chan struct{}),
		handlers: NewPacketHandler(nbtns),
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

	log.Printf("NBTNS UDP server listening on %s", s.addr)
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
	var packet NBTNSPacket
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
	response := &NBTNSPacket{
		Header: NBTNSHeader{
			TransactionID: packet.Header.TransactionID,
			Flags:         FlagResponse | FlagAuthoritative,
			Questions:     packet.Header.Questions,
		},
	}

	// Process based on operation code
	switch packet.Header.Flags & 0xF000 {
	case OpNameQuery:
		s.handlers.handleNameQuery(&packet, response)
	case OpRegistration:
		s.handlers.handleRegistration(&packet, response)
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
