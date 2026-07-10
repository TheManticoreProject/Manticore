package nbns

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

// NAME_FLAGS bits carried in each NODE_NAME entry of a NODE STATUS RESPONSE
// (RFC 1002 4.2.18). The field is a 16-bit big-endian word laid out, from the
// most-significant bit, as: G | ONT | DRG | CNF | ACT | PRM | RESERVED. The
// Group (G) bit is the most significant bit, matching the NB_FLAGS layout of an
// NB resource record.
const (
	NameFlagGroup      uint16 = 0x8000 // G:   set for a group name, clear for a unique name
	NameFlagONTMask    uint16 = 0x6000 // ONT: owner node type (see OwnerNodeType)
	NameFlagDeregister uint16 = 0x1000 // DRG: the name is in the process of being deregistered
	NameFlagConflict   uint16 = 0x0800 // CNF: the name is in conflict
	NameFlagActive     uint16 = 0x0400 // ACT: the name is active
	NameFlagPermanent  uint16 = 0x0200 // PRM: the name is a permanent node name

	// nameFlagONTShift right-aligns the two ONT bits after masking with
	// NameFlagONTMask, yielding the 0..3 owner-node-type code.
	nameFlagONTShift = 13
)

// nodeNameEntryLength is the fixed on-the-wire size of a single NODE_NAME entry
// in the NBSTAT RDATA: a 16-byte NetBIOS name (15 characters plus the 1-byte
// suffix) followed by the 2-byte NAME_FLAGS field (RFC 1002 4.2.18).
const nodeNameEntryLength = NetBIOSNameLength + 2

// unitIDLength is the size of the STATISTICS block's leading UNIT_ID field,
// which carries the adapter's 6-byte unit id (its MAC address) (RFC 1002
// 4.2.18).
const unitIDLength = 6

// nodeStatusReadBufferSize bounds a single NODE STATUS RESPONSE datagram. A
// response grows with the number of registered names and always trails the full
// STATISTICS block, so it can exceed the RFC 1001 15.1.1 minimum reassembly
// size; a full 16-bit datagram buffer avoids truncating a busy host's reply.
const nodeStatusReadBufferSize = 0xFFFF

// nodeStatusStatisticsLength is the fixed on-the-wire size of the STATISTICS
// block that trails the NODE_NAME array in an NBSTAT RESPONSE (RFC 1002
// 4.2.18): the 6-byte UNIT_ID, a 1-byte JUMPERS and 1-byte TEST_RESULT, and a
// run of 16-bit counters with two 32-bit send/receive counters, totalling 46
// bytes. The counters carry no security-relevant data, so the responder emits
// them as zero and only the UNIT_ID (adapter MAC) is populated.
const nodeStatusStatisticsLength = 46

// NodeName is one entry of a remote node's NetBIOS name table as returned in a
// NODE STATUS RESPONSE: the (space-trimmed) 15-character base name, its 16th
// service suffix byte, and the raw NAME_FLAGS word. The helper predicates below
// decode the individual NAME_FLAGS bits.
type NodeName struct {
	Name   string // base name with trailing padding spaces trimmed
	Suffix byte   // the 16th name byte selecting the registered service
	Flags  uint16 // raw NAME_FLAGS (see the NameFlag* constants)
}

// IsGroup reports whether the entry is a group name (the G bit is set).
func (n NodeName) IsGroup() bool { return n.Flags&NameFlagGroup != 0 }

// IsActive reports whether the entry's ACT (active) bit is set.
func (n NodeName) IsActive() bool { return n.Flags&NameFlagActive != 0 }

// IsPermanent reports whether the entry's PRM (permanent) bit is set.
func (n NodeName) IsPermanent() bool { return n.Flags&NameFlagPermanent != 0 }

// IsConflict reports whether the entry's CNF (conflict) bit is set.
func (n NodeName) IsConflict() bool { return n.Flags&NameFlagConflict != 0 }

// IsDeregistering reports whether the entry's DRG (deregister-in-progress) bit
// is set.
func (n NodeName) IsDeregistering() bool { return n.Flags&NameFlagDeregister != 0 }

