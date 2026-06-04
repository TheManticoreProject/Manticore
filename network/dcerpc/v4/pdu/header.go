// Package pdu models connectionless (datagram) DCE/RPC protocol data units (PDUs):
// the fixed 80-octet common header and the bodies of the connectionless PDU types
// (request, ping, response, working, nocall, reject, ack, fault, cl_cancel, fack,
// cancel_ack) ([C706] chapter 12).
//
// The codec is two-layered. The generic PDU type (pdu.go) frames any connectionless
// PDU as a Header plus an opaque body whose length is the header's len field; this
// round-trips every PDU type at the wire level. The bodies that have a defined
// internal structure get their own typed codecs: FackBody (also used by nocall),
// CancelBody (cl_cancel), CancelAckBody (cancel_ack), and the status-code body
// shared by fault and reject (fault.go).
//
// Only the little-endian data representation (the high nibble of drep[0] == 1) is
// supported, matching the connection-oriented codec under network/dcerpc/v5/pdu; a
// big-endian drep is rejected on Unmarshal rather than silently mis-parsed. The
// three header UUIDs are encoded with the standard DCE uuid_t little-endian layout
// via windows/guid.
//
// References:
//   - [C706] section 12.6.3 (connectionless PDU common header fields):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [C706] chapter 10 (connectionless RPC protocol machines):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap10.htm
package pdu

