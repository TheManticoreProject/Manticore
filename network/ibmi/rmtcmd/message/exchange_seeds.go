package message

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Header byte 4 of 0x7001 advertises the client's password capabilities:
// the JTOpen 4.8 client sends SeedAttributes (SHA-1 capable); current JTOpen
// sends SeedAttributesLevel4, which also advertises the password level 4
// algorithm (iSeriesNavigatorReverse docs/04).
const (
	SeedAttributes       byte = 0x01
	SeedAttributesLevel4 byte = 0x03
)

// ExchangeSeedsFrame encodes request 0x7001 with SeedAttributes: the 8-byte
// client seed is the whole template.
func ExchangeSeedsFrame(clientSeed [8]byte) hostserver.Frame {
	return ExchangeSeedsFrameWithAttributes(clientSeed, SeedAttributes)
}

// ExchangeSeedsFrameWithAttributes encodes request 0x7001 with the given
// client attributes byte.
func ExchangeSeedsFrameWithAttributes(clientSeed [8]byte, attributes byte) hostserver.Frame {
	return hostserver.Frame{
		Header: hostserver.Header{
			Attributes: uint16(attributes) << 8,
			ServerID:   hostserver.ServerIDRemoteCommand,
			RequestID:  RequestExchangeSeeds,
		},
		Template: clientSeed[:],
	}
}

// ExchangeSeedsReply is reply 0xF001.
type ExchangeSeedsReply struct {
	ReturnCode uint32
	ServerSeed [8]byte
	// ServerAttributes is header byte 5 (the server's password-level
	// attributes on V5R1 and later). It is informational.
	ServerAttributes byte
	// Fields holds any optional LL/CP fields after the seed.
	Fields hostserver.Fields
}

// ParseExchangeSeedsReply decodes reply 0xF001: a 4-byte return code, then
// with a zero return code the 8-byte server seed and optional LL/CP fields.
func ParseExchangeSeedsReply(f hostserver.Frame) (*ExchangeSeedsReply, error) {
	b := body(f)
	if len(b) < 4 {
		return nil, short("0xF001 reply", len(b), 4)
	}
	r := &ExchangeSeedsReply{ReturnCode: binary.BigEndian.Uint32(b), ServerAttributes: byte(f.Header.Attributes)}
	if r.ReturnCode != 0 {
		return r, nil
	}
	if len(b) < 12 {
		return nil, short("0xF001 reply", len(b), 12)
	}
	copy(r.ServerSeed[:], b[4:12])
	fields, err := hostserver.ParseFields(b[12:], hostserver.FieldPolicy{MaxFields: maxReplyFields, Singletons: map[uint16]bool{}})
	if err != nil {
		return nil, err
	}
	r.Fields = fields
	return r, nil
}