// OwnerNodeType decodes the two ONT bits into their RFC 1002 node-type label:
// "B", "P", "M", or "reserved".
func (n NodeName) OwnerNodeType() string {
	switch (n.Flags & NameFlagONTMask) >> nameFlagONTShift {
	case 0:
		return "B"
	case 1:
		return "P"
	case 2:
		return "M"
	default:
		return "reserved"
	}
}

// SuffixLabel returns a human-readable description of the entry's service
// suffix, distinguishing the group and unique forms of the ambiguous suffixes.
func (n NodeName) SuffixLabel() string {
	return SuffixLabel(n.Suffix, n.IsGroup())
}

// String renders the entry the way an interactive node-status listing would:
// name, suffix (hex + label), and the unique/group flavour.
func (n NodeName) String() string {
	flavour := "UNIQUE"
	if n.IsGroup() {
		flavour = "GROUP"
	}
	return fmt.Sprintf("%-15s <0x%02x> %s (%s)", n.Name, n.Suffix, flavour, n.SuffixLabel())
}

// NodeStatusResult is the parsed outcome of a NODE STATUS RESPONSE: the remote
// node's full NetBIOS name table and the adapter MAC address taken from the
// STATISTICS block's UNIT_ID field.
type NodeStatusResult struct {
	Names []NodeName       // the remote node's registered NetBIOS names
	MAC   net.HardwareAddr // adapter unit id (MAC) from the STATISTICS block, nil if absent
}

// SuffixLabel maps a NetBIOS name suffix (the 16th name byte) to a
// human-readable service description. group selects between the unique and
// group interpretations of the suffixes whose meaning depends on the G bit
// (notably 0x1C, a domain group, versus the unique services). Unknown suffixes
// are rendered as their hex value.
func SuffixLabel(suffix byte, group bool) string {
	switch suffix {
	case 0x00:
		if group {
			return "Domain/Workgroup Name"
		}
		return "Workstation Service"
	case 0x03:
		return "Messenger Service"
	case 0x06:
		return "RAS Server Service"
	case 0x1B:
		return "Domain Master Browser"
	case 0x1C:
		return "Domain Controllers"
	case 0x1D:
		return "Master Browser"
	case 0x1E:
		return "Browser Service Elections"
	case 0x1F:
		return "NetDDE Service"
	case 0x20:
		return "File Server Service"
	case 0x21:
		return "RAS Client Service"
	case 0xBE:
		return "Network Monitor Agent"
	case 0xBF:
		return "Network Monitor Application"
	default:
		return fmt.Sprintf("Unknown service (0x%02x)", suffix)
	}
}

// encodedWildcardName returns the first-level encoding of the reserved "*"
// wildcard name used by a NODE STATUS REQUEST (RFC 1002 4.2.17): the single
// byte '*' (0x2A) followed by fifteen NUL bytes. Unlike an ordinary name this
// is NUL-padded rather than space-padded and begins with '*', so it cannot go
// through NetBIOSName.FirstLevelEncode (which space-pads and rejects a leading
// '*'); the half-byte encoding is therefore applied directly here, mirroring
// that codec.
func encodedWildcardName() string {
	raw := make([]byte, NetBIOSNameLength)
	raw[0] = '*' // the remaining 15 bytes stay 0x00

	encoded := make([]byte, EncodedNameLength)
	for i := 0; i < NetBIOSNameLength; i++ {
		encoded[i*2] = ((raw[i] >> 4) & 0x0F) + ASCII_A
		encoded[i*2+1] = (raw[i] & 0x0F) + ASCII_A
	}
	return string(encoded)
}

// buildNodeStatusRequest assembles the wire bytes of a NODE STATUS REQUEST (RFC
// 1002 4.2.17) carrying the given transaction ID. Node status is an OPCODE
// query distinguished from a name query only by its NBSTAT (0x0021) question
// type, so the header flags are a plain query (no RD/B) and the single question
// asks for the "*" wildcard name in class IN with QDCOUNT=1. The wildcard name
// is framed as an RFC 1002 4.2.1.2 label sequence: one 32-byte first-level
// label terminated by the zero-length root label.
func buildNodeStatusRequest(trnID uint16) []byte {
	encoded := encodedWildcardName()

	buf := make([]byte, 12, 12+1+len(encoded)+1+4)

	// Header: TRN_ID, flags (OPCODE query), QDCOUNT=1; the remaining counts
	// stay zero.
	binary.BigEndian.PutUint16(buf[0:2], trnID)
	binary.BigEndian.PutUint16(buf[2:4], OpNameQuery)
	binary.BigEndian.PutUint16(buf[4:6], 1)

	// Question name as a label sequence, then the NBSTAT type and IN class.
	buf = append(buf, byte(len(encoded)))
	buf = append(buf, encoded...)
	buf = append(buf, 0x00) // root label / terminator
	buf = binary.BigEndian.AppendUint16(buf, QuestionTypeNBSTAT)
	buf = binary.BigEndian.AppendUint16(buf, QuestionClassIn)

	return buf
}

