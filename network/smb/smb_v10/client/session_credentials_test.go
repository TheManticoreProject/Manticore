package client_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

const (
	statusMoreProcessingRequired = 0xC0000016
	statusLogonFailure           = 0xC000006D
)

// ntlmSetupServer is a scripted transport playing the server side of an
// extended-security (NTLMSSP) SESSION_SETUP_ANDX exchange. It answers the
// NEGOTIATE leg with a real CHALLENGE under uid, then answers the AUTHENTICATE
// leg with finalStatus, recording the user name the AUTHENTICATE carried.
type ntlmSetupServer struct {
	uid         uint16
	finalStatus uint32

	accept        *spnego.AcceptContext
	authenticated string
	pending       []byte
}

func (s *ntlmSetupServer) Connect(net.IP, int) error { return nil }
func (s *ntlmSetupServer) Close() error              { return nil }
func (s *ntlmSetupServer) IsConnected() bool         { return true }
func (s *ntlmSetupServer) SetTimeout(time.Duration)  {}

func (s *ntlmSetupServer) Send(data []byte) (int, error) {
	request := message.NewMessage()
	if err := request.Unmarshal(data); err != nil {
		return 0, err
	}
	setup, ok := request.Command.(*commands.SessionSetupAndxRequest)
	if !ok {
		return 0, fmt.Errorf("unexpected request command %T", request.Command)
	}

	response := commands.NewSessionSetupAndxResponse()
	status := s.finalStatus
	if s.accept == nil {
		s.accept = spnego.NewAcceptContext("LAB", challenge.TargetTypeDomain, nil)
		blob, err := s.accept.AcceptNegotiateToken([]byte(setup.SecurityBlob))
		if err != nil {
			return 0, err
		}
		response.SecurityBlob = []types.UCHAR(blob)
		response.SecurityBlobLength = types.USHORT(len(blob))
		status = statusMoreProcessingRequired
	} else {
		if err := s.accept.AcceptAuthenticateToken([]byte(setup.SecurityBlob)); err != nil {
			return 0, err
		}
		_, s.authenticated, _ = s.accept.Identity()
	}

	reply := message.NewMessage()
	reply.Header.SetFlags(flags.FLAGS_REPLY)
	reply.Header.UID = s.uid
	reply.Header.Status = status
	reply.AddCommand(response)
	raw, err := reply.Marshal()
	if err != nil {
		return 0, err
	}
	s.pending = raw
	return len(data), nil
}

func (s *ntlmSetupServer) Receive() ([]byte, error) {
	if s.pending == nil {
		return nil, fmt.Errorf("no reply queued")
	}
	raw := s.pending
	s.pending = nil
	return raw, nil
}

// ntlmSetupClient returns a client negotiated for user-level, extended-security
// authentication with the given server signing policy.
func ntlmSetupClient(signingState string) *client.Client {
	mode := securitymode.SecurityMode(securitymode.NEGOTIATE_USER_SECURITY | securitymode.NEGOTIATE_ENCRYPT_PASSWORDS)
	if signingState != client.SigningStateDisabled {
		mode |= securitymode.NEGOTIATE_SECURITY_SIGNATURES_ENABLED
	}
	return &client.Client{
		Connection: &client.Connection{
			Server: &client.Server{
				SecurityMode: mode,
				Capabilities: capabilities.CAP_UNICODE | capabilities.CAP_EXTENDED_SECURITY,
				SigningState: signingState,
			},
		},
	}
}

