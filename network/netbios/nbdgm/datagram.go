// Package nbdgm implements the NetBIOS Datagram Service (RFC 1002 4.4, UDP
// port 138): the datagram header codec, the DIRECT_UNIQUE / DIRECT_GROUP /
// BROADCAST datagrams, the DATAGRAM ERROR packet, the DATAGRAM QUERY REQUEST
// and POSITIVE / NEGATIVE DATAGRAM QUERY RESPONSE packets, and FIRST/MORE
// fragmentation and reassembly of the USER_DATA payload.
//
// This is the connectionless transport that the Browser protocol and mailslots
// ride on top of; only the datagram transport layer lives here. The NetBIOS
// name encoding for the SOURCE_NAME / DESTINATION_NAME fields is reused from
// the sibling name-service package (network/netbios/nbns).
package nbdgm

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbns"
)

// DefaultDatagramPort is the well-known UDP port of the NetBIOS Datagram
// Service (RFC 1002 4.4 / RFC 1001 5.4). All datagrams are exchanged on it.
const DefaultDatagramPort = 138

// Datagram MSG_TYPE values (RFC 1002 4.4.1). The message type is the first byte
// of every datagram and selects both the semantics and the trailer layout.
const (
	MsgTypeDirectUnique          uint8 = 0x10 // DIRECT_UNIQUE DATAGRAM
	MsgTypeDirectGroup           uint8 = 0x11 // DIRECT_GROUP DATAGRAM
	MsgTypeBroadcast             uint8 = 0x12 // BROADCAST DATAGRAM
	MsgTypeError                 uint8 = 0x13 // DATAGRAM ERROR
	MsgTypeQueryRequest          uint8 = 0x14 // DATAGRAM QUERY REQUEST
	MsgTypePositiveQueryResponse uint8 = 0x15 // DATAGRAM POSITIVE QUERY RESPONSE
	MsgTypeNegativeQueryResponse uint8 = 0x16 // DATAGRAM NEGATIVE QUERY RESPONSE
)

// FLAGS field bit layout (RFC 1002 4.4.1). The one-byte FLAGS field is drawn in
// the RFC as
//
//	  0   1   2   3   4   5   6   7
//	+---+---+---+---+---+---+---+---+
//	| 0 | 0 | 0 | 0 |  SNT  | F | M |
//	+---+---+---+---+---+---+---+---+
//
// where bit 0 is the most-significant bit. In terms of the numeric byte value
// M (MORE) is therefore the least-significant bit (0x01), F (FIRST) is 0x02,
// and SNT (source end-node type) occupies bits 2-3 (mask 0x0C).
const (
	FlagMore  uint8 = 0x01 // M: set when more fragments of this datagram follow
	FlagFirst uint8 = 0x02 // F: set on the first (or only) fragment of a datagram

	sntShift = 2
	sntMask  = 0x0C // SNT occupies the two bits above F
)

// SNT source end-node type values carried in the FLAGS field (RFC 1002 4.4.1).
const (
	NodeTypeB    uint8 = 0x00 // B node (broadcast)
	NodeTypeP    uint8 = 0x01 // P node (point-to-point)
	NodeTypeM    uint8 = 0x02 // M node (mixed)
	NodeTypeNBDD uint8 = 0x03 // NetBIOS Datagram Distribution server
)

// DATAGRAM ERROR ERROR_CODE values (RFC 1002 4.4.2).
const (
	ErrorDestinationNameNotPresent uint8 = 0x82 // DESTINATION NAME NOT PRESENT
	ErrorInvalidSourceNameFormat   uint8 = 0x83 // INVALID SOURCE NAME FORMAT
	ErrorInvalidDestNameFormat     uint8 = 0x84 // INVALID DESTINATION NAME FORMAT
)

// baseHeaderLen is the size of the header common to every datagram: MSG_TYPE(1)
// + FLAGS(1) + DGM_ID(2) + SOURCE_IP(4) + SOURCE_PORT(2). The DIRECT/BROADCAST
// datagrams extend it with DGM_LENGTH(2) + PACKET_OFFSET(2).
const (
	baseHeaderLen   = 10
	directHeaderLen = 14
)

