package filesystem

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

// encodeUTF16LE encodes s as little-endian UTF-16, the on-wire form of the
// variable name fields in MS-FSCC information-class structures.
func encodeUTF16LE(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	b := make([]byte, len(u16)*2)
	for i, u := range u16 {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	return b
}

// decodeUTF16LE decodes a little-endian UTF-16 byte buffer into a string,
// trimming any trailing NUL units.
func decodeUTF16LE(b []byte) string {
	u16 := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u16 = append(u16, binary.LittleEndian.Uint16(b[i:i+2]))
	}
	return strings.TrimRight(string(utf16.Decode(u16)), "\x00")
}

// nextEntry returns the offset of the entry after the one at pos in a chained
// information-class buffer, or -1 when the chain ends there. fixedSize is the
// entry's fixed part and nameLenOffset the offset of its 4-byte name length.
//
// The chain ends when NextEntryOffset is zero, runs past buf, or points inside
// the current entry's fixed part and name. Following such an offset would make
// entries overlap, and each would decode most of the remaining buffer again,
// so a small hostile buffer could expand into gigabytes of names. An offset
// that is not 8-byte aligned, which [MS-FSCC] forbids, is still followed: it
// cannot cause overlap, and refusing it would only lose entries from a
// careless server. Offsets are compared in uint64 so a 2 GiB+ value cannot
// wrap pos on 32-bit platforms.
func nextEntry(buf []byte, pos, fixedSize, nameLenOffset int) int {
	next := uint64(binary.LittleEndian.Uint32(buf[pos:]))
	nameLen := uint64(binary.LittleEndian.Uint32(buf[pos+nameLenOffset:]))
	if next == 0 || next > uint64(len(buf)-pos) || next < uint64(fixedSize)+nameLen {
		return -1
	}
	return pos + int(next)
}
