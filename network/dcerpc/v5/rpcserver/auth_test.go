package rpcserver

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// loopTransport hands every PDU the client sends to an association and queues
// whatever it answers for the client to read.
type loopTransport struct {
	association *Association
	replies     [][]byte
	// tamper, when set, is applied to each request PDU before the server sees it.
	tamper func([]byte)
}

func (l *loopTransport) Connect() error { return nil }
func (l *loopTransport) Send(p []byte) error {
	p = append([]byte(nil), p...)
	if l.tamper != nil && len(p) > 2 && pdu.PacketType(p[2]) == pdu.PacketTypeRequest {
		l.tamper(p)
	}
	reply, err := l.association.Handle(p)
	if err != nil {
		return err
	}
	if len(reply) > 0 {
		l.replies = append(l.replies, reply)
	}
	return nil
}
func (l *loopTransport) Recv() ([]byte, error) {
	if len(l.replies) == 0 {
		return nil, errors.New("nothing to read")
	}
	reply := l.replies[0]
	l.replies = l.replies[1:]
	return reply, nil
}
func (l *loopTransport) Close() error        { return nil }
func (l *loopTransport) MaxXmitFrag() uint16 { return 4280 }
func (l *loopTransport) MaxRecvFrag() uint16 { return 4280 }

// ntlmEndpoint serves service with NTLM-authenticated binds for alice.
func ntlmEndpoint(t *testing.T, service Service) *Dispatcher {
	t.Helper()
	info, err := targetinfo.BuildServerTargetInfo("SERVER", "LAB", "server.lab.local", "lab.local", nil)
	if err != nil {
		t.Fatal(err)
	}
	d := New(service)
	d.EnableNTLM(NTLMAuth{
		CredentialLookup: func(domain, username string) ([16]byte, bool) {
			if strings.EqualFold(username, "alice") {
				return nt.NTHash("Passw0rd!"), true
			}
			return [16]byte{}, false
		},
		TargetName: "LAB",
		TargetType: challenge.TargetTypeDomain,
		TargetInfo: info,
	})
	return d
}

