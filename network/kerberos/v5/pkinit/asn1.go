package pkinit

import (
	"encoding/asn1"
	"fmt"
	"math/big"
	"time"
)

// PKINIT object identifiers (RFC 4556 §3.2.4 and RFC 3279).
var (
	// oidIDPKINITAuthData is id-pkinit-authData (the eContentType of the client
	// AuthPack SignedData).
	oidIDPKINITAuthData = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 1}
	// oidIDPKINITDHKeyData is id-pkinit-DHKeyData (the eContentType of the KDC's
	// KDCDHKeyInfo SignedData in the reply).
	oidIDPKINITDHKeyData = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 2}
	// oidDHPublicNumber is dhpublicnumber, the algorithm OID of the DH public
	// value carried in clientPublicValue (RFC 3279 §2.3.3).
	oidDHPublicNumber = asn1.ObjectIdentifier{1, 2, 840, 10046, 2, 1}
	// oidSignedData is the CMS id-signedData content type (RFC 5652).
	oidSignedData = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	// oidContentType and oidMessageDigest are the CMS signed attributes.
	oidContentType   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	// oidSHA1 is the id-sha1 digest algorithm (RFC 3370).
	oidSHA1 = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
	// oidSHA256 is the id-sha256 digest algorithm (RFC 5754 / NIST).
	oidSHA256 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	// oidRSAEncryption is the rsaEncryption signature/key algorithm.
	oidRSAEncryption = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	// oidSHA1WithRSA and oidSHA256WithRSA are the combined RSA signature
	// algorithms a KDC may name in the SignerInfo signatureAlgorithm field.
	oidSHA1WithRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}
	oidSHA256WithRSA = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}

	// RFC 8636 PKINIT algorithm-agility KDF identifiers, under
	// id-pkinit-kdf ::= { id-pkinit kdf(6) } = 1.3.6.1.5.2.3.6. These name the
	// SP800-56A concatenation KDFs a client advertises in AuthPack.supportedKDFs
	// and a KDC selects in the PA-PK-AS-REP DHRepInfo.kdf field. Note the final
	// arc ordering from the RFC: sha512 is 3 and sha384 is 4.
	oidPKINITKDFSHA1   = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 6, 1} // id-pkinit-kdf-ah-sha1
	oidPKINITKDFSHA256 = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 6, 2} // id-pkinit-kdf-ah-sha256
	oidPKINITKDFSHA512 = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 6, 3} // id-pkinit-kdf-ah-sha512
	oidPKINITKDFSHA384 = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 6, 4} // id-pkinit-kdf-ah-sha384
)

// supportedKDFOIDs is the ordered set of RFC 8636 KDFs this client advertises in
// AuthPack.supportedKDFs, strongest first. Advertising the SHA-2 KDFs lets a
// conforming KDC derive an AES-SHA2 (etype 19/20) reply key; the SHA-1 agility
// KDF is offered for completeness. A KDC that does not implement RFC 8636 simply
// omits the DHRepInfo.kdf field and the client falls back to the RFC 4556 SHA-1
// OctetString2Key (see DeriveReplyKey).
var supportedKDFOIDs = []asn1.ObjectIdentifier{
	oidPKINITKDFSHA384,
	oidPKINITKDFSHA256,
	oidPKINITKDFSHA1,
}

// pkAuthenticator is RFC 4556 PKAuthenticator.
type pkAuthenticator struct {
	CUSec      int       `asn1:"explicit,tag:0"`
	CTime      time.Time `asn1:"explicit,tag:1,generalized"`
	Nonce      int       `asn1:"explicit,tag:2"`
	PAChecksum []byte    `asn1:"explicit,tag:3,optional"`
}

// algorithmIdentifier is the X.509 AlgorithmIdentifier with an opaque parameters
// field (DomainParameters for DH).
type algorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

// subjectPublicKeyInfo is the X.509 SubjectPublicKeyInfo carrying the DH public
// value (clientPublicValue in the AuthPack).
type subjectPublicKeyInfo struct {
	Algorithm        algorithmIdentifier
	SubjectPublicKey asn1.BitString
}

// authPack is RFC 4556 AuthPack extended per RFC 8636 with supportedKDFs [4].
// supportedCMSTypes [2] is omitted. SupportedKDFs is a SEQUENCE OF KDFAlgorithmId
// advertising the algorithm-agility KDFs the client accepts; it is optional and
// omitted when empty (leaving a plain RFC 4556 AuthPack).
type authPack struct {
	PKAuthenticator   pkAuthenticator      `asn1:"explicit,tag:0"`
	ClientPublicValue subjectPublicKeyInfo `asn1:"explicit,tag:1,optional"`
	ClientDHNonce     []byte               `asn1:"explicit,tag:3,optional"`
	SupportedKDFs     []kdfAlgorithmID     `asn1:"explicit,tag:4,optional,omitempty"`
}

