package pdu

import (
	"fmt"
)

// Auth3 is an rpc_auth_3 PDU: the optional third leg of an authenticated bind. After a
// bind that carried an authentication token and a bind_ack that returned the security
// provider's challenge, the client sends an auth3 carrying the final token (for NTLM,
// the AUTHENTICATE message) to complete the exchange. The server sends no reply.
//
// The PDU is the common header, a 4-byte pad (a vestigial max_xmit_frag/max_recv_frag
// that the server ignores), then the sec_trailer and the auth_value. The header's
// auth_length counts only AuthValue.
//
// References:
//   - [C706] section 12.6.4.2 ("The rpc_auth_3 PDU"):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.10 PDU Types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/cef9d684-f09f-4533-a54c-9255079d3e1d
type Auth3 struct {
	Header     Header
	SecTrailer SecTrailer
	AuthValue  []byte
}

// Marshal serializes the complete auth3 PDU, filling in the header's packet type,
// frag_length, and auth_length.
func (a *Auth3) Marshal() ([]byte, error) {
	if len(a.AuthValue) == 0 {
		return nil, fmt.Errorf("auth3 PDU has no auth_value")
	}

	body := make([]byte, 4) // pad (ignored max_xmit_frag/max_recv_frag)
	body = append(body, a.SecTrailer.Marshal()...)
	body = append(body, a.AuthValue...)

	if a.Header.RPCVersion == 0 && a.Header.DataRepresentation == ([4]byte{}) {
		a.Header = NewHeader(PacketTypeAuth3, a.Header.PacketFlags, a.Header.CallID)
	}
	a.Header.PacketType = PacketTypeAuth3
	a.Header.AuthLength = uint16(len(a.AuthValue))
	a.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := a.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete auth3 PDU and returns the bytes consumed.
func (a *Auth3) Unmarshal(data []byte) (int, error) {
	pos, err := a.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if a.Header.PacketType != PacketTypeAuth3 {
		return 0, fmt.Errorf("not an auth3 PDU: packet type is %s", a.Header.PacketType)
	}
	pos += 4 // skip the 4-byte pad
	if len(data) < pos+SecTrailerSize {
		return 0, fmt.Errorf("auth3 PDU sec_trailer truncated")
	}
	if err := a.SecTrailer.Unmarshal(data[pos:]); err != nil {
		return 0, fmt.Errorf("auth3 PDU: %w", err)
	}
	pos += SecTrailerSize

	authLen := int(a.Header.AuthLength)
	if len(data) < pos+authLen {
		return 0, fmt.Errorf("auth3 PDU auth_value truncated: have %d bytes, need %d", len(data)-pos, authLen)
	}
	a.AuthValue = append([]byte(nil), data[pos:pos+authLen]...)
	return pos + authLen, nil
}
