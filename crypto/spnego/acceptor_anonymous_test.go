package spnego

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

// TestAcceptorAcceptsMinimalAnonymousAuthenticate drives an anonymous AUTHENTICATE
// with no domain and no workstation -- 73 bytes on the wire, with Version but no
// MIC -- into the acceptor. It used to be refused as "data too short", because
// the parser demanded the 88 bytes of a fixed part carrying a MIC.
func TestAcceptorAcceptsMinimalAnonymousAuthenticate(t *testing.T) {
	ctx := NewAcceptContext("LAB", challenge.TargetTypeDomain, serverTargetInfo(t, nil))

	client := NewAuthContext(AuthTypeNTLM, testDomain, testUsername, testPassword, testWorkstation, true)
	v := version.DefaultVersion()
	negotiateToken, err := client.CreateNegotiateToken(clientNegotiateFlags, &v)
	if err != nil {
		t.Fatalf("client CreateNegotiateToken() error = %v", err)
	}
	if _, err := ctx.AcceptNegotiateToken(negotiateToken); err != nil {
		t.Fatalf("AcceptNegotiateToken() error = %v", err)
	}

	msg, err := authenticate.CreateAnonymousAuthenticateMessage(ctx.Challenge, "", "")
	if err != nil {
		t.Fatalf("CreateAnonymousAuthenticateMessage() error = %v", err)
	}
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if len(raw) >= 88 {
		t.Fatalf("anonymous AUTHENTICATE is %d bytes; the test needs one shorter than 88", len(raw))
	}

	if err := ctx.AcceptAuthenticateToken(raw); err != nil {
		t.Fatalf("AcceptAuthenticateToken() of a %d-byte anonymous AUTHENTICATE error = %v", len(raw), err)
	}
	if !ctx.Authenticate.IsAnonymous() {
		t.Error("accepted AUTHENTICATE lost NTLMSSP_NEGOTIATE_ANONYMOUS")
	}
	if ctx.Authenticate.NeedsMIC {
		t.Error("NeedsMIC set on an anonymous AUTHENTICATE")
	}
	if domain, username, workstation := ctx.Identity(); domain != "" || username != "" || workstation != "" {
		t.Errorf("Identity() = %q, %q, %q; want all empty", domain, username, workstation)
	}
}