// Name is a NetBIOS name as carried in a datagram SOURCE_NAME or
// DESTINATION_NAME field: the up-to-15-character base name, the one-byte
// service suffix (the 16th byte of the padded NetBIOS name), and an optional
// scope. On the wire it is the RFC 1002 4.2.1.2 second-level encoding.
type Name struct {
	Name   string // base name, up to 15 characters (may be the "*" wildcard)
	Suffix byte   // service suffix (16th byte of the padded NetBIOS name)
	Scope  string // optional NetBIOS scope ID, dot-separated ("" for the default)
}

// encode builds the on-the-wire form of the name: the 0x20 length byte, the
// 32-byte first-level encoding of the 16-byte (15 chars + suffix) name, any
// dot-separated scope labels, and a 0x00 terminator. The first-level half-byte
// encoding is reused from the name-service codec (nbns.EncodeSessionServiceName)
// so the datagram and name services agree on the wire form.
func (n Name) encode() ([]byte, error) {
	// EncodeSessionServiceName returns exactly [0x20][32 encoded bytes][0x00];
	// that is the datagram name form for the default (empty) scope.
	base, err := nbns.EncodeSessionServiceName(n.Name, n.Suffix)
	if err != nil {
		return nil, err
	}
	if n.Scope == "" {
		return base, nil
	}

	// With a scope, drop the trailing 0x00 root label and append the scope
	// labels before a fresh terminator.
	out := make([]byte, 0, len(base)+len(n.Scope)+8)
	out = append(out, base[:len(base)-1]...)
	for _, label := range strings.Split(n.Scope, ".") {
		if len(label) == 0 || len(label) > 63 {
			return nil, fmt.Errorf("invalid scope label %q", label)
		}
		out = append(out, byte(len(label)))
		out = append(out, label...)
	}
	out = append(out, 0x00)
	return out, nil
}

// decodeName decodes a datagram name at offset in data and returns the name and
// the number of bytes it consumed. It mirrors the second-level encoding: a
// 0x20 length byte, the 32 first-level-encoded characters (half-byte pairs, each
// character being the nibble plus 'A'), and a 0x00-terminated scope label
// sequence. The suffix byte is preserved (unlike nbns.FirstLevelDecode, which
// trims it when it is a space).
func decodeName(data []byte, offset int) (Name, int, error) {
	pos := offset
	if pos >= len(data) {
		return Name{}, 0, fmt.Errorf("truncated name: no length byte")
	}

	l := int(data[pos])
	pos++
	if l != nbns.EncodedNameLength {
		return Name{}, 0, fmt.Errorf("unexpected first label length %d (want %d)", l, nbns.EncodedNameLength)
	}
	if pos+l > len(data) {
		return Name{}, 0, fmt.Errorf("truncated name label")
	}

	// Decode the 32 printable characters back into the 16-byte NetBIOS name.
	encoded := data[pos : pos+l]
	pos += l
	decoded := make([]byte, nbns.NetBIOSNameLength)
	for i := 0; i < nbns.NetBIOSNameLength; i++ {
		high := encoded[i*2] - 'A'
		low := encoded[i*2+1] - 'A'
		if high > 0x0F || low > 0x0F {
			return Name{}, 0, fmt.Errorf("invalid name encoding character")
		}
		decoded[i] = (high << 4) | low
	}

	// Read the scope label sequence up to the 0x00 root label.
	var scope []string
	for {
		if pos >= len(data) {
			return Name{}, 0, fmt.Errorf("truncated name: missing scope terminator")
		}
		sl := int(data[pos])
		pos++
		if sl == 0 {
			break
		}
		if pos+sl > len(data) {
			return Name{}, 0, fmt.Errorf("truncated scope label")
		}
		scope = append(scope, string(data[pos:pos+sl]))
		pos += sl
	}

	return Name{
		Name:   strings.TrimRight(string(decoded[:nbns.NetBIOSNameLength-1]), " "),
		Suffix: decoded[nbns.NetBIOSNameLength-1],
		Scope:  strings.Join(scope, "."),
	}, pos - offset, nil
}

