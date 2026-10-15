package hostserver

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// HeaderLength is the size of the common host-server header.
const HeaderLength = 20

// Server IDs carried at header offset 6.
const (
	ServerIDFile          uint16 = 0xE002
	ServerIDDataQueue     uint16 = 0xE007
	ServerIDRemoteCommand uint16 = 0xE008
	ServerIDSignon        uint16 = 0xE009
)

// ErrMalformed is matched (via errors.Is) by every header, frame and field
// decoding failure.
var ErrMalformed = errors.New("hostserver: malformed data")

// MalformedError reports where decoding failed, without echoing content.
type MalformedError struct {
	What   string // "header", "frame", "field", ...
	Offset int
	Reason string
}

func (e *MalformedError) Error() string {
	return fmt.Sprintf("hostserver: malformed %s at offset %d: %s", e.What, e.Offset, e.Reason)
}

// Is makes errors.Is(err, ErrMalformed) true.
func (e *MalformedError) Is(target error) bool { return target == ErrMalformed }

func malformed(what string, offset int, format string, args ...any) error {
	return &MalformedError{What: what, Offset: offset, Reason: fmt.Sprintf(format, args...)}
}

// Header is the fixed 20-byte prefix of every host-server frame. All fields
// are big-endian on the wire.
type Header struct {
	Length         uint32 // offset 0: total frame length, header included
	Attributes     uint16 // offset 4: request/header attributes
	ServerID       uint16 // offset 6
	Instance       uint32 // offset 8: server instance / reserved
	CorrelationID  uint32 // offset 12
	TemplateLength uint16 // offset 16: fixed-template length after the header
	RequestID      uint16 // offset 18: request or reply ID
}

// validate checks the length invariants shared by Marshal and Unmarshal.
func (h Header) validate() error {
	if h.Length < HeaderLength {
		return malformed("header", 0, "length %d is below %d", h.Length, HeaderLength)
	}
	if uint64(h.TemplateLength) > uint64(h.Length)-HeaderLength {
		return malformed("header", 16, "template length %d exceeds the %d-byte body", h.TemplateLength, h.Length-HeaderLength)
	}
	return nil
}

// MarshalBinary returns the 20-byte encoding of h after validating it.
func (h Header) MarshalBinary() ([]byte, error) {
	if err := h.validate(); err != nil {
		return nil, err
	}
	b := make([]byte, HeaderLength)
	binary.BigEndian.PutUint32(b[0:], h.Length)
	binary.BigEndian.PutUint16(b[4:], h.Attributes)
	binary.BigEndian.PutUint16(b[6:], h.ServerID)
	binary.BigEndian.PutUint32(b[8:], h.Instance)
	binary.BigEndian.PutUint32(b[12:], h.CorrelationID)
	binary.BigEndian.PutUint16(b[16:], h.TemplateLength)
	binary.BigEndian.PutUint16(b[18:], h.RequestID)
	return b, nil
}

// UnmarshalHeader decodes exactly HeaderLength bytes.
func UnmarshalHeader(b []byte) (Header, error) {
	if len(b) != HeaderLength {
		return Header{}, malformed("header", 0, "%d bytes, want %d", len(b), HeaderLength)
	}
	h := Header{
		Length:         binary.BigEndian.Uint32(b[0:]),
		Attributes:     binary.BigEndian.Uint16(b[4:]),
		ServerID:       binary.BigEndian.Uint16(b[6:]),
		Instance:       binary.BigEndian.Uint32(b[8:]),
		CorrelationID:  binary.BigEndian.Uint32(b[12:]),
		TemplateLength: binary.BigEndian.Uint16(b[16:]),
		RequestID:      binary.BigEndian.Uint16(b[18:]),
	}
	if err := h.validate(); err != nil {
		return Header{}, err
	}
	return h, nil
}