// kdfAlgorithmID is RFC 8636 KDFAlgorithmId:
//
//	KDFAlgorithmId ::= SEQUENCE { kdf-id [0] OBJECT IDENTIFIER, ... }
//
// It appears both in the request (AuthPack.supportedKDFs) and the reply
// (DHRepInfo.kdf). The tag [0] is EXPLICIT, matching the reference (MIT krb5)
// encoding.
type kdfAlgorithmID struct {
	KDFID asn1.ObjectIdentifier `asn1:"explicit,tag:0"`
}

// kdfAlgorithmIdentifier is the X.509 AlgorithmIdentifier used as the OtherInfo
// algorithmID: the KDF OID with parameters ABSENT (RFC 8636 Errata 8639). It is a
// dedicated type — rather than reusing algorithmIdentifier — so the optional
// parameters field is never emitted.
type kdfAlgorithmIdentifier struct {
	Algorithm asn1.ObjectIdentifier
}

// kdfOtherInfo is the SP800-56A OtherInfo structure keyed into the RFC 8636
// agility KDF (RFC 8636 §3):
//
//	OtherInfo ::= SEQUENCE {
//	    algorithmID   AlgorithmIdentifier,
//	    partyUInfo    [0] OCTET STRING,
//	    partyVInfo    [1] OCTET STRING,
//	    suppPubInfo   [2] OCTET STRING OPTIONAL,
//	    suppPrivInfo  [3] OCTET STRING OPTIONAL }
//
// partyUInfo/partyVInfo carry the DER of a KRB5PrincipalName (client / TGS) and
// suppPubInfo carries the DER of a PkinitSuppPubInfo. All context tags are
// EXPLICIT and the OCTET STRING wrappers are present, matching the reference
// encoding; suppPrivInfo is never used.
type kdfOtherInfo struct {
	AlgorithmID kdfAlgorithmIdentifier
	PartyUInfo  []byte `asn1:"explicit,tag:0"`
	PartyVInfo  []byte `asn1:"explicit,tag:1"`
	SuppPubInfo []byte `asn1:"explicit,tag:2,optional"`
}

// pkinitSuppPubInfo is RFC 8636 PkinitSuppPubInfo, the suppPubInfo payload of
// OtherInfo:
//
//	PkinitSuppPubInfo ::= SEQUENCE {
//	    enctype    [0] Int32,
//	    as-REQ     [1] OCTET STRING,
//	    pk-as-rep  [2] OCTET STRING, ... }
//
// enctype is the AS reply-key enctype; as-REQ is the DER of the AS-REQ sent to
// the KDC (without any TCP length prefix); pk-as-rep is the DER of the
// PA-PK-AS-REP value from the reply. This binds the derived key to the exact
// request/reply pair.
type pkinitSuppPubInfo struct {
	EType   int    `asn1:"explicit,tag:0"`
	ASReq   []byte `asn1:"explicit,tag:1"`
	PKASRep []byte `asn1:"explicit,tag:2"`
}

// krb5PrincipalNameMarshal is the RFC 4556 KRB5PrincipalName used as party info
// in the KDF OtherInfo:
//
//	KRB5PrincipalName ::= SEQUENCE { realm [0] Realm, principalName [1] PrincipalName }
//
// Realm is pre-encoded as a [0] EXPLICIT { GeneralString } RawValue because Go's
// encoding/asn1 ignores struct tags on asn1.RawValue and would otherwise emit a
// PrintableString (Kerberos requires GeneralString, RFC 4120).
type krb5PrincipalNameMarshal struct {
	Realm         asn1.RawValue        // pre-encoded [0] EXPLICIT { GeneralString }
	PrincipalName principalNameMarshal `asn1:"explicit,tag:1"`
}

// principalNameMarshal is RFC 4120 PrincipalName with GeneralString components:
//
//	PrincipalName ::= SEQUENCE { name-type [0] Int32, name-string [1] SEQUENCE OF KerberosString }
type principalNameMarshal struct {
	NameType   int             `asn1:"explicit,tag:0"`
	NameString []asn1.RawValue `asn1:"explicit,tag:1"` // SEQUENCE OF GeneralString
}

