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
)

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

// authPack is RFC 4556 AuthPack. supportedCMSTypes [2] is omitted.
type authPack struct {
	PKAuthenticator   pkAuthenticator      `asn1:"explicit,tag:0"`
	ClientPublicValue subjectPublicKeyInfo `asn1:"explicit,tag:1,optional"`
	ClientDHNonce     []byte               `asn1:"explicit,tag:3,optional"`
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
