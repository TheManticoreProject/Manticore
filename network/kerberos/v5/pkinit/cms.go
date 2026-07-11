package pkinit

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
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

// VerifyOptions controls verification of the KDC's CMS SignedData signature on
// the PKINIT reply (RFC 4556 §3.2.4). Verification is performed whenever a trust
// anchor is present and InsecureSkipSignatureCheck is not set: the KDC's
// signature over the KDCDHKeyInfo is checked and the signer certificate must
// chain to (or be pinned by) one of Anchors. A nil *VerifyOptions means no
// verification (the legacy behaviour, kept for callers that cannot supply an
// anchor).
type VerifyOptions struct {
	// Anchors are the trusted certificates the KDC signer certificate must
	// chain to — the issuing CA (or root) certificate — or, when pinned
	// directly, the KDC signing certificate itself (which also covers a
	// self-signed KDC certificate). At least one is required unless
	// InsecureSkipSignatureCheck is set.
	Anchors []*x509.Certificate
	// InsecureSkipSignatureCheck disables signature and chain verification for
	// the anonymous / self-signed lab case where no anchor can be pinned. It is
	// insecure: it removes the RFC 4556 §3.2.4 protection against a substituted
	// KDC DH public value, so set it only knowingly.
	InsecureSkipSignatureCheck bool
}

// verifySignedDataEContent parses a CMS ContentInfo (SignedData) from a PKINIT
// reply and returns the encapsulated eContent (the DER of KDCDHKeyInfo). When
// opts enables verification, the KDC's signature over the eContent is checked
// and the signer certificate is chained/pinned to a trust anchor per RFC 4556
// §3.2.4 before the eContent is returned; otherwise the eContent is returned
// without verifying the signature (successful AS-REP decryption still gates the
// exchange).
func verifySignedDataEContent(contentInfoDER []byte, opts *VerifyOptions) ([]byte, error) {
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
		return nil, fmt.Errorf("pkinit: unexpected reply eContentType %v (want id-pkinit-DHKeyData)", sd.EncapContentInfo.EContentType)
	}
	if len(sd.EncapContentInfo.EContent) == 0 {
		return nil, fmt.Errorf("pkinit: reply SignedData has no eContent")
	}

	if opts != nil && !opts.InsecureSkipSignatureCheck {
		if len(opts.Anchors) == 0 {
			return nil, fmt.Errorf("pkinit: KDC signature verification enabled but no trust anchor supplied; provide one or set InsecureSkipSignatureCheck")
		}
		if err := verifyKDCSignedData(&sd, opts); err != nil {
			return nil, err
		}
	}
	return sd.EncapContentInfo.EContent, nil
}

