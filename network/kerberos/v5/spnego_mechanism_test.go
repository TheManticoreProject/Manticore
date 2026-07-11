package kerberos

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// silverArmedClient forges a silver service ticket for spn (signed with an RC4
// service key) and preloads it into a fresh client, so the SPNEGO mechanism can
// build an AP-REQ with neither a TGT nor a KDC — exactly the silver-ticket
// pass-the-ticket path. It returns the client and the ticket session key/etype
// the acceptor would key against.
func silverArmedClient(t *testing.T, spn string) (*KerberosClient, []byte, int) {
	t.Helper()
	sessionKey := bytes.Repeat([]byte{0x42}, 32)
	ft, err := ForgeSilver(ForgeOptions{
		Realm:           testRealm,
		Username:        "Administrator",
		DomainSID:       testDomainSID,
		UserRID:         500,
		LogonDomainName: "CORP",
		Key:             bytes.Repeat([]byte{0xCD}, 16), // service RC4 NT hash
		KeyEType:        messages.ETypeRC4HMAC,
		SessionKey:      sessionKey,
		SessionEType:    messages.ETypeAES256CTSHMACSHA196,
	}, spn)
	if err != nil {
		t.Fatalf("ForgeSilver: %v", err)
	}
	c := NewClient("Administrator", "corp.local", "") // no KDC host: must not be contacted
	if err := c.LoadForgedServiceTicket(ft); err != nil {
		t.Fatalf("LoadForgedServiceTicket: %v", err)
	}
	return c, ft.SessionKey, ft.SessionEType
}

