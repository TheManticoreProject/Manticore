package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

func TestConnectWithProfileToken(t *testing.T) {
	srv := operationServer()
	issuer := newClient(t, srv)
	token, err := issuer.GenerateProfileToken(context.Background(), credential(t, testPassword), signon.TokenRequest{Type: ibminet.TokenMultipleUse, Timeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer token.Destroy()
	if srv.Tokens() != 1 || token.Type() != ibminet.TokenMultipleUse || time.Until(token.ExpiresAt()) <= 0 {
		t.Fatalf("issued %d tokens, type %s, expires %s", srv.Tokens(), token.Type(), token.ExpiresAt())
	}

	c := newClient(t, srv)
	if err := c.Connect(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	if _, err := c.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	var tokenAuths int
	for _, r := range srv.Requests() {
		switch r.RequestID {
		case 0x7007:
			if r.CredentialType != 0x01 || len(r.Substitute) != 8 {
				t.Errorf("0x7007 must authenticate with a fresh password substitute: type 0x%02X, %d bytes", r.CredentialType, len(r.Substitute))
			}
		case 0x7004, 0x7002:
			if r.CredentialType == 0x02 && len(r.Substitute) == 32 {
				tokenAuths++
			}
		}
	}
	if tokenAuths != 2 {
		t.Errorf("%d token authentications, want Sign-on and Remote Command", tokenAuths)
	}
}

func TestConnectRejectsSingleUseTokenBeforeIO(t *testing.T) {
	srv := newServer()
	token, err := newClient(t, srv).GenerateProfileToken(context.Background(), credential(t, testPassword), signon.TokenRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if token.Type() != ibminet.TokenSingleUse {
		t.Fatalf("default token type = %s", token.Type())
	}
	before := len(srv.Requests())
	if err := newClient(t, srv).Connect(context.Background(), token); !errors.Is(err, ErrSingleUseToken) {
		t.Fatalf("Connect = %v", err)
	}
	if len(srv.Requests()) != before {
		t.Error("Connect sent requests with a single-use token")
	}
	// The unspent single-use token still authenticates one connection.
	ep := ibminet.Endpoint{Service: ibminet.ServiceSignon, Name: "as-signon", Host: "ibmi.test", Port: hosttest.SignonPort, Source: ibminet.EndpointExplicit}
	opts := signon.Options{Dialer: newClient(t, srv).cfg.dialer()}
	if _, err := signon.Validate(context.Background(), ep, token, opts); err != nil {
		t.Fatalf("single-use Validate = %v", err)
	}
	if _, err := signon.Validate(context.Background(), ep, token, opts); !errors.Is(err, ibminet.ErrTokenConsumed) {
		t.Errorf("second use = %v; it must fail locally before sending", err)
	}
}

func TestServerRejectsUnknownToken(t *testing.T) {
	srv := newServer()
	forged, _ := ibminet.NewProfileToken(make([]byte, 32), ibminet.TokenMultipleUse, time.Time{})
	err := newClient(t, srv).Connect(context.Background(), forged)
	var se *signon.Error
	if !errors.As(err, &se) || se.Reason() != signon.ReasonTokenFailure || !errors.Is(err, signon.ErrAuthenticationRejected) {
		t.Fatalf("Connect with an unknown token = %v", err)
	}
	if srv.Count(0x7004) != 1 || srv.Count(0x7002) != 0 {
		t.Errorf("requests = %+v", srv.Requests())
	}
}

func TestGenerateProfileTokenFailures(t *testing.T) {
	srv := newServer()
	c := newClient(t, srv)
	if _, err := c.GenerateProfileToken(context.Background(), credential(t, "WRONGPW1"), signon.TokenRequest{}); !errors.Is(err, signon.ErrAuthenticationRejected) {
		t.Errorf("wrong password = %v", err)
	}
	if n := srv.Count(0x7007); n != 1 {
		t.Errorf("0x7007 sent %d times; a rejected password is never retried", n)
	}
	srv.TokenRC = 0x00060002 // maximum number of tokens generated
	_, err := c.GenerateProfileToken(context.Background(), credential(t, testPassword), signon.TokenRequest{})
	var se *signon.Error
	if !errors.As(err, &se) || se.Code != 0x00060002 {
		t.Errorf("token store full = %v", err)
	}
	for name, req := range map[string]signon.TokenRequest{
		"timeout too long":  {Timeout: 2 * time.Hour},
		"fractional second": {Timeout: 1500 * time.Millisecond},
		"undefined type":    {Type: 9},
	} {
		before := len(srv.Requests())
		if _, err := c.GenerateProfileToken(context.Background(), credential(t, testPassword), req); err == nil || len(srv.Requests()) != before {
			t.Errorf("%s: %v after %d requests", name, err, len(srv.Requests())-before)
		}
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestTokenNeverAppearsInErrorsOrInfo(t *testing.T) {
	srv := newServer()
	c := newClient(t, srv)
	token, err := c.GenerateProfileToken(context.Background(), credential(t, testPassword), signon.TokenRequest{Type: ibminet.TokenMultipleUse})
	if err != nil {
		t.Fatal(err)
	}
	var issued []byte
	for _, r := range srv.Requests() {
		if r.RequestID == 0x7004 || r.RequestID == 0x7002 {
			issued = r.Substitute
		}
	}
	srv.StartRC = 0x00020002
	err = newClient(t, srv).Connect(context.Background(), token)
	if err == nil {
		t.Fatal("expected a start-server failure")
	}
	for _, r := range srv.Requests() {
		if r.RequestID == 0x7004 {
			issued = r.Substitute
		}
	}
	for _, out := range []string{err.Error(), fmt.Sprintf("%+v %#v %x", token, token, token)} {
		if len(issued) > 0 && strings.Contains(strings.ToLower(out), fmt.Sprintf("%x", issued)) {
			t.Errorf("output reveals the token: %s", out)
		}
	}
}