// NodeStatus performs a NetBIOS node-status query (the equivalent of
// "nbtstat -A") against target: it unicasts a NODE STATUS REQUEST to
// target:137/udp and parses the NODE STATUS RESPONSE into the remote node's
// NetBIOS name table and adapter MAC address. target may be a bare host
// ("10.0.0.1") or host:port; a missing port defaults to DefaultNBNSUDPPort (137).
//
// The request is transmitted, then retransmitted up to the client's Retransmit
// count if no matching response arrives within the per-transmission timeout
// (RFC 1001 15.1.1 retry model). Only a response echoing the request's
// transaction ID is accepted, so unrelated datagrams are ignored.
func (c *Client) NodeStatus(target string) (*NodeStatusResult, error) {
	if target == "" {
		return nil, fmt.Errorf("node status target is empty")
	}

	// Accept both a bare host and a host:port; default the port to 137 when the
	// caller supplied only a host.
	if _, _, err := net.SplitHostPort(target); err != nil {
		target = net.JoinHostPort(target, fmt.Sprintf("%d", DefaultNBNSUDPPort))
	}
	addr, err := net.ResolveUDPAddr("udp4", target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node status target %q: %w", target, err)
	}

	trnID := uint16(rand.Intn(0x10000))
	req := buildNodeStatusRequest(trnID)

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return nil, fmt.Errorf("failed to open UDP socket: %w", err)
	}
	defer conn.Close()

	buf := make([]byte, nodeStatusReadBufferSize)
	attempts := c.retransmits() + 1
	for attempt := 0; attempt < attempts; attempt++ {
		if _, err := conn.WriteToUDP(req, addr); err != nil {
			return nil, fmt.Errorf("failed to send NODE STATUS REQUEST: %w", err)
		}

		deadline := time.Now().Add(c.timeout())
		if err := conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}

		// Drain datagrams until this transmission's deadline, ignoring anything
		// that is not the matching NODE STATUS RESPONSE.
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					break // deadline reached: fall through to the next retransmission
				}
				return nil, fmt.Errorf("failed to read NODE STATUS RESPONSE: %w", err)
			}

			result, matched, err := parseNodeStatusResponse(buf[:n], trnID)
			if err != nil || !matched {
				continue // not ours (or undecodable): keep waiting for a valid reply
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no response to NODE STATUS REQUEST for %q", target)
}

// parseNodeStatusResponse decodes a NODE STATUS RESPONSE datagram and, if it is
// the response to our query, returns the parsed name table and MAC. matched is
// true only when the datagram is a response echoing wantTRN; a request, or a
// response carrying a different transaction ID, yields matched=false so the
// caller keeps waiting.
func parseNodeStatusResponse(data []byte, wantTRN uint16) (*NodeStatusResult, bool, error) {
	var resp NBNSPacket
	if _, err := resp.Unmarshal(data); err != nil {
		return nil, false, err
	}

	if resp.Header.Flags&FlagResponse == 0 {
		return nil, false, nil // a request (e.g. our own datagram echoed back)
	}
	if resp.Header.TransactionID != wantTRN {
		return nil, false, nil // response to some other query
	}

	// The name table and STATISTICS ride in the NBSTAT answer record's RDATA.
	for _, rr := range resp.Answers {
		if rr.Type != QuestionTypeNBSTAT {
			continue
		}
		result, err := parseNodeStatusRData(rr.RData)
		if err != nil {
			return nil, true, err
		}
		return result, true, nil
	}

	return nil, true, fmt.Errorf("NODE STATUS RESPONSE carried no NBSTAT record")
}

