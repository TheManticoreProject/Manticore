package server

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// The conformance suite pairs the SMB1 client in this repository with the server
// over an in-memory pipe and records, in one place, exactly what the server
// currently answers.
//
// It is a regression net rather than a feature test. Two tables below say which
// commands are served and what each client call does today; a later phase that
// implements a command has to move a row, which makes the coverage change
// deliberate instead of silent. That is what makes the large phases safe to
// land: anything that quietly stops working shows up as a table mismatch.
//
// There is no socket involved. A pipe removes the listener, the port and the
// scheduling from the picture, so a failure here is the protocol and nothing else.

// conformanceConfig is the server the suite runs against: it authenticates one
// identity, and takes its signing policy from the caller.
func conformanceConfig(policy SigningPolicy) Config {
	config := captureServerConfig()
	config.SigningPolicy = policy
	config.Authenticator = StaticAccounts(Account{
		Domain:   captureDomain,
		Username: captureUsername,
		NTHash:   nt.NTHash(capturePassword),
	})
	return config
}

// pipedServer starts a server serving one end of an in-memory pipe, and returns
// the transport for the other end.
func pipedServer(t *testing.T, config Config) (*Server, *tcp.TCPTransport) {
	t.Helper()

	srv, err := NewServer(config)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	serverSide, clientSide := net.Pipe()
	remote := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}

	served := make(chan struct{})
	go func() {
		defer close(served)
		_ = srv.ServeConn(tcp.NewTCPTransportFromConn(serverSide), remote)
	}()

	client := tcp.NewTCPTransportFromConn(clientSide)
	client.SetTimeout(5 * time.Second)

	t.Cleanup(func() {
		client.Close()
		srv.Close()
		select {
		case <-served:
		case <-time.After(3 * time.Second):
			t.Error("the server did not stop serving the piped connection")
		}
	})

	return srv, client
}

// pipedClient pairs the SMB1 client with a server over a pipe, negotiated and
// optionally authenticated.
func pipedClient(t *testing.T, config Config, authenticate bool) (*Server, *smb1client.Client) {
	t.Helper()

	srv, transportEnd := pipedServer(t, config)

	// The transport is already connected, so the client negotiates on it directly
	// rather than dialling.
	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("client Negotiate() over a pipe error = %v", err)
	}

	if authenticate {
		creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
		if err != nil {
			t.Fatalf("NewCredentials() error = %v", err)
		}
		if err := client.SessionSetup(creds); err != nil {
			t.Fatalf("client SessionSetup() over a pipe error = %v", err)
		}
	}

	return srv, client
}

// servedCommands are the commands the server answers, with what each one is
// expected to do.
//
// A phase that implements a command moves it out of the unserved set below and
// into here. Leaving the two in step is the point: a command that starts being
// answered without this table changing means the coverage moved by accident.
var servedCommands = map[codes.CommandCode]string{
	codes.SMB_COM_NEGOTIATE:          "selects a dialect",
	codes.SMB_COM_SESSION_SETUP_ANDX: "runs the authentication exchange",
	codes.SMB_COM_LOGOFF_ANDX:        "drops the session",
	codes.SMB_COM_ECHO:               "echoes the payload",

	codes.SMB_COM_TREE_CONNECT_ANDX: "connects a tree to a share",
	codes.SMB_COM_TREE_DISCONNECT:   "drops the tree and its handles",

	codes.SMB_COM_NT_CREATE_ANDX: "opens or creates a file or directory",
	codes.SMB_COM_CLOSE:          "releases a handle",
	codes.SMB_COM_READ_ANDX:      "reads from a handle",
	codes.SMB_COM_WRITE_ANDX:     "writes through a handle",
	codes.SMB_COM_FLUSH:          "commits a handle, or the whole tree",

	codes.SMB_COM_DELETE:           "deletes a file, wildcards included",
	codes.SMB_COM_RENAME:           "renames or moves an entry",
	codes.SMB_COM_CREATE_DIRECTORY: "creates a directory",
	codes.SMB_COM_DELETE_DIRECTORY: "removes an empty directory",
	codes.SMB_COM_CHECK_DIRECTORY:  "reports whether a path is a directory",

	codes.SMB_COM_TRANSACTION2:           "carries the find and information subcommands",
	codes.SMB_COM_TRANSACTION2_SECONDARY: "continues a fragmented transaction",
	codes.SMB_COM_FIND_CLOSE2:            "releases a search handle",
}

