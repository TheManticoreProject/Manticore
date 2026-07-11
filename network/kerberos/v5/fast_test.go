package kerberos

import (
	"bytes"
	"encoding/hex"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// fastArmoredClient returns a client with a FAST armor TGT configured (a dummy
// ticket plus a known AES256 armor session key) and a password credential, so
// the FAST assembly helpers can run without a KDC.
func fastArmoredClient(t *testing.T, armorKey []byte, armorEType int) *KerberosClient {
	t.Helper()
	tkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   testRealm,
		SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", testRealm}},
		EncPart: messages.EncryptedData{EType: armorEType, Cipher: bytes.Repeat([]byte{0x7a}, 32)},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("Passw0rd!")
	c.WithFASTArmor("armor$", testRealm, tkt, raw, armorKey, armorEType)
	return c
}

// TestFASTArmorKeyDerivation pins the FAST armor-key derivation: fast.go combines
// a fresh subkey with the armor TGT session key via KRB-FX-CF2 under the RFC 6113
// "subkeyarmor"/"ticketarmor" peppers. This confirms fast.go's pepper constants
// match the RFC and reproduce the crypto package's known-answer value.
func TestFASTArmorKeyDerivation(t *testing.T) {
	if fastPepperSubkeyArmor != "subkeyarmor" || fastPepperTicketArmor != "ticketarmor" {
		t.Fatalf("armor peppers drifted: %q / %q", fastPepperSubkeyArmor, fastPepperTicketArmor)
	}
	// Same inputs as the crypto KAT "aes256-subkeyarmor" (prf_test.go).
	subkey, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	armorSession, _ := hex.DecodeString("2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40")
	want, _ := hex.DecodeString("18b1ecfa39a1bf373bcbc098aae658fe35082a00d7d6660e6cb6fd6138404d21")

	got, etype, err := kerbcrypto.KRBFXCF2(
		subkey, messages.ETypeAES256CTSHMACSHA196,
		armorSession, messages.ETypeAES256CTSHMACSHA196,
		fastPepperSubkeyArmor, fastPepperTicketArmor)
	if err != nil {
		t.Fatalf("KRBFXCF2: %v", err)
	}
	if etype != messages.ETypeAES256CTSHMACSHA196 {
		t.Errorf("armor key etype = %d, want AES256", etype)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("armor key = %x, want %x", got, want)
	}
}

// TestBuildArmorAPReq confirms the FX_FAST_ARMOR_AP_REQUEST assembly: the AP-REQ
// carries the armor TGT verbatim, and its authenticator (decryptable with the
// armor session key at key usage 11) carries the subkey the KDC needs to
// reconstruct the armor key, keyed to the armor principal.
func TestBuildArmorAPReq(t *testing.T) {
	armorKey := bytes.Repeat([]byte{0x33}, 32)
	c := fastArmoredClient(t, armorKey, messages.ETypeAES256CTSHMACSHA196)

	subkey := bytes.Repeat([]byte{0x5c}, kerbcrypto.KeyLen(messages.ETypeAES256CTSHMACSHA196))
	apReqBytes, err := c.buildArmorAPReq(subkey)
	if err != nil {
		t.Fatalf("buildArmorAPReq: %v", err)
	}

	var apReq messages.APReq
	if _, err := apReq.Unmarshal(apReqBytes); err != nil {
		t.Fatalf("parse armor AP-REQ: %v", err)
	}
	if !bytes.Equal(apReq.TicketRaw, c.fast.ticketRaw) {
		t.Error("armor AP-REQ does not carry the armor TGT verbatim")
	}

	authBytes, err := kerbcrypto.Decrypt(messages.ETypeAES256CTSHMACSHA196, armorKey, kerbcrypto.KeyUsageAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		t.Fatalf("decrypt armor authenticator: %v", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(authBytes); err != nil {
		t.Fatalf("parse armor authenticator: %v", err)
	}
	if auth.SubKey == nil || !bytes.Equal(auth.SubKey.KeyValue, subkey) {
		t.Errorf("armor authenticator subkey mismatch: %+v", auth.SubKey)
	}
	if auth.CRealm != testRealm || len(auth.CName.NameString) != 1 || auth.CName.NameString[0] != "armor$" {
		t.Errorf("armor authenticator principal = %q/%+v, want %s/armor$", auth.CRealm, auth.CName, testRealm)
	}
}

// TestBuildArmorAPReqRejectsBadEType covers the error path: an unsupported armor
// session etype fails encryption of the authenticator.
func TestBuildArmorAPReqRejectsBadEType(t *testing.T) {
	c := fastArmoredClient(t, bytes.Repeat([]byte{0x33}, 32), 9999) // bogus etype
	if _, err := c.buildArmorAPReq(bytes.Repeat([]byte{0x5c}, 32)); err == nil {
		t.Fatal("expected buildArmorAPReq to fail for an unsupported armor etype")
	}
}

// TestFASTASReqBody verifies the inner AS-REQ body FAST wraps: it requests a TGT
// (krbtgt/REALM) for the client, carries the credential's etype list, and sets
// the forwardable/proxiable/renewable options a real AD client sends.
func TestFASTASReqBody(t *testing.T) {
	c := fastArmoredClient(t, bytes.Repeat([]byte{0x33}, 32), messages.ETypeAES256CTSHMACSHA196)
	body := c.asReqBody(0x2233)

	if len(body.CName.NameString) != 1 || body.CName.NameString[0] != "alice" {
		t.Errorf("cname = %+v, want alice", body.CName)
	}
	if body.Realm != testRealm {
		t.Errorf("realm = %q, want %q", body.Realm, testRealm)
	}
	if len(body.SName.NameString) != 2 || body.SName.NameString[0] != "krbtgt" || body.SName.NameString[1] != testRealm {
		t.Errorf("sname = %+v, want krbtgt/%s", body.SName, testRealm)
	}
	if body.Nonce != 0x2233 {
		t.Errorf("nonce = %d, want 0x2233", body.Nonce)
	}
	if len(body.EType) == 0 || body.EType[0] != messages.ETypeAES256CTSHMACSHA196 {
		t.Errorf("etype list = %v, want AES256 first", body.EType)
	}
	for _, bit := range []int{kdcOptionForwardable, kdcOptionProxiable, kdcOptionRenewable} {
		if body.KDCOptions.At(bit) != 1 {
			t.Errorf("AS-REQ option bit %d not set", bit)
		}
	}
}

// TestFASTEnabledAndSelfArmor covers the FAST configuration accessors:
// FASTEnabled tracks whether an armor TGT is set, and WithFASTArmorFromClient
// copies another client's TGT state as the armor (the self-armor pattern).
func TestFASTEnabledAndSelfArmor(t *testing.T) {
	plain := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	if plain.FASTEnabled() {
		t.Error("FASTEnabled should be false before an armor is configured")
	}

	armorClient := fakeTGTClient(t) // a client holding a usable TGT
	plain.WithFASTArmorFromClient(armorClient)
	if !plain.FASTEnabled() {
		t.Fatal("FASTEnabled should be true after WithFASTArmorFromClient")
	}
	if plain.fast.cname != armorClient.username || plain.fast.realm != armorClient.realm {
		t.Errorf("self-armor identity = %q@%q, want %q@%q", plain.fast.cname, plain.fast.realm, armorClient.username, armorClient.realm)
	}
	if !bytes.Equal(plain.fast.sessionKey, armorClient.sessionKey) || plain.fast.sessionEType != armorClient.sessionEType {
		t.Error("self-armor did not adopt the armor client's session key/etype")
	}
	if !bytes.Equal(plain.fast.ticketRaw, armorClient.tgtTicketRaw) {
		t.Error("self-armor did not adopt the armor client's raw TGT")
	}
}
