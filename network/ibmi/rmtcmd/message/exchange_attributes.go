package message

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// Values sent in the 0x1001 template.
const (
	// AttributesClientCCSID is UTF-16 (13488), as JTOpen 4.8 sends.
	AttributesClientCCSID uint32 = 13488
	// AttributesClientVersion is the client version.
	AttributesClientVersion uint32 = 1
	// RequestedDataStreamLevel 0 asks the server for its own level.
	RequestedDataStreamLevel uint16 = 0
	// DefaultNLV is US English.
	DefaultNLV = "2924"
)

// ExchangeAttributesFrame encodes request 0x1001: a 14-byte template of
// client CCSID, NLV as four EBCDIC zoned digits, client version and
// requested data-stream level. nlv must be four ASCII digits.
func ExchangeAttributesFrame(correlation uint32, nlv string) (hostserver.Frame, error) {
	if len(nlv) != 4 {
		return hostserver.Frame{}, fmt.Errorf("message: NLV %q must be four digits", nlv)
	}
	t := make([]byte, 14)
	binary.BigEndian.PutUint32(t[0:], AttributesClientCCSID)
	for i := 0; i < 4; i++ {
		if nlv[i] < '0' || nlv[i] > '9' {
			return hostserver.Frame{}, fmt.Errorf("message: NLV %q must be four digits", nlv)
		}
		t[4+i] = 0xF0 | (nlv[i] - '0')
	}
	binary.BigEndian.PutUint32(t[8:], AttributesClientVersion)
	binary.BigEndian.PutUint16(t[12:], RequestedDataStreamLevel)
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, CorrelationID: correlation, RequestID: RequestExchangeAttributes},
		Template: t,
	}, nil
}

// exchangeAttributesMinBody covers offsets 20..35 of reply 0x8001.
const exchangeAttributesMinBody = 16

// ExchangeAttributesReply is reply 0x8001. Its body is fixed-position, not
// LL/CP: return code at offset 20, job CCSID at 22, data-stream level at 34.
type ExchangeAttributesReply struct {
	// ReturnCode is the 2-byte Remote Command return code.
	ReturnCode uint16
	JobCCSID   uint32
	// DataStreamLevel is the Remote Command level that selects
	// capabilities. It is unrelated to the Sign-on level.
	DataStreamLevel uint16
	// Reserved holds offsets 26..33, whose meaning is not established; it is
	// kept opaque.
	Reserved [8]byte
}

// ParseExchangeAttributesReply decodes reply 0x8001. Bytes after offset 35
// are not interpreted.
func ParseExchangeAttributesReply(f hostserver.Frame) (*ExchangeAttributesReply, error) {
	b := body(f)
	if len(b) < exchangeAttributesMinBody {
		return nil, short("0x8001 reply", len(b), exchangeAttributesMinBody)
	}
	r := &ExchangeAttributesReply{
		ReturnCode:      binary.BigEndian.Uint16(b[0:]),
		JobCCSID:        binary.BigEndian.Uint32(b[2:]),
		DataStreamLevel: binary.BigEndian.Uint16(b[14:]),
	}
	copy(r.Reserved[:], b[6:14])
	return r, nil
}