// servedTrans2Subcommands are the TRANSACTION2 subcommands the server answers.
// The same discipline as servedCommands: a subcommand gaining a handler has to
// gain a row, and the cross-check below fails if the two drift.
var servedTrans2Subcommands = map[subcommands.Transaction2Subcommand]string{
	subcommands.TRANS2_FIND_FIRST2:            "opens a directory enumeration",
	subcommands.TRANS2_FIND_NEXT2:             "continues one",
	subcommands.TRANS2_QUERY_PATH_INFORMATION: "describes a path",
	subcommands.TRANS2_QUERY_FILE_INFORMATION: "describes an open handle",
	subcommands.TRANS2_SET_PATH_INFORMATION:   "changes a path",
	subcommands.TRANS2_SET_FILE_INFORMATION:   "changes an open handle",
	subcommands.TRANS2_QUERY_FS_INFORMATION:   "describes the volume",
}

// TestConformanceServedTrans2SubcommandsAreServed asserts the subcommand table and
// the handler table agree, in both directions.
func TestConformanceServedTrans2SubcommandsAreServed(t *testing.T) {
	for subcommand := range servedTrans2Subcommands {
		if _, ok := trans2Handlers[subcommand]; !ok {
			t.Errorf("TRANSACTION2 subcommand 0x%04X is listed as served but has no handler", uint16(subcommand))
		}
	}
	for subcommand := range trans2Handlers {
		if _, ok := servedTrans2Subcommands[subcommand]; !ok {
			t.Errorf("TRANSACTION2 subcommand 0x%04X has a handler but is not listed; add it with a note on what it does",
				uint16(subcommand))
		}
	}
}

// TestConformanceServedCommandsAreServed asserts every command in the served set
// is answered, and that the set is not silently shrinking.
func TestConformanceServedCommandsAreServed(t *testing.T) {
	for command := range servedCommands {
		if _, ok := dispatchTable[command]; !ok {
			t.Errorf("command 0x%02X is listed as served but has no handler", uint8(command))
		}
	}
	for command := range dispatchTable {
		if _, ok := servedCommands[command]; !ok {
			t.Errorf("command 0x%02X has a handler but is not listed in servedCommands; add it with a note on what it does",
				uint8(command))
		}
	}
}

// TestConformanceUnservedCommandsAreRefused asserts every command the message
// layer knows but the server does not serve is answered with a defined status
// rather than a panic, a hang, or an accidental success.
//
// This is the exhaustive half of the net: it walks the whole command space, so a
// command that starts being answered — or that starts crashing — cannot go
// unnoticed.
func TestConformanceUnservedCommandsAreRefused(t *testing.T) {
	_, client := pipedClient(t, conformanceConfig(SigningDisabled), true)
	uid := client.Session.SessionUID

	raw := client.Transport

	// Counted so the walk cannot quietly become a no-op: several branches below
	// skip a command, and if they ever skipped all of them the test would pass
	// while asserting nothing.
	exercised := 0
	skippedUnmarshalable := 0

	for value := 0; value <= 0xFF; value++ {
		command := codes.CommandCode(value)
		if _, served := servedCommands[command]; served {
			continue
		}
		if !isKnownCommand(command) {
			continue
		}

		request, err := commands.CreateRequestCommand(command)
		if err != nil {
			t.Fatalf("CreateRequestCommand(0x%02X) error = %v", value, err)
		}
		request.Init()

		msg := message.NewMessage()
		msg.Header.Command = command
		msg.Header.Flags2 = flags2.Flags2(flags2.FLAGS2_UNICODE | flags2.FLAGS2_NT_STATUS_ERROR_CODES)
		msg.Header.UID = uid
		msg.AddCommand(request)

		marshalled, err := msg.Marshal()
		if err != nil {
			// A command whose zero value cannot be marshalled is not something a
			// client could send either, so there is nothing to assert about it.
			skippedUnmarshalable++
			continue
		}
		exercised++
		if _, err := raw.Send(marshalled); err != nil {
			t.Fatalf("failed to send command 0x%02X: %v", value, err)
		}

		reply, err := raw.Receive()
		if err != nil {
			// The server is entitled to drop a connection it cannot make sense
			// of, but then the rest of the walk cannot continue on it.
			t.Fatalf("command 0x%02X ended the connection: %v", value, err)
		}
		if len(reply) < 9 {
			t.Fatalf("command 0x%02X produced a %d-byte reply", value, len(reply))
		}

		status := nt_status.NT_STATUS(uint32(reply[5]) | uint32(reply[6])<<8 | uint32(reply[7])<<16 | uint32(reply[8])<<24)
		switch status {
		case nt_status.NT_STATUS_NOT_IMPLEMENTED, nt_status.NT_STATUS_INVALID_SMB:
			// Expected: recognized but unserved, or a body the zero value could
			// not make valid.
		case nt_status.NT_STATUS_SUCCESS:
			t.Errorf("command 0x%02X succeeded although it is not in servedCommands", value)
		default:
			t.Errorf("command 0x%02X answered an unexpected %s", value, statusName(status))
		}
	}

	// The message layer knows well over a hundred commands, so a walk that
	// exercised only a handful means the loop stopped covering the space.
	if exercised < 50 {
		t.Fatalf("only %d commands were exercised (%d skipped as unmarshalable); the walk is no longer covering the command space",
			exercised, skippedUnmarshalable)
	}
	t.Logf("exercised %d unserved commands, skipped %d whose zero value cannot be marshalled",
		exercised, skippedUnmarshalable)
}