// TestSPNEGOInitTokenBuildsAPReq drives InitToken over a preloaded silver ticket
// and confirms the emitted SPNEGO mechToken is a well-formed KRB_AP_REQ: the
// GSS-wrapped AP-REQ carries the ticket verbatim, mutual-required is set, and the
// authenticator (decryptable with the ticket session key at key usage 11) holds
// the 0x8003 GSS checksum with the mutual flag and names the client.
func TestSPNEGOInitTokenBuildsAPReq(t *testing.T) {
	const spn = "cifs/host.corp.local"
	c, sessionKey, sessionEType := silverArmedClient(t, spn)

	m := NewSPNEGOMechanism(c, spn)
	token, err := m.InitToken()
	if err != nil {
		t.Fatalf("InitToken: %v", err)
	}

	tokID, apReqBytes, err := gssapi.UnwrapToken(token)
	if err != nil {
		t.Fatalf("UnwrapToken: %v", err)
	}
	if tokID != gssapi.TokIDAPReq {
		t.Fatalf("TOK_ID = %02x %02x, want AP-REQ (01 00)", tokID[0], tokID[1])
	}

	var apReq messages.APReq
	if _, err := apReq.Unmarshal(apReqBytes); err != nil {
		t.Fatalf("parse AP-REQ: %v", err)
	}
	// mutual-required (AP-options bit 2 -> byte 0, 0x20) must be set.
	if apReq.APOptions.Bytes[0]&0x20 == 0 {
		t.Errorf("mutual-required AP option not set: % X", apReq.APOptions.Bytes)
	}
	if apReq.Authenticator.EType != sessionEType {
		t.Errorf("authenticator etype = %d, want session etype %d", apReq.Authenticator.EType, sessionEType)
	}

	authBytes, err := kerbcrypto.Decrypt(sessionEType, sessionKey, kerbcrypto.KeyUsageAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		t.Fatalf("decrypt authenticator: %v", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(authBytes); err != nil {
		t.Fatalf("parse authenticator: %v", err)
	}
	if auth.Cksum == nil || auth.Cksum.CKSumType != gssapi.ChecksumTypeGSSAPI {
		t.Fatalf("authenticator checksum type = %v, want 0x8003", auth.Cksum)
	}
	if len(auth.Cksum.Checksum) != 24 {
		t.Fatalf("0x8003 checksum length = %d, want 24", len(auth.Cksum.Checksum))
	}
	if flags := binary.LittleEndian.Uint32(auth.Cksum.Checksum[20:]); flags&gssapi.GSSMutualFlag == 0 {
		t.Errorf("checksum flags 0x%x missing mutual bit", flags)
	}
	if len(auth.CName.NameString) != 1 || auth.CName.NameString[0] != "Administrator" {
		t.Errorf("authenticator cname = %+v, want Administrator", auth.CName)
	}
}

// TestSPNEGOSessionKeyReturnsSubkey confirms SessionKey exposes the negotiated
// initiator subkey after InitToken (used to sign/seal SMB and RPC traffic), and
// returns nil before a context exists.
func TestSPNEGOSessionKeyReturnsSubkey(t *testing.T) {
	const spn = "cifs/host.corp.local"
	c, _, sessionEType := silverArmedClient(t, spn)

	m := NewSPNEGOMechanism(c, spn)
	if m.SessionKey() != nil {
		t.Error("SessionKey should be nil before InitToken establishes a context")
	}
	if _, err := m.InitToken(); err != nil {
		t.Fatalf("InitToken: %v", err)
	}
	key := m.SessionKey()
	if len(key) != kerbcrypto.KeyLen(sessionEType) {
		t.Fatalf("SessionKey length = %d, want %d", len(key), kerbcrypto.KeyLen(sessionEType))
	}
	// InitToken asserts an initiator subkey, so SessionKey returns it (not the
	// bare ticket session key).
	if m.ctx == nil || !bytes.Equal(key, m.ctx.SubKey) {
		t.Error("SessionKey should return the negotiated context subkey")
	}
}

// TestSPNEGOInitTokenNoCredential covers the error path: with no TGT, no
// preloaded ticket, and no credential, InitToken fails while acquiring a TGT
// rather than contacting the network.
func TestSPNEGOInitTokenNoCredential(t *testing.T) {
	c := NewClient("bob", "corp.local", "") // no credential, no ticket
	m := NewSPNEGOMechanism(c, "cifs/host.corp.local")
	if _, err := m.InitToken(); err == nil {
		t.Fatal("expected InitToken to fail without a credential or preloaded ticket")
	}
}

// TestSPNEGOAcceptResponseToken exercises the AcceptResponseToken paths that do
// not require a live acceptor: an empty token is a no-op, a token with no
// established context is rejected, and a GSS-wrapped KRB-ERROR is surfaced with
// its error code.
func TestSPNEGOAcceptResponseToken(t *testing.T) {
	const spn = "cifs/host.corp.local"

	// Empty token is a no-op even before a context exists.
	m := NewSPNEGOMechanism(NewClient("bob", "corp.local", ""), spn)
	if err := m.AcceptResponseToken(nil); err != nil {
		t.Errorf("empty token should be a no-op, got %v", err)
	}

	// A non-empty token with no established context is rejected.
	if err := m.AcceptResponseToken([]byte{0x01, 0x02}); err == nil {
		t.Error("expected error with no established context")
	}

	// With a context established, a GSS-wrapped KRB-ERROR must be surfaced.
	c, _, _ := silverArmedClient(t, spn)
	m2 := NewSPNEGOMechanism(c, spn)
	if _, err := m2.InitToken(); err != nil {
		t.Fatalf("InitToken: %v", err)
	}
	krbErr := &messages.KRBError{
		STime:     time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC),
		ErrorCode: messages.ErrSkew,
		Realm:     testRealm,
		SName:     messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}},
		EText:     "clock skew too great",
	}
	krbErrBytes, err := krbErr.Marshal()
	if err != nil {
		t.Fatalf("KRBError.Marshal: %v", err)
	}
	errToken, err := gssapi.WrapToken(gssapi.TokIDError, krbErrBytes)
	if err != nil {
		t.Fatalf("WrapToken: %v", err)
	}
	if err := m2.AcceptResponseToken(errToken); err == nil {
		t.Error("expected AcceptResponseToken to surface a server KRB-ERROR")
	}
}