// Datagram is a decoded NetBIOS datagram. Which trailer fields are meaningful
// depends on MsgType: the DIRECT/BROADCAST types carry DgmLength, PacketOffset,
// SourceName, DestinationName and UserData; DATAGRAM ERROR carries ErrorCode;
// the query request/response types carry only DestinationName.
type Datagram struct {
	MsgType    uint8
	Flags      uint8
	DgmID      uint16
	SourceIP   net.IP
	SourcePort uint16

	// DIRECT_UNIQUE / DIRECT_GROUP / BROADCAST fields.
	DgmLength       uint16 // length of the trailer (names + user data) in this packet
	PacketOffset    uint16 // offset of this fragment's USER_DATA in the reassembled datagram
	SourceName      Name
	DestinationName Name
	UserData        []byte

	// DATAGRAM ERROR field.
	ErrorCode uint8
}

// NodeType returns the SNT source end-node type encoded in the FLAGS field.
func (d *Datagram) NodeType() uint8 { return (d.Flags & sntMask) >> sntShift }

// SetNodeType sets the SNT source end-node type bits of the FLAGS field.
func (d *Datagram) SetNodeType(nt uint8) {
	d.Flags = (d.Flags &^ sntMask) | ((nt << sntShift) & sntMask)
}

// IsFirst reports whether the FIRST (F) flag is set.
func (d *Datagram) IsFirst() bool { return d.Flags&FlagFirst != 0 }

// HasMore reports whether the MORE (M) flag is set, i.e. further fragments of
// this datagram follow.
func (d *Datagram) HasMore() bool { return d.Flags&FlagMore != 0 }

// isDirect reports whether the message type is one of the DIRECT/BROADCAST
// datagrams, which carry the extended (14-byte) header and a name+user-data
// trailer.
func isDirect(msgType uint8) bool {
	return msgType == MsgTypeDirectUnique ||
		msgType == MsgTypeDirectGroup ||
		msgType == MsgTypeBroadcast
}

// isQuery reports whether the message type is a DATAGRAM QUERY REQUEST or a
// POSITIVE/NEGATIVE DATAGRAM QUERY RESPONSE, all of which carry only a
// DESTINATION_NAME after the base header.
func isQuery(msgType uint8) bool {
	return msgType == MsgTypeQueryRequest ||
		msgType == MsgTypePositiveQueryResponse ||
		msgType == MsgTypeNegativeQueryResponse
}

// Marshal serialises the datagram to its RFC 1002 4.4 wire form. For the
// DIRECT/BROADCAST types the DGM_LENGTH field is computed as the length of the
// emitted trailer (names, when the FIRST flag is set, plus USER_DATA), matching
// how a receiver validates it; the caller-supplied DgmLength is ignored.
func (d *Datagram) Marshal() ([]byte, error) {
	ip := d.SourceIP.To4()
	if ip == nil {
		return nil, fmt.Errorf("SOURCE_IP must be an IPv4 address, got %v", d.SourceIP)
	}

	buf := make([]byte, baseHeaderLen)
	buf[0] = d.MsgType
	buf[1] = d.Flags
	binary.BigEndian.PutUint16(buf[2:4], d.DgmID)
	copy(buf[4:8], ip)
	binary.BigEndian.PutUint16(buf[8:10], d.SourcePort)

	switch {
	case isDirect(d.MsgType):
		// The trailer holds the two names (only on the FIRST fragment) followed
		// by this packet's USER_DATA. A non-first fragment repeats neither name.
		var trailer []byte
		if d.IsFirst() {
			src, err := d.SourceName.encode()
			if err != nil {
				return nil, fmt.Errorf("encoding SOURCE_NAME: %w", err)
			}
			dst, err := d.DestinationName.encode()
			if err != nil {
				return nil, fmt.Errorf("encoding DESTINATION_NAME: %w", err)
			}
			trailer = append(trailer, src...)
			trailer = append(trailer, dst...)
		}
		trailer = append(trailer, d.UserData...)

		ext := make([]byte, 4)
		binary.BigEndian.PutUint16(ext[0:2], uint16(len(trailer)))
		binary.BigEndian.PutUint16(ext[2:4], d.PacketOffset)
		buf = append(buf, ext...)
		buf = append(buf, trailer...)

	case d.MsgType == MsgTypeError:
		buf = append(buf, d.ErrorCode)

	case isQuery(d.MsgType):
		dst, err := d.DestinationName.encode()
		if err != nil {
			return nil, fmt.Errorf("encoding DESTINATION_NAME: %w", err)
		}
		buf = append(buf, dst...)

	default:
		return nil, fmt.Errorf("unknown datagram MSG_TYPE 0x%02x", d.MsgType)
	}

	return buf, nil
}