// parseNodeStatusRData decodes the RDATA of an NBSTAT resource record (RFC 1002
// 4.2.18): a 1-byte NUM_NAMES count, that many 18-byte NODE_NAME entries (a
// 16-byte name plus 2 NAME_FLAGS bytes), then the STATISTICS block whose
// leading 6 bytes are the UNIT_ID (the adapter MAC address). The remaining
// STATISTICS counters are not needed for enumeration and are skipped; a reply
// that is truncated before the UNIT_ID simply yields a nil MAC.
func parseNodeStatusRData(rdata []byte) (*NodeStatusResult, error) {
	if len(rdata) < 1 {
		return nil, fmt.Errorf("NBSTAT RDATA too short for NUM_NAMES")
	}

	numNames := int(rdata[0])
	offset := 1

	result := &NodeStatusResult{Names: make([]NodeName, 0, numNames)}
	for i := 0; i < numNames; i++ {
		if offset+nodeNameEntryLength > len(rdata) {
			return nil, fmt.Errorf("truncated NODE_NAME entry %d of %d", i, numNames)
		}

		raw := rdata[offset : offset+NetBIOSNameLength]
		flags := binary.BigEndian.Uint16(rdata[offset+NetBIOSNameLength : offset+nodeNameEntryLength])
		offset += nodeNameEntryLength

		result.Names = append(result.Names, NodeName{
			Name:   strings.TrimRight(string(raw[:NetBIOSNameLength-1]), " "),
			Suffix: raw[NetBIOSNameLength-1],
			Flags:  flags,
		})
	}

	// STATISTICS block: the UNIT_ID (MAC) is its first 6 bytes.
	if offset+unitIDLength <= len(rdata) {
		mac := make(net.HardwareAddr, unitIDLength)
		copy(mac, rdata[offset:offset+unitIDLength])
		result.MAC = mac
	}

	return result, nil
}

// marshal encodes a NodeName as a single 18-byte NODE_NAME entry for the RDATA
// of an NBSTAT RESPONSE (RFC 1002 4.2.18): the 16-byte NetBIOS name (the base
// name space-padded to 15 characters followed by the 1-byte service suffix) and
// the 2-byte big-endian NAME_FLAGS word. It is the inverse of the per-entry
// decode in parseNodeStatusRData.
func (n NodeName) marshal() []byte {
	buf := make([]byte, nodeNameEntryLength)

	// Bytes 0..14 are the base name, right-padded with spaces; byte 15 is the
	// service suffix; a base name longer than 15 characters is truncated.
	base := n.Name
	if len(base) > NetBIOSNameLength-1 {
		base = base[:NetBIOSNameLength-1]
	}
	copy(buf, base)
	for i := len(base); i < NetBIOSNameLength-1; i++ {
		buf[i] = ' '
	}
	buf[NetBIOSNameLength-1] = n.Suffix

	binary.BigEndian.PutUint16(buf[NetBIOSNameLength:nodeNameEntryLength], n.Flags)
	return buf
}

// buildNodeStatusRData assembles the RDATA of an NBSTAT RESPONSE record (RFC
// 1002 4.2.18): a 1-byte NUM_NAMES count, that many 18-byte NODE_NAME entries,
// then the fixed-length STATISTICS block whose leading 6-byte UNIT_ID carries
// the adapter MAC (zero-filled when mac is nil or not 6 bytes) and whose
// remaining counters are emitted as zero. The layout is the exact inverse of
// parseNodeStatusRData, so this responder and the sibling NodeStatus client
// agree on the wire format.
func buildNodeStatusRData(names []NodeName, mac net.HardwareAddr) []byte {
	// NUM_NAMES is a single unsigned byte, so at most 255 entries fit in one
	// response; a larger table is truncated rather than misreported.
	if len(names) > 0xFF {
		names = names[:0xFF]
	}

	buf := make([]byte, 0, 1+len(names)*nodeNameEntryLength+nodeStatusStatisticsLength)
	buf = append(buf, byte(len(names)))
	for _, n := range names {
		buf = append(buf, n.marshal()...)
	}

	stats := make([]byte, nodeStatusStatisticsLength)
	if len(mac) == unitIDLength {
		copy(stats[:unitIDLength], mac)
	}
	buf = append(buf, stats...)

	return buf
}
