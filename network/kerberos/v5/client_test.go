package kerberos

import (
	"bytes"
	"encoding/asn1"
	"encoding/hex"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// buildASRep encodes an AS-REP whose enc-part is an EncASRepPart (with the given
// nonce and session key) encrypted under key/etype at key-usage 3, so it can be
// fed to processASRep without a live KDC.
func buildASRep(t *testing.T, etype int, key []byte, nonce int, sessionKey []byte) []byte {
	t.Helper()
	enc := &messages.EncASRepPart{
		Key:      messages.EncryptionKey{KeyType: etype, KeyValue: sessionKey},
		Nonce:    nonce,
		Flags:    messages.NewKerberosFlags(messages.TicketFlagForwardable, messages.TicketFlagInitial),
		AuthTime: time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC),
		EndTime:  time.Date(2026, 7, 10, 20, 0, 0, 0, time.UTC),
		SRealm:   "CORP.LOCAL",
		SName:    messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
	}
	plain, err := enc.Marshal()
	if err != nil {
		t.Fatalf("EncASRepPart.Marshal: %v", err)
	}
	cipher, err := kerbcrypto.Encrypt(etype, key, kerbcrypto.KeyUsageASRepEncPart, plain)
	if err != nil {
		t.Fatalf("Encrypt enc-part: %v", err)
	}
	rep := &messages.ASRep{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeASRep,
		CRealm:  "CORP.LOCAL",
		CName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		Ticket: messages.Ticket{
			TktVno:  messages.KerberosV5,
			Realm:   "CORP.LOCAL",
			SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
			EncPart: messages.EncryptedData{EType: etype, Cipher: bytes.Repeat([]byte{0x5a}, 16)},
		},
		EncPart: messages.EncryptedData{EType: etype, Cipher: cipher},
	}
	wire, err := rep.Marshal()
	if err != nil {
		t.Fatalf("ASRep.Marshal: %v", err)
	}
	return wire
}

// TestProcessASRepSuccess drives processASRep with a hand-built AS-REP encrypted
// under a known AES256 key, confirming it decrypts the enc-part, accepts the
// matching nonce, and stores the TGT session key/etype on the client.
func TestProcessASRepSuccess(t *testing.T) {
	keyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	key, _ := hex.DecodeString(keyHex)
	etype := messages.ETypeAES256CTSHMACSHA196
	nonce := 0x33445566
	sessionKey := bytes.Repeat([]byte{0x42}, 32)

	c := NewClient("alice", "corp.local", "10.0.0.1")
	if err := c.WithAESKey(keyHex); err != nil {
		t.Fatalf("WithAESKey: %v", err)
	}

	wire := buildASRep(t, etype, key, nonce, sessionKey)
	if err := c.processASRep(wire, etype, "", nil, nonce); err != nil {
		t.Fatalf("processASRep: %v", err)
	}
	if !c.hasTGT {
		t.Fatal("hasTGT not set after processASRep")
	}
	if !bytes.Equal(c.sessionKey, sessionKey) {
		t.Errorf("session key = %X, want %X", c.sessionKey, sessionKey)
	}
	if c.sessionEType != etype {
		t.Errorf("session etype = %d, want %d", c.sessionEType, etype)
	}
	if c.tgtEnc.SRealm != "CORP.LOCAL" {
		t.Errorf("tgtEnc.SRealm = %q", c.tgtEnc.SRealm)
	}
}

// TestProcessASRepNonceMismatch confirms an AS-REP whose enc-part nonce differs
// from the request nonce is rejected (RFC 4120 3.1.3 replay defense).
func TestProcessASRepNonceMismatch(t *testing.T) {
	keyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	key, _ := hex.DecodeString(keyHex)
	etype := messages.ETypeAES256CTSHMACSHA196

	c := NewClient("alice", "corp.local", "10.0.0.1")
	if err := c.WithAESKey(keyHex); err != nil {
		t.Fatalf("WithAESKey: %v", err)
	}
	wire := buildASRep(t, etype, key, 111, bytes.Repeat([]byte{0x42}, 32))
	if err := c.processASRep(wire, etype, "", nil, 222); err == nil {
		t.Fatal("expected nonce-mismatch error, got nil")
	}
	if c.hasTGT {
		t.Error("hasTGT must not be set on a rejected AS-REP")
	}
}

