// Package pdu models connection-oriented DCE/RPC protocol data units (PDUs): the
// 16-byte common header and the Bind, Bind_Ack, Bind_Nak, Request, Response, and
// Fault bodies ([C706] section 12, [MS-RPCE] section 2.2).
//
// Each PDU type embeds the common Header and represents a whole PDU. Marshal returns
// the complete on-the-wire bytes (header included) and sets the header's frag_length;
// Unmarshal parses a complete PDU and returns the number of bytes consumed.
//
// Integers are read in the byte order the sender's data representation (DREP)
// announces: little-endian (DREP first byte 0x10), which is what Windows and Samba
// always send, or big-endian (high nibble 0), which [C706] permits and some
// non-Windows DCE peers use. A DREP whose integer order is neither is rejected on
// Unmarshal. PDUs are always marshalled little-endian.
package pdu

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5"
)

// HeaderSize is the size, in bytes, of the connection-oriented common header.
const HeaderSize = 16

// DataRepresentationLittleEndian is the canonical packed_drep for little-endian
// integers, ASCII characters, and IEEE floating point.
var DataRepresentationLittleEndian = [4]byte{0x10, 0x00, 0x00, 0x00}

// Header is the connection-oriented common header present at the start of every PDU.
//
// References:
//   - [C706] section 12.6.3.1 (common fields of the connection-oriented PDU):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.10 PDU Types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/cef9d684-f09f-4533-a54c-9255079d3e1d
type Header struct {
	RPCVersion         uint8
	RPCVersionMinor    uint8
	PacketType         PacketType
	PacketFlags        PFCFlags
	DataRepresentation [4]byte
	FragLength         uint16
	AuthLength         uint16
	CallID             uint32
}

// NewHeader returns a header initialized for protocol version 5.0 with little-endian
// data representation and the given packet type, flags, and call id.
func NewHeader(pt PacketType, flags PFCFlags, callID uint32) Header {
	return Header{
		RPCVersion:         dcerpc.MajorVersion,
		RPCVersionMinor:    dcerpc.MinorVersion,
		PacketType:         pt,
		PacketFlags:        flags,
		DataRepresentation: DataRepresentationLittleEndian,
		CallID:             callID,
	}
}

// Marshal serializes the header into its 16-byte wire form.
func (h *Header) Marshal() ([]byte, error) {
	buf := make([]byte, HeaderSize)
	buf[0] = h.RPCVersion
	buf[1] = h.RPCVersionMinor
	buf[2] = byte(h.PacketType)
	buf[3] = byte(h.PacketFlags)
	copy(buf[4:8], h.DataRepresentation[:])
	binary.LittleEndian.PutUint16(buf[8:10], h.FragLength)
	binary.LittleEndian.PutUint16(buf[10:12], h.AuthLength)
	binary.LittleEndian.PutUint32(buf[12:16], h.CallID)
	return buf, nil
}

// Unmarshal parses a header from the start of data and returns the bytes consumed
// (always HeaderSize on success).
func (h *Header) Unmarshal(data []byte) (int, error) {
	if len(data) < HeaderSize {
		return 0, fmt.Errorf("PDU header truncated: have %d bytes, need %d", len(data), HeaderSize)
	}
	h.RPCVersion = data[0]
	h.RPCVersionMinor = data[1]
	h.PacketType = PacketType(data[2])
	h.PacketFlags = PFCFlags(data[3])
	copy(h.DataRepresentation[:], data[4:8])
	// The high nibble of the first DREP byte encodes the integer byte order: 1 is
	// little-endian (Windows, Samba), 0 is big-endian ([C706] section 14.1). Any
	// other value is not a defined integer representation.
	switch h.DataRepresentation[0] >> 4 {
	case drepLittleEndian, drepBigEndian:
	default:
		return 0, fmt.Errorf("unsupported data representation 0x%02x: integer order must be little- or big-endian", h.DataRepresentation[0])
	}
	order := h.ByteOrder()
	h.FragLength = order.Uint16(data[8:10])
	h.AuthLength = order.Uint16(data[10:12])
	h.CallID = order.Uint32(data[12:16])
	return HeaderSize, nil
}

// Integer byte-order codes carried in the high nibble of the first DREP octet
// ([C706] section 14.1).
const (
	drepBigEndian    = 0
	drepLittleEndian = 1
)

// ByteOrder returns the integer byte order the header's data representation
// selects: big-endian when the high nibble of the first DREP octet is 0,
// little-endian otherwise. A PDU body reads its own integer fields in this order,
// and the stub is decoded in it under the NDR receiver-makes-right rule.
func (h *Header) ByteOrder() binary.ByteOrder {
	if h.DataRepresentation[0]>>4 == drepBigEndian {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// String returns a human-readable one-line summary of the header.
func (h *Header) String() string {
	return fmt.Sprintf("DCE/RPC v%d.%d %s flags=%s frag_len=%d auth_len=%d call_id=%d",
		h.RPCVersion, h.RPCVersionMinor, h.PacketType, h.PacketFlags, h.FragLength, h.AuthLength, h.CallID)
}

// PeekHeader parses just the common header from the front of a complete PDU buffer.
// It is a convenience for dispatching on PacketType before fully unmarshalling.
func PeekHeader(data []byte) (*Header, error) {
	h := &Header{}
	if _, err := h.Unmarshal(data); err != nil {
		return nil, err
	}
	return h, nil
}
