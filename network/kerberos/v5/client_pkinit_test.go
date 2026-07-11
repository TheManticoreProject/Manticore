package kerberos

import (
	"bytes"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pkinit"
)

// CMS OIDs used to hand-build a synthetic PKINIT KDC reply (RFC 5652 / RFC 4556).
var (
	testOIDSignedData        = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	testOIDIDPKINITDHKeyData = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 2}
)

// derRaw marshals a single ASN.1 element with the given class/tag/compound flag.
func derRaw(t *testing.T, class, tag int, compound bool, content []byte) []byte {
	t.Helper()
	b, err := asn1.Marshal(asn1.RawValue{Class: class, Tag: tag, IsCompound: compound, Bytes: content})
	if err != nil {
		t.Fatalf("marshal raw (class %d tag %d): %v", class, tag, err)
	}
	return b
}

// derMarshal marshals any value, failing the test on error.
func derMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := asn1.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	return b
}

// synthPKINITReplyPA constructs the PA-PK-AS-REP (dhInfo variant) a Windows KDC
// emits, carrying the KDC's DH public value inside a minimal CMS SignedData whose
// eContent is a KDCDHKeyInfo. The parser does not verify the KDC signature, so no
// real signer is needed. The layout mirrors the pkinit package's own test harness.
func synthPKINITReplyPA(t *testing.T, kdcY *big.Int, serverNonce []byte) []byte {
	t.Helper()

	// KDCDHKeyInfo ::= SEQUENCE { subjectPublicKey [0] BIT STRING, nonce [1] INTEGER }
	yDER := derMarshal(t, kdcY)
	yBitString := derMarshal(t, asn1.BitString{Bytes: yDER, BitLength: len(yDER) * 8})
	subjPub := derRaw(t, asn1.ClassContextSpecific, 0, true, yBitString)
	nonceField := derRaw(t, asn1.ClassContextSpecific, 1, true, derMarshal(t, 7))
	infoDER := derRaw(t, asn1.ClassUniversal, asn1.TagSequence, true, append(subjPub, nonceField...))

	// EncapsulatedContentInfo ::= SEQUENCE { eContentType OID, eContent [0] OCTET STRING }
	eContentField := derRaw(t, asn1.ClassContextSpecific, 0, true, derMarshal(t, infoDER))
	encap := derRaw(t, asn1.ClassUniversal, asn1.TagSequence, true,
		append(derMarshal(t, testOIDIDPKINITDHKeyData), eContentField...))

	// SignedData ::= SEQUENCE { version, digestAlgorithms SET, encapContentInfo, signerInfos SET }
	emptySet := derRaw(t, asn1.ClassUniversal, asn1.TagSet, true, nil)
	var sdParts []byte
	sdParts = append(sdParts, derMarshal(t, 3)...)
	sdParts = append(sdParts, emptySet...)
	sdParts = append(sdParts, encap...)
	sdParts = append(sdParts, emptySet...)
	sdDER := derRaw(t, asn1.ClassUniversal, asn1.TagSequence, true, sdParts)

	// ContentInfo ::= SEQUENCE { contentType OID, content [0] SignedData }. The
	// parser captures the [0] element as a RawValue and reads its .Bytes as the
	// SignedData DER, so a single [0] wrapping the SignedData is what it expects.
	contentField := derRaw(t, asn1.ClassContextSpecific, 0, true, sdDER)
	ciDER := derRaw(t, asn1.ClassUniversal, asn1.TagSequence, true,
		append(derMarshal(t, testOIDSignedData), contentField...))

	// DHRepInfo: dhSignedData [0] IMPLICIT OCTET STRING || serverDHNonce [1].
	dhSigned := derRaw(t, asn1.ClassContextSpecific, 0, false, ciDER)
	srvNonce := derRaw(t, asn1.ClassContextSpecific, 1, false, serverNonce)

	// dhInfo [0] wrapping the DHRepInfo fields.
	return derRaw(t, asn1.ClassContextSpecific, 0, true, append(dhSigned, srvNonce...))
}