// TestPickETypeFromError exercises the PREAUTH_REQUIRED etype/salt negotiation:
// with a password credential (AES256/AES128/RC4) the strongest advertised etype
// and its salt/s2kparams are chosen, whether the e-data is a SEQUENCE OF PA-DATA
// or bare ETYPE-INFO2.
func TestPickETypeFromError(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("Passw0rd!")

	info := messages.ETypeInfo2{
		{EType: messages.ETypeRC4HMAC},
		{EType: messages.ETypeAES256CTSHMACSHA196, Salt: "CORP.LOCALalice", S2KParams: []byte{0, 0, 0x10, 0}},
	}
	infoDER, err := info.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// e-data as a SEQUENCE OF PA-DATA carrying PA-ETYPE-INFO2.
	paList := []messages.PAData{{PADataType: messages.PAETypeInfo2, PADataValue: infoDER}}
	paDER, err := asn1.Marshal(paList)
	if err != nil {
		t.Fatal(err)
	}
	for _, edata := range [][]byte{paDER, infoDER} {
		etype, salt, s2k := c.pickETypeFromError(messages.KRBError{EData: edata})
		if etype != messages.ETypeAES256CTSHMACSHA196 {
			t.Errorf("etype = %d, want AES256", etype)
		}
		if salt != "CORP.LOCALalice" {
			t.Errorf("salt = %q, want CORP.LOCALalice", salt)
		}
		if !bytes.Equal(s2k, []byte{0, 0, 0x10, 0}) {
			t.Errorf("s2kparams = %X", s2k)
		}
	}

	// No e-data: fall back to the credential's strongest etype and default salt.
	etype, salt, _ := c.pickETypeFromError(messages.KRBError{})
	if etype != c.cred.SupportedETypes()[0] || salt != c.cred.DefaultSalt() {
		t.Errorf("fallback etype/salt = %d/%q", etype, salt)
	}
}

// TestPickBestETypeHonoursCredential confirms an NT-hash credential (RC4-only)
// never selects an advertised AES etype it cannot key.
func TestPickBestETypeHonoursCredential(t *testing.T) {
	info := messages.ETypeInfo2{
		{EType: messages.ETypeAES256CTSHMACSHA196, Salt: "x"},
		{EType: messages.ETypeRC4HMAC},
	}
	etype, _, _ := pickBestEType(info, []int{messages.ETypeRC4HMAC}, "default")
	if etype != messages.ETypeRC4HMAC {
		t.Errorf("etype = %d, want RC4 (credential cannot do AES)", etype)
	}
}

// TestKDCOptionsEncoding verifies the KDCOptions bit positions produced for AS-REQ
// and TGS-REQ match the documented flags (RFC 4120 5.4.1 / RFC 6806).
func TestKDCOptionsEncoding(t *testing.T) {
	as := kdcOptionsForASReq()
	for _, bit := range []int{kdcOptionForwardable, kdcOptionProxiable, kdcOptionRenewable} {
		if as.At(bit) != 1 {
			t.Errorf("AS-REQ option bit %d not set", bit)
		}
	}
	if as.At(kdcOptionCanonicalize) == 1 {
		t.Error("AS-REQ must not set canonicalize")
	}

	tgs := kdcOptionsForTGSReq()
	for _, bit := range []int{kdcOptionForwardable, kdcOptionRenewable, kdcOptionCanonicalize, kdcOptionRenewableOK} {
		if tgs.At(bit) != 1 {
			t.Errorf("TGS-REQ option bit %d not set", bit)
		}
	}

	// encodeKDCOptions must place bit N in byte N/8 at position 7-(N%8).
	bs := encodeKDCOptions(kdcOptionEncTktInSKey) // bit 28 -> byte 3, 0x08
	if !bytes.Equal(bs.Bytes, []byte{0x00, 0x00, 0x00, 0x08}) {
		t.Errorf("encodeKDCOptions(28) = % X, want 00 00 00 08", bs.Bytes)
	}
}

// TestParseSPN covers the SPN parsing accepted by GetTGS and the S4U/silver paths.
func TestParseSPN(t *testing.T) {
	ok := []struct {
		in      string
		service string
		host    string
	}{
		{"cifs/dc01.corp.local", "cifs", "dc01.corp.local"},
		{"ldap/dc01.corp.local@CORP.LOCAL", "ldap", "dc01.corp.local"},
		{"http/web", "http", "web"},
	}
	for _, tc := range ok {
		pn, err := parseSPN(tc.in, "CORP.LOCAL")
		if err != nil {
			t.Errorf("parseSPN(%q): %v", tc.in, err)
			continue
		}
		if pn.NameType != messages.NameTypeSRVInst || pn.NameString[0] != tc.service || pn.NameString[1] != tc.host {
			t.Errorf("parseSPN(%q) = %+v", tc.in, pn)
		}
	}
	for _, bad := range []string{"noslash", "cifs/", "/host", ""} {
		if _, err := parseSPN(bad, "CORP.LOCAL"); err == nil {
			t.Errorf("parseSPN(%q): expected error", bad)
		}
	}
}