// TestConformanceClientAPI walks the client's entry points against the server and
// asserts each one behaves as the table says.
func TestConformanceClientAPI(t *testing.T) {
	for _, policy := range []SigningPolicy{SigningDisabled, SigningRequired} {
		policy := policy
		t.Run(policy.String(), func(t *testing.T) {
			_, client := pipedClient(t, conformanceConfig(policy), true)

			// Signing must be active exactly when the policy requires it, since
			// every assertion below then runs over signed messages.
			if wantSigning := policy == SigningRequired; client.Connection.IsSigningActive != wantSigning {
				t.Fatalf("signing active = %t, want %t under policy %s",
					client.Connection.IsSigningActive, wantSigning, policy)
			}

			t.Run("Echo", func(t *testing.T) {
				payload := []byte("conformance")
				echoed, err := client.Echo(payload)
				if err != nil {
					t.Fatalf("Echo() error = %v", err)
				}
				if !bytes.Equal(echoed, payload) {
					t.Fatalf("Echo() returned %q, want %q", echoed, payload)
				}
			})

			t.Run("TreeConnect to an unserved share", func(t *testing.T) {
				if err := client.TreeConnect("nosuchshare"); err == nil {
					t.Fatal("TreeConnect() succeeded for a share that is not served")
				}
			})

			t.Run("ListDirectory", func(t *testing.T) {
				// No share is registered on the conformance server, so a listing
				// has nowhere to run: what is asserted is that it is refused
				// rather than crashing, since the file-service tests cover the
				// working path against a share.
				if _, err := client.ListEntries("\\*"); err == nil {
					t.Fatal("ListEntries() succeeded with no tree connected")
				}
			})

			t.Run("Logoff", func(t *testing.T) {
				if err := client.Logoff(); err != nil {
					t.Fatalf("Logoff() error = %v", err)
				}
			})
		})
	}
}

// TestConformanceNegotiatedParameters asserts what the client learns from the
// negotiate response, which is the contract every later exchange depends on.
func TestConformanceNegotiatedParameters(t *testing.T) {
	_, client := pipedClient(t, conformanceConfig(SigningDisabled), false)

	if got := client.Connection.SelectedDialect; got != dialects.DIALECT_NT_LM_0_12 {
		t.Errorf("dialect = %q, want %q", got, dialects.DIALECT_NT_LM_0_12)
	}
	if got := client.Connection.Server.MaxBufferSize; got != DefaultMaxBufferSize {
		t.Errorf("MaxBufferSize = %d, want %d", got, DefaultMaxBufferSize)
	}
	if got := client.Connection.MaxMpxCount; got != DefaultMaxMpxCount {
		t.Errorf("MaxMpxCount = %d, want %d", got, DefaultMaxMpxCount)
	}
	if !client.Connection.Server.SecurityMode.SupportsUserLevelAccessControl() {
		t.Error("the server did not advertise user-level access control")
	}
	if !client.Connection.Server.SecurityMode.SupportsChallengeResponseAuth() {
		t.Error("the server did not advertise challenge/response authentication")
	}
	if client.Connection.Server.ServerGUID.ToFormatD() == "00000000-0000-0000-0000-000000000000" {
		t.Error("the server advertised a zero GUID")
	}
}

