package pdu

import "fmt"

// PDU is a complete connectionless DCE/RPC protocol data unit: the common Header
// followed by an opaque body of Header.BodyLength octets. It is the generic framing
// layer that round-trips every connectionless PDU type; the structured contents of
// the bodies that have a defined layout are parsed by the typed body codecs
// (FackBody, CancelBody, CancelAckBody) and the fault/reject status helpers.
//
// The ping, ack, and working PDUs carry no body, so Body is nil for them. The
// request and response bodies are NDR-marshalled stub data (opaque to this layer).
//
// Authentication verifiers are not modelled: a non-zero auth_proto means an
// auth_verifier trailer follows the body, which this codec does not interpret.
type PDU struct {
	Header Header
	Body   []byte
}

// Marshal serializes the PDU into its wire form. It sets Header.BodyLength from the
// body length before encoding the header, so callers do not need to maintain it.
func (p *PDU) Marshal() ([]byte, error) {
	if len(p.Body) > 0xFFFF {
		return nil, fmt.Errorf("connectionless PDU body too large: %d bytes, max %d", len(p.Body), 0xFFFF)
	}
	p.Header.BodyLength = uint16(len(p.Body))
	hdr, err := p.Header.Marshal()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(hdr)+len(p.Body))
	out = append(out, hdr...)
	out = append(out, p.Body...)
	return out, nil
}

// Unmarshal parses a complete PDU (header plus the Header.BodyLength body octets that
// follow) from data and returns the number of bytes consumed. Any trailing bytes
// beyond the body (for example an authentication verifier when auth_proto is
// non-zero) are left for the caller.
func (p *PDU) Unmarshal(data []byte) (int, error) {
	n, err := p.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	bodyLen := int(p.Header.BodyLength)
	if len(data) < n+bodyLen {
		return 0, fmt.Errorf("connectionless PDU body truncated: have %d body bytes, header declares %d", len(data)-n, bodyLen)
	}
	if bodyLen == 0 {
		p.Body = nil
	} else {
		p.Body = append([]byte(nil), data[n:n+bodyLen]...)
	}
	return n + bodyLen, nil
}

// String returns a human-readable one-line summary of the PDU.
func (p *PDU) String() string {
	return fmt.Sprintf("%s body=%d", p.Header.String(), len(p.Body))
}
