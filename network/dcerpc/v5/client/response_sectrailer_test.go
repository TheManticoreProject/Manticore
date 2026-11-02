package client

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// TestUnprotectResponseRejectsAnotherSecurityContext checks a protected response
// whose sec_trailer names another auth_type, auth_level or auth_context_id than
// the bound context is refused before the provider sees its token: otherwise a
// response could be accepted under weaker protection than was negotiated.
func TestUnprotectResponseRejectsAnotherSecurityContext(t *testing.T) {
	const tokenLen = 16
	for _, tc := range []struct {
		name   string
		mutate func(trailer []byte)
	}{
		{"auth_type", func(tr []byte) { tr[0] = pdu.AuthTypeGSSKerberos }},
		{"lower auth_level", func(tr []byte) { tr[1] = pdu.AuthLevelPktIntegrity }},
		{"auth_context_id", func(tr []byte) { binary.LittleEndian.PutUint32(tr[4:8], 99) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeSecCtx{tokenLen: tokenLen}
			c := &Client{authType: pdu.AuthTypeNTLMSSP, authLevel: pdu.AuthLevelPktPrivacy, authContextID: 0, sec: fake}
			frag := buildResponseFrag(t, []byte("resp-stub-data"), pdu.AuthLevelPktPrivacy, bytes.Repeat([]byte{0xCD}, tokenLen))
			trailer := frag[len(frag)-tokenLen-pdu.SecTrailerSize : len(frag)-tokenLen]
			tc.mutate(trailer)

			if _, err := c.unprotectResponseStub(frag); err == nil {
				t.Fatal("a response under another security context was accepted")
			}
			if fake.unprotectCalled {
				t.Error("the provider was handed a response under another security context")
			}
		})
	}

	// The matching trailer still passes.
	fake := &fakeSecCtx{tokenLen: tokenLen}
	c := &Client{authType: pdu.AuthTypeNTLMSSP, authLevel: pdu.AuthLevelPktPrivacy, sec: fake}
	frag := buildResponseFrag(t, []byte("resp-stub-data"), pdu.AuthLevelPktPrivacy, bytes.Repeat([]byte{0xCD}, tokenLen))
	if _, err := c.unprotectResponseStub(frag); err != nil {
		t.Fatalf("the matching response was refused: %v", err)
	}
}

// TestUnprotectResponseAcceptsAKerberosStyleContext pins a context like the
// Kerberos client's (auth_type 9, auth_context_id 79231): a response echoing it
// is accepted.
func TestUnprotectResponseAcceptsAKerberosStyleContext(t *testing.T) {
	const tokenLen = 16
	fake := &fakeSecCtx{tokenLen: tokenLen}
	c := &Client{authType: pdu.AuthTypeGSSNegotiate, authLevel: pdu.AuthLevelPktPrivacy, authContextID: 79231, sec: fake}
	frag := buildResponseFrag(t, []byte("resp-stub-data"), pdu.AuthLevelPktPrivacy, bytes.Repeat([]byte{0xCD}, tokenLen))
	trailer := frag[len(frag)-tokenLen-pdu.SecTrailerSize : len(frag)-tokenLen]
	trailer[0] = pdu.AuthTypeGSSNegotiate
	binary.LittleEndian.PutUint32(trailer[4:8], 79231)
	if _, err := c.unprotectResponseStub(frag); err != nil {
		t.Fatalf("a response echoing the bound Kerberos-style context was refused: %v", err)
	}
}

// TestHandshakeTrailerMustMatchTheBind checks the bind_ack trailer rule.
func TestHandshakeTrailerMustMatchTheBind(t *testing.T) {
	c := &Client{authType: pdu.AuthTypeNTLMSSP, authContextID: 0}
	if err := c.checkHandshakeTrailer(&pdu.SecTrailer{AuthType: pdu.AuthTypeNTLMSSP}); err != nil {
		t.Errorf("the matching bind_ack trailer was refused: %v", err)
	}
	if err := c.checkHandshakeTrailer(&pdu.SecTrailer{AuthType: pdu.AuthTypeGSSKerberos}); err == nil {
		t.Error("a bind_ack trailer for another auth_type was accepted")
	}
	if err := c.checkHandshakeTrailer(&pdu.SecTrailer{AuthType: pdu.AuthTypeNTLMSSP, AuthContextID: 5}); err == nil {
		t.Error("a bind_ack trailer for another auth_context_id was accepted")
	}
}
