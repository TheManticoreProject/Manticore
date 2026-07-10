package pkinit

import (
	"bytes"
	"crypto/sha1"
	"encoding/asn1"
	"testing"
	"time"
)

func TestDHSharedSecretRoundTrip(t *testing.T) {
	group := MODPGroup2()
	a, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("GenerateDHKeyPair(a): %v", err)
	}
	b, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("GenerateDHKeyPair(b): %v", err)
	}
	sa, err := a.SharedSecret(b.Y)
	if err != nil {
		t.Fatalf("a.SharedSecret: %v", err)
	}
	sb, err := b.SharedSecret(a.Y)
	if err != nil {
		t.Fatalf("b.SharedSecret: %v", err)
	}
	if !bytes.Equal(sa, sb) {
		t.Fatal("DH shared secrets differ")
	}
	if len(sa) != group.modulusLen() {
		t.Fatalf("shared secret not padded to modulus length: got %d want %d", len(sa), group.modulusLen())
	}
}

func TestMODPPrimesBitLength(t *testing.T) {
	if bl := MODPGroup2().P.BitLen(); bl != 1024 {
		t.Fatalf("group 2 modulus is %d bits, want 1024", bl)
	}
	if bl := MODPGroup14().P.BitLen(); bl != 2048 {
		t.Fatalf("group 14 modulus is %d bits, want 2048", bl)
	}
}

func TestOctetString2KeyKATAndLength(t *testing.T) {
	x := []byte("shared-secret-material")
	// The first 20 bytes must equal SHA1(0x00 | x) (RFC 4556 §3.2.3.1).
	first := sha1.Sum(append([]byte{0x00}, x...))
	got := OctetString2Key(x, 32)
	if len(got) != 32 {
		t.Fatalf("key length %d, want 32", len(got))
	}
	if !bytes.Equal(got[:20], first[:]) {
		t.Fatal("first 20 bytes of octetstring2key != SHA1(0x00|x)")
	}
	// The next block must be SHA1(0x01 | x).
	second := sha1.Sum(append([]byte{0x01}, x...))
	if !bytes.Equal(got[20:32], second[:12]) {
		t.Fatal("bytes 20..32 != SHA1(0x01|x)[:12]")
	}
	// A 16-byte request must be a strict prefix of the 32-byte one.
	if !bytes.Equal(OctetString2Key(x, 16), got[:16]) {
		t.Fatal("16-byte key is not a prefix of the 32-byte key")
	}
}

func TestAuthPackRoundTrip(t *testing.T) {
	group := MODPGroup2()
	kp, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("GenerateDHKeyPair: %v", err)
	}
	cpv, err := buildClientPublicValue(kp)
	if err != nil {
		t.Fatalf("buildClientPublicValue: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	sum := sha1.Sum([]byte("req-body"))
	ap := authPack{
		PKAuthenticator: pkAuthenticator{
			CUSec:      123456,
			CTime:      now,
			Nonce:      42,
			PAChecksum: sum[:],
		},
		ClientPublicValue: cpv,
		ClientDHNonce:     bytes.Repeat([]byte{0xAB}, DHNonceLen),
	}
	der, err := asn1.Marshal(ap)
	if err != nil {
		t.Fatalf("marshal AuthPack: %v", err)
	}
	var got authPack
	if _, err := asn1.Unmarshal(der, &got); err != nil {
		t.Fatalf("unmarshal AuthPack: %v", err)
	}
	if got.PKAuthenticator.Nonce != 42 || got.PKAuthenticator.CUSec != 123456 {
		t.Fatal("PKAuthenticator scalar fields did not round-trip")
	}
	if !bytes.Equal(got.PKAuthenticator.PAChecksum, sum[:]) {
		t.Fatal("paChecksum did not round-trip")
	}
	if !got.ClientPublicValue.Algorithm.Algorithm.Equal(oidDHPublicNumber) {
		t.Fatal("clientPublicValue algorithm OID wrong")
	}
	// The subjectPublicKey BIT STRING must decode back to the DH public value Y.
	y, err := parseKDCPublicValue(kdcDHKeyInfo{SubjectPublicKey: got.ClientPublicValue.SubjectPublicKey})
	if err != nil {
		t.Fatalf("parse public value: %v", err)
	}
	if y.Cmp(kp.Y) != 0 {
		t.Fatal("recovered DH public value != original Y")
	}
}

// TestFullExchangeRoundTrip builds a real PA-PK-AS-REQ, then constructs the
// matching KDC-side PA-PK-AS-REP with an independent KDC DH key, and verifies
// that both sides derive the identical AS reply key.
func TestFullExchangeRoundTrip(t *testing.T) {
	group := MODPGroup2()
	priv, certDER, err := GenerateSelfSignedCert(2048, "shadowcred-test")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}
	reqBody := []byte("this-stands-in-for-the-kdc-req-body")

	_, pkReq, err := BuildASReqPAData(reqBody, priv, certDER, group, 0x11223344, time.Now())
	if err != nil {
		t.Fatalf("BuildASReqPAData: %v", err)
	}

	// KDC side: its own ephemeral key, and a server DH nonce.
	kdcKP, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("KDC GenerateDHKeyPair: %v", err)
	}
	serverNonce := bytes.Repeat([]byte{0x5A}, DHNonceLen)

	replyPA := buildSyntheticASRep(t, kdcKP, serverNonce)

	reply, err := ParseASRepPAData(replyPA)
	if err != nil {
		t.Fatalf("ParseASRepPAData: %v", err)
	}
	if reply.KDCPublicValue.Cmp(kdcKP.Y) != 0 {
		t.Fatal("parsed KDC public value != KDC Y")
	}
	if !bytes.Equal(reply.ServerDHNonce, serverNonce) {
		t.Fatal("parsed serverDHNonce mismatch")
	}

	const keyLen = 32
	clientKey, err := pkReq.DeriveReplyKey(reply, keyLen)
	if err != nil {
		t.Fatalf("client DeriveReplyKey: %v", err)
	}

	// KDC-side derivation of the same key.
	kdcShared, err := kdcKP.SharedSecret(pkReq.KeyPair.Y)
	if err != nil {
		t.Fatalf("KDC SharedSecret: %v", err)
	}
	kdcInput := append(append(append([]byte{}, kdcShared...), pkReq.ClientDHNonce...), serverNonce...)
	kdcKey := OctetString2Key(kdcInput, keyLen)
	if !bytes.Equal(clientKey, kdcKey) {
		t.Fatal("client and KDC derived different reply keys")
	}
}

