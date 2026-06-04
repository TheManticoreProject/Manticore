// Package pdu models connection-oriented DCE/RPC protocol data units (PDUs): the
// 16-byte common header and the Bind, Bind_Ack, Bind_Nak, Request, Response, and
// Fault bodies ([C706] section 12, [MS-RPCE] section 2.2).
//
// Each PDU type embeds the common Header and represents a whole PDU. Marshal returns
// the complete on-the-wire bytes (header included) and sets the header's frag_length;
// Unmarshal parses a complete PDU and returns the number of bytes consumed.
//
// Only the little-endian data representation (DREP first byte 0x10) is supported,
// which is what Windows and Samba always send; a big-endian DREP is rejected on
// Unmarshal rather than silently mis-parsed.
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
	// Only little-endian integer representation is supported. The high nibble of the
	// first DREP byte encodes the integer byte order (1 = little-endian).
	if h.DataRepresentation[0]>>4 != 1 {
		return 0, fmt.Errorf("unsupported data representation 0x%02x: only little-endian is supported", h.DataRepresentation[0])
	}
	h.FragLength = binary.LittleEndian.Uint16(data[8:10])
	h.AuthLength = binary.LittleEndian.Uint16(data[10:12])
	h.CallID = binary.LittleEndian.Uint32(data[12:16])
	return HeaderSize, nil
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
