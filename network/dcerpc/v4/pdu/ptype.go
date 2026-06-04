package pdu

// PacketType identifies the kind of a connectionless DCE/RPC PDU, as carried in the
// 5 least significant bits of the ptype field of the common header.
//
// The connectionless ptype values differ from the connection-oriented ones (which
// reuse 0-10 for an overlapping but not identical set and add bind/alter_context at
// 11+). The values below are the complete connectionless set.
//
// References:
//   - [C706] section 12.6.4 (PDU types) and table in section 12.5:
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type PacketType uint8

const (
	PacketTypeRequest   PacketType = 0
	PacketTypePing      PacketType = 1
	PacketTypeResponse  PacketType = 2
	PacketTypeFault     PacketType = 3
	PacketTypeWorking   PacketType = 4
	PacketTypeNoCall    PacketType = 5
	PacketTypeReject    PacketType = 6
	PacketTypeAck       PacketType = 7
	PacketTypeClCancel  PacketType = 8
	PacketTypeFack      PacketType = 9
	PacketTypeCancelAck PacketType = 10
)

// String returns the mnemonic name of the packet type.
func (t PacketType) String() string {
	switch t {
	case PacketTypeRequest:
		return "request"
	case PacketTypePing:
		return "ping"
	case PacketTypeResponse:
		return "response"
	case PacketTypeFault:
		return "fault"
	case PacketTypeWorking:
		return "working"
	case PacketTypeNoCall:
		return "nocall"
	case PacketTypeReject:
		return "reject"
	case PacketTypeAck:
		return "ack"
	case PacketTypeClCancel:
		return "cl_cancel"
	case PacketTypeFack:
		return "fack"
	case PacketTypeCancelAck:
		return "cancel_ack"
	default:
		return "unknown"
	}
}
