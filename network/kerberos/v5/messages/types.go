package messages

import (
	"encoding/asn1"
	"time"
)

// KerberosTime represents a Kerberos timestamp (GeneralizedTime without fractional seconds).
// It is stored as a standard Go time.Time value.
type KerberosTime = time.Time

// PrincipalName contains a name-type and a sequence of name strings,
// as defined in RFC 4120 Section 5.2.2.
type PrincipalName struct {
	// NameType specifies the type of name (e.g. NT-PRINCIPAL = 1).
	NameType int `asn1:"explicit,tag:0"`
	// NameString contains the sequence of name components.
	NameString []string `asn1:"explicit,tag:1"`
}

// EncryptedData holds a Kerberos encrypted blob, as defined in RFC 4120 Section 5.2.9.
// The actual encryption algorithm and key are identified by EType.
type EncryptedData struct {
	// EType identifies the encryption algorithm used.
	EType int `asn1:"explicit,tag:0"`
	// KvNo is the optional key version number.
	KvNo int `asn1:"explicit,tag:1,optional"`
	// Cipher contains the encrypted bytes.
	Cipher []byte `asn1:"explicit,tag:2"`
}

// HostAddress represents a network address, as defined in RFC 4120 Section 5.2.5.
type HostAddress struct {
	// AddrType identifies the address type (e.g. 2 = IPv4, 24 = IPv6).
	AddrType int `asn1:"explicit,tag:0"`
	// Address contains the raw address bytes.
	Address []byte `asn1:"explicit,tag:1"`
}

// AuthorizationData is an authorization-data element, as defined in RFC 4120 Section 5.2.6.
type AuthorizationData struct {
	// ADType identifies the authorization-data type.
	ADType int `asn1:"explicit,tag:0"`
	// ADData contains the type-specific authorization data.
	ADData []byte `asn1:"explicit,tag:1"`
}

// PAData is a pre-authentication data element, as defined in RFC 4120 Section 5.2.7.
type PAData struct {
	// PADataType identifies the pre-authentication data type.
	PADataType int `asn1:"explicit,tag:1"`
	// PADataValue contains the pre-authentication data bytes.
	PADataValue []byte `asn1:"explicit,tag:2"`
}

// KDCOptions is a bit string encoding KDC request options flags,
// as defined in RFC 4120 Section 5.4.1.
type KDCOptions = asn1.BitString

// generalStringRaw encodes a string as an ASN.1 GeneralString (tag 27 = 0x1B).
// Go's encoding/asn1 ignores the "generalstring" struct tag and always produces
// PrintableString for ASCII-safe content. Kerberos requires GeneralString for
// realm names and principal name components (RFC 4120).
func generalStringRaw(s string) asn1.RawValue {
	return asn1.RawValue{Class: asn1.ClassUniversal, Tag: 27, Bytes: []byte(s)}
}

// realmExplicit encodes s as an ASN.1 [tag] EXPLICIT { GeneralString } context element.
// Go's asn1.Marshal ignores explicit,tag:N struct tags for asn1.RawValue fields, so we
// pre-encode the context wrapper (a0|tag constructed) directly in the returned RawValue.
func realmExplicit(tag int, s string) asn1.RawValue {
	gs := generalStringRaw(s)
	gsBytes, err := asn1.Marshal(gs)
	if err != nil {
		panic("messages: failed to marshal realm GeneralString: " + err.Error())
	}
	return asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: tag, IsCompound: true, Bytes: gsBytes}
}

// PrincipalNameMarshal is the wire representation of PrincipalName for marshaling.
// It uses []asn1.RawValue (GeneralString) instead of []string, which Go's asn1
// would incorrectly encode as PrintableString.
type PrincipalNameMarshal struct {
	NameType   int             `asn1:"explicit,tag:0"`
	NameString []asn1.RawValue `asn1:"explicit,tag:1"`
}

// MarshalPrincipalName converts a PrincipalName to its GeneralString-encoded form.
func MarshalPrincipalName(pn PrincipalName) PrincipalNameMarshal {
	strs := make([]asn1.RawValue, len(pn.NameString))
	for i, s := range pn.NameString {
		strs[i] = generalStringRaw(s)
	}
	return PrincipalNameMarshal{NameType: pn.NameType, NameString: strs}
}

// wrapApplication wraps seqBytes (a full DER-encoded SEQUENCE) in an ASN.1 APPLICATION tag,
// producing APPLICATION[tag] { SEQUENCE { ... } } as required by RFC 4120.
func wrapApplication(tag int, seqBytes []byte) ([]byte, error) {
	return asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassApplication,
		Tag:        tag,
		IsCompound: true,
		Bytes:      seqBytes,
	})
}

// unwrapApplication unwraps an ASN.1 APPLICATION tag from data and verifies the tag.
// Returns the inner bytes and the number of bytes consumed from data.
func unwrapApplication(data []byte, expected_tag int) (inner []byte, consumed int, err error) {
	var raw asn1.RawValue
	rest, err := asn1.Unmarshal(data, &raw)
	if err != nil {
		return nil, 0, err
	}
	if raw.Class != asn1.ClassApplication || raw.Tag != expected_tag {
		return nil, 0, asn1.StructuralError{Msg: "wrong APPLICATION tag"}
	}
	return raw.Bytes, len(data) - len(rest), nil
}
