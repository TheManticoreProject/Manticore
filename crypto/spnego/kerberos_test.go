package spnego

import (
	"bytes"
	"encoding/asn1"
	"testing"
)

// stubKerberos is a KerberosProvider test double.
type stubKerberos struct {
	initToken  []byte
	accepted   []byte
	sessionKey []byte
	acceptErr  error
}

func (s *stubKerberos) InitToken() ([]byte, error)         { return s.initToken, nil }
func (s *stubKerberos) AcceptResponseToken(t []byte) error { s.accepted = t; return s.acceptErr }
func (s *stubKerberos) SessionKey() []byte                 { return s.sessionKey }

func TestKerberosNegotiateTokenUsesKerberosMech(t *testing.T) {
	apReq := []byte("fake-gss-ap-req-token")
	ctx := NewAuthContext(AuthTypeKerberos, "CORP.LOCAL", "alice", "", "WS", true)
	ctx.Kerberos = &stubKerberos{initToken: apReq}

	tok, err := ctx.CreateNegotiateToken(0, nil)
	if err != nil {
		t.Fatalf("CreateNegotiateToken: %v", err)
	}

	// The SPNEGO NegTokenInit must advertise the Kerberos mechanism OID and carry
	// the provider's AP-REQ as the mechToken.
	oidDER, _ := asn1.Marshal(KerberosOID)
	if !bytes.Contains(tok, oidDER) {
		t.Errorf("negotiate token does not advertise the Kerberos mech OID")
	}
	if !bytes.Contains(tok, apReq) {
		t.Errorf("negotiate token does not carry the AP-REQ mechToken")
	}
}

func TestKerberosNegotiateRequiresProvider(t *testing.T) {
	ctx := NewAuthContext(AuthTypeKerberos, "R", "u", "", "WS", true)
	if _, err := ctx.CreateNegotiateToken(0, nil); err == nil {
		t.Error("expected error when no Kerberos provider is configured")
	}
}

func TestKerberosChallengeVerifiesAPRepAndCapturesKey(t *testing.T) {
	apRep := []byte("fake-ap-rep")
	key := []byte("session-key-bytes-32-............")
	stub := &stubKerberos{sessionKey: key}
	ctx := NewAuthContext(AuthTypeKerberos, "R", "u", "", "WS", true)
	ctx.Kerberos = stub

	// Build a server NegTokenResp advertising Kerberos with the AP-REP.
	resp := NegTokenResp{
		NegState:      NegStateAcceptCompleted,
		SupportedMech: KerberosOID,
		ResponseToken: apRep,
	}
	respSeq, err := resp.Marshal()
	if err != nil {
		t.Fatalf("marshal NegTokenResp: %v", err)
	}
	// The server delivers the NegTokenResp inside a [1]-tagged SecurityBlob.
	challengeToken, err := (&SecurityBlob{Data: respSeq}).Marshal()
	if err != nil {
		t.Fatalf("wrap SecurityBlob: %v", err)
	}

	out, err := ctx.CreateAuthenticateTokenFromChallengeToken(challengeToken)
	if err != nil {
		t.Fatalf("CreateAuthenticateTokenFromChallengeToken: %v", err)
	}
	if out != nil {
		t.Errorf("Kerberos challenge should produce no follow-up token, got %d bytes", len(out))
	}
	if !bytes.Equal(stub.accepted, apRep) {
		t.Errorf("provider did not receive the AP-REP response token")
	}
	if !bytes.Equal(ctx.GetSessionKey(), key) {
		t.Errorf("session key not captured from the Kerberos provider")
	}
}
