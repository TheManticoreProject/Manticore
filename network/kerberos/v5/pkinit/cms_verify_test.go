package pkinit

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"
)

// digestOIDFor maps a crypto.Hash to the matching CMS digest algorithm OID.
func digestOIDFor(h crypto.Hash) asn1.ObjectIdentifier {
	if h == crypto.SHA256 {
		return oidSHA256
	}
	return oidSHA1
}

// signReplyContentInfo builds a KDC-style PKINIT reply CMS ContentInfo
// (SignedData) over embedEContent, computing the messageDigest and signature
// over signedEContent (equal to embedEContent for a well-formed reply; distinct
// to model a tampered eContent) and signing with signKey (the signer
// certificate's own key for a valid reply; a foreign key to model a bad
// signature). It mirrors the wire layout BuildSignedAuthPack produces for the
// request, but with the reply's id-pkinit-DHKeyData eContentType.
func signReplyContentInfo(t *testing.T, signKey *rsa.PrivateKey, signerCert *x509.Certificate, digest crypto.Hash, signedEContent, embedEContent []byte) []byte {
	t.Helper()

	// signedAttrs: content-type = id-pkinit-DHKeyData, message-digest =
	// H(signedEContent), each wrapped as a SET OF with one value.
	ctSet := marshalSetOf(mustMarshal(t, oidIDPKINITDHKeyData))
	mdSet := marshalSetOf(mustMarshal(t, hashBytes(digest, signedEContent)))
	attrs := []attribute{
		{Type: oidContentType, Values: asn1.RawValue{FullBytes: ctSet}},
		{Type: oidMessageDigest, Values: asn1.RawValue{FullBytes: mdSet}},
	}

	// Signature is over the DER of signedAttrs as an EXPLICIT SET OF.
	signedAttrsSet, err := marshalAttributesSet(attrs)
	if err != nil {
		t.Fatalf("marshalAttributesSet: %v", err)
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, signKey, digest, hashBytes(digest, signedAttrsSet))
	if err != nil {
		t.Fatalf("SignPKCS1v15: %v", err)
	}

	// Re-tag signedAttrs as [0] IMPLICIT for the SignerInfo.
	var tmp asn1.RawValue
	if _, err := asn1.Unmarshal(signedAttrsSet, &tmp); err != nil {
		t.Fatalf("unmarshal signedAttrs: %v", err)
	}
	signedAttrsImplicit := mustMarshal(t, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: tmp.Bytes})

	digOID := digestOIDFor(digest)
	si := signerInfo{
		Version: 1,
		SID: issuerAndSerialNumber{
			Issuer:       asn1.RawValue{FullBytes: signerCert.RawIssuer},
			SerialNumber: signerCert.SerialNumber,
		},
		DigestAlgorithm:    algorithmIdentifier{Algorithm: digOID, Parameters: asn1.RawValue{FullBytes: nullDER}},
		SignedAttrs:        asn1.RawValue{FullBytes: signedAttrsImplicit},
		SignatureAlgorithm: algorithmIdentifier{Algorithm: oidRSAEncryption, Parameters: asn1.RawValue{FullBytes: nullDER}},
		Signature:          signature,
	}
	digAlg := mustMarshal(t, algorithmIdentifier{Algorithm: digOID, Parameters: asn1.RawValue{FullBytes: nullDER}})
	sd := signedData{
		Version:          3,
		DigestAlgorithms: asn1.RawValue{FullBytes: marshalSetOf(digAlg)},
		EncapContentInfo: encapsulatedContentInfo{EContentType: oidIDPKINITDHKeyData, EContent: embedEContent},
		Certificates:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: signerCert.Raw},
		SignerInfos:      asn1.RawValue{FullBytes: marshalSetOf(mustMarshal(t, si))},
	}
	ci := contentInfo{
		ContentType: oidSignedData,
		Content:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: mustMarshal(t, sd)},
	}
	return mustMarshal(t, ci)
}

// makeCAAndLeaf generates a CA certificate and a leaf certificate signed by it,
// returning the CA cert, the leaf cert and the leaf's private key.
func makeCAAndLeaf(t *testing.T) (caCert, leafCert *x509.Certificate, leafKey *rsa.PrivateKey) {
	t.Helper()
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test PKINIT CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	caCert, err = x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}

	leafKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "kdc.test.local"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create leaf cert: %v", err)
	}
	leafCert, err = x509.ParseCertificate(leafDER)
	if err != nil {
		t.Fatalf("parse leaf cert: %v", err)
	}
	return caCert, leafCert, leafKey
}

