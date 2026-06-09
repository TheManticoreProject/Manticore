package client

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// fakeTransport is an in-memory transport that records the last sent message and
// returns a queued canned response, so the client can be exercised without a
// real server.
type fakeTransport struct {
	connected bool
	sent      [][]byte
	responses [][]byte
}

func (f *fakeTransport) Connect(net.IP, int) error { f.connected = true; return nil }
func (f *fakeTransport) Close() error              { f.connected = false; return nil }
func (f *fakeTransport) IsConnected() bool         { return f.connected }
func (f *fakeTransport) Send(data []byte) (int, error) {
	f.sent = append(f.sent, append([]byte{}, data...))
	return len(data), nil
}
func (f *fakeTransport) Receive() ([]byte, error) {
	resp := f.responses[0]
	f.responses = f.responses[1:]
	return resp, nil
}

// cannedNegotiateResponse builds the wire bytes of an SMB2 NEGOTIATE response.
func cannedNegotiateResponse(t *testing.T) []byte {
	t.Helper()
	resp := commands.NewNegotiateResponse()
	resp.DialectRevision = dialects.SMB2_DIALECT_2_0_2
	resp.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	resp.MaxTransactSize = 0x10000
	resp.MaxReadSize = 0x10000
	resp.MaxWriteSize = 0x10000
	resp.SystemTime = 0x01D0000000000000
	resp.SecurityBuffer = []byte{0x60, 0x40, 0x06, 0x06} // pretend SPNEGO token

	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.SetCommand(resp)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building canned response: %v", err)
	}
	return wire
}

func newTestClient(ft *fakeTransport) *Client {
	c := newClient(ft, net.ParseIP("127.0.0.1"), 445)
	ft.connected = true
	return c
}

func TestNewRequestSequencing(t *testing.T) {
	c := newTestClient(&fakeTransport{})

	m0 := c.newRequest(commands.NewNegotiateRequest())
	if m0.Header.MessageId != 0 {
		t.Errorf("first MessageId = %d, want 0", m0.Header.MessageId)
	}
	if m0.Header.CreditCharge != 0 {
		t.Errorf("CreditCharge = %d, want 0 (SMB 2.0.2)", m0.Header.CreditCharge)
	}
	if m0.Header.Credit != 1 {
		t.Errorf("CreditRequest = %d, want 1", m0.Header.Credit)
	}

	m1 := c.newRequest(commands.NewNegotiateRequest())
	if m1.Header.MessageId != 1 {
		t.Errorf("second MessageId = %d, want 1", m1.Header.MessageId)
	}
}

func TestNegotiateExchange(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponse(t)}}
	c := newTestClient(ft)

	if err := c.Negotiate(); err != nil {
		t.Fatalf("Negotiate: %v", err)
	}

	// Server properties must be captured from the response.
	if c.Connection.Dialect != dialects.SMB2_DIALECT_2_0_2 {
		t.Errorf("Dialect = %v, want SMB 2.0.2", c.Connection.Dialect)
	}
	if c.Connection.Server.MaxReadSize != 0x10000 {
		t.Errorf("MaxReadSize = 0x%x, want 0x10000", c.Connection.Server.MaxReadSize)
	}
	if !c.Connection.Server.SecurityMode.IsSigningEnabled() {
		t.Errorf("expected server signing-enabled to be captured")
	}
	if len(c.Connection.Server.SecurityBuffer) != 4 {
		t.Errorf("SecurityBuffer len = %d, want 4", len(c.Connection.Server.SecurityBuffer))
	}

	// NEGOTIATE must have used MessageId 0 and advanced the counter to 1.
	if c.Connection.MessageId != 1 {
		t.Errorf("MessageId after negotiate = %d, want 1", c.Connection.MessageId)
	}

	// The sent request must be a NEGOTIATE with SessionId 0.
	sentReq := message.NewMessage()
	if _, err := sentReq.Unmarshal(ft.sent[0]); err != nil {
		t.Fatalf("parsing sent request: %v", err)
	}
	if sentReq.Header.MessageId != 0 || sentReq.Header.SessionId != 0 {
		t.Errorf("sent NEGOTIATE MessageId/SessionId = %d/%d, want 0/0", sentReq.Header.MessageId, sentReq.Header.SessionId)
	}
}