// buildSyntheticASRep constructs a PA-PK-AS-REP (dhInfo variant) in the wire form
// a Windows KDC emits: the dhInfo [0] tag replaces the DHRepInfo SEQUENCE, with
// dhSignedData [0] IMPLICIT OCTET STRING and serverDHNonce [1] IMPLICIT.
func buildSyntheticASRep(t *testing.T, kdcKP *DHKeyPair, serverNonce []byte) []byte {
	t.Helper()
	yBytes, err := asn1.Marshal(kdcKP.Y)
	if err != nil {
		t.Fatalf("marshal KDC Y: %v", err)
	}
	info := kdcDHKeyInfo{
		SubjectPublicKey: asn1.BitString{Bytes: yBytes, BitLength: len(yBytes) * 8},
		Nonce:            7,
	}
	infoDER, err := asn1.Marshal(info)
	if err != nil {
		t.Fatalf("marshal KDCDHKeyInfo: %v", err)
	}
	// Minimal SignedData carrying infoDER as eContent (no signer; the parser does
	// not verify the KDC signature).
	sd := signedData{
		Version:          3,
		DigestAlgorithms: asn1.RawValue{FullBytes: marshalSetOf(mustMarshal(t, algorithmIdentifier{Algorithm: oidSHA1, Parameters: asn1.RawValue{FullBytes: nullDER}}))},
		EncapContentInfo: encapsulatedContentInfo{EContentType: oidIDPKINITDHKeyData, EContent: infoDER},
		SignerInfos:      asn1.RawValue{FullBytes: marshalSetOf(nil)},
	}
	sdDER := mustMarshal(t, sd)
	ci := contentInfo{
		ContentType: oidSignedData,
		Content:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: sdDER},
	}
	ciDER := mustMarshal(t, ci)

	// DHRepInfo content: dhSignedData [0] IMPLICIT OCTET STRING || serverDHNonce [1].
	dhSigned := mustMarshal(t, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: false, Bytes: ciDER})
	srvNonce := mustMarshal(t, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 1, IsCompound: false, Bytes: serverNonce})
	dhRepContent := append(dhSigned, srvNonce...)

	// dhInfo [0] wrapping the DHRepInfo fields directly.
	return mustMarshal(t, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: dhRepContent})
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := asn1.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	return b
}

func TestBuildSignedAuthPackExtracts(t *testing.T) {
	// Round-trip a SignedData whose eContent is DHKeyData through the parser used
	// on the reply path, confirming the hand-built CMS structure is decodable.
	priv, certDER, err := GenerateSelfSignedCert(2048, "cms-test")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}
	_ = priv
	// A SignedData over authData is what the request builds; verify a request
	// blob parses as a CMS ContentInfo with the signedData OID.
	authPackDER := []byte{0x30, 0x00}
	ci, err := BuildSignedAuthPack(authPackDER, priv, certDER)
	if err != nil {
		t.Fatalf("BuildSignedAuthPack: %v", err)
	}
	var parsed contentInfo
	if _, err := asn1.Unmarshal(ci, &parsed); err != nil {
		t.Fatalf("parse built ContentInfo: %v", err)
	}
	if !parsed.ContentType.Equal(oidSignedData) {
		t.Fatal("built ContentInfo is not signedData")
	}
	var sd signedData
	if _, err := asn1.Unmarshal(parsed.Content.Bytes, &sd); err != nil {
		t.Fatalf("parse built SignedData: %v", err)
	}
	if !sd.EncapContentInfo.EContentType.Equal(oidIDPKINITAuthData) {
		t.Fatal("built eContentType != id-pkinit-authData")
	}
	if !bytes.Equal(sd.EncapContentInfo.EContent, authPackDER) {
		t.Fatal("eContent did not round-trip")
	}
	if sd.Version != 3 {
		t.Fatalf("SignedData version %d, want 3", sd.Version)
	}
}
