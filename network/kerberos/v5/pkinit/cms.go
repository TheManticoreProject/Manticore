package pkinit

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"math/big"
)

// CMS SignedData structures (RFC 5652), hand-built with encoding/asn1 so PKINIT
// needs no external PKCS#7/CMS dependency.

// contentInfo is the CMS ContentInfo wrapper. Content is [0] EXPLICIT ANY.
type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,tag:0"`
}

// encapsulatedContentInfo is EncapsulatedContentInfo. EContent is
// [0] EXPLICIT OCTET STRING.
type encapsulatedContentInfo struct {
	EContentType asn1.ObjectIdentifier
	EContent     []byte `asn1:"explicit,tag:0,optional"`
}

// issuerAndSerialNumber identifies the signer's certificate. Issuer is embedded
// verbatim (RawValue) from the certificate's RawIssuer.
type issuerAndSerialNumber struct {
	Issuer       asn1.RawValue
	SerialNumber *big.Int
}

// attribute is a CMS Attribute (attrType, SET OF attrValues).
type attribute struct {
	Type   asn1.ObjectIdentifier
	Values asn1.RawValue // a SET OF value (built pre-encoded)
}

// signerInfo is CMS SignerInfo (version 1 = issuerAndSerialNumber). SignedAttrs
// is [0] IMPLICIT SET OF Attribute.
type signerInfo struct {
	Version            int
	SID                issuerAndSerialNumber
	DigestAlgorithm    algorithmIdentifier
	SignedAttrs        asn1.RawValue `asn1:"tag:0"`
	SignatureAlgorithm algorithmIdentifier
	Signature          []byte
}

// signedData is CMS SignedData. Certificates is [0] IMPLICIT SET OF Certificate.
type signedData struct {
	Version          int
	DigestAlgorithms asn1.RawValue // SET OF AlgorithmIdentifier (pre-encoded)
	EncapContentInfo encapsulatedContentInfo
	Certificates     asn1.RawValue `asn1:"tag:0,optional"`
	SignerInfos      asn1.RawValue // SET OF SignerInfo (pre-encoded)
}

// BuildSignedAuthPack builds the CMS SignedData ContentInfo that goes into the
// PA-PK-AS-REQ signedAuthPack field: it signs the DER-encoded AuthPack with the
// client RSA private key, embedding certDER as the signer certificate. The
// digest algorithm is SHA-1 and the signature is RSA PKCS#1 v1.5, matching the
// algorithms Windows KDCs accept for PKINIT.
func BuildSignedAuthPack(authPackDER []byte, priv *rsa.PrivateKey, certDER []byte) ([]byte, error) {
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("pkinit: parse signer certificate: %w", err)
	}

	sha1Digest := sha1.Sum(authPackDER)

	// signedAttrs: content-type (= id-pkinit-authData) and message-digest
	// (= SHA1(eContent)). Each attribute's value is a SET OF containing one item.
	ctVal, err := asn1.Marshal(oidIDPKINITAuthData)
	if err != nil {
		return nil, err
	}
	ctSet := marshalSetOf(ctVal)
	mdVal, err := asn1.Marshal(sha1Digest[:])
	if err != nil {
		return nil, err
	}
	mdSet := marshalSetOf(mdVal)

	attrs := []attribute{
		{Type: oidContentType, Values: asn1.RawValue{FullBytes: ctSet}},
		{Type: oidMessageDigest, Values: asn1.RawValue{FullBytes: mdSet}},
	}

	// The signature is computed over the DER of the signedAttrs as an explicit
	// SET OF (tag 0x31), not the [0] IMPLICIT tag they carry in the SignerInfo
	// (RFC 5652 §5.4).
	signedAttrsForSig, err := marshalAttributesSet(attrs)
	if err != nil {
		return nil, err
	}
	attrDigest := sha1.Sum(signedAttrsForSig)
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA1, attrDigest[:])
	if err != nil {
		return nil, fmt.Errorf("pkinit: sign AuthPack: %w", err)
	}

	// Re-tag the signedAttrs as [0] IMPLICIT for the SignerInfo (same content,
	// context tag 0 instead of the universal SET tag).
	signedAttrsImplicit := asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true}
	{
		var tmp asn1.RawValue
		if _, err := asn1.Unmarshal(signedAttrsForSig, &tmp); err != nil {
			return nil, err
		}
		signedAttrsImplicit.Bytes = tmp.Bytes
	}
	signedAttrsImplicitBytes, err := asn1.Marshal(signedAttrsImplicit)
	if err != nil {
		return nil, err
	}

	si := signerInfo{
		Version: 1,
		SID: issuerAndSerialNumber{
			Issuer:       asn1.RawValue{FullBytes: cert.RawIssuer},
			SerialNumber: cert.SerialNumber,
		},
		DigestAlgorithm:    algorithmIdentifier{Algorithm: oidSHA1, Parameters: asn1.RawValue{FullBytes: nullDER}},
		SignedAttrs:        asn1.RawValue{FullBytes: signedAttrsImplicitBytes},
		SignatureAlgorithm: algorithmIdentifier{Algorithm: oidRSAEncryption, Parameters: asn1.RawValue{FullBytes: nullDER}},
		Signature:          signature,
	}
	siBytes, err := asn1.Marshal(si)
	if err != nil {
		return nil, fmt.Errorf("pkinit: marshal SignerInfo: %w", err)
	}

	// digestAlgorithms SET OF { sha1 }.
	digAlg, err := asn1.Marshal(algorithmIdentifier{Algorithm: oidSHA1, Parameters: asn1.RawValue{FullBytes: nullDER}})
	if err != nil {
		return nil, err
	}

	sd := signedData{
		Version:          3,
		DigestAlgorithms: asn1.RawValue{FullBytes: marshalSetOf(digAlg)},
		EncapContentInfo: encapsulatedContentInfo{
			EContentType: oidIDPKINITAuthData,
			EContent:     authPackDER,
		},
		Certificates: asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: cert.Raw},
		SignerInfos:  asn1.RawValue{FullBytes: marshalSetOf(siBytes)},
	}
	sdBytes, err := asn1.Marshal(sd)
	if err != nil {
		return nil, fmt.Errorf("pkinit: marshal SignedData: %w", err)
	}

	// Go's encoding/asn1 ignores the explicit,tag:0 struct tag for an
	// asn1.RawValue on marshal, so build the [0] EXPLICIT wrapper by hand.
	ci := contentInfo{
		ContentType: oidSignedData,
		Content:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: sdBytes},
	}
	ciBytes, err := asn1.Marshal(ci)
	if err != nil {
		return nil, fmt.Errorf("pkinit: marshal ContentInfo: %w", err)
	}
	return ciBytes, nil
}

