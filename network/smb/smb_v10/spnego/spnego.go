package spnego

import (
	"bytes"
	"encoding/asn1"
	"errors"
	"fmt"
)

// OIDs for various authentication mechanisms
var (
	// SPNEGO OID: 1.3.6.1.5.5.2
	SpnegoOID = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 2}

	// NTLM OID: 1.3.6.1.4.1.311.2.2.10
	NtlmOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 2, 10}

	// Kerberos OID: 1.2.840.113554.1.2.2
	KerberosOID = asn1.ObjectIdentifier{1, 2, 840, 113554, 1, 2, 2}
)

// NegTokenInit is the initial SPNEGO token sent by the client
type NegTokenInit struct {
	MechTypes    []asn1.ObjectIdentifier `asn1:"explicit,tag:0"`
	ReqFlags     asn1.BitString          `asn1:"explicit,optional,tag:1"`
	MechToken    []byte                  `asn1:"explicit,optional,tag:2"`
	MechTokenMIC []byte                  `asn1:"explicit,optional,tag:3"`
}

// NegTokenResp is the response token sent by the server
type NegTokenResp struct {
	NegState      asn1.Enumerated       `asn1:"explicit,optional,tag:0"`
	SupportedMech asn1.ObjectIdentifier `asn1:"explicit,optional,tag:1"`
	ResponseToken []byte                `asn1:"explicit,optional,tag:2"`
	MechListMIC   []byte                `asn1:"explicit,optional,tag:3"`
}

// NegState values
const (
	Accept           asn1.Enumerated = 0
	AcceptIncomplete asn1.Enumerated = 1
	Reject           asn1.Enumerated = 2
	RequestMIC       asn1.Enumerated = 3
)

// GSS API constants
const (
	GSS_API_SPNEGO = 0x60
)

// CreateNegTokenInit creates an ASN.1 encoded SPNEGO NegTokenInit
func CreateNegTokenInit(ntlmToken []byte) ([]byte, error) {
	// Create the NegTokenInit structure
	token := NegTokenInit{
		MechTypes: []asn1.ObjectIdentifier{NtlmOID},
		MechToken: ntlmToken,
	}

	// Encode the NegTokenInit
	tokenBytes, err := asn1.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NegTokenInit: %v", err)
	}

	// Create the SPNEGO header
	spnegoBytes, err := asn1.Marshal(SpnegoOID)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SPNEGO OID: %v", err)
	}

	// Combine into GSS-API format
	buffer := new(bytes.Buffer)

	// GSS-API header
	buffer.WriteByte(GSS_API_SPNEGO)

	// Length of the remaining data
	totalLen := len(spnegoBytes) + len(tokenBytes)
	if totalLen < 128 {
		buffer.WriteByte(byte(totalLen))
	} else {
		// For longer lengths, we need to use the long form
		lenBytes := encodeLength(totalLen)
		buffer.WriteByte(byte(0x80 | len(lenBytes)))
		buffer.Write(lenBytes)
	}

	// Write the SPNEGO OID and token
	buffer.Write(spnegoBytes)
	buffer.Write(tokenBytes)

	return buffer.Bytes(), nil
}

// ParseNegTokenResp parses a server's NegTokenResp
func ParseNegTokenResp(data []byte) (*NegTokenResp, error) {
	// Skip GSS-API header
	if len(data) < 2 || data[0] != GSS_API_SPNEGO {
		return nil, errors.New("invalid GSS-API header")
	}

	// Skip length bytes
	offset := 2
	if data[1]&0x80 != 0 {
		numLenBytes := int(data[1] & 0x7F)
		offset = 2 + numLenBytes
	}

	// Skip OID
	var oid asn1.ObjectIdentifier
	rest, err := asn1.Unmarshal(data[offset:], &oid)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OID: %v", err)
	}

	// Parse the NegTokenResp
	var resp NegTokenResp
	_, err = asn1.Unmarshal(rest, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal NegTokenResp: %v", err)
	}

	return &resp, nil
}

// CreateNegTokenResp creates an ASN.1 encoded SPNEGO NegTokenResp
func CreateNegTokenResp(state asn1.Enumerated, mech asn1.ObjectIdentifier, token []byte) ([]byte, error) {
	resp := NegTokenResp{
		NegState:      state,
		SupportedMech: mech,
		ResponseToken: token,
	}

	// Encode the NegTokenResp
	respBytes, err := asn1.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NegTokenResp: %v", err)
	}

	// Create the SPNEGO header
	spnegoBytes, err := asn1.Marshal(SpnegoOID)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SPNEGO OID: %v", err)
	}

	// Combine into GSS-API format
	buffer := new(bytes.Buffer)

	// GSS-API header
	buffer.WriteByte(GSS_API_SPNEGO)

	// Length of the remaining data
	totalLen := len(spnegoBytes) + len(respBytes)
	if totalLen < 128 {
		buffer.WriteByte(byte(totalLen))
	} else {
		// For longer lengths, we need to use the long form
		lenBytes := encodeLength(totalLen)
		buffer.WriteByte(byte(0x80 | len(lenBytes)))
		buffer.Write(lenBytes)
	}

	// Write the SPNEGO OID and token
	buffer.Write(spnegoBytes)
	buffer.Write(respBytes)

	return buffer.Bytes(), nil
}

// encodeLength encodes a length in ASN.1 DER format
func encodeLength(length int) []byte {
	if length < 128 {
		return []byte{byte(length)}
	}

	// Calculate how many bytes we need
	numBytes := 0
	temp := length
	for temp > 0 {
		temp >>= 8
		numBytes++
	}

	// Encode the length
	result := make([]byte, numBytes)
	for i := numBytes - 1; i >= 0; i-- {
		result[i] = byte(length & 0xFF)
		length >>= 8
	}

	return result
}

// ExtractNTLMToken extracts the NTLM token from a SPNEGO token
func ExtractNTLMToken(spnegoToken []byte) ([]byte, error) {
	// Skip GSS-API header
	if len(spnegoToken) < 2 || spnegoToken[0] != GSS_API_SPNEGO {
		return nil, errors.New("invalid GSS-API header")
	}

	// Skip length bytes
	offset := 2
	if spnegoToken[1]&0x80 != 0 {
		numLenBytes := int(spnegoToken[1] & 0x7F)
		offset = 2 + numLenBytes
	}

	// Skip OID
	var oid asn1.ObjectIdentifier
	rest, err := asn1.Unmarshal(spnegoToken[offset:], &oid)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OID: %v", err)
	}

	// Try to parse as NegTokenInit
	var init NegTokenInit
	_, err = asn1.Unmarshal(rest, &init)
	if err == nil && len(init.MechToken) > 0 {
		return init.MechToken, nil
	}

	// Try to parse as NegTokenResp
	var resp NegTokenResp
	_, err = asn1.Unmarshal(rest, &resp)
	if err == nil && len(resp.ResponseToken) > 0 {
		return resp.ResponseToken, nil
	}

	return nil, errors.New("no NTLM token found in SPNEGO message")
}