// verifyKDCSignedData verifies the KDC's CMS signature over a PKINIT reply
// SignedData (RFC 4556 §3.2.4, RFC 5652 §5.4): it locates the signer
// certificate, confirms the signed content-type and messageDigest attributes,
// verifies the RSA signature over the DER re-encoding of signedAttrs, and
// confirms the signer certificate is pinned to or chains to a trust anchor.
func verifyKDCSignedData(sd *signedData, opts *VerifyOptions) error {
	certs, err := parseSignedDataCertificates(sd.Certificates)
	if err != nil {
		return err
	}
	si, err := parseSingleSignerInfo(sd.SignerInfos)
	if err != nil {
		return err
	}
	signer := findSignerCert(certs, si.SID)
	if signer == nil {
		return fmt.Errorf("pkinit: KDC SignedData does not carry the signer certificate")
	}

	digestHash, err := digestForOID(si.DigestAlgorithm.Algorithm)
	if err != nil {
		return err
	}
	sigHash, err := rsaHashForSigAlg(si.SignatureAlgorithm.Algorithm, digestHash)
	if err != nil {
		return err
	}

	// RFC 4556 §3.2.4 requires signed attributes on the reply. Confirm the
	// content-type attribute names id-pkinit-DHKeyData and that the messageDigest
	// attribute equals the digest of the eContent before trusting the signature
	// (RFC 5652 §5.4).
	attrs, err := parseSignedAttrs(si.SignedAttrs)
	if err != nil {
		return err
	}
	ct, err := signedAttrOID(attrs, oidContentType)
	if err != nil {
		return err
	}
	if !ct.Equal(oidIDPKINITDHKeyData) {
		return fmt.Errorf("pkinit: signed content-type attribute %v != id-pkinit-DHKeyData", ct)
	}
	md, err := signedAttrOctets(attrs, oidMessageDigest)
	if err != nil {
		return err
	}
	if !bytes.Equal(md, hashBytes(digestHash, sd.EncapContentInfo.EContent)) {
		return fmt.Errorf("pkinit: KDC messageDigest attribute does not match the eContent digest")
	}

	// The signature covers the DER of signedAttrs re-tagged as an EXPLICIT SET OF
	// (0x31), not the [0] IMPLICIT tag they carry in the SignerInfo (RFC 5652
	// §5.4) — the classic CMS re-encoding gotcha.
	signedAttrsForSig, err := reencodeSignedAttrsSet(si.SignedAttrs)
	if err != nil {
		return err
	}
	pub, ok := signer.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("pkinit: KDC signer key is %T, only RSA is supported", signer.PublicKey)
	}
	if err := rsa.VerifyPKCS1v15(pub, sigHash, hashBytes(sigHash, signedAttrsForSig), si.Signature); err != nil {
		return fmt.Errorf("pkinit: KDC SignedData signature verification failed: %w", err)
	}

	// Anchor check: accept immediately if the signer certificate is itself
	// pinned (byte-identical to a supplied anchor, which also covers a
	// self-signed KDC certificate); otherwise require an X.509 chain to one of
	// the anchors, using any other certificates in the SignedData as
	// intermediates.
	for _, a := range opts.Anchors {
		if bytes.Equal(a.Raw, signer.Raw) {
			return nil
		}
	}
	roots := x509.NewCertPool()
	for _, a := range opts.Anchors {
		roots.AddCert(a)
	}
	inter := x509.NewCertPool()
	for _, c := range certs {
		if !bytes.Equal(c.Raw, signer.Raw) {
			inter.AddCert(c)
		}
	}
	if _, err := signer.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: inter,
		// The KDC certificate carries the id-pkinit-KPKdc EKU, not the default
		// serverAuth; do not reject on EKU here.
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return fmt.Errorf("pkinit: KDC certificate does not chain to a trusted anchor: %w", err)
	}
	return nil
}

// parseSignedDataCertificates parses the X.509 certificates carried in a
// SignedData certificates [0] field (a SET OF CertificateChoices); non-X.509
// choices are skipped.
func parseSignedDataCertificates(raw asn1.RawValue) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	b := raw.Bytes
	for len(b) > 0 {
		var rv asn1.RawValue
		rest, err := asn1.Unmarshal(b, &rv)
		if err != nil {
			return nil, fmt.Errorf("pkinit: walk SignedData certificates: %w", err)
		}
		b = rest
		if rv.Class == asn1.ClassUniversal && rv.Tag == asn1.TagSequence {
			cert, err := x509.ParseCertificate(rv.FullBytes)
			if err != nil {
				return nil, fmt.Errorf("pkinit: parse SignedData certificate: %w", err)
			}
			certs = append(certs, cert)
		}
	}
	return certs, nil
}

// parseSingleSignerInfo parses the first SignerInfo from the SET OF SignerInfo
// (a PKINIT reply carries exactly one).
func parseSingleSignerInfo(raw asn1.RawValue) (signerInfo, error) {
	var si signerInfo
	if len(raw.Bytes) == 0 {
		return si, fmt.Errorf("pkinit: KDC SignedData has no SignerInfo")
	}
	if _, err := asn1.Unmarshal(raw.Bytes, &si); err != nil {
		return si, fmt.Errorf("pkinit: parse SignerInfo: %w", err)
	}
	return si, nil
}

// findSignerCert returns the certificate matching the SignerInfo's
// issuerAndSerialNumber, or nil.
func findSignerCert(certs []*x509.Certificate, sid issuerAndSerialNumber) *x509.Certificate {
	for _, c := range certs {
		if c.SerialNumber != nil && sid.SerialNumber != nil &&
			c.SerialNumber.Cmp(sid.SerialNumber) == 0 &&
			bytes.Equal(c.RawIssuer, sid.Issuer.FullBytes) {
			return c
		}
	}
	return nil
}

