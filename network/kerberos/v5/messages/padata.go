package messages

import (
	"encoding/asn1"
	"time"
)

// PAEncTSEnc is the plaintext body of a PA-ENC-TIMESTAMP pre-authentication element,
// as defined in RFC 4120 Section 5.2.7.2.
// It is encrypted with the client's key and used to prove knowledge of the password.
type PAEncTSEnc struct {
	// PATimestamp is the client's current time.
	PATimestamp time.Time `asn1:"explicit,tag:0,generalized"`
	// PAUSec is the optional microseconds component of PATimestamp.
	PAUSec int `asn1:"explicit,tag:1,optional"`
}

// Marshal encodes PAEncTSEnc as a plain ASN.1 SEQUENCE (no APPLICATION wrapper).
func (p *PAEncTSEnc) Marshal() ([]byte, error) {
	return asn1.Marshal(*p)
}

// Unmarshal decodes PAEncTSEnc from a plain ASN.1 SEQUENCE.
// Returns the number of bytes consumed from data.
func (p *PAEncTSEnc) Unmarshal(data []byte) (int, error) {
	rest, err := asn1.Unmarshal(data, p)
	if err != nil {
		return 0, err
	}
	return len(data) - len(rest), nil
}

// ETypeInfo2Entry is a single entry in a PA-ETYPE-INFO2 pre-authentication element,
// as defined in RFC 4120 Section 5.2.7.5.
// It specifies an encryption type and optional salt/parameters for string-to-key derivation.
type ETypeInfo2Entry struct {
	// EType identifies the encryption type.
	EType int `asn1:"explicit,tag:0"`
	// Salt is the optional salt string for string-to-key derivation.
	Salt string `asn1:"explicit,tag:1,optional,utf8"`
	// S2KParams contains optional string-to-key parameters (e.g. iteration count).
	S2KParams []byte `asn1:"explicit,tag:2,optional"`
}

// ETypeInfo2 is a sequence of ETypeInfo2Entry values returned in PA-ETYPE-INFO2.
// The KDC uses this to tell the client which encryption types and salts to use.
type ETypeInfo2 []ETypeInfo2Entry

// Marshal encodes ETypeInfo2 as an ASN.1 SEQUENCE OF.
func (e ETypeInfo2) Marshal() ([]byte, error) {
	return asn1.Marshal([]ETypeInfo2Entry(e))
}

// Unmarshal decodes ETypeInfo2 from an ASN.1 SEQUENCE OF.
// Returns the number of bytes consumed from data.
func (e *ETypeInfo2) Unmarshal(data []byte) (int, error) {
	var entries []ETypeInfo2Entry
	rest, err := asn1.Unmarshal(data, &entries)
	if err != nil {
		return 0, err
	}
	*e = ETypeInfo2(entries)
	return len(data) - len(rest), nil
}
