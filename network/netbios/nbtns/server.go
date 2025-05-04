package nbtns

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// Default ports for NetBIOS name service
	DefaultNBTNSPort = 137

	// Timeouts and retry counts
	ReadTimeout  = 5 * time.Second
	WriteTimeout = 5 * time.Second
)

// Server represents a NetBIOS Name Server
type Server struct {
	nbtns    *NetBIOSNameServer
	listener *net.UDPConn
	addr     *net.UDPAddr
	wg       sync.WaitGroup
	quit     chan struct{}
}

// NewServer creates a new NBTNS server instance
func NewServer(addr string, secured bool) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	return &Server{
		nbtns: NewNetBIOSNameServer(secured),
		addr:  udpAddr,
		quit:  make(chan struct{}),
	}, nil
}

// Start begins listening for NBTNS requests
func (s *Server) Start() error {
	var err error
	s.listener, err = net.ListenUDP("udp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}

	s.wg.Add(1)
	go s.serve()

	log.Printf("NBTNS server listening on %s", s.addr)
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

// serve handles incoming NBTNS requests
func (s *Server) serve() {
	defer s.wg.Done()

	buf := make([]byte, 1024)
	for {
		select {
		case <-s.quit:
			return
		default:
			if err := s.listener.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
				log.Printf("Failed to set read deadline: %v", err)
				continue
			}

			n, remoteAddr, err := s.listener.ReadFromUDP(buf)
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

// handlePacket processes a single NBTNS packet
func (s *Server) handlePacket(data []byte, remoteAddr *net.UDPAddr) {
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
		s.handleNameQuery(&packet, response)
	case OpRegistration:
		s.handleRegistration(&packet, response)
	case OpRelease:
		s.handleRelease(&packet, response)
	case OpRefresh:
		s.handleRefresh(&packet, response)
	default:
		response.Header.Flags |= RcodeNotImpl
	}

	// Send response
	responseData, err := response.Marshal()
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	if err := s.listener.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		log.Printf("Failed to set write deadline: %v", err)
		return
	}

	if _, err := s.listener.WriteToUDP(responseData, remoteAddr); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

// handleNameQuery processes a name query request
func (s *Server) handleNameQuery(request *NBTNSPacket, response *NBTNSPacket) {
	for _, q := range request.Questions {
		owners, nameType, err := s.nbtns.QueryName(q.Name.Name)
		if err != nil {
			response.Header.Flags |= RcodeNameError
			return
		}

		// Create resource record for each owner
		for _, owner := range owners {
			rr := NBTNSResourceRecord{
				Name:     q.Name,
				Type:     q.Type,
				Class:    q.Class,
				TTL:      uint32(24 * time.Hour.Seconds()), // 24 hour TTL
				RDLength: uint16(len(owner)),
				RData:    owner,
			}
			response.Answers = append(response.Answers, rr)
		}

		response.Header.Answers = uint16(len(response.Answers))

		// Set group bit if this is a group name
		if nameType == Group {
			response.Header.Flags |= 0x0080 // Group name bit
		}
	}
}

// handleRegistration processes a name registration request
func (s *Server) handleRegistration(request *NBTNSPacket, response *NBTNSPacket) {
	for _, rr := range request.Answers {
		nameType := Unique
		if request.Header.Flags&0x0080 != 0 {
			nameType = Group
		}

		err := s.nbtns.RegisterName(
			rr.Name.Name,
			nameType,
			net.IP(rr.RData),
			time.Duration(rr.TTL)*time.Second,
		)

		if err != nil {
			response.Header.Flags |= RcodeConflict
			return
		}
	}
}

// handleRelease processes a name release request
func (s *Server) handleRelease(request *NBTNSPacket, response *NBTNSPacket) {
	for _, rr := range request.Answers {
		if err := s.nbtns.ReleaseName(rr.Name.Name, net.IP(rr.RData)); err != nil {
			response.Header.Flags |= RcodeServerError
			return
		}
	}
}

// handleRefresh processes a name refresh request
func (s *Server) handleRefresh(request *NBTNSPacket, response *NBTNSPacket) {
	for _, rr := range request.Answers {
		if err := s.nbtns.RefreshName(rr.Name.Name, net.IP(rr.RData)); err != nil {
			response.Header.Flags |= RcodeServerError
			return
		}
	}
}
