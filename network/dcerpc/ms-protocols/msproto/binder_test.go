package msproto

import (
	"errors"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
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

// nakTransport is a transport whose peer rejects every bind with a bind_nak. It records
// whether it was connected and whether it was closed, so tests can detect a leaked handle.
type nakTransport struct {
	t         *testing.T
	connected bool
	closed    bool
}

func (f *nakTransport) Connect() error    { f.connected = true; return nil }
func (f *nakTransport) Send([]byte) error { return nil }
func (f *nakTransport) Recv() ([]byte, error) {
	nak := &pdu.BindNak{Header: pdu.NewHeader(pdu.PacketTypeBindNak, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1)}
	b, err := nak.Marshal()
	if err != nil {
		f.t.Fatalf("marshal bind_nak: %v", err)
	}
	return b, nil
}
func (f *nakTransport) Close() error        { f.closed = true; return nil }
func (f *nakTransport) MaxXmitFrag() uint16 { return 4280 }
func (f *nakTransport) MaxRecvFrag() uint16 { return 4280 }

var _ dcerpctransport.Transport = (*nakTransport)(nil)

// stubTCPTransport replaces newTCPTransport for the duration of the test, handing out a
// fresh nakTransport per dial and recording each dialed port alongside its transport.
func stubTCPTransport(t *testing.T) map[int]*nakTransport {
	t.Helper()
	dialed := map[int]*nakTransport{}
	orig := newTCPTransport
	newTCPTransport = func(host string, port int, timeout time.Duration) dcerpctransport.Transport {
		tr := &nakTransport{t: t}
		dialed[port] = tr
		return tr
	}
	t.Cleanup(func() { newTCPTransport = orig })
	return dialed
}

// TestPipeBinderClosesTransportOnBindFailure verifies that a rejected bind closes the named
// pipe the binder opened instead of leaking its handle.
func TestPipeBinderClosesTransportOnBindFailure(t *testing.T) {
	tr := &nakTransport{t: t}
	b := NewPipeBinder(&fakeDialer{tr: tr}, `\srvsvc`)

	rpc, closeFn, err := b.Bind(syntax.NDRTransferSyntax())
	if err == nil {
		t.Fatal("expected an error when the server rejects the bind")
	}
	if rpc != nil || closeFn != nil {
		t.Error("no client or closer should be returned on bind failure")
	}
	if !tr.connected {
		t.Fatal("transport was never connected; the test does not exercise the leak")
	}
	if !tr.closed {
		t.Error("named-pipe transport left open after a failed bind")
	}
}

// TestTCPBinderClosesTransportOnBindFailure verifies that a rejected bind on an explicit
// port closes the TCP connection the binder opened.
func TestTCPBinderClosesTransportOnBindFailure(t *testing.T) {
	dialed := stubTCPTransport(t)
	creds, err := credentials.NewCredentials("DOMAIN", "user", "password", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	b := NewTCPBinder("dc.example", 49667, creds, 0)

	rpc, closeFn, err := b.Bind(syntax.NDRTransferSyntax())
	if err == nil {
		t.Fatal("expected an error when the server rejects the bind")
	}
	if rpc != nil || closeFn != nil {
		t.Error("no client or closer should be returned on bind failure")
	}
	tr := dialed[49667]
	if tr == nil || !tr.connected {
		t.Fatal("target transport was never connected; the test does not exercise the leak")
	}
	if !tr.closed {
		t.Error("TCP transport left open after a failed bind")
	}
}

// TestTCPBinderClosesEndpointMapperOnBindFailure verifies that a rejected endpoint-mapper
// bind during port resolution closes the TCP/135 connection.
func TestTCPBinderClosesEndpointMapperOnBindFailure(t *testing.T) {
	dialed := stubTCPTransport(t)
	b := NewTCPBinder("dc.example", 0, nil, 0)

	_, _, err := b.Bind(syntax.NDRTransferSyntax())
	if err == nil {
		t.Fatal("expected an error when the endpoint mapper rejects the bind")
	}
	if errors.Unwrap(err) == nil {
		t.Errorf("error %q does not wrap the endpoint-mapper failure", err)
	}
	tr := dialed[tcp.EndpointMapperPort]
	if tr == nil || !tr.connected {
		t.Fatal("endpoint-mapper transport was never connected; the test does not exercise the leak")
	}
	if !tr.closed {
		t.Error("endpoint-mapper transport left open after a failed bind")
	}
}