// TestVerifySignedDataEContent exercises the RFC 4556 §3.2.4 KDC signature
// verifier over crafted CMS SignedData replies.
func TestVerifySignedDataEContent(t *testing.T) {
	// A realistic eContent (a KDCDHKeyInfo DER).
	group := MODPGroup2()
	kdcKP, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("GenerateDHKeyPair: %v", err)
	}
	yBytes := mustMarshal(t, kdcKP.Y)
	eContent := mustMarshal(t, kdcDHKeyInfo{
		SubjectPublicKey: asn1.BitString{Bytes: yBytes, BitLength: len(yBytes) * 8},
		Nonce:            7,
	})

	// A self-signed KDC certificate (the pin / self-signed lab case).
	selfKey, selfCertDER, err := GenerateSelfSignedCert(2048, "self-signed-kdc")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}
	selfCert, err := x509.ParseCertificate(selfCertDER)
	if err != nil {
		t.Fatalf("parse self-signed cert: %v", err)
	}

	// A CA-issued KDC certificate (the enterprise-CA chain case).
	caCert, leafCert, leafKey := makeCAAndLeaf(t)

	// A foreign key/cert used as a wrong anchor.
	_, otherCertDER, err := GenerateSelfSignedCert(2048, "other-kdc")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert(other): %v", err)
	}
	otherCert, err := x509.ParseCertificate(otherCertDER)
	if err != nil {
		t.Fatalf("parse other cert: %v", err)
	}

	tamperedEContent := append([]byte(nil), eContent...)
	tamperedEContent[len(tamperedEContent)-1] ^= 0xFF

	tests := []struct {
		name    string
		ci      []byte
		opts    *VerifyOptions
		wantErr bool
	}{
		{
			name: "valid self-signed pinned",
			ci:   signReplyContentInfo(t, selfKey, selfCert, crypto.SHA1, eContent, eContent),
			opts: &VerifyOptions{Anchors: []*x509.Certificate{selfCert}},
		},
		{
			name: "valid sha256 self-signed pinned",
			ci:   signReplyContentInfo(t, selfKey, selfCert, crypto.SHA256, eContent, eContent),
			opts: &VerifyOptions{Anchors: []*x509.Certificate{selfCert}},
		},
		{
			name: "valid chain to CA",
			ci:   signReplyContentInfo(t, leafKey, leafCert, crypto.SHA1, eContent, eContent),
			opts: &VerifyOptions{Anchors: []*x509.Certificate{caCert}},
		},
		{
			name: "nil opts skips verification",
			ci:   signReplyContentInfo(t, selfKey, selfCert, crypto.SHA1, eContent, eContent),
			opts: nil,
		},
		{
			name: "explicit insecure skip",
			ci:   signReplyContentInfo(t, selfKey, selfCert, crypto.SHA1, eContent, eContent),
			opts: &VerifyOptions{InsecureSkipSignatureCheck: true},
		},
		{
			name:    "bad signature (foreign signing key)",
			ci:      signReplyContentInfo(t, leafKey, selfCert, crypto.SHA1, eContent, eContent),
			opts:    &VerifyOptions{Anchors: []*x509.Certificate{selfCert}},
			wantErr: true,
		},
		{
			name:    "tampered eContent (messageDigest mismatch)",
			ci:      signReplyContentInfo(t, selfKey, selfCert, crypto.SHA1, eContent, tamperedEContent),
			opts:    &VerifyOptions{Anchors: []*x509.Certificate{selfCert}},
			wantErr: true,
		},
		{
			name:    "wrong anchor (no chain)",
			ci:      signReplyContentInfo(t, leafKey, leafCert, crypto.SHA1, eContent, eContent),
			opts:    &VerifyOptions{Anchors: []*x509.Certificate{otherCert}},
			wantErr: true,
		},
		{
			name:    "verification enabled but no anchor",
			ci:      signReplyContentInfo(t, selfKey, selfCert, crypto.SHA1, eContent, eContent),
			opts:    &VerifyOptions{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := verifySignedDataEContent(tc.ci, tc.opts)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (eContent len %d)", len(got))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) == 0 {
				t.Fatal("returned empty eContent")
			}
		})
	}
}
