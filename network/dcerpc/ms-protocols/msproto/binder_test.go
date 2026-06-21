package msproto

import (
	"errors"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// compile-time assertions that both binders satisfy the Binder contract.
var (
	_ Binder = (*PipeBinder)(nil)
	_ Binder = (*TCPBinder)(nil)
)

// fakeDialer records the pipe it was asked to open and returns a canned result.
type fakeDialer struct {
	gotPipe string
	tr      dcerpctransport.Transport
	err     error
}

func (d *fakeDialer) RPCTransport(pipeName string) (dcerpctransport.Transport, error) {
	d.gotPipe = pipeName
	return d.tr, d.err
}

// TestPipeBinderDialError verifies that a dialer error is wrapped (not swallowed) and that
// the binder asks for the exact pipe it was constructed with, before any bind is attempted.
func TestPipeBinderDialError(t *testing.T) {
	dialErr := errors.New("boom")
	d := &fakeDialer{err: dialErr}
	b := NewPipeBinder(d, `\srvsvc`)

	rpc, closeFn, err := b.Bind(syntax.NDRTransferSyntax())
	if err == nil {
		t.Fatal("expected an error when the dialer fails")
	}
	if !errors.Is(err, dialErr) {
		t.Errorf("error %q does not wrap the dialer error", err)
	}
	if rpc != nil || closeFn != nil {
		t.Error("no client or closer should be returned on dial failure")
	}
	if d.gotPipe != `\srvsvc` {
		t.Errorf("dialed pipe = %q, want %q", d.gotPipe, `\srvsvc`)
	}
}

// TestNewTCPBinderDefaults verifies the constructor fills in NTLM packet-privacy auth and
// the default timeout, and preserves an explicitly supplied timeout and port.
func TestNewTCPBinderDefaults(t *testing.T) {
	b := NewTCPBinder("dc.example", 0, nil, 0)
	if b.AuthType != pdu.AuthTypeNTLMSSP {
		t.Errorf("AuthType = 0x%02x, want NTLMSSP 0x%02x", b.AuthType, pdu.AuthTypeNTLMSSP)
	}
	if b.AuthLevel != pdu.AuthLevelPktPrivacy {
		t.Errorf("AuthLevel = 0x%02x, want PktPrivacy 0x%02x", b.AuthLevel, pdu.AuthLevelPktPrivacy)
	}
	if b.Timeout != DefaultTCPTimeout {
		t.Errorf("Timeout = %v, want default %v", b.Timeout, DefaultTCPTimeout)
	}
	if b.Port != 0 {
		t.Errorf("Port = %d, want 0 (resolve via endpoint mapper)", b.Port)
	}

	custom := 25 * time.Second
	b2 := NewTCPBinder("dc.example", 9389, nil, custom)
	if b2.Timeout != custom {
		t.Errorf("Timeout = %v, want %v", b2.Timeout, custom)
	}
	if b2.Port != 9389 {
		t.Errorf("Port = %d, want 9389", b2.Port)
	}
}