// synthPKINITReplyPAWithKDF is synthPKINITReplyPA plus an RFC 8636 kdf [2]
// KDFAlgorithmId { kdf-id [0] OID } appended to the DHRepInfo, so the reply
// negotiates an algorithm-agility KDF.
func synthPKINITReplyPAWithKDF(t *testing.T, kdcY *big.Int, serverNonce []byte, kdfID asn1.ObjectIdentifier) []byte {
	t.Helper()
	base := synthPKINITReplyPA(t, kdcY, serverNonce)
	// Unwrap the dhInfo [0] to recover the DHRepInfo content, append the kdf
	// element, and rewrap.
	var dhInfo asn1.RawValue
	if _, err := asn1.Unmarshal(base, &dhInfo); err != nil {
		t.Fatalf("unwrap dhInfo: %v", err)
	}
	kdfIDSeq := derMarshal(t, struct {
		KDFID asn1.ObjectIdentifier `asn1:"explicit,tag:0"`
	}{KDFID: kdfID})
	kdfElem := derRaw(t, asn1.ClassContextSpecific, 2, true, kdfIDSeq)
	return derRaw(t, asn1.ClassContextSpecific, 0, true, append(dhInfo.Bytes, kdfElem...))
}

// TestProcessPKINITASRepAgilityKDF drives AS-REP processing when the KDC
// negotiates an RFC 8636 SHA-256 agility KDF and issues an AES128-SHA256
// (etype 19) reply key. The reply key is derived via the agility KDF on both
// sides through the exported DeriveReplyKeyAgility path, proving processPKINITASRep
// selects the agility KDF (not the legacy octetstring2key) from the kdfID and
// decrypts the enc-part with the AES-SHA2 reply key.
func TestProcessPKINITASRepAgilityKDF(t *testing.T) {
	group := pkinit.MODPGroup2()
	priv, certDER, err := pkinit.GenerateSelfSignedCert(2048, "unit-test")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}
	reqBytes := []byte("synthetic-as-req-DER")
	_, pkReq, err := pkinit.BuildASReqPAData([]byte("req-body"), priv, certDER, group, 0x11223344, time.Now())
	if err != nil {
		t.Fatalf("BuildASReqPAData: %v", err)
	}

	kdcKP, err := pkinit.GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("KDC GenerateDHKeyPair: %v", err)
	}
	serverNonce := bytes.Repeat([]byte{0x5A}, pkinit.DHNonceLen)
	paValue := synthPKINITReplyPAWithKDF(t, kdcKP.Y, serverNonce, asn1.ObjectIdentifier{1, 3, 6, 1, 5, 2, 3, 6, 2})

	c := NewClient("alice", "corp.local", "")

	const etype = messages.ETypeAES128CTSHMACSHA256 // 19, AES-SHA2
	keyLen := kerbcrypto.KeyLen(etype)

	// KDC-side derivation of the reply key via the same agility KDF and identical
	// OtherInfo the client will rebuild inside processPKINITASRep.
	reply, err := pkinit.ParseASRepPAData(paValue, nil)
	if err != nil {
		t.Fatalf("ParseASRepPAData: %v", err)
	}
	if len(reply.KDFID) == 0 {
		t.Fatal("reply did not carry a negotiated kdfID")
	}
	replyKey, err := pkReq.DeriveReplyKeyAgility(reply, keyLen, pkinit.KDFInputs{
		ClientRealm: c.realm,
		ClientName:  pkinit.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{c.username}},
		ServerRealm: c.realm,
		ServerName:  pkinit.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", c.realm}},
		EType:       etype,
		ASReq:       reqBytes,
		PKASRep:     paValue,
	})
	if err != nil {
		t.Fatalf("KDC DeriveReplyKeyAgility: %v", err)
	}

	const nonce = 0x11223344
	sessionKey := bytes.Repeat([]byte{0x42}, 16)
	wire := buildPKINITASRep(t, etype, replyKey, nonce, sessionKey, paValue)

	if err := c.processPKINITASRep(wire, pkReq, nonce, reqBytes); err != nil {
		t.Fatalf("processPKINITASRep (agility KDF): %v", err)
	}
	if !c.hasTGT {
		t.Fatal("hasTGT not set after a successful agility-KDF PKINIT AS-REP")
	}
	key, gotEType := c.PKINITReplyKey()
	if !bytes.Equal(key, replyKey) || gotEType != etype {
		t.Errorf("PKINITReplyKey = (%x, %d), want (%x, %d)", key, gotEType, replyKey, etype)
	}
}