// extractSignedDataEContent parses a CMS ContentInfo (SignedData) and returns
// the encapsulated eContent (for a PKINIT reply, the DER of KDCDHKeyInfo). The
// KDC's signature is not verified here; successful decryption of the AS-REP
// enc-part with the DH-derived reply key authenticates the exchange.
func extractSignedDataEContent(contentInfoDER []byte) ([]byte, error) {
	var ci contentInfo
	if _, err := asn1.Unmarshal(contentInfoDER, &ci); err != nil {
		return nil, fmt.Errorf("pkinit: parse reply ContentInfo: %w", err)
	}
	if !ci.ContentType.Equal(oidSignedData) {
		return nil, fmt.Errorf("pkinit: reply is not CMS SignedData (OID %v)", ci.ContentType)
	}
	var sd signedData
	if _, err := asn1.Unmarshal(ci.Content.Bytes, &sd); err != nil {
		return nil, fmt.Errorf("pkinit: parse reply SignedData: %w", err)
	}
	if !sd.EncapContentInfo.EContentType.Equal(oidIDPKINITDHKeyData) {
		return nil, fmt.Errorf("pkinit: unexpected reply eContentType %v", sd.EncapContentInfo.EContentType)
	}
	if len(sd.EncapContentInfo.EContent) == 0 {
		return nil, fmt.Errorf("pkinit: reply SignedData has no eContent")
	}
	return sd.EncapContentInfo.EContent, nil
}

// nullDER is the DER encoding of ASN.1 NULL, used for RSA/SHA-1 algorithm
// parameters.
var nullDER = []byte{0x05, 0x00}

// marshalSetOf wraps one pre-encoded DER element in a universal SET OF (tag 0x31).
func marshalSetOf(elem []byte) []byte {
	out, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: elem})
	if err != nil {
		panic("pkinit: marshal SET OF: " + err.Error())
	}
	return out
}

// marshalAttributesSet encodes attrs as a universal SET OF Attribute (tag 0x31),
// the form over which the CMS signature is computed.
func marshalAttributesSet(attrs []attribute) ([]byte, error) {
	var content []byte
	for _, a := range attrs {
		b, err := asn1.Marshal(a)
		if err != nil {
			return nil, err
		}
		content = append(content, b...)
	}
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: content})
}
