package nbdgm

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Listener socket timeouts and buffer size. MaxUDPSize is the RFC 1001 minimum
// reassembly buffer, the largest single UDP datagram a conforming datagram
// service must accept; larger logical datagrams arrive fragmented.
const (
	UDPReadTimeout = 5 * time.Second
	MaxUDPSize     = 576
)

// DatagramHandler is invoked for each fully received datagram. For the
// DIRECT/BROADCAST types the datagram is delivered only once fully reassembled,
// with the complete USER_DATA and decoded SOURCE_NAME/DESTINATION_NAME; the
// query request/response and error types are delivered as they arrive. source
// is the sender's UDP address.
type DatagramHandler func(source *net.UDPAddr, d *Datagram)

// Listener receives NetBIOS datagrams on a UDP socket (port 138 by default),
// reassembles fragmented DIRECT/BROADCAST datagrams, and hands each complete
// datagram to a callback. It mirrors the nbns UDP server: Start spawns a serve
// goroutine and Stop shuts it down. A Listener is the inbound counterpart to
// Sender.
type Listener struct {
	addr    *net.UDPAddr
	conn    *net.UDPConn
	handler DatagramHandler
	reasm   *Reassembler
	wg      sync.WaitGroup
	quit    chan struct{}
}

// NewListener creates a datagram Listener bound to addr (e.g. ":138" or
// "127.0.0.1:0" for an ephemeral port) that delivers received datagrams to
// handler.
func NewListener(addr string, handler DatagramHandler) (*Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}
	return &Listener{
		addr:    udpAddr,
		handler: handler,
		reasm:   NewReassembler(),
		quit:    make(chan struct{}),
	}, nil
}

// Start begins listening for datagrams.
func (l *Listener) Start() error {
	conn, err := net.ListenUDP("udp", l.addr)
	if err != nil {
		return fmt.Errorf("failed to start datagram listener: %w", err)
	}
	l.conn = conn
	l.wg.Add(1)
	go l.serve()
	return nil
}

// LocalAddr returns the address the listener is bound to, useful when it was
// started on an ephemeral port ("...:0").
func (l *Listener) LocalAddr() *net.UDPAddr {
	if l.conn == nil {
		return l.addr
	}
	return l.conn.LocalAddr().(*net.UDPAddr)
}

// Stop gracefully shuts down the listener.
func (l *Listener) Stop() {
	close(l.quit)
	if l.conn != nil {
		l.conn.Close()
	}
	l.wg.Wait()
}

// serve reads and dispatches inbound datagrams until Stop is called.
func (l *Listener) serve() {
	defer l.wg.Done()

	buf := make([]byte, MaxUDPSize)
	for {
		select {
		case <-l.quit:
			return
		default:
			if err := l.conn.SetReadDeadline(time.Now().Add(UDPReadTimeout)); err != nil {
				log.Printf("nbdgm: failed to set read deadline: %v", err)
				continue
			}

			n, remoteAddr, err := l.conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				// A read that fails because Stop closed the socket is an
				// expected part of shutdown, not an error to report.
				select {
				case <-l.quit:
					return
				default:
				}
				log.Printf("nbdgm: failed to read datagram: %v", err)
				continue
			}

			l.handlePacket(append([]byte(nil), buf[:n]...), remoteAddr)
		}
	}
}

// handlePacket parses one received UDP payload and dispatches it, reassembling
// fragmented DIRECT/BROADCAST datagrams via the Reassembler.
func (l *Listener) handlePacket(data []byte, remoteAddr *net.UDPAddr) {
	var d Datagram
	if _, err := d.Unmarshal(data); err != nil {
		log.Printf("nbdgm: failed to parse datagram from %s: %v", remoteAddr, err)
		return
	}

	if isDirect(d.MsgType) {
		assembled, done, err := l.reasm.Add(remoteAddr.String(), &d)
		if err != nil {
			log.Printf("nbdgm: reassembly error from %s: %v", remoteAddr, err)
			return
		}
		if !done {
			return
		}
		l.handler(remoteAddr, assembled)
		return
	}

	// ERROR and query request/response datagrams are complete in one packet.
	l.handler(remoteAddr, &d)
}