// authenticatedClient binds an in-repo client to a new association of d at the
// given level with the given password.
func authenticatedClient(t *testing.T, d *Dispatcher, level uint8, password string) (*client.Client, *loopTransport, error) {
	t.Helper()
	transport := &loopTransport{association: d.Open()}
	c := client.NewClient(transport)
	creds, err := credentials.NewCredentials("LAB", "alice", password, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetAuth(pdu.AuthTypeNTLMSSP, level, creds); err != nil {
		t.Fatal(err)
	}
	return c, transport, c.Bind(testSyntax())
}

// TestNTLMBindAtEveryLevel binds with NTLM at the connect, packet, integrity
// and privacy levels and makes a call whose answer spans several fragments, so
// every response fragment is protected and verified by the client.
func TestNTLMBindAtEveryLevel(t *testing.T) {
	answer := bytes.Repeat([]byte("0123456789abcdef"), 1000) // 16 000 bytes, four fragments
	for _, level := range []uint8{pdu.AuthLevelConnect, pdu.AuthLevelPkt, pdu.AuthLevelPktIntegrity, pdu.AuthLevelPktPrivacy} {
		service := &stubService{answer: answer}
		c, _, err := authenticatedClient(t, ntlmEndpoint(t, service), level, "Passw0rd!")
		if err != nil {
			t.Fatalf("level %d: Bind() error = %v", level, err)
		}
		for call := 0; call < 3; call++ {
			stub, err := c.Call(0, []byte("request stub"))
			if err != nil {
				t.Fatalf("level %d call %d: Call() error = %v", level, call, err)
			}
			if !bytes.Equal(stub, answer) {
				t.Fatalf("level %d call %d: the answer came back as %d bytes", level, call, len(stub))
			}
			if got := string(service.lastStub); got != "request stub" {
				t.Fatalf("level %d call %d: the interface received %q", level, call, got)
			}
		}
	}
}

// TestNTLMBindWithTheWrongPasswordIsRefused checks the AUTH3 is verified: with
// a wrong password every request is faulted and never reaches the interface.
func TestNTLMBindWithTheWrongPasswordIsRefused(t *testing.T) {
	for _, level := range []uint8{pdu.AuthLevelConnect, pdu.AuthLevelPktIntegrity, pdu.AuthLevelPktPrivacy} {
		service := &stubService{answer: []byte("secret")}
		c, _, err := authenticatedClient(t, ntlmEndpoint(t, service), level, "wrong")
		if err != nil {
			continue // a refusal at bind time is as good
		}
		if _, err := c.Call(0, []byte("in")); err == nil {
			t.Errorf("level %d: a call succeeded after a failed authentication", level)
		}
		if service.callCount() != 0 {
			t.Errorf("level %d: the interface was called", level)
		}
	}
}

// TestNTLMTamperedRequestIsRefused flips one stub byte of each protected request
// in transit: the signature no longer verifies and the request is faulted.
func TestNTLMTamperedRequestIsRefused(t *testing.T) {
	for _, level := range []uint8{pdu.AuthLevelPktIntegrity, pdu.AuthLevelPktPrivacy} {
		service := &stubService{answer: []byte("x")}
		c, transport, err := authenticatedClient(t, ntlmEndpoint(t, service), level, "Passw0rd!")
		if err != nil {
			t.Fatalf("level %d: Bind() error = %v", level, err)
		}
		transport.tamper = func(p []byte) { p[pdu.HeaderSize+8] ^= 0x01 }
		_, err = c.Call(0, []byte("request stub"))
		var fault *pdu.Fault
		if !errors.As(err, &fault) || fault.Status != pdu.NCASFaultAccessDenied {
			t.Errorf("level %d: a tampered request: err = %v, want an access-denied fault", level, err)
		}
		if service.callCount() != 0 {
			t.Errorf("level %d: the interface was called with a tampered stub", level)
		}
	}
}

// TestUnprotectedRequestOnAProtectedAssociationIsRefused sends a request without
// a verifier on an association bound at the integrity level.
func TestUnprotectedRequestOnAProtectedAssociationIsRefused(t *testing.T) {
	service := &stubService{answer: []byte("x")}
	_, transport, err := authenticatedClient(t, ntlmEndpoint(t, service), pdu.AuthLevelPktIntegrity, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	reply, err := transport.association.Handle(callRequest(t, 9, 0, []byte("in")))
	if err != nil {
		t.Fatal(err)
	}
	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil || fault.Status != pdu.NCASFaultAccessDenied {
		t.Fatalf("an unprotected request was answered with %x, %v", reply, err)
	}
	if service.callCount() != 0 {
		t.Error("the interface was called")
	}
}

// TestAuthenticatedBindRefusedWithoutNTLM keeps the old behaviour for an
// endpoint that did not enable NTLM.
func TestAuthenticatedBindRefusedWithoutNTLM(t *testing.T) {
	transport := &loopTransport{association: New(&stubService{}).Open()}
	c := client.NewClient(transport)
	creds, _ := credentials.NewCredentials("LAB", "alice", "Passw0rd!", "")
	if err := c.SetAuth(pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, creds); err != nil {
		t.Fatal(err)
	}
	if err := c.Bind(testSyntax()); err == nil {
		t.Fatal("an authenticated bind succeeded on an endpoint without NTLM")
	}
}

// TestAuthenticatedAssociationIsSafeUnderConcurrentHandle replays a captured
// AUTH3 and signed requests through Handle from several goroutines at once.
// The stateful security context must never be used concurrently; run with
// -race.
func TestAuthenticatedAssociationIsSafeUnderConcurrentHandle(t *testing.T) {
	service := &stubService{answer: []byte("x")}
	c, transport, err := authenticatedClient(t, ntlmEndpoint(t, service), pdu.AuthLevelPktPrivacy, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	var captured []byte
	transport.tamper = func(p []byte) { captured = append([]byte(nil), p...) }
	if _, err := c.Call(0, []byte("in")); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			transport.association.Handle(captured) // a replay: faulted, but never racing
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
	if service.callCount() != 1 {
		t.Errorf("replayed requests reached the interface: %d calls", service.callCount())
	}
}

// TestUnauthenticatedRebindOfAProtectedAssociationIsRefused checks security
// cannot be dropped by binding again without it.
func TestUnauthenticatedRebindOfAProtectedAssociationIsRefused(t *testing.T) {
	service := &stubService{answer: []byte("x")}
	_, transport, err := authenticatedClient(t, ntlmEndpoint(t, service), pdu.AuthLevelPktPrivacy, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	reply, err := transport.association.Handle(bindRequest(t, 20, 4280, testSyntax()))
	if err != nil {
		t.Fatal(err)
	}
	nak := &pdu.BindNak{}
	if _, err := nak.Unmarshal(reply); err != nil {
		t.Fatalf("an unauthenticated rebind was answered with %x, %v; want a bind_nak", reply, err)
	}
	reply, _ = transport.association.Handle(callRequest(t, 21, 0, []byte("in")))
	fault := &pdu.Fault{}
	if _, err := fault.Unmarshal(reply); err != nil || service.callCount() != 0 {
		t.Fatalf("an unprotected request after the refused rebind was served: %x", reply)
	}
}

// TestCallLevelIsRefused checks RPC_C_AUTHN_LEVEL_CALL (3) is refused at bind
// instead of being served unprotected.
func TestCallLevelIsRefused(t *testing.T) {
	if _, err := rawNTLMBind(t, pdu.AuthLevelCall, 4280); err == nil {
		t.Fatal("a call-level bind succeeded")
	}
}

// TestSmallFragmentIsRaisedForProtectedReplies checks a bind asking for a
// fragment too small for a signed reply is answered with the size replies are
// really framed at, instead of announcing one the replies then exceed.
func TestSmallFragmentIsRaisedForProtectedReplies(t *testing.T) {
	ack, err := rawNTLMBind(t, pdu.AuthLevelPktIntegrity, 64)
	if err != nil {
		t.Fatal(err)
	}
	if ack.MaxXmitFrag != DefaultMaxFragment {
		t.Errorf("a 64-byte fragment request was answered with max_xmit_frag %d, want %d", ack.MaxXmitFrag, DefaultMaxFragment)
	}
}

// rawNTLMBind sends a hand-built NTLM bind at level, since the in-repo client
// promotes the call level to the packet level, and returns the bind_ack.
func rawNTLMBind(t *testing.T, level uint8, fragment uint16) (*pdu.BindAck, error) {
	t.Helper()
	association := ntlmEndpoint(t, &stubService{}).Open()
	negotiate := []byte("NTLMSSP\x00\x01\x00\x00\x00\x35\x82\x08\x60\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1),
		MaxXmitFrag: fragment,
		MaxRecvFrag: fragment,
		ContextList: []pdu.ContextElement{{ContextID: 0, AbstractSyntax: testSyntax(), TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()}}},
		SecTrailer:  pdu.SecTrailer{AuthType: pdu.AuthTypeNTLMSSP, AuthLevel: level, AuthContextID: 1},
		AuthValue:   negotiate,
	}
	encoded, err := bind.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	reply, err := association.Handle(encoded)
	if err != nil {
		return nil, err
	}
	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(reply); err != nil {
		return nil, fmt.Errorf("the bind was not acknowledged: %w", err)
	}
	return ack, nil
}
