package gssapi

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func TestGSSChecksumValue(t *testing.T) {
	flags := uint32(GSSMutualFlag | GSSIntegFlag | GSSConfFlag)
	cksum := GSSChecksumValue(flags, nil)
	if len(cksum) != 24 {
		t.Fatalf("checksum length = %d, want 24", len(cksum))
	}
	if binary.LittleEndian.Uint32(cksum[0:]) != 16 {
		t.Errorf("Lgth field = %d, want 16", binary.LittleEndian.Uint32(cksum[0:]))
	}
	for _, b := range cksum[4:20] {
		if b != 0 {
			t.Errorf("Bnd should be zero for no channel bindings: % X", cksum[4:20])
			break
		}
	}
	if got := binary.LittleEndian.Uint32(cksum[20:]); got != flags {
		t.Errorf("Flags = 0x%x, want 0x%x", got, flags)
	}
	// With channel bindings the Bnd field is a nonzero MD5.
	cb := GSSChecksumValue(flags, []byte("chan-binding"))
	allZero := true
	for _, b := range cb[4:20] {
		if b != 0 {
			allZero = false
		}
	}
	if allZero {
		t.Error("Bnd should be a nonzero MD5 when channel bindings are supplied")
	}
}

func TestGSSChannelBindings(t *testing.T) {
	app := []byte("tls-server-end-point:hashbytes")
	cb := GSSChannelBindings(app)

	// 16 bytes of zeroed initiator/acceptor addrtype+length, then the LE
	// application-data length, then the application data itself.
	if len(cb) != 20+len(app) {
		t.Fatalf("length = %d, want %d", len(cb), 20+len(app))
	}
	for i, b := range cb[:16] {
		if b != 0 {
			t.Errorf("byte %d of address block should be zero, got 0x%02x", i, b)
		}
	}
	if got := binary.LittleEndian.Uint32(cb[16:20]); got != uint32(len(app)) {
		t.Errorf("application-data length = %d, want %d", got, len(app))
	}
	if !bytes.Equal(cb[20:], app) {
		t.Errorf("application data = %q, want %q", cb[20:], app)
	}

	// Empty bindings still carry the 16-byte address block and a zero length.
	if got := GSSChannelBindings(nil); len(got) != 20 {
		t.Errorf("empty bindings length = %d, want 20", len(got))
	}
}

func TestWrapUnwrapToken(t *testing.T) {
	msg := []byte("fake-krb-ap-req")
	tok, err := WrapToken(TokIDAPReq, msg)
	if err != nil {
		t.Fatal(err)
	}
	if tok[0] != 0x60 {
		t.Errorf("token should start with 0x60 (APPLICATION 0), got 0x%02x", tok[0])
	}
	id, got, err := UnwrapToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if id != TokIDAPReq {
		t.Errorf("TOK_ID = %02x %02x, want 01 00", id[0], id[1])
	}
	if !bytes.Equal(got, msg) {
		t.Errorf("inner message not preserved")
	}
	// A non-Kerberos OID must be rejected.
	bad := append([]byte(nil), tok...)
	// flip a byte in the OID region
	bad[5] ^= 0xff
	if _, _, err := UnwrapToken(bad); err == nil {
		t.Error("expected error for a corrupted mech OID")
	}
}

func fakeTicketAndKey(t *testing.T) ([]byte, []byte) {
	t.Helper()
	tkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   messages.PrincipalName{NameType: iana.NameTypeSRVInst, NameString: []string{"cifs", "dc01.corp.local"}},
		EncPart: messages.EncryptedData{EType: iana.ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0x5a}, 32)},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return raw, bytes.Repeat([]byte{0x11}, 32)
}

