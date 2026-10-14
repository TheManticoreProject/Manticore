package message

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ErrRLE is matched (via errors.Is) by every malformed RLE stream.
var ErrRLE = errors.New("message: malformed RLE data")

const rleEscape = 0x1B

// DecodeRLE expands Remote Command RLE data to exactly length bytes:
//
//	1b 1b            a literal 0x1b byte
//	1b RR RR CC CC   the two bytes RR RR repeated CC CC (big-endian) times
//	any other byte   itself
//
// Expansion past length, a zero repeat count, a truncated escape, or a
// result shorter than length is rejected. The caller bounds length before
// calling, so allocation never exceeds it.
func DecodeRLE(src []byte, length uint32) ([]byte, error) {
	out := make([]byte, 0, length)
	remaining := func() uint64 { return uint64(length) - uint64(len(out)) }
	for i := 0; i < len(src); {
		if src[i] != rleEscape {
			if remaining() < 1 {
				return nil, fmt.Errorf("%w: literal at offset %d expands past %d bytes", ErrRLE, i, length)
			}
			out = append(out, src[i])
			i++
			continue
		}
		if i+1 >= len(src) {
			return nil, fmt.Errorf("%w: truncated escape at offset %d", ErrRLE, i)
		}
		if src[i+1] == rleEscape {
			if remaining() < 1 {
				return nil, fmt.Errorf("%w: escaped literal at offset %d expands past %d bytes", ErrRLE, i, length)
			}
			out = append(out, rleEscape)
			i += 2
			continue
		}
		if i+5 > len(src) {
			return nil, fmt.Errorf("%w: truncated repeat record at offset %d", ErrRLE, i)
		}
		count := uint64(binary.BigEndian.Uint16(src[i+3:]))
		if count == 0 {
			return nil, fmt.Errorf("%w: zero repeat count at offset %d", ErrRLE, i)
		}
		if 2*count > remaining() {
			return nil, fmt.Errorf("%w: repeat record at offset %d expands past %d bytes", ErrRLE, i, length)
		}
		for j := uint64(0); j < count; j++ {
			out = append(out, src[i+1], src[i+2])
		}
		i += 5
	}
	if uint64(len(out)) != uint64(length) {
		return nil, fmt.Errorf("%w: expanded to %d bytes, advertised %d", ErrRLE, len(out), length)
	}
	return out, nil
}