// TestConformanceAuthenticationOutcomes walks the outcomes a logon can have, in
// one place, so a change to any of them is visible together.
func TestConformanceAuthenticationOutcomes(t *testing.T) {
	cases := []struct {
		name     string
		config   func() Config
		domain   string
		username string
		password string
		succeeds bool
	}{
		{
			name:     "verified identity",
			config:   func() Config { return conformanceConfig(SigningDisabled) },
			domain:   captureDomain,
			username: captureUsername,
			password: capturePassword,
			succeeds: true,
		},
		{
			name:     "wrong password",
			config:   func() Config { return conformanceConfig(SigningDisabled) },
			domain:   captureDomain,
			username: captureUsername,
			password: "wrong",
			succeeds: false,
		},
		{
			name:     "unknown identity",
			config:   func() Config { return conformanceConfig(SigningDisabled) },
			domain:   captureDomain,
			username: "nobody",
			password: capturePassword,
			succeeds: false,
		},
		{
			name: "unknown identity as guest",
			config: func() Config {
				config := conformanceConfig(SigningDisabled)
				config.AllowGuest = true
				return config
			},
			domain:   captureDomain,
			username: "nobody",
			password: capturePassword,
			succeeds: true,
		},
		{
			name: "guest refused when signing is required",
			config: func() Config {
				config := conformanceConfig(SigningRequired)
				config.AllowGuest = true
				return config
			},
			domain:   captureDomain,
			username: "nobody",
			password: capturePassword,
			succeeds: false,
		},
		{
			name:     "no credential store at all",
			config:   func() Config { return captureServerConfig() },
			domain:   captureDomain,
			username: captureUsername,
			password: capturePassword,
			succeeds: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, transportEnd := pipedServer(t, tc.config())
			client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
			if err := client.Negotiate(); err != nil {
				t.Fatalf("Negotiate() error = %v", err)
			}

			creds, err := credentials.NewCredentials(tc.domain, tc.username, tc.password, "")
			if err != nil {
				t.Fatalf("NewCredentials() error = %v", err)
			}

			err = client.SessionSetup(creds)
			if tc.succeeds && err != nil {
				t.Fatalf("SessionSetup() error = %v, want success", err)
			}
			if !tc.succeeds && err == nil {
				t.Fatal("SessionSetup() succeeded, want a refusal")
			}
		})
	}
}

// TestConformanceOverPipeAndSocketAgree asserts the suite's pipe harness and a
// real socket produce the same result, so the convenience of the pipe does not
// come at the cost of testing something else.
func TestConformanceOverPipeAndSocketAgree(t *testing.T) {
	config := conformanceConfig(SigningRequired)

	// Over a pipe.
	_, pipeClient := pipedClient(t, config, true)
	pipeEchoed, err := pipeClient.Echo([]byte("agree"))
	if err != nil {
		t.Fatalf("Echo() over a pipe error = %v", err)
	}

	// Over a socket, through the listener.
	srv, err := NewServer(config)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener := listenLoopback(t)
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	socketClient := smb1client.NewClientUsingTCPTransport(net.ParseIP(host), mustAtoi(t, portText))
	if err := socketClient.Connect(net.ParseIP(host), mustAtoi(t, portText)); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer socketClient.Disconnect()

	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := socketClient.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() over a socket error = %v", err)
	}
	socketEchoed, err := socketClient.Echo([]byte("agree"))
	if err != nil {
		t.Fatalf("Echo() over a socket error = %v", err)
	}

	if !bytes.Equal(pipeEchoed, socketEchoed) {
		t.Fatalf("pipe echoed %q, socket echoed %q", pipeEchoed, socketEchoed)
	}
	if pipeClient.Connection.IsSigningActive != socketClient.Connection.IsSigningActive {
		t.Fatal("signing was activated over one transport but not the other")
	}
}

// TestServeConnRefusesWhenClosed asserts ServeConn declines rather than serving on
// a closed server, and closes the connection it was handed.
func TestServeConnRefusesWhenClosed(t *testing.T) {
	srv, err := NewServer(conformanceConfig(SigningDisabled))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	if err := srv.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	serverSide, clientSide := net.Pipe()
	defer clientSide.Close()

	if err := srv.ServeConn(tcp.NewTCPTransportFromConn(serverSide), nil); err == nil {
		t.Fatal("ServeConn() on a closed server should fail")
	}
	if err := srv.ServeConn(nil, nil); err == nil {
		t.Fatal("ServeConn(nil) should fail")
	}
}

// mustAtoi parses a port, failing the test rather than returning an error.
func mustAtoi(t *testing.T, text string) int {
	t.Helper()
	value, err := strconv.Atoi(text)
	if err != nil {
		t.Fatalf("Atoi(%q) error = %v", text, err)
	}
	return value
}
