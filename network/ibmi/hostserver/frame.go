package hostserver

import (
	"errors"
	"fmt"
	"math"
)

// Frame size limits. DefaultMaxFrame applies when a limit of 0 is given.
const (
	MinMaxFrame     uint32 = HeaderLength
	DefaultMaxFrame uint32 = 16 << 20
	MaxMaxFrame     uint32 = 64 << 20
)

// ErrFrameTooLarge is matched (via errors.Is) when a frame exceeds the
// configured maximum, whether it is being built or received.
var ErrFrameTooLarge = errors.New("hostserver: frame exceeds maximum size")

// ResolveMaxFrame returns the effective maximum frame size: DefaultMaxFrame
// for 0, max itself within [MinMaxFrame, MaxMaxFrame], an error otherwise.
func ResolveMaxFrame(max uint32) (uint32, error) {
	if max == 0 {
		return DefaultMaxFrame, nil
	}
	if max < MinMaxFrame || max > MaxMaxFrame {
		return 0, fmt.Errorf("hostserver: maximum frame %d outside %d..%d", max, MinMaxFrame, MaxMaxFrame)
	}
	return max, nil
}

// Frame is one complete host-server message: header routing fields, the
// fixed template and the variable (LL/CP) area. Length and TemplateLength
// are derived from the slices; the values in Header are ignored when
// marshaling and set by UnmarshalFrame.
type Frame struct {
	Header   Header
	Template []byte
	Variable []byte
}

// Size returns the encoded length, or an error if it overflows the 32-bit
// length field or the template overflows its 16-bit length.
func (f Frame) Size() (uint32, error) {
	if len(f.Template) > math.MaxUint16 {
		return 0, fmt.Errorf("%w: template of %d bytes exceeds %d", ErrFrameTooLarge, len(f.Template), math.MaxUint16)
	}
	total := uint64(HeaderLength) + uint64(len(f.Template)) + uint64(len(f.Variable))
	if total > math.MaxUint32 {
		return 0, fmt.Errorf("%w: %d bytes", ErrFrameTooLarge, total)
	}
	return uint32(total), nil
}

// Marshal encodes f, deriving Length and TemplateLength. maxFrame is
// resolved with ResolveMaxFrame.
func (f Frame) Marshal(maxFrame uint32) ([]byte, error) {
	max, err := ResolveMaxFrame(maxFrame)
	if err != nil {
		return nil, err
	}
	size, err := f.Size()
	if err != nil {
		return nil, err
	}
	if size > max {
		return nil, fmt.Errorf("%w: %d bytes, maximum %d", ErrFrameTooLarge, size, max)
	}
	h := f.Header
	h.Length = size
	h.TemplateLength = uint16(len(f.Template))
	hb, err := h.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, size)
	out = append(out, hb...)
	out = append(out, f.Template...)
	out = append(out, f.Variable...)
	return out, nil
}

// UnmarshalFrame decodes exactly one frame occupying all of b. The returned
// Template and Variable are copies, so b may be reused.
func UnmarshalFrame(b []byte, maxFrame uint32) (Frame, error) {
	max, err := ResolveMaxFrame(maxFrame)
	if err != nil {
		return Frame{}, err
	}
	if len(b) < HeaderLength {
		return Frame{}, malformed("frame", 0, "%d bytes is shorter than the header", len(b))
	}
	h, err := UnmarshalHeader(b[:HeaderLength])
	if err != nil {
		return Frame{}, err
	}
	if h.Length > max {
		return Frame{}, fmt.Errorf("%w: header advertises %d bytes, maximum %d", ErrFrameTooLarge, h.Length, max)
	}
	if uint64(len(b)) != uint64(h.Length) {
		return Frame{}, malformed("frame", 0, "header advertises %d bytes, got %d", h.Length, len(b))
	}
	return frameFromBody(h, b[HeaderLength:]), nil
}

// frameFromBody splits a body already validated against h.
func frameFromBody(h Header, body []byte) Frame {
	t := int(h.TemplateLength)
	return Frame{
		Header:   h,
		Template: append([]byte{}, body[:t]...),
		Variable: append([]byte{}, body[t:]...),
	}
}
