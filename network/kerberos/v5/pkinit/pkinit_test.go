package pkinit

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"hash"
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

// unhex parses a whitespace-separated hex string into bytes for the KAT vectors.
func unhex(t *testing.T, s string) []byte {
	t.Helper()
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if c := s[i]; c != ' ' && c != '\n' && c != '\t' {
			clean = append(clean, c)
		}
	}
	b, err := hex.DecodeString(string(clean))
	if err != nil {
		t.Fatalf("bad hex vector: %v", err)
	}
	return b
}

// TestAgilityKDFRFC8636KAT reproduces the RFC 8636 §8 known-answer test vectors
// for the algorithm-agility KDF, exercising the full OtherInfo construction plus
// the SP800-56A concatenation KDF against the RFC's published wire outputs. The
// common inputs (§8.1) are: Z = 256 zero octets, client "lha@SU.SE", server
// "krbtgt/SU.SE@SU.SE" (both NT-PRINCIPAL, per the reference implementation),
// as-REQ = 10 octets of 0xAA, pk-as-rep = 9 octets of 0xBB, reply-key enctype 18.
func TestAgilityKDFRFC8636KAT(t *testing.T) {
	z := make([]byte, 256) // §8.1: 256 zero octets
	in := KDFInputs{
		ClientRealm: "SU.SE",
		ClientName:  PrincipalName{NameType: 1, NameString: []string{"lha"}},
		ServerRealm: "SU.SE",
		ServerName:  PrincipalName{NameType: 1, NameString: []string{"krbtgt", "SU.SE"}},
		EType:       18,
		ASReq:       bytes.Repeat([]byte{0xAA}, 10),
		PKASRep:     bytes.Repeat([]byte{0xBB}, 9),
	}

	cases := []struct {
		name    string
		oid     asn1.ObjectIdentifier
		newHash func() hash.Hash
		keyLen  int
		want    string
	}{
		{
			name:    "sha1-enctype18", // RFC 8636 §8.2
			oid:     oidPKINITKDFSHA1,
			newHash: sha1.New,
			keyLen:  32,
			want:    "E6AB38C9 413E035B B079201E D0B6B73D 8D49A814 A737C04E E6649614 206F73AD",
		},
		{
			name:    "sha256-enctype18", // RFC 8636 §8.3
			oid:     oidPKINITKDFSHA256,
			newHash: sha256.New,
			keyLen:  32,
			want:    "77EF4E48 C420AE3F EC75109D 7981697E ED5D295C 90C62564 F7BFD101 FA9BC1D5",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			otherInfo, err := buildKDFOtherInfo(tc.oid, in)
			if err != nil {
				t.Fatalf("buildKDFOtherInfo: %v", err)
			}
			got := AgilityKDF(tc.newHash, z, otherInfo, tc.keyLen)
			want := unhex(t, tc.want)
			if !bytes.Equal(got, want) {
				t.Fatalf("KDF output mismatch\n got: %X\nwant: %X\nOtherInfo: %X", got, want, otherInfo)
			}
			// The hash constructor must also be the one kdfHash selects for the OID.
			sel, ok := kdfHash(tc.oid)
			if !ok {
				t.Fatalf("kdfHash(%v) not found", tc.oid)
			}
			if !bytes.Equal(AgilityKDF(sel, z, otherInfo, tc.keyLen), want) {
				t.Fatal("kdfHash-selected hash produced a different key")
			}
		})
	}
}

// TestAgilityKDFDiffersFromOctetString2Key confirms the RFC 8636 SHA-1 agility
// KDF is a distinct construction from the legacy RFC 4556 SHA-1 octetstring2key:
// the same inputs must not collide (different counter width and byte order, and
// the OtherInfo mixed in).
func TestAgilityKDFDiffersFromOctetString2Key(t *testing.T) {
	z := bytes.Repeat([]byte{0x11}, 32)
	otherInfo := []byte{0x30, 0x00}
	agility := AgilityKDF(sha1.New, z, otherInfo, 16)
	legacy := OctetString2Key(z, 16)
	if bytes.Equal(agility, legacy) {
		t.Fatal("agility SHA-1 KDF collided with octetstring2key")
	}
}

// TestKDFSelectionFromReply checks that ParseASRepPAData records the KDC-selected
// kdfID (sha256 / sha384) from the DHRepInfo.kdf field and leaves KDFID nil when
// the field is absent (the legacy octetstring2key fallback trigger).
func TestKDFSelectionFromReply(t *testing.T) {
	group := MODPGroup2()
	kdcKP, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("GenerateDHKeyPair: %v", err)
	}
	serverNonce := bytes.Repeat([]byte{0x5A}, DHNonceLen)

	cases := []struct {
		name string
		kdf  asn1.ObjectIdentifier // nil => no kdf field
		want asn1.ObjectIdentifier
	}{
		{"absent", nil, nil},
		{"sha256", oidPKINITKDFSHA256, oidPKINITKDFSHA256},
		{"sha384", oidPKINITKDFSHA384, oidPKINITKDFSHA384},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pa := buildSyntheticASRepWithKDF(t, kdcKP, serverNonce, tc.kdf)
			reply, err := ParseASRepPAData(pa, nil)
			if err != nil {
				t.Fatalf("ParseASRepPAData: %v", err)
			}
			if tc.want == nil {
				if len(reply.KDFID) != 0 {
					t.Fatalf("KDFID = %v, want nil", reply.KDFID)
				}
				return
			}
			if !reply.KDFID.Equal(tc.want) {
				t.Fatalf("KDFID = %v, want %v", reply.KDFID, tc.want)
			}
		})
	}
}

// TestDeriveReplyKeyAgilityRejectsUnknownKDF confirms an unsupported/empty kdfID
// is refused rather than silently mis-derived.
func TestDeriveReplyKeyAgilityRejectsUnknownKDF(t *testing.T) {
	group := MODPGroup2()
	kp, err := GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("GenerateDHKeyPair: %v", err)
	}
	r := &Request{KeyPair: kp, ClientDHNonce: bytes.Repeat([]byte{0x1}, DHNonceLen)}
	reply := &Reply{KDCPublicValue: kp.Y, KDFID: asn1.ObjectIdentifier{1, 2, 3, 4}}
	if _, err := r.DeriveReplyKeyAgility(reply, 32, KDFInputs{}); err == nil {
		t.Fatal("expected error for unsupported kdfID")
	}
	reply.KDFID = nil
	if _, err := r.DeriveReplyKeyAgility(reply, 32, KDFInputs{}); err == nil {
		t.Fatal("expected error for empty kdfID")
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

	reply, err := ParseASRepPAData(replyPA, nil)
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
	return buildSyntheticASRepWithKDF(t, kdcKP, serverNonce, nil)
}

// buildSyntheticASRepWithKDF is buildSyntheticASRep with an optional RFC 8636
// kdf [2] KDFAlgorithmId element appended to the DHRepInfo (nil kdfID omits it).
func buildSyntheticASRepWithKDF(t *testing.T, kdcKP *DHKeyPair, serverNonce []byte, kdfID asn1.ObjectIdentifier) []byte {
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

	// kdf [2] EXPLICIT KDFAlgorithmId { kdf-id [0] OID }, when negotiated.
	if len(kdfID) > 0 {
		algDER := mustMarshal(t, kdfAlgorithmID{KDFID: kdfID})
		kdfElem := mustMarshal(t, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 2, IsCompound: true, Bytes: algDER})
		dhRepContent = append(dhRepContent, kdfElem...)
	}

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