// buildPKINITASRep assembles an AS-REP whose enc-part (EncASRepPart with the given
// nonce and session key) is encrypted under replyKey/replyEType, carrying the
// supplied PA-PK-AS-REP so processPKINITASRep can derive the reply key and decrypt.
func buildPKINITASRep(t *testing.T, replyEType int, replyKey []byte, nonce int, sessionKey, paValue []byte) []byte {
	t.Helper()
	enc := &messages.EncASRepPart{
		Key:      messages.EncryptionKey{KeyType: messages.ETypeAES256CTSHMACSHA196, KeyValue: sessionKey},
		Nonce:    nonce,
		Flags:    messages.NewKerberosFlags(messages.TicketFlagForwardable, messages.TicketFlagInitial),
		AuthTime: time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC),
		EndTime:  time.Date(2026, 7, 11, 20, 0, 0, 0, time.UTC),
		SRealm:   testRealm,
		SName:    messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", testRealm}},
	}
	plain, err := enc.Marshal()
	if err != nil {
		t.Fatalf("EncASRepPart.Marshal: %v", err)
	}
	cipher, err := kerbcrypto.Encrypt(replyEType, replyKey, kerbcrypto.KeyUsageASRepEncPart, plain)
	if err != nil {
		t.Fatalf("encrypt enc-part: %v", err)
	}
	rep := &messages.ASRep{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeASRep,
		CRealm:  testRealm,
		CName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		PAData:  []messages.PAData{{PADataType: messages.PAPKASRep, PADataValue: paValue}},
		Ticket: messages.Ticket{
			TktVno:  messages.KerberosV5,
			Realm:   testRealm,
			SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", testRealm}},
			EncPart: messages.EncryptedData{EType: replyEType, Cipher: bytes.Repeat([]byte{0x5a}, 32)},
		},
		EncPart: messages.EncryptedData{EType: replyEType, Cipher: cipher},
	}
	wire, err := rep.Marshal()
	if err != nil {
		t.Fatalf("ASRep.Marshal: %v", err)
	}
	return wire
}

// pkinitExchange sets up a matched client request / KDC reply pair: it builds a
// real client PKINIT request, derives the shared reply key on the KDC side, and
// returns the client, its request, the synthetic PA-PK-AS-REP, and the AES256
// reply key both sides agree on.
func pkinitExchange(t *testing.T) (c *KerberosClient, pkReq *pkinit.Request, paValue, replyKey []byte) {
	t.Helper()
	group := pkinit.MODPGroup2()
	priv, certDER, err := pkinit.GenerateSelfSignedCert(2048, "unit-test")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}
	_, pkReq, err = pkinit.BuildASReqPAData([]byte("req-body"), priv, certDER, group, 0x11223344, time.Now())
	if err != nil {
		t.Fatalf("BuildASReqPAData: %v", err)
	}

	kdcKP, err := pkinit.GenerateDHKeyPair(group)
	if err != nil {
		t.Fatalf("KDC GenerateDHKeyPair: %v", err)
	}
	serverNonce := bytes.Repeat([]byte{0x5A}, pkinit.DHNonceLen)
	paValue = synthPKINITReplyPA(t, kdcKP.Y, serverNonce)

	// KDC-side derivation of the reply key (must equal the client's candidate[0]).
	shared, err := kdcKP.SharedSecret(pkReq.KeyPair.Y)
	if err != nil {
		t.Fatalf("KDC SharedSecret: %v", err)
	}
	keyMaterial := append(append(append([]byte{}, shared...), pkReq.ClientDHNonce...), serverNonce...)
	replyKey = pkinit.OctetString2Key(keyMaterial, kerbcrypto.KeyLen(messages.ETypeAES256CTSHMACSHA196))

	c = NewClient("alice", "corp.local", "")
	return c, pkReq, paValue, replyKey
}

// TestProcessPKINITASRepSuccess drives the PKINIT AS-REP processing end to end
// with a synthetic KDC reply: the DH reply key is reconstructed, the AS-REP
// enc-part decrypts, and the TGT plus the reply key (needed for UnPAC-the-hash)
// are stored on the client.
func TestProcessPKINITASRepSuccess(t *testing.T) {
	c, pkReq, paValue, replyKey := pkinitExchange(t)
	const nonce = 0x11223344
	sessionKey := bytes.Repeat([]byte{0x42}, 32)
	wire := buildPKINITASRep(t, messages.ETypeAES256CTSHMACSHA196, replyKey, nonce, sessionKey, paValue)

	if err := c.processPKINITASRep(wire, pkReq, nonce, []byte("req-body")); err != nil {
		t.Fatalf("processPKINITASRep: %v", err)
	}
	if !c.hasTGT {
		t.Fatal("hasTGT not set after a successful PKINIT AS-REP")
	}
	if !bytes.Equal(c.sessionKey, sessionKey) {
		t.Errorf("session key mismatch: %x", c.sessionKey)
	}
	key, etype := c.PKINITReplyKey()
	if !bytes.Equal(key, replyKey) || etype != messages.ETypeAES256CTSHMACSHA196 {
		t.Errorf("PKINITReplyKey = (%x, %d), want (%x, AES256)", key, etype, replyKey)
	}
}