// Unmarshal parses a datagram from data and returns the number of bytes
// consumed. It never panics on short or malformed input, reporting an error
// instead. For a DIRECT/BROADCAST datagram only the DGM_LENGTH trailer bytes
// are consumed, so trailing padding in an oversized buffer is ignored.
func (d *Datagram) Unmarshal(data []byte) (int, error) {
	if len(data) < baseHeaderLen {
		return 0, fmt.Errorf("datagram too short: need %d header bytes, got %d", baseHeaderLen, len(data))
	}

	d.MsgType = data[0]
	d.Flags = data[1]
	d.DgmID = binary.BigEndian.Uint16(data[2:4])
	d.SourceIP = net.IP(append([]byte(nil), data[4:8]...))
	d.SourcePort = binary.BigEndian.Uint16(data[8:10])

	switch {
	case isDirect(d.MsgType):
		if len(data) < directHeaderLen {
			return 0, fmt.Errorf("DIRECT datagram too short: need %d header bytes, got %d", directHeaderLen, len(data))
		}
		d.DgmLength = binary.BigEndian.Uint16(data[10:12])
		d.PacketOffset = binary.BigEndian.Uint16(data[12:14])

		trailer := data[directHeaderLen:]
		if int(d.DgmLength) > len(trailer) {
			return 0, fmt.Errorf("DGM_LENGTH %d exceeds %d available trailer bytes", d.DgmLength, len(trailer))
		}
		trailer = trailer[:d.DgmLength]

		if d.IsFirst() {
			src, n1, err := decodeName(trailer, 0)
			if err != nil {
				return 0, fmt.Errorf("decoding SOURCE_NAME: %w", err)
			}
			dst, n2, err := decodeName(trailer, n1)
			if err != nil {
				return 0, fmt.Errorf("decoding DESTINATION_NAME: %w", err)
			}
			d.SourceName = src
			d.DestinationName = dst
			d.UserData = append([]byte(nil), trailer[n1+n2:]...)
		} else {
			d.UserData = append([]byte(nil), trailer...)
		}
		return directHeaderLen + int(d.DgmLength), nil

	case d.MsgType == MsgTypeError:
		if len(data) < baseHeaderLen+1 {
			return 0, fmt.Errorf("DATAGRAM ERROR too short: need %d bytes, got %d", baseHeaderLen+1, len(data))
		}
		d.ErrorCode = data[baseHeaderLen]
		return baseHeaderLen + 1, nil

	case isQuery(d.MsgType):
		dst, n, err := decodeName(data, baseHeaderLen)
		if err != nil {
			return 0, fmt.Errorf("decoding DESTINATION_NAME: %w", err)
		}
		d.DestinationName = dst
		return baseHeaderLen + n, nil

	default:
		return 0, fmt.Errorf("unknown datagram MSG_TYPE 0x%02x", d.MsgType)
	}
}
