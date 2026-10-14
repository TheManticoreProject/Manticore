package message

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// CPParameter is the code point of every program-call parameter envelope.
const CPParameter uint16 = 0x1103

// ParameterEnvelopeLength is the fixed prefix of a parameter envelope.
const ParameterEnvelopeLength = 12

// Request usages.
const (
	UsageInput          uint16 = 11
	UsageOutput         uint16 = 22 // output allocation
	UsageInputOutput    uint16 = 13 // below Remote Command level 5
	UsageInputOutputNew uint16 = 33 // level 5 and later
)

// Reply usages accepted for returned parameters. 22 and 23 mark
// RLE-compressed data in a reply.
const (
	ReplyUsageOutputBase            uint16 = 2
	ReplyUsageInputOutputBase       uint16 = 3
	ReplyUsageOutput                uint16 = 12
	ReplyUsageOutputCompressed      uint16 = 22
	ReplyUsageInputOutputCompressed uint16 = 23
)

// RequestParameter is one parameter envelope in a 0x1003 request.
type RequestParameter struct {
	Usage uint16
	// Maximum is the receiver (or uncompressed input) length.
	Maximum uint32
	// Data is the transmitted payload. For uncompressed input the caller
	// passes the data with trailing zero bytes removed (TrimTrailingZeros);
	// the server restores them up to Maximum.
	Data []byte
}

// TrimTrailingZeros returns b without its trailing zero bytes.
func TrimTrailingZeros(b []byte) []byte {
	end := len(b)
	for end > 0 && b[end-1] == 0 {
		end--
	}
	return b[:end]
}

// appendParameter appends one envelope: LL, CP 0x1103, maximum, usage, data.
func appendParameter(dst []byte, p RequestParameter) ([]byte, error) {
	ll := uint64(ParameterEnvelopeLength) + uint64(len(p.Data))
	if ll > math.MaxUint32 {
		return nil, fmt.Errorf("%w: parameter of %d bytes", hostserver.ErrFrameTooLarge, ll)
	}
	var prefix [ParameterEnvelopeLength]byte
	binary.BigEndian.PutUint32(prefix[0:], uint32(ll))
	binary.BigEndian.PutUint16(prefix[4:], CPParameter)
	binary.BigEndian.PutUint32(prefix[6:], p.Maximum)
	binary.BigEndian.PutUint16(prefix[10:], p.Usage)
	return append(append(dst, prefix[:]...), p.Data...), nil
}

// ReplyParameter is one returned parameter envelope.
type ReplyParameter struct {
	Usage uint16
	// Length is the advertised (uncompressed) output length.
	Length uint32
	// Payload is the transmitted bytes; a sub-slice of the reply.
	Payload []byte
}

// ReadParameter decodes the envelope at the start of b and returns it with
// the number of bytes it occupies.
func ReadParameter(b []byte) (ReplyParameter, int, error) {
	if len(b) < ParameterEnvelopeLength {
		return ReplyParameter{}, 0, &hostserver.MalformedError{What: "parameter", Reason: fmt.Sprintf("%d bytes cannot hold a %d-byte envelope", len(b), ParameterEnvelopeLength)}
	}
	ll := binary.BigEndian.Uint32(b)
	if ll < ParameterEnvelopeLength || uint64(ll) > uint64(len(b)) {
		return ReplyParameter{}, 0, &hostserver.MalformedError{What: "parameter", Reason: fmt.Sprintf("length %d outside %d..%d", ll, ParameterEnvelopeLength, len(b))}
	}
	if cp := binary.BigEndian.Uint16(b[4:]); cp != CPParameter {
		return ReplyParameter{}, 0, &hostserver.MalformedError{What: "parameter", Offset: 4, Reason: fmt.Sprintf("code point 0x%04X, want 0x%04X", cp, CPParameter)}
	}
	return ReplyParameter{
		Length:  binary.BigEndian.Uint32(b[6:]),
		Usage:   binary.BigEndian.Uint16(b[10:]),
		Payload: b[ParameterEnvelopeLength:ll:ll],
	}, int(ll), nil
}