// TestProcessPKINITASRepRejectsBadEType confirms an AS-REP whose reply-key etype
// is unsupported (KeyLen 0) is rejected before any decryption.
func TestProcessPKINITASRepRejectsBadEType(t *testing.T) {
	c, pkReq, paValue, replyKey := pkinitExchange(t)
	// A parseable reply, but the enc-part advertises an unusable reply-key etype.
	wire := buildPKINITASRep(t, messages.ETypeAES256CTSHMACSHA196, replyKey, 1, bytes.Repeat([]byte{0x42}, 32), paValue)
	var rep messages.ASRep
	if _, err := rep.Unmarshal(wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rep.EncPart.EType = 0 // no key length -> rejected
	bad, err := rep.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := c.processPKINITASRep(bad, pkReq, 1, []byte("req-body")); err == nil {
		t.Fatal("expected an error for an unsupported reply-key etype")
	}
	if c.hasTGT {
		t.Error("hasTGT must not be set on a rejected AS-REP")
	}
}

// TestProcessPKINITASRepNonceMismatch confirms an enc-part whose nonce differs
// from the request nonce is rejected (no candidate key decrypts to a matching
// nonce), and no TGT is stored.
func TestProcessPKINITASRepNonceMismatch(t *testing.T) {
	c, pkReq, paValue, replyKey := pkinitExchange(t)
	wire := buildPKINITASRep(t, messages.ETypeAES256CTSHMACSHA196, replyKey, 999, bytes.Repeat([]byte{0x42}, 32), paValue)
	if err := c.processPKINITASRep(wire, pkReq, 12345, []byte("req-body")); err == nil {
		t.Fatal("expected a nonce-mismatch rejection")
	}
	if c.hasTGT {
		t.Error("hasTGT must not be set on a nonce mismatch")
	}
}

// TestProcessPKINITASRepNoPAData confirms an AS-REP lacking a PA-PK-AS-REP element
// is rejected (the KDC did not return its DH contribution).
func TestProcessPKINITASRepNoPAData(t *testing.T) {
	c, pkReq, _, replyKey := pkinitExchange(t)
	// Reply with an unrelated padata type only.
	wire := buildPKINITASRep(t, messages.ETypeAES256CTSHMACSHA196, replyKey, 1, bytes.Repeat([]byte{0x42}, 32), nil)
	var rep messages.ASRep
	if _, err := rep.Unmarshal(wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rep.PAData = []messages.PAData{pacRequestPA()} // no PA-PK-AS-REP
	stripped, err := rep.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := c.processPKINITASRep(stripped, pkReq, 1, []byte("req-body")); err == nil {
		t.Fatal("expected an error when the AS-REP has no PA-PK-AS-REP element")
	}
}

// TestWithPKINITConfiguration covers the PKINIT configuration accessors: WithPKINIT
// records the key/cert and installs the default DH groups, pkinitConfigured
// reflects that, WithPKINITGroups overrides the group list, and the reply key is
// empty until a successful exchange.
func TestWithPKINITConfiguration(t *testing.T) {
	priv, certDER, err := pkinit.GenerateSelfSignedCert(2048, "config-test")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}

	c := NewClient("alice", "corp.local", "10.0.0.1")
	if c.pkinitConfigured() {
		t.Error("pkinitConfigured should be false before WithPKINIT")
	}
	c.WithPKINIT(priv, certDER)
	if !c.pkinitConfigured() {
		t.Fatal("pkinitConfigured should be true after WithPKINIT")
	}
	if len(c.pkinitGroups) != 2 {
		t.Errorf("default PKINIT groups = %d, want 2 (MODP14, MODP2)", len(c.pkinitGroups))
	}
	if key, etype := c.PKINITReplyKey(); key != nil || etype != 0 {
		t.Errorf("PKINITReplyKey should be empty before a successful exchange, got (%x, %d)", key, etype)
	}

	c.WithPKINITGroups(pkinit.MODPGroup2())
	if len(c.pkinitGroups) != 1 {
		t.Errorf("WithPKINITGroups override = %d groups, want 1", len(c.pkinitGroups))
	}
}
