package pdu

import (
	"encoding/binary"
	"fmt"
)

// fackFixedSize is the size of the fixed portion of a fack body, preceding the
// variable-length selack array: vers(1) + pad1(1) + window_size(2) + max_tsdu(4) +
// max_frag_size(4) + serial_num(2) + selack_len(2).
const fackFixedSize = 16

// FackBody is the body of a fack PDU (and the optional body of a nocall PDU),
// carrying flow-control and selective-acknowledgement information ([C706] section
// 12.6.4.5, fack_body_t at body format version 0).
//
// Wire layout (little-endian):
//
//	0      vers          uint8     (body format version, 0)
//	1      pad1          uint8     (unused)
//	2-3    window_size   uint16    (receiver window, in fragments)
//	4-7    max_tsdu      uint32    (largest local transport service data unit)
//	8-11   max_frag_size uint32    (largest accepted fragment size)
//	12-13  serial_num    uint16    (serial number being acknowledged)
//	14-15  selack_len    uint16    (number of selack elements that follow)
//	16...  selack[]      uint32[]  (selective-ack bitmasks, selack_len elements)
type FackBody struct {
	Version     uint8
	Pad1        uint8
	WindowSize  uint16
	MaxTSDU     uint32
	MaxFragSize uint32
	SerialNum   uint16
	// SelAck holds the selective-acknowledgement bitmasks. Its length is encoded on
	// the wire as selack_len.
	SelAck []uint32
}

// Marshal serializes the fack body into its wire form.
func (b *FackBody) Marshal() ([]byte, error) {
	if len(b.SelAck) > 0xFFFF {
		return nil, fmt.Errorf("fack selack array too large: %d elements, max %d", len(b.SelAck), 0xFFFF)
	}
	buf := make([]byte, fackFixedSize+4*len(b.SelAck))
	buf[0] = b.Version
	buf[1] = b.Pad1
	binary.LittleEndian.PutUint16(buf[2:4], b.WindowSize)
	binary.LittleEndian.PutUint32(buf[4:8], b.MaxTSDU)
	binary.LittleEndian.PutUint32(buf[8:12], b.MaxFragSize)
	binary.LittleEndian.PutUint16(buf[12:14], b.SerialNum)
	binary.LittleEndian.PutUint16(buf[14:16], uint16(len(b.SelAck)))
	for i, mask := range b.SelAck {
		binary.LittleEndian.PutUint32(buf[fackFixedSize+4*i:], mask)
	}
	return buf, nil
}

// Unmarshal parses a fack body from data and returns the number of bytes consumed.
func (b *FackBody) Unmarshal(data []byte) (int, error) {
	if len(data) < fackFixedSize {
		return 0, fmt.Errorf("fack body truncated: have %d bytes, need at least %d", len(data), fackFixedSize)
	}
	b.Version = data[0]
	b.Pad1 = data[1]
	b.WindowSize = binary.LittleEndian.Uint16(data[2:4])
	b.MaxTSDU = binary.LittleEndian.Uint32(data[4:8])
	b.MaxFragSize = binary.LittleEndian.Uint32(data[8:12])
	b.SerialNum = binary.LittleEndian.Uint16(data[12:14])
	selackLen := int(binary.LittleEndian.Uint16(data[14:16]))
	// Reject a declared count that cannot fit in the remaining bytes before
	// allocating, so a malformed length cannot drive a large allocation.
	if avail := (len(data) - fackFixedSize) / 4; selackLen > avail {
		return 0, fmt.Errorf("fack selack_len %d exceeds %d available elements", selackLen, avail)
	}
	b.SelAck = make([]uint32, selackLen)
	for i := range b.SelAck {
		b.SelAck[i] = binary.LittleEndian.Uint32(data[fackFixedSize+4*i:])
	}
	return fackFixedSize + 4*selackLen, nil
}