// TestSessionSetupNewCredentialsGetOwnSession verifies that a second SessionSetup
// with different credentials authenticates with those credentials and yields a
// session of its own, leaving the first session and its table entry intact.
func TestSessionSetupNewCredentialsGetOwnSession(t *testing.T) {
	c := ntlmSetupClient(client.SigningStateDisabled)

	c.Transport = &ntlmSetupServer{uid: 0x0A00}
	if err := c.SessionSetup(&credentials.Credentials{Username: "alice", Password: "pw"}); err != nil {
		t.Fatalf("first SessionSetup: %v", err)
	}
	alice := c.Session

	server := &ntlmSetupServer{uid: 0x0B00}
	c.Transport = server
	if err := c.SessionSetup(&credentials.Credentials{Username: "carol", Password: "pw"}); err != nil {
		t.Fatalf("second SessionSetup: %v", err)
	}

	if server.authenticated != "carol" {
		t.Errorf("second setup authenticated as %q, want %q", server.authenticated, "carol")
	}
	if c.Session == alice {
		t.Fatal("second setup reused the first session object")
	}
	if got := c.Session.Credentials.Username; got != "carol" {
		t.Errorf("current session carries credentials of %q, want %q", got, "carol")
	}
	if alice.SessionUID != 0x0A00 || alice.Credentials.Username != "alice" {
		t.Errorf("first session altered: UID 0x%04x, user %q", alice.SessionUID, alice.Credentials.Username)
	}
	if c.Connection.SessionTable[0x0A00] != alice {
		t.Error("UID 0x0A00 no longer maps to the first session")
	}
	if c.Connection.SessionTable[0x0B00] != c.Session {
		t.Error("UID 0x0B00 does not map to the second session")
	}
}

// TestSessionSetupFailureKeepsPreviousSession verifies that a rejected setup
// with different credentials neither replaces the current session nor alters the
// session already registered in the table.
func TestSessionSetupFailureKeepsPreviousSession(t *testing.T) {
	c := ntlmSetupClient(client.SigningStateDisabled)

	c.Transport = &ntlmSetupServer{uid: 0x0A00}
	if err := c.SessionSetup(&credentials.Credentials{Username: "alice", Password: "pw"}); err != nil {
		t.Fatalf("first SessionSetup: %v", err)
	}
	alice := c.Session

	server := &ntlmSetupServer{uid: 0x0B00, finalStatus: statusLogonFailure}
	c.Transport = server
	if err := c.SessionSetup(&credentials.Credentials{Username: "carol", Password: "bad"}); err == nil {
		t.Fatal("expected the rejected SessionSetup to fail")
	}

	if server.authenticated != "carol" {
		t.Errorf("rejected setup authenticated as %q, want %q", server.authenticated, "carol")
	}
	if c.Session != alice {
		t.Error("a failed setup replaced the current session")
	}
	if alice.SessionUID != 0x0A00 || alice.Credentials.Username != "alice" {
		t.Errorf("first session altered: UID 0x%04x, user %q", alice.SessionUID, alice.Credentials.Username)
	}
	if len(c.Connection.SessionTable) != 1 || c.Connection.SessionTable[0x0A00] != alice {
		t.Errorf("session table altered by a failed setup: %v", c.Connection.SessionTable)
	}
}

// TestSessionSetupFailureLeavesSigningInactive verifies that signing armed for
// the AUTHENTICATE leg is withdrawn when the server rejects the setup, so the
// connection does not go on signing with the key of a session that does not exist.
func TestSessionSetupFailureLeavesSigningInactive(t *testing.T) {
	c := ntlmSetupClient(client.SigningStateEnabled)
	c.Transport = &ntlmSetupServer{uid: 0x0C00, finalStatus: statusLogonFailure}

	if err := c.SessionSetup(&credentials.Credentials{Username: "alice", Password: "bad"}); err == nil {
		t.Fatal("expected the rejected SessionSetup to fail")
	}

	if c.Connection.IsSigningActive {
		t.Error("signing left active after a failed session setup")
	}
	if c.Connection.SigningSessionKey != nil {
		t.Error("signing key left set after a failed session setup")
	}
	if c.Session != nil {
		t.Error("a failed setup left a session as current")
	}
	if len(c.Connection.SessionTable) != 0 {
		t.Errorf("a failed setup registered a session: %v", c.Connection.SessionTable)
	}
}