// parseSignedAttrs decodes the individual attributes carried in the SignerInfo
// signedAttrs field.
func parseSignedAttrs(raw asn1.RawValue) ([]attribute, error) {
	if len(raw.Bytes) == 0 {
		return nil, fmt.Errorf("pkinit: KDC SignerInfo carries no signed attributes")
	}
	var attrs []attribute
	b := raw.Bytes
	for len(b) > 0 {
		var a attribute
		rest, err := asn1.Unmarshal(b, &a)
		if err != nil {
			return nil, fmt.Errorf("pkinit: parse signed attribute: %w", err)
		}
		b = rest
		attrs = append(attrs, a)
	}
	return attrs, nil
}

// findSignedAttr returns the attribute with the given type OID, or an error.
func findSignedAttr(attrs []attribute, oid asn1.ObjectIdentifier) (*attribute, error) {
	for i := range attrs {
		if attrs[i].Type.Equal(oid) {
			return &attrs[i], nil
		}
	}
	return nil, fmt.Errorf("pkinit: signed attribute %v absent", oid)
}

// signedAttrOID decodes the single OID value of a signed attribute (e.g.
// content-type).
func signedAttrOID(attrs []attribute, oid asn1.ObjectIdentifier) (asn1.ObjectIdentifier, error) {
	a, err := findSignedAttr(attrs, oid)
	if err != nil {
		return nil, err
	}
	var v asn1.ObjectIdentifier
	if _, err := asn1.Unmarshal(a.Values.Bytes, &v); err != nil {
		return nil, fmt.Errorf("pkinit: parse OID signed attribute %v: %w", oid, err)
	}
	return v, nil
}

// signedAttrOctets decodes the single OCTET STRING value of a signed attribute
// (e.g. message-digest).
func signedAttrOctets(attrs []attribute, oid asn1.ObjectIdentifier) ([]byte, error) {
	a, err := findSignedAttr(attrs, oid)
	if err != nil {
		return nil, err
	}
	var v []byte
	if _, err := asn1.Unmarshal(a.Values.Bytes, &v); err != nil {
		return nil, fmt.Errorf("pkinit: parse octet-string signed attribute %v: %w", oid, err)
	}
	return v, nil
}

// reencodeSignedAttrsSet re-tags the [0] IMPLICIT signedAttrs as a universal
// SET OF (tag 0x31), the form over which the CMS signature is computed (RFC 5652
// §5.4).
func reencodeSignedAttrsSet(raw asn1.RawValue) ([]byte, error) {
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: raw.Bytes})
}

// digestForOID maps a CMS digestAlgorithm OID to a crypto.Hash.
func digestForOID(oid asn1.ObjectIdentifier) (crypto.Hash, error) {
	switch {
	case oid.Equal(oidSHA1):
		return crypto.SHA1, nil
	case oid.Equal(oidSHA256):
		return crypto.SHA256, nil
	}
	return 0, fmt.Errorf("pkinit: unsupported CMS digest algorithm %v", oid)
}

// rsaHashForSigAlg maps a SignerInfo signatureAlgorithm OID to the hash used
// with RSA PKCS#1 v1.5. For rsaEncryption the hash is the one named by the
// digestAlgorithm (RFC 5652); the combined sha*WithRSAEncryption OIDs name it
// explicitly.
func rsaHashForSigAlg(oid asn1.ObjectIdentifier, digestHash crypto.Hash) (crypto.Hash, error) {
	switch {
	case oid.Equal(oidRSAEncryption):
		return digestHash, nil
	case oid.Equal(oidSHA1WithRSA):
		return crypto.SHA1, nil
	case oid.Equal(oidSHA256WithRSA):
		return crypto.SHA256, nil
	}
	return 0, fmt.Errorf("pkinit: unsupported CMS signature algorithm %v", oid)
}

// hashBytes returns the digest of data under the given hash.
func hashBytes(h crypto.Hash, data []byte) []byte {
	switch h {
	case crypto.SHA256:
		sum := sha256.Sum256(data)
		return sum[:]
	default: // crypto.SHA1
		sum := sha1.Sum(data)
		return sum[:]
	}
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
