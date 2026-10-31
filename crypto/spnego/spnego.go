package spnego

import (
	"bytes"
	"encoding/asn1"
	"errors"
	"fmt"
)

// GSS API constants
const (
	GSS_API_SPNEGO = 0x60 // [APPLICATION 0]
)

// wrapSPNEGO wraps an already-DER-marshaled NegToken... bytes with the SPNEGO OID and [0] tag.
func wrapSPNEGO(innerBytes []byte) ([]byte, error) {
	// Marshal the SPNEGO OID
	oidBytes, err := asn1.Marshal(SpnegoOID)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SPNEGO OID: %v", err)
	}
	// Wrap inner in [0] EXPLICIT
	raw := asn1.RawValue{
		Class:      asn1.ClassContextSpecific,
		Tag:        0,
		IsCompound: true,
		Bytes:      innerBytes,
	}
	wrappedInner, err := asn1.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap inner token: %v", err)
	}

	// Build final buffer: GSS tag, length, OID, wrappedInner
	buf := &bytes.Buffer{}
	buf.WriteByte(GSS_API_SPNEGO)

	totalLen := len(oidBytes) + len(wrappedInner)
	writeLength(buf, totalLen)

	buf.Write(oidBytes)
	buf.Write(wrappedInner)

	return buf.Bytes(), nil
}

// writeLength writes DER length octets for the given length
func writeLength(buf *bytes.Buffer, length int) {
	if length < 128 {
		buf.WriteByte(byte(length))
		return
	}
	// long form
	// determine number of bytes
	var lenBytes []byte
	tmp := length
	for tmp > 0 {
		lenBytes = append([]byte{byte(tmp & 0xFF)}, lenBytes...)
		tmp >>= 8
	}
	buf.WriteByte(0x80 | byte(len(lenBytes)))
	buf.Write(lenBytes)
}

// ExtractNTLMToken extracts the NTLM token from a SPNEGO token (init or resp)
func ExtractNTLMToken(spnegoToken []byte) ([]byte, error) {
	// A continuation token is the bare negTokenResp [1] choice, without the GSS
	// header (RFC 4178 section 4.2).
	var blob SecurityBlob
	if _, err := blob.Unmarshal(spnegoToken); err == nil {
		var resp NegTokenResp
		if _, err := resp.Unmarshal(blob.Data); err == nil && len(resp.ResponseToken) > 0 {
			return resp.ResponseToken, nil
		}
		return nil, errors.New("no NTLM token found in SPNEGO message")
	}

	inner, err := unwrapSPNEGO(spnegoToken)
	if err != nil {
		return nil, err
	}
	// Try NegTokenInit
	var init NegTokenInit
	if _, err := asn1.Unmarshal(inner, &init); err == nil && len(init.MechToken) > 0 {
		return init.MechToken, nil
	}
	// Try NegTokenResp. Its fields are EXPLICIT-tagged, which the struct tags on
	// NegTokenResp do not describe, so it is decoded by its own Unmarshal.
	var resp NegTokenResp
	if _, err := resp.Unmarshal(inner); err == nil && len(resp.ResponseToken) > 0 {
		return resp.ResponseToken, nil
	}
	return nil, errors.New("no NTLM token found in SPNEGO message")
}

// ParseNegTokenInit decodes the GSS-wrapped NegTokenInit an initiator sends
// first, so an acceptor can tell which mechanism its optimistic token is for.
func ParseNegTokenInit(spnegoToken []byte) (*NegTokenInit, error) {
	inner, err := unwrapSPNEGO(spnegoToken)
	if err != nil {
		return nil, err
	}
	var init NegTokenInit
	if _, err := asn1.Unmarshal(inner, &init); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NegTokenInit: %v", err)
	}
	return &init, nil
}

// unwrapSPNEGO strips the GSS header, checks the SPNEGO OID and returns the
// content of the [0] context tag that follows it.
func unwrapSPNEGO(spnegoToken []byte) ([]byte, error) {
	naked, err := stripGSSHeader(spnegoToken)
	if err != nil {
		return nil, err
	}
	var oid asn1.ObjectIdentifier
	rest, err := asn1.Unmarshal(naked, &oid)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OID: %v", err)
	}
	if !oid.Equal(SpnegoOID) {
		return nil, fmt.Errorf("unexpected SPNEGO OID: %v", oid)
	}
	var raw asn1.RawValue
	if _, err := asn1.Unmarshal(rest, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context tag: %v", err)
	}
	if raw.Class != asn1.ClassContextSpecific || raw.Tag != 0 {
		return nil, errors.New("missing SPNEGO inner context tag")
	}
	return raw.Bytes, nil
}

// stripGSSHeader removes the [APPLICATION 0] tag and returns the inner content
func stripGSSHeader(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != GSS_API_SPNEGO {
		return nil, errors.New("invalid GSS-API header")
	}
	// parse length
	length, lenLen, err := parseLength(data[1:])
	if err != nil {
		return nil, err
	}
	if 1+lenLen+length > len(data) {
		return nil, errors.New("invalid length in GSS header")
	}
	return data[1+lenLen : 1+lenLen+length], nil
}

// parseLength reads DER length from the start of b, returns (length, totalBytesConsumed, error)
func parseLength(b []byte) (int, int, error) {
	if len(b) < 1 {
		return 0, 0, errors.New("length bytes missing")
	}
	if b[0]&0x80 == 0 {
		// short form
		return int(b[0]), 1, nil
	}
	// long form
	n := int(b[0] & 0x7F)
	if n == 0 || n > len(b)-1 {
		return 0, 0, errors.New("invalid long-form length")
	}
	// The length is bounded by the bytes that follow it, which also keeps it
	// from overflowing int.
	var length uint64
	for i := 0; i < n; i++ {
		length = (length << 8) | uint64(b[1+i])
		if length > uint64(len(b)-1-n) {
			return 0, 0, errors.New("long-form length exceeds the data")
		}
	}
	return int(length), 1 + n, nil
}
