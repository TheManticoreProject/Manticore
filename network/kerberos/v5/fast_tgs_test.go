package kerberos

import (
	"bytes"
	"encoding/asn1"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// fastTGSClient returns a client holding a usable (hand-built) TGT with a known
// AES256 session key and FAST enabled, so the TGS-exchange FAST assembly helpers
// can run without a KDC. The TGT session key doubles as the armor "ticket"
// session key per RFC 6113 §5.4.2.
func fastTGSClient(t *testing.T, sessionKey []byte, sessionEType int) *KerberosClient {
	t.Helper()
	tkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   testRealm,
		SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", testRealm}},
		EncPart: messages.EncryptedData{EType: sessionEType, Cipher: bytes.Repeat([]byte{0x7a}, 32)},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}
	c := NewClient("alice", testRealm, "10.0.0.1").WithPassword("Passw0rd!")
	c.tgtTicket = tkt
	c.tgtTicketRaw = raw
	c.sessionKey = sessionKey
	c.sessionEType = sessionEType
	c.hasTGT = true
	// Enabling FAST via a self-armor is only what flips GetTGS onto the FAST path;
	// the TGS armor itself is the client's own TGT (set above), not this field.
	c.WithFASTArmor(c.username, c.realm, tkt, raw, sessionKey, sessionEType)
	return c
}

// parseFASTTGSReq decodes a FAST-armored TGS-REQ, returning the parsed outer
// request, the AP-REQ bytes carried in the PA-TGS-REQ, and the KrbFastArmoredReq
// unwrapped from PA-FX-FAST.
func parseFASTTGSReq(t *testing.T, reqBytes []byte) (messages.TGSReq, []byte, messages.KrbFastArmoredReq) {
	t.Helper()
	var tgsReq messages.TGSReq
	if _, err := tgsReq.Unmarshal(reqBytes); err != nil {
		t.Fatalf("parse FAST TGS-REQ: %v", err)
	}

	var apReqBytes, fastValue []byte
	for _, pa := range tgsReq.PAData {
		switch pa.PADataType {
		case messages.PATGSReq:
			apReqBytes = pa.PADataValue
		case messages.PAFXFast:
			fastValue = pa.PADataValue
		}
	}
	if apReqBytes == nil {
		t.Fatal("FAST TGS-REQ missing PA-TGS-REQ")
	}
	if fastValue == nil {
		t.Fatal("FAST TGS-REQ missing PA-FX-FAST")
	}

	// PA-FX-FAST-REQUEST is a CHOICE: [0] EXPLICIT KrbFastArmoredReq.
	var choice asn1.RawValue
	if _, err := asn1.Unmarshal(fastValue, &choice); err != nil {
		t.Fatalf("parse PA-FX-FAST CHOICE: %v", err)
	}
	if choice.Class != asn1.ClassContextSpecific || choice.Tag != 0 {
		t.Fatalf("PA-FX-FAST alternative class=%d tag=%d, want context [0]", choice.Class, choice.Tag)
	}
	var armored messages.KrbFastArmoredReq
	if _, err := armored.Unmarshal(choice.Bytes); err != nil {
		t.Fatalf("parse KrbFastArmoredReq: %v", err)
	}
	return tgsReq, apReqBytes, armored
}

// TestBuildFASTTGSReqStructure verifies the wire shape of a FAST-armored TGS-REQ:
// the outer request carries both PA-TGS-REQ and PA-FX-FAST, and the armored
// request omits the armor field (RFC 6113 §5.4.2: the armor is the TGT itself).
func TestBuildFASTTGSReqStructure(t *testing.T) {
	sessionKey := bytes.Repeat([]byte{0x24}, kerbcrypto.KeyLen(messages.ETypeAES256CTSHMACSHA196))
	c := fastTGSClient(t, sessionKey, messages.ETypeAES256CTSHMACSHA196)
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}}

	reqBytes, ctx, err := c.buildFASTTGSReq(testRealm, sname, true, c.tgtTicket, c.tgtTicketRaw, c.sessionKey, c.sessionEType, nil)
	if err != nil {
		t.Fatalf("buildFASTTGSReq: %v", err)
	}

	_, apReqBytes, armored := parseFASTTGSReq(t, reqBytes)

	if armored.Armor != nil {
		t.Errorf("TGS KrbFastArmoredReq must omit the armor field, got %+v", armored.Armor)
	}
	if len(apReqBytes) == 0 {
		t.Fatal("PA-TGS-REQ carried no AP-REQ bytes")
	}
	if len(armored.EncFastReq.Cipher) == 0 {
		t.Fatal("enc-fast-req is empty")
	}
	if ctx.armorEType != c.sessionEType {
		t.Errorf("armor etype = %d, want TGT session etype %d", ctx.armorEType, c.sessionEType)
	}
	if len(ctx.subkey) != kerbcrypto.KeyLen(c.sessionEType) {
		t.Errorf("subkey length = %d, want %d", len(ctx.subkey), kerbcrypto.KeyLen(c.sessionEType))
	}
}

