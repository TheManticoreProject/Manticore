package pdu

import (
	"encoding/binary"
	"fmt"
)

// Common connection-oriented RPC fault status codes (nca_* / rpc_s_*). This is a
// subset; unknown codes are reported numerically by FaultStatus.String.
//
// References:
//   - [C706] section 12.6.3.1 ("The fault PDU") and Appendix E (reject status codes):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.5 Fault PDU:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/55c10f44-9037-4d51-aaff-c146bd0f1988
const (
	NCASStatusOK             uint32 = 0x00000000
	NCASOpRngError           uint32 = 0x1C010002 // nca_s_op_rng_error: invalid opnum
	NCASUnkIf                uint32 = 0x1C010003 // nca_s_unk_if: unknown interface
	NCASProtoError           uint32 = 0x1C01000B // nca_s_proto_error
	NCASFaultIntDivByZero    uint32 = 0x1C000001 // nca_s_fault_int_div_by_zero
	NCASFaultAddrError       uint32 = 0x1C000002 // nca_s_fault_addr_error
	NCASFaultNDR             uint32 = 0x000006F7 // nca_s_fault_ndr: NDR could not decode
	NCASFaultAccessDenied    uint32 = 0x00000005 // nca_s_fault_access_denied
	NCASFaultContextMismatch uint32 = 0x1C00001A // nca_s_fault_context_mismatch
)

// FaultStatus is a fault PDU status code, with a descriptive String.
type FaultStatus uint32

// String returns a mnemonic for known status codes, otherwise the hex value.
func (s FaultStatus) String() string {
	switch uint32(s) {
	case NCASStatusOK:
		return "ok"
	case NCASOpRngError:
		return "nca_s_op_rng_error"
	case NCASUnkIf:
		return "nca_s_unk_if"
	case NCASProtoError:
		return "nca_s_proto_error"
	case NCASFaultIntDivByZero:
		return "nca_s_fault_int_div_by_zero"
	case NCASFaultAddrError:
		return "nca_s_fault_addr_error"
	case NCASFaultNDR:
		return "nca_s_fault_ndr"
	case NCASFaultAccessDenied:
		return "nca_s_fault_access_denied"
	case NCASFaultContextMismatch:
		return "nca_s_fault_context_mismatch"
	default:
		return fmt.Sprintf("0x%08x", uint32(s))
	}
}

// Fault is a fault PDU: it reports that a call failed, carrying a status code and any
// error stub data.
//
// References:
//   - [C706] section 12.6.3.1 ("The fault PDU"):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.5 Fault PDU:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/55c10f44-9037-4d51-aaff-c146bd0f1988
type Fault struct {
	Header      Header
	AllocHint   uint32
	ContextID   uint16
	CancelCount uint8
	Status      uint32
	Stub        []byte
}

// Marshal serializes the complete fault PDU.
func (f *Fault) Marshal() ([]byte, error) {
	body := make([]byte, 16)
	binary.LittleEndian.PutUint32(body[0:4], f.AllocHint)
	binary.LittleEndian.PutUint16(body[4:6], f.ContextID)
	body[6] = f.CancelCount
	body[7] = 0 // reserved
	binary.LittleEndian.PutUint32(body[8:12], f.Status)
	// body[12:16] reserved2
	body = append(body, f.Stub...)

	if f.Header.RPCVersion == 0 && f.Header.DataRepresentation == ([4]byte{}) {
		f.Header = NewHeader(PacketTypeFault, f.Header.PacketFlags, f.Header.CallID)
	}
	f.Header.PacketType = PacketTypeFault
	f.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := f.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete fault PDU and returns the bytes consumed.
func (f *Fault) Unmarshal(data []byte) (int, error) {
	pos, err := f.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if f.Header.PacketType != PacketTypeFault {
		return 0, fmt.Errorf("not a fault PDU: packet type is %s", f.Header.PacketType)
	}
	if len(data) < pos+16 {
		return 0, fmt.Errorf("fault PDU truncated")
	}
	f.AllocHint = binary.LittleEndian.Uint32(data[pos : pos+4])
	f.ContextID = binary.LittleEndian.Uint16(data[pos+4 : pos+6])
	f.CancelCount = data[pos+6]
	f.Status = binary.LittleEndian.Uint32(data[pos+8 : pos+12])
	pos += 16

	end, err := stubEnd(&f.Header, len(data))
	if err != nil {
		return 0, err
	}
	if end < pos {
		end = pos
	}
	f.Stub = append([]byte(nil), data[pos:end]...)
	return end, nil
}

// Error implements the error interface so a Fault can be returned directly from a
// call. The message includes the mnemonic status code.
func (f *Fault) Error() string {
	return fmt.Sprintf("DCE/RPC fault: status %s (ctx=%d)", FaultStatus(f.Status), f.ContextID)
}

// String returns a one-line summary.
func (f *Fault) String() string {
	return fmt.Sprintf("fault ctx=%d cancel_count=%d status=%s stub=%d bytes", f.ContextID, f.CancelCount, FaultStatus(f.Status), len(f.Stub))
}
