package nbdgm

import (
	"fmt"
	"math/rand"
	"net"
)

// LimitedBroadcastAddr is the IPv4 limited-broadcast address a BROADCAST
// datagram is sent to (RFC 1002 5.4). With DefaultDatagramPort it forms the
// destination 255.255.255.255:138.
const LimitedBroadcastAddr = "255.255.255.255"

// DefaultMaxFragmentSize is the maximum size of a single outbound UDP datagram
// (header plus trailer) when the sender fragments an oversized USER_DATA. It is
// the RFC 1001 minimum reassembly buffer size (576), which every conforming
// receiver must accept.
const DefaultMaxFragmentSize = 576

// Sender emits NetBIOS datagrams. It carries the source identity stamped into
// every datagram's header (SOURCE_NAME, SOURCE_IP, SOURCE_PORT and the SNT
// node type) and fragments a USER_DATA payload too large for a single UDP
// datagram. A Sender is a lightweight value: each send opens its own
// short-lived UDP socket, so one Sender is safe to reuse across sequential
// sends.
type Sender struct {
	// SourceName is the sending application's NetBIOS name, carried as the
	// SOURCE_NAME of DIRECT/BROADCAST datagrams.
	SourceName Name

	// SourceIP and SourcePort populate the datagram header's SOURCE_IP and
	// SOURCE_PORT. SourceIP must be an IPv4 address. SourcePort is typically
	// DefaultDatagramPort.
	SourceIP   net.IP
	SourcePort uint16

	// NodeType is the SNT source end-node type stamped into the FLAGS field
	// (NodeTypeB by default).
	NodeType uint8

	// MaxFragmentSize bounds the size of each emitted UDP datagram. A
	// non-positive value uses DefaultMaxFragmentSize.
	MaxFragmentSize int
}

// maxFragment returns the effective per-packet size limit.
func (s *Sender) maxFragment() int {
	if s.MaxFragmentSize <= 0 {
		return DefaultMaxFragmentSize
	}
	return s.MaxFragmentSize
}

// SendDirectUnique sends a DIRECT_UNIQUE datagram carrying userData to dest,
// which may be a bare host or host:port (the port defaults to 138). userData is
// fragmented across as many UDP datagrams as needed.
func (s *Sender) SendDirectUnique(dest string, destName Name, userData []byte) error {
	return s.sendDirect(MsgTypeDirectUnique, dest, destName, userData, false)
}

// SendDirectGroup sends a DIRECT_GROUP datagram carrying userData to dest (see
// SendDirectUnique for the dest and fragmentation semantics).
func (s *Sender) SendDirectGroup(dest string, destName Name, userData []byte) error {
	return s.sendDirect(MsgTypeDirectGroup, dest, destName, userData, false)
}

// SendBroadcast sends a BROADCAST datagram carrying userData to the IPv4
// limited-broadcast address on port 138. userData is fragmented as needed.
func (s *Sender) SendBroadcast(destName Name, userData []byte) error {
	dest := net.JoinHostPort(LimitedBroadcastAddr, fmt.Sprintf("%d", DefaultDatagramPort))
	return s.sendDirect(MsgTypeBroadcast, dest, destName, userData, true)
}

// sendDirect builds, fragments and transmits a DIRECT/BROADCAST datagram.
func (s *Sender) sendDirect(msgType uint8, dest string, destName Name, userData []byte, broadcast bool) error {
	if s.SourceIP.To4() == nil {
		return fmt.Errorf("Sender.SourceIP must be an IPv4 address, got %v", s.SourceIP)
	}

	addr, err := resolveDest(dest)
	if err != nil {
		return err
	}

	fragments, err := s.Fragment(msgType, destName, userData)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return fmt.Errorf("failed to open UDP socket: %w", err)
	}
	defer conn.Close()

	if broadcast {
		if err := enableBroadcast(conn); err != nil {
			return fmt.Errorf("failed to enable broadcast on UDP socket: %w", err)
		}
	}

	for i, frag := range fragments {
		if _, err := conn.WriteToUDP(frag, addr); err != nil {
			return fmt.Errorf("failed to send datagram fragment %d/%d: %w", i+1, len(fragments), err)
		}
	}
	return nil
}

// Fragment builds the wire-form UDP payloads for a DIRECT/BROADCAST datagram,
// splitting userData across as many packets as the per-packet size limit
// requires. The first fragment carries SOURCE_NAME and DESTINATION_NAME and the
// FIRST flag; every fragment but the last sets the MORE flag; PACKET_OFFSET on
// each fragment is the byte offset of its USER_DATA within the whole payload.
// All fragments share one freshly generated DGM_ID.
func (s *Sender) Fragment(msgType uint8, destName Name, userData []byte) ([][]byte, error) {
	if !isDirect(msgType) {
		return nil, fmt.Errorf("Fragment: MSG_TYPE 0x%02x is not a DIRECT/BROADCAST datagram", msgType)
	}

	// Compute the USER_DATA capacity of the first packet (which also carries the
	// two names) and of the subsequent name-free packets.
	src, err := s.SourceName.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding SOURCE_NAME: %w", err)
	}
	dst, err := destName.encode()
	if err != nil {
		return nil, fmt.Errorf("encoding DESTINATION_NAME: %w", err)
	}

	limit := s.maxFragment()
	firstCap := limit - directHeaderLen - len(src) - len(dst)
	restCap := limit - directHeaderLen
	if firstCap <= 0 || restCap <= 0 {
		return nil, fmt.Errorf("MaxFragmentSize %d too small for datagram header and names", limit)
	}

	dgmID := uint16(rand.Intn(0x10000))
	flagsBase := (s.NodeType << sntShift) & sntMask

	var fragments [][]byte
	offset := 0
	first := true
	for {
		chunkCap := restCap
		if first {
			chunkCap = firstCap
		}

		end := offset + chunkCap
		if end > len(userData) {
			end = len(userData)
		}
		chunk := userData[offset:end]

		flags := flagsBase
		if first {
			flags |= FlagFirst
		}
		if end < len(userData) {
			flags |= FlagMore
		}

		d := &Datagram{
			MsgType:      msgType,
			Flags:        flags,
			DgmID:        dgmID,
			SourceIP:     s.SourceIP,
			SourcePort:   s.SourcePort,
			PacketOffset: uint16(offset),
			UserData:     chunk,
		}
		if first {
			d.SourceName = s.SourceName
			d.DestinationName = destName
		}

		wire, err := d.Marshal()
		if err != nil {
			return nil, err
		}
		fragments = append(fragments, wire)

		offset = end
		first = false
		if offset >= len(userData) {
			break
		}
	}

	return fragments, nil
}

// resolveDest resolves a datagram destination, accepting a bare host or a
// host:port pair and defaulting the port to 138.
func resolveDest(dest string) (*net.UDPAddr, error) {
	if _, _, err := net.SplitHostPort(dest); err != nil {
		dest = net.JoinHostPort(dest, fmt.Sprintf("%d", DefaultDatagramPort))
	}
	addr, err := net.ResolveUDPAddr("udp4", dest)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve datagram destination %q: %w", dest, err)
	}
	return addr, nil
}