import (
	"encoding/binary"
	"fmt"

	dcerpccl "github.com/TheManticoreProject/Manticore/network/dcerpc/v4"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// HeaderSize is the size, in bytes, of the connectionless common header.
const HeaderSize = 80

// NoHint is the reserved "no hint available" value for the ihint and ahint fields.
const NoHint uint16 = 0xFFFF

// DataRepresentationLittleEndian is the canonical 3-octet connectionless drep for
// little-endian integers, ASCII characters, and IEEE floating point. Note the
// connectionless drep is 3 octets, unlike the 4-octet connection-oriented packed
// drep.
var DataRepresentationLittleEndian = [3]byte{0x10, 0x00, 0x00}

// Header is the connectionless common header present at the start of every PDU
// ([C706] section 12.6.3.1, IDL type dc_rpc_cl_pkt_hdr_t).
//
// Wire layout (offsets in octets):
//
//	0      rpc_vers      uint8     (version in 4 LSBs)
//	1      ptype         uint8     (type in 5 LSBs)
//	2      flags1        uint8
//	3      flags2        uint8
//	4-6    drep          byte[3]
//	7      serial_hi     uint8
//	8-23   object        uuid_t    (16)
//	24-39  if_id         uuid_t    (16)
//	40-55  act_id        uuid_t    (16)
//	56-59  server_boot   uint32
//	60-63  if_vers       uint32
//	64-67  seqnum        uint32
//	68-69  opnum         uint16
//	70-71  ihint         uint16
//	72-73  ahint         uint16
//	74-75  len           uint16    (length of the body that follows)
//	76-77  fragnum       uint16
//	78     auth_proto    uint8
//	79     serial_lo     uint8
type Header struct {
	RPCVersion         uint8
	PacketType         PacketType
	Flags1             Flags1
	Flags2             Flags2
	DataRepresentation [3]byte
	SerialHi           uint8
	ObjectID           guid.GUID
	InterfaceID        guid.GUID
	ActivityID         guid.GUID
	ServerBoot         uint32
	InterfaceVersion   uint32
	SequenceNumber     uint32
	OpNum              uint16
	InterfaceHint      uint16
	ActivityHint       uint16
	// BodyLength is the len field: the size, in octets, of the body following the
	// header. PDU.Marshal sets it from the body length, so callers normally leave it
	// zero.
	BodyLength     uint16
	FragmentNumber uint16
	AuthProto      uint8
	SerialLo       uint8
}

// NewHeader returns a header initialized for connectionless protocol version 4 with
// little-endian data representation, the given packet type, and no interface/activity
// hint. The caller fills in the UUIDs, sequence number, opnum, and fragmentation
// fields as needed.
func NewHeader(pt PacketType) Header {
	return Header{
		RPCVersion:         dcerpccl.ProtocolVersion,
		PacketType:         pt,
		DataRepresentation: DataRepresentationLittleEndian,
		InterfaceHint:      NoHint,
		ActivityHint:       NoHint,
	}
}

// Serial reassembles the 16-bit serial number from its high and low halves, which
// are stored at opposite ends of the header (serial_hi at octet 7, serial_lo at
// octet 79).
func (h *Header) Serial() uint16 {
	return uint16(h.SerialHi)<<8 | uint16(h.SerialLo)
}

// Marshal serializes the header into its 80-octet wire form.
func (h *Header) Marshal() ([]byte, error) {
	buf := make([]byte, HeaderSize)
	buf[0] = h.RPCVersion
	buf[1] = byte(h.PacketType)
	buf[2] = byte(h.Flags1)
	buf[3] = byte(h.Flags2)
	copy(buf[4:7], h.DataRepresentation[:])
	buf[7] = h.SerialHi
	copy(buf[8:24], h.ObjectID.ToBytes())
	copy(buf[24:40], h.InterfaceID.ToBytes())
	copy(buf[40:56], h.ActivityID.ToBytes())
	binary.LittleEndian.PutUint32(buf[56:60], h.ServerBoot)
	binary.LittleEndian.PutUint32(buf[60:64], h.InterfaceVersion)
	binary.LittleEndian.PutUint32(buf[64:68], h.SequenceNumber)
	binary.LittleEndian.PutUint16(buf[68:70], h.OpNum)
	binary.LittleEndian.PutUint16(buf[70:72], h.InterfaceHint)
	binary.LittleEndian.PutUint16(buf[72:74], h.ActivityHint)
	binary.LittleEndian.PutUint16(buf[74:76], h.BodyLength)
	binary.LittleEndian.PutUint16(buf[76:78], h.FragmentNumber)
	buf[78] = h.AuthProto
	buf[79] = h.SerialLo
	return buf, nil
}

// Unmarshal parses a header from the start of data and returns the bytes consumed
// (always HeaderSize on success).
func (h *Header) Unmarshal(data []byte) (int, error) {
	if len(data) < HeaderSize {
		return 0, fmt.Errorf("connectionless PDU header truncated: have %d bytes, need %d", len(data), HeaderSize)
	}
	h.RPCVersion = data[0]
	h.PacketType = PacketType(data[1])
	h.Flags1 = Flags1(data[2])
	h.Flags2 = Flags2(data[3])
	copy(h.DataRepresentation[:], data[4:7])
	// Only little-endian integer representation is supported. The high nibble of the
	// first drep octet encodes the integer byte order (1 = little-endian).
	if h.DataRepresentation[0]>>4 != 1 {
		return 0, fmt.Errorf("unsupported data representation 0x%02x: only little-endian is supported", h.DataRepresentation[0])
	}
	h.SerialHi = data[7]
	h.ObjectID.FromRawBytes(data[8:24])
	h.InterfaceID.FromRawBytes(data[24:40])
	h.ActivityID.FromRawBytes(data[40:56])
	h.ServerBoot = binary.LittleEndian.Uint32(data[56:60])
	h.InterfaceVersion = binary.LittleEndian.Uint32(data[60:64])
	h.SequenceNumber = binary.LittleEndian.Uint32(data[64:68])
	h.OpNum = binary.LittleEndian.Uint16(data[68:70])
	h.InterfaceHint = binary.LittleEndian.Uint16(data[70:72])
	h.ActivityHint = binary.LittleEndian.Uint16(data[72:74])
	h.BodyLength = binary.LittleEndian.Uint16(data[74:76])
	h.FragmentNumber = binary.LittleEndian.Uint16(data[76:78])
	h.AuthProto = data[78]
	h.SerialLo = data[79]
	return HeaderSize, nil
}

// String returns a human-readable one-line summary of the header.
func (h *Header) String() string {
	return fmt.Sprintf("DCE/RPC(cl) v%d %s flags1=%s flags2=%s if=%s act=%s seq=%d op=%d frag=%d len=%d",
		h.RPCVersion, h.PacketType, h.Flags1, h.Flags2,
		h.InterfaceID.ToFormatD(), h.ActivityID.ToFormatD(),
		h.SequenceNumber, h.OpNum, h.FragmentNumber, h.BodyLength)
}