// TestFASTTGSArmorKeyFromAuthenticatorSubkey is the crux of the TGS FAST rule
// (RFC 6113 §5.4.1.1 / §5.4.2): the armor key MUST be derived from the subkey in
// the PA-TGS-REQ authenticator combined with the TGT session key via KRB-FX-CF2
// under the "subkeyarmor"/"ticketarmor" peppers. It recovers the subkey from the
// AP-REQ authenticator (decrypting with the TGT session key at key usage 7) and
// reproduces the armor key the builder returned.
func TestFASTTGSArmorKeyFromAuthenticatorSubkey(t *testing.T) {
	sessionKey := bytes.Repeat([]byte{0x51}, kerbcrypto.KeyLen(messages.ETypeAES256CTSHMACSHA196))
	c := fastTGSClient(t, sessionKey, messages.ETypeAES256CTSHMACSHA196)
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}}

	reqBytes, ctx, err := c.buildFASTTGSReq(testRealm, sname, true, c.tgtTicket, c.tgtTicketRaw, c.sessionKey, c.sessionEType, nil)
	if err != nil {
		t.Fatalf("buildFASTTGSReq: %v", err)
	}
	_, apReqBytes, armored := parseFASTTGSReq(t, reqBytes)

	// Recover the authenticator subkey (encrypted under the TGT session key, key
	// usage 7 — the PA-TGS-REQ authenticator usage).
	var apReq messages.APReq
	if _, err := apReq.Unmarshal(apReqBytes); err != nil {
		t.Fatalf("parse AP-REQ: %v", err)
	}
	authBytes, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSReqPAAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		t.Fatalf("decrypt TGS authenticator: %v", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(authBytes); err != nil {
		t.Fatalf("parse TGS authenticator: %v", err)
	}
	if auth.SubKey == nil {
		t.Fatal("PA-TGS-REQ authenticator carries no subkey (required for TGS FAST)")
	}
	if !bytes.Equal(auth.SubKey.KeyValue, ctx.subkey) {
		t.Errorf("authenticator subkey %x != builder subkey %x", auth.SubKey.KeyValue, ctx.subkey)
	}

	// The armor key must equal KRB-FX-CF2(subkey, TGT-session-key, peppers).
	wantArmor, wantEType, err := kerbcrypto.KRBFXCF2(
		auth.SubKey.KeyValue, auth.SubKey.KeyType,
		c.sessionKey, c.sessionEType,
		fastPepperSubkeyArmor, fastPepperTicketArmor)
	if err != nil {
		t.Fatalf("KRBFXCF2: %v", err)
	}
	if !bytes.Equal(wantArmor, ctx.armorKey) || wantEType != ctx.armorEType {
		t.Errorf("armor key mismatch: got %x/%d, want %x/%d", ctx.armorKey, ctx.armorEType, wantArmor, wantEType)
	}

	// The req-checksum MUST be a keyed checksum over the AP-REQ (not the
	// KDC-REQ-BODY) computed with the armor key at key usage 50.
	if !kerbcrypto.VerifyChecksum(armored.ReqChecksum.CKSumType, ctx.armorKey, kerbcrypto.KeyUsageFASTReqChksum, apReqBytes, armored.ReqChecksum.Checksum) {
		t.Error("req-checksum does not verify over the AP-REQ with the armor key")
	}
}

