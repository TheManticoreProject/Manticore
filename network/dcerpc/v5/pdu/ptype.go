package pdu

// PacketType identifies the kind of a connection-oriented DCE/RPC PDU, as carried in
// the ptype field of the common header.
//
// References:
//   - [C706] section 12.6.4 (PDU types):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.10 PDU Types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/cef9d684-f09f-4533-a54c-9255079d3e1d
type PacketType uint8

const (
	PacketTypeRequest          PacketType = 0
	PacketTypePing             PacketType = 1
	PacketTypeResponse         PacketType = 2
	PacketTypeFault            PacketType = 3
	PacketTypeWorking          PacketType = 4
	PacketTypeNoCall           PacketType = 5
	PacketTypeReject           PacketType = 6
	PacketTypeAck              PacketType = 7
	PacketTypeClCancel         PacketType = 8
	PacketTypeFack             PacketType = 9
	PacketTypeCancelAck        PacketType = 10
	PacketTypeBind             PacketType = 11
	PacketTypeBindAck          PacketType = 12
	PacketTypeBindNak          PacketType = 13
	PacketTypeAlterContext     PacketType = 14
	PacketTypeAlterContextResp PacketType = 15
	PacketTypeAuth3            PacketType = 16
	PacketTypeShutdown         PacketType = 17
	PacketTypeCoCancel         PacketType = 18
	PacketTypeOrphaned         PacketType = 19
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
	case PacketTypeBind:
		return "bind"
	case PacketTypeBindAck:
		return "bind_ack"
	case PacketTypeBindNak:
		return "bind_nak"
	case PacketTypeAlterContext:
		return "alter_context"
	case PacketTypeAlterContextResp:
		return "alter_context_resp"
	case PacketTypeAuth3:
		return "auth3"
	case PacketTypeShutdown:
		return "shutdown"
	case PacketTypeCoCancel:
		return "co_cancel"
	case PacketTypeOrphaned:
		return "orphaned"
	default:
		return "unknown"
	}
}
