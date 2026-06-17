package pdu

import (
	"encoding/binary"
	"fmt"
)

// Authentication service identifiers (auth_type), as carried in the sec_trailer of an
// authenticated PDU.
//
// Reference: [MS-RPCE] 2.2.1.1.7 Security Providers:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/d4097450-c62f-484b-872f-ddf59a7a0d36
const (
	AuthTypeNone         uint8 = 0x00
	AuthTypeGSSNegotiate uint8 = 0x09 // SPNEGO
	AuthTypeNTLMSSP      uint8 = 0x0A // RPC_C_AUTHN_WINNT
	AuthTypeGSSKerberos  uint8 = 0x10
)

// Authentication levels (auth_level), as carried in the sec_trailer. Higher levels
// subsume the protection of lower ones.
//
// Reference: [MS-RPCE] 2.2.1.1.8 Authentication Levels:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/425a7c53-c33a-4868-8e5b-2a850d40dc73
const (
	AuthLevelDefault      uint8 = 0x00
	AuthLevelNone         uint8 = 0x01
	AuthLevelConnect      uint8 = 0x02
	AuthLevelCall         uint8 = 0x03
	AuthLevelPkt          uint8 = 0x04
	AuthLevelPktIntegrity uint8 = 0x05
	AuthLevelPktPrivacy   uint8 = 0x06
)

// SecTrailerSize is the size, in bytes, of the sec_trailer (auth_verifier_co_t) header
// that precedes the auth_value in an authenticated PDU.
const SecTrailerSize = 8

// SecTrailer is the fixed 8-byte security trailer (auth_verifier_co_t) that sits
// between a PDU's body and its authentication token (auth_value). The header's
// auth_length counts only the auth_value that follows this trailer, not the trailer
// itself.
//
// References:
//   - [C706] section 13.2.6.1 (the connection-oriented authentication verifier):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap13.htm
//   - [MS-RPCE] 2.2.2.11 auth_verifier_co (sec_trailer):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/c0fe71f3-0fd1-4a0d-aa3f-b6a36e3fbe81
type SecTrailer struct {
	AuthType      uint8  // security provider (auth_type)
	AuthLevel     uint8  // protection level (auth_level)
	AuthPadLength uint8  // bytes of padding inserted before the trailer to 4-byte-align it
	AuthReserved  uint8  // must be zero
	AuthContextID uint32 // identifies the security context within the connection
}

// Marshal serializes the 8-byte sec_trailer.
func (s *SecTrailer) Marshal() []byte {
	buf := make([]byte, SecTrailerSize)
	buf[0] = s.AuthType
	buf[1] = s.AuthLevel
	buf[2] = s.AuthPadLength
	buf[3] = s.AuthReserved
	binary.LittleEndian.PutUint32(buf[4:8], s.AuthContextID)
	return buf
}

// Unmarshal parses an 8-byte sec_trailer from the front of data.
func (s *SecTrailer) Unmarshal(data []byte) error {
	if len(data) < SecTrailerSize {
		return fmt.Errorf("sec_trailer truncated: have %d bytes, need %d", len(data), SecTrailerSize)
	}
	s.AuthType = data[0]
	s.AuthLevel = data[1]
	s.AuthPadLength = data[2]
	s.AuthReserved = data[3]
	s.AuthContextID = binary.LittleEndian.Uint32(data[4:8])
	return nil
}

// appendAuthVerifier appends an authentication verifier to a PDU body in place: it
// pads body to a 4-byte boundary (measured from the start of the PDU, i.e. including
// the 16-byte header), records that padding in the sec_trailer's auth_pad_length, then
// appends the sec_trailer and the auth_value. When authValue is empty it returns body
// unchanged. The caller is responsible for setting the header's auth_length to
// len(authValue).
func appendAuthVerifier(body []byte, st *SecTrailer, authValue []byte) []byte {
	if len(authValue) == 0 {
		return body
	}
	pad := align4(HeaderSize + len(body))
	if pad > 0 {
		body = append(body, make([]byte, pad)...)
	}
	st.AuthPadLength = uint8(pad)
	body = append(body, st.Marshal()...)
	body = append(body, authValue...)
	return body
}

// ExtractAuthVerifier returns the sec_trailer and auth_value from the tail of a
// complete PDU, located using the common header's auth_length: the last auth_length
// bytes are the auth_value, preceded by the 8-byte sec_trailer. It returns nil, nil
// when the PDU carries no authentication verifier (auth_length == 0).
func ExtractAuthVerifier(raw []byte) (*SecTrailer, []byte, error) {
	hdr, err := PeekHeader(raw)
	if err != nil {
		return nil, nil, err
	}
	authLen := int(hdr.AuthLength)
	if authLen == 0 {
		return nil, nil, nil
	}
	fragLen := int(hdr.FragLength)
	if fragLen > len(raw) {
		fragLen = len(raw)
	}
	trailerStart := fragLen - authLen - SecTrailerSize
	if trailerStart < HeaderSize {
		return nil, nil, fmt.Errorf("auth verifier overruns PDU: frag_length=%d auth_length=%d", hdr.FragLength, authLen)
	}
	var st SecTrailer
	if err := st.Unmarshal(raw[trailerStart:]); err != nil {
		return nil, nil, err
	}
	authValue := append([]byte(nil), raw[trailerStart+SecTrailerSize:fragLen]...)
	return &st, authValue, nil
}
