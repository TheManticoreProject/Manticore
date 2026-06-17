package pdu

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Request is a request PDU: it carries a method call (opnum) and its marshalled
// arguments (stub data) to the server.
//
// When the PFC_OBJECT_UUID flag is set, a 16-byte object UUID precedes the stub data;
// otherwise it is absent. Stub data is 8-octet aligned, which the fixed 8-byte body
// (plus an optional 16-byte object UUID) already satisfies.
//
// Reference: [C706] section 12.6.3.1 ("The request PDU"):
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type Request struct {
	Header     Header
	AllocHint  uint32
	ContextID  uint16
	Opnum      uint16
	ObjectUUID *guid.GUID // present iff PFC_OBJECT_UUID is set
	Stub       []byte
}

// Marshal serializes the complete request PDU. If ObjectUUID is non-nil the
// PFC_OBJECT_UUID flag is set automatically. AllocHint defaults to len(Stub) when
// left zero.
func (r *Request) Marshal() ([]byte, error) {
	allocHint := r.AllocHint
	if allocHint == 0 {
		allocHint = uint32(len(r.Stub))
	}

	body := make([]byte, 8)
	binary.LittleEndian.PutUint32(body[0:4], allocHint)
	binary.LittleEndian.PutUint16(body[4:6], r.ContextID)
	binary.LittleEndian.PutUint16(body[6:8], r.Opnum)

	if r.ObjectUUID != nil {
		body = append(body, r.ObjectUUID.ToBytes()...)
		r.Header.PacketFlags |= PFCObjectUuid
	}
	body = append(body, r.Stub...)

	if r.Header.RPCVersion == 0 && r.Header.DataRepresentation == ([4]byte{}) {
		flags := r.Header.PacketFlags
		r.Header = NewHeader(PacketTypeRequest, flags, r.Header.CallID)
	}
	r.Header.PacketType = PacketTypeRequest
	r.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := r.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete request PDU and returns the bytes consumed.
func (r *Request) Unmarshal(data []byte) (int, error) {
	pos, err := r.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if r.Header.PacketType != PacketTypeRequest {
		return 0, fmt.Errorf("not a request PDU: packet type is %s", r.Header.PacketType)
	}
	if len(data) < pos+8 {
		return 0, fmt.Errorf("request PDU truncated")
	}
	r.AllocHint = binary.LittleEndian.Uint32(data[pos : pos+4])
	r.ContextID = binary.LittleEndian.Uint16(data[pos+4 : pos+6])
	r.Opnum = binary.LittleEndian.Uint16(data[pos+6 : pos+8])
	pos += 8

	if r.Header.PacketFlags.Has(PFCObjectUuid) {
		if len(data) < pos+16 {
			return 0, fmt.Errorf("request PDU object UUID truncated")
		}
		r.ObjectUUID = &guid.GUID{}
		r.ObjectUUID.FromRawBytes(data[pos : pos+16])
		pos += 16
	}

	end, err := stubEnd(&r.Header, len(data))
	if err != nil {
		return 0, err
	}
	if end < pos {
		return 0, fmt.Errorf("request PDU stub bounds invalid")
	}
	r.Stub = append([]byte(nil), data[pos:end]...)
	return end, nil
}

// String returns a one-line summary.
func (r *Request) String() string {
	return fmt.Sprintf("request ctx=%d opnum=%d alloc_hint=%d stub=%d bytes", r.ContextID, r.Opnum, r.AllocHint, len(r.Stub))
}

// Response is a response PDU: it carries the marshalled return values (stub data) of
// a successful call back to the client.
//
// Reference: [C706] section 12.6.3.1 ("The response PDU"):
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type Response struct {
	Header      Header
	AllocHint   uint32
	ContextID   uint16
	CancelCount uint8
	Stub        []byte
}

// Marshal serializes the complete response PDU.
func (r *Response) Marshal() ([]byte, error) {
	allocHint := r.AllocHint
	if allocHint == 0 {
		allocHint = uint32(len(r.Stub))
	}

	body := make([]byte, 8)
	binary.LittleEndian.PutUint32(body[0:4], allocHint)
	binary.LittleEndian.PutUint16(body[4:6], r.ContextID)
	body[6] = r.CancelCount
	body[7] = 0 // reserved
	body = append(body, r.Stub...)

	if r.Header.RPCVersion == 0 && r.Header.DataRepresentation == ([4]byte{}) {
		r.Header = NewHeader(PacketTypeResponse, r.Header.PacketFlags, r.Header.CallID)
	}
	r.Header.PacketType = PacketTypeResponse
	r.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := r.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete response PDU and returns the bytes consumed.
func (r *Response) Unmarshal(data []byte) (int, error) {
	pos, err := r.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if r.Header.PacketType != PacketTypeResponse {
		return 0, fmt.Errorf("not a response PDU: packet type is %s", r.Header.PacketType)
	}
	if len(data) < pos+8 {
		return 0, fmt.Errorf("response PDU truncated")
	}
	r.AllocHint = binary.LittleEndian.Uint32(data[pos : pos+4])
	r.ContextID = binary.LittleEndian.Uint16(data[pos+4 : pos+6])
	r.CancelCount = data[pos+6]
	pos += 8

	end, err := stubEnd(&r.Header, len(data))
	if err != nil {
		return 0, err
	}
	if end < pos {
		return 0, fmt.Errorf("response PDU stub bounds invalid")
	}
	r.Stub = append([]byte(nil), data[pos:end]...)
	return end, nil
}

// String returns a one-line summary.
func (r *Response) String() string {
	return fmt.Sprintf("response ctx=%d cancel_count=%d alloc_hint=%d stub=%d bytes", r.ContextID, r.CancelCount, r.AllocHint, len(r.Stub))
}

// stubEnd returns the index at which the stub data ends, derived from the header's
// frag_length and auth_length. When the PDU carries an authentication verifier, the
// trailing bytes are laid out as [stub][auth_pad][sec_trailer][auth_value], so the stub
// ends auth_length + SecTrailerSize bytes before frag_length. The caller strips the
// auth padding (recorded in the sec_trailer's auth_pad_length) separately, since this
// helper does not parse the trailer. It clamps to the available buffer so a truncated
// read does not produce an out-of-bounds slice.
func stubEnd(h *Header, available int) (int, error) {
	end := int(h.FragLength)
	if h.AuthLength > 0 {
		end -= int(h.AuthLength) + SecTrailerSize
	}
	if end > available {
		end = available
	}
	if end < HeaderSize {
		return 0, fmt.Errorf("frag_length %d shorter than header", h.FragLength)
	}
	return end, nil
}