func TestInitSecContextProducesDecryptableAPReq(t *testing.T) {
	ticketRaw, sessionKey := fakeTicketAndKey(t)

	token, ctx, err := InitSecContext(InitOptions{
		TicketRaw:    ticketRaw,
		SessionKey:   sessionKey,
		SessionEType: iana.ETypeAES256CTSHMACSHA196,
		ClientName:   messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"alice"}},
		ClientRealm:  "CORP.LOCAL",
		Flags:        GSSIntegFlag | GSSConfFlag,
		Mutual:       true,
	})
	if err != nil {
		t.Fatalf("InitSecContext: %v", err)
	}

	id, apReqBytes, err := UnwrapToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if id != TokIDAPReq {
		t.Fatalf("TOK_ID = %02x %02x", id[0], id[1])
	}

	var apReq messages.APReq
	if _, err := apReq.Unmarshal(apReqBytes); err != nil {
		t.Fatalf("parse AP-REQ: %v", err)
	}
	// mutual-required (APOptions bit 2) must be set.
	if apReq.APOptions.Bytes[0]&0x20 == 0 {
		t.Errorf("mutual-required AP option not set: % X", apReq.APOptions.Bytes)
	}
	if !bytes.Equal(apReq.TicketRaw, ticketRaw) {
		t.Error("ticket not carried verbatim")
	}

	// Decrypt the authenticator with the session key at key usage 11.
	authBytes, err := kerbcrypto.Decrypt(iana.ETypeAES256CTSHMACSHA196, sessionKey, kerbcrypto.KeyUsageAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		t.Fatalf("decrypt authenticator: %v", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(authBytes); err != nil {
		t.Fatalf("parse authenticator: %v", err)
	}
	if auth.Cksum == nil || auth.Cksum.CKSumType != ChecksumTypeGSSAPI {
		t.Fatalf("authenticator checksum type = %v, want 0x8003", auth.Cksum)
	}
	if len(auth.Cksum.Checksum) != 24 {
		t.Fatalf("0x8003 checksum length = %d, want 24", len(auth.Cksum.Checksum))
	}
	gotFlags := binary.LittleEndian.Uint32(auth.Cksum.Checksum[20:])
	if gotFlags&GSSMutualFlag == 0 || gotFlags&GSSIntegFlag == 0 || gotFlags&GSSConfFlag == 0 {
		t.Errorf("checksum flags 0x%x missing expected bits", gotFlags)
	}
	if auth.SeqNumber != ctx.SeqNumber {
		t.Errorf("authenticator seq %d != ctx seq %d", auth.SeqNumber, ctx.SeqNumber)
	}
	if auth.CName.NameString[0] != "alice" {
		t.Errorf("cname wrong: %+v", auth.CName)
	}
}

// makeAPRepToken builds an AP-REP GSS token echoing ctime/cusec, encrypted under
// the session key, as an acceptor would for mutual authentication.
func makeAPRepToken(t *testing.T, sessionKey []byte, etype int, ctime time.Time, cusec int) []byte {
	t.Helper()
	enc := messages.EncAPRepPart{CTime: ctime, CUSec: cusec}
	encBytes, err := enc.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := kerbcrypto.Encrypt(etype, sessionKey, kerbcrypto.KeyUsageAPRepEncPart, encBytes)
	if err != nil {
		t.Fatal(err)
	}
	apRep := messages.APRep{PVNO: messages.KerberosV5, MsgType: messages.MsgTypeAPRep, EncPart: messages.EncryptedData{EType: etype, Cipher: cipher}}
	apRepBytes, err := apRep.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	token, err := WrapToken(TokIDAPRep, apRepBytes)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestAcceptAPRepMutualAuth(t *testing.T) {
	ticketRaw, sessionKey := fakeTicketAndKey(t)
	_, ctx, err := InitSecContext(InitOptions{
		TicketRaw:    ticketRaw,
		SessionKey:   sessionKey,
		SessionEType: iana.ETypeAES256CTSHMACSHA196,
		ClientName:   messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"alice"}},
		ClientRealm:  "CORP.LOCAL",
		Mutual:       true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Correct AP-REP verifies.
	good := makeAPRepToken(t, sessionKey, iana.ETypeAES256CTSHMACSHA196, ctx.ctime, ctx.cusec)
	if err := ctx.AcceptAPRep(good); err != nil {
		t.Errorf("AcceptAPRep rejected a valid AP-REP: %v", err)
	}

	// Wrong ctime fails (replay / mismatch).
	bad := makeAPRepToken(t, sessionKey, iana.ETypeAES256CTSHMACSHA196, ctx.ctime.Add(time.Second), ctx.cusec)
	if err := ctx.AcceptAPRep(bad); err == nil {
		t.Error("AcceptAPRep accepted an AP-REP with mismatched ctime")
	}

	// An AP-REQ token (wrong TOK_ID) is rejected.
	apreqTok, _ := WrapToken(TokIDAPReq, []byte("x"))
	if err := ctx.AcceptAPRep(apreqTok); err == nil {
		t.Error("AcceptAPRep accepted a non-AP-REP token")
	}
}