// TestFASTTGSReqRoundTrip round-trips the armored KrbFastReq: decrypting
// enc-fast-req with the armor key (key usage 51) must recover the inner request
// with the requested SPN and the inner nonce the reply is expected to echo.
func TestFASTTGSReqRoundTrip(t *testing.T) {
	sessionKey := bytes.Repeat([]byte{0x66}, kerbcrypto.KeyLen(messages.ETypeAES256CTSHMACSHA196))
	c := fastTGSClient(t, sessionKey, messages.ETypeAES256CTSHMACSHA196)
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}}

	reqBytes, ctx, err := c.buildFASTTGSReq(testRealm, sname, true, c.tgtTicket, c.tgtTicketRaw, c.sessionKey, c.sessionEType, nil)
	if err != nil {
		t.Fatalf("buildFASTTGSReq: %v", err)
	}
	_, _, armored := parseFASTTGSReq(t, reqBytes)

	plain, err := kerbcrypto.Decrypt(armored.EncFastReq.EType, ctx.armorKey, kerbcrypto.KeyUsageFASTEnc, armored.EncFastReq.Cipher)
	if err != nil {
		t.Fatalf("decrypt enc-fast-req: %v", err)
	}
	var fastReq messages.KrbFastReq
	if _, err := fastReq.Unmarshal(plain); err != nil {
		t.Fatalf("parse KrbFastReq: %v", err)
	}
	if !principalNameEqualFold(fastReq.ReqBody.SName, sname) {
		t.Errorf("inner SName = %+v, want %+v", fastReq.ReqBody.SName, sname)
	}
	if fastReq.ReqBody.Realm != testRealm {
		t.Errorf("inner realm = %q, want %q", fastReq.ReqBody.Realm, testRealm)
	}
	if fastReq.ReqBody.Nonce != ctx.nonce {
		t.Errorf("inner nonce = %d, want ctx.nonce %d", fastReq.ReqBody.Nonce, ctx.nonce)
	}
	// The armored (inner) padata must carry PA-PAC-REQUEST.
	var sawPAC bool
	for _, pa := range fastReq.PAData {
		if pa.PADataType == messages.PAPACRequest {
			sawPAC = true
		}
	}
	if !sawPAC {
		t.Error("inner padata missing PA-PAC-REQUEST")
	}
}

// TestFASTTGSCookieEchoed confirms an FX-COOKIE handed to the builder is placed
// in the armored (inner) padata (RFC 6113 §5.4.4.3 replay cookie echo).
func TestFASTTGSCookieEchoed(t *testing.T) {
	sessionKey := bytes.Repeat([]byte{0x77}, kerbcrypto.KeyLen(messages.ETypeAES256CTSHMACSHA196))
	c := fastTGSClient(t, sessionKey, messages.ETypeAES256CTSHMACSHA196)
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}}
	cookie := &messages.PAData{PADataType: messages.PAFXCookie, PADataValue: []byte("opaque-cookie")}

	reqBytes, ctx, err := c.buildFASTTGSReq(testRealm, sname, true, c.tgtTicket, c.tgtTicketRaw, c.sessionKey, c.sessionEType, cookie)
	if err != nil {
		t.Fatalf("buildFASTTGSReq: %v", err)
	}
	_, _, armored := parseFASTTGSReq(t, reqBytes)

	plain, err := kerbcrypto.Decrypt(armored.EncFastReq.EType, ctx.armorKey, kerbcrypto.KeyUsageFASTEnc, armored.EncFastReq.Cipher)
	if err != nil {
		t.Fatalf("decrypt enc-fast-req: %v", err)
	}
	var fastReq messages.KrbFastReq
	if _, err := fastReq.Unmarshal(plain); err != nil {
		t.Fatalf("parse KrbFastReq: %v", err)
	}
	var sawCookie bool
	for _, pa := range fastReq.PAData {
		if pa.PADataType == messages.PAFXCookie && bytes.Equal(pa.PADataValue, cookie.PADataValue) {
			sawCookie = true
		}
	}
	if !sawCookie {
		t.Error("inner padata did not echo the FX-COOKIE")
	}
}