// PrincipalName identifies a Kerberos principal (name-type and components) for
// the RFC 8636 KDF party info, without the pkinit package depending on the
// messages package.
type PrincipalName struct {
	NameType   int
	NameString []string
}

// generalString encodes s as an ASN.1 GeneralString (tag 27), the string type
// Kerberos requires for realm and principal-name components (RFC 4120). Go's
// encoding/asn1 emits PrintableString for the "generalstring" struct tag, so the
// RawValue is built by hand.
func generalString(s string) asn1.RawValue {
	return asn1.RawValue{Class: asn1.ClassUniversal, Tag: 27, Bytes: []byte(s)}
}

// buildKRB5PrincipalName returns the DER of a KRB5PrincipalName (RFC 4556) for
// the given realm and principal, with GeneralString realm and name components.
func buildKRB5PrincipalName(realm string, name PrincipalName) ([]byte, error) {
	realmGS, err := asn1.Marshal(generalString(realm))
	if err != nil {
		return nil, fmt.Errorf("pkinit: marshal KDF realm: %w", err)
	}
	comps := make([]asn1.RawValue, len(name.NameString))
	for i, s := range name.NameString {
		comps[i] = generalString(s)
	}
	m := krb5PrincipalNameMarshal{
		Realm:         asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: realmGS},
		PrincipalName: principalNameMarshal{NameType: name.NameType, NameString: comps},
	}
	return asn1.Marshal(m)
}

// domainParameters is the DH DomainParameters (RFC 3279 §2.3.3) carried in the
// AlgorithmIdentifier parameters of clientPublicValue.
type domainParameters struct {
	P *big.Int
	G *big.Int
	Q *big.Int
}

// paPKASReq is RFC 4556 PA-PK-AS-REQ. signedAuthPack is the DER of a CMS
// ContentInfo (SignedData) over the AuthPack.
type paPKASReq struct {
	SignedAuthPack []byte `asn1:"implicit,tag:0"`
}

// kdcDHKeyInfo is RFC 4556 KDCDHKeyInfo (the eContent of the KDC reply's
// SignedData). subjectPublicKey is a BIT STRING whose content is the DER
// encoding of an INTEGER (the KDC's DH public value).
type kdcDHKeyInfo struct {
	SubjectPublicKey asn1.BitString `asn1:"explicit,tag:0"`
	Nonce            int            `asn1:"explicit,tag:1"`
	DHKeyExpiration  time.Time      `asn1:"explicit,tag:2,optional,generalized"`
}

// buildClientPublicValue encodes the client's ephemeral DH public value as a
// SubjectPublicKeyInfo: algorithm dhpublicnumber with the group's DomainParameters,
// and subjectPublicKey a BIT STRING wrapping DER(INTEGER Y) per RFC 3279 §2.3.3.
func buildClientPublicValue(kp *DHKeyPair) (subjectPublicKeyInfo, error) {
	dp := domainParameters{P: kp.Group.P, G: kp.Group.G, Q: kp.Group.Q}
	dpBytes, err := asn1.Marshal(dp)
	if err != nil {
		return subjectPublicKeyInfo{}, fmt.Errorf("marshal DomainParameters: %w", err)
	}
	// The DH public value is ASN.1-encoded as an INTEGER; that encoding becomes
	// the content of the subjectPublicKey BIT STRING (RFC 3279 §2.3.3).
	yBytes, err := asn1.Marshal(kp.Y)
	if err != nil {
		return subjectPublicKeyInfo{}, fmt.Errorf("marshal DH public INTEGER: %w", err)
	}
	return subjectPublicKeyInfo{
		Algorithm: algorithmIdentifier{
			Algorithm:  oidDHPublicNumber,
			Parameters: asn1.RawValue{FullBytes: dpBytes},
		},
		SubjectPublicKey: asn1.BitString{Bytes: yBytes, BitLength: len(yBytes) * 8},
	}, nil
}

// parseKDCPublicValue extracts the KDC's DH public value from a KDCDHKeyInfo's
// subjectPublicKey BIT STRING (which wraps a DER INTEGER, RFC 3279 §2.3.3).
func parseKDCPublicValue(info kdcDHKeyInfo) (*big.Int, error) {
	var y *big.Int
	if _, err := asn1.Unmarshal(info.SubjectPublicKey.Bytes, &y); err != nil {
		return nil, fmt.Errorf("parse KDC DH public INTEGER: %w", err)
	}
	return y, nil
}
