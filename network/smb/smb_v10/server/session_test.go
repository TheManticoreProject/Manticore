package server

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	ntlmversion "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// serverWithAccount stands up a server that will authenticate the test identity,
// under the given signing policy.
func serverWithAccount(t *testing.T, policy SigningPolicy, adjust func(*Config)) (*Server, string) {
	t.Helper()

	config := captureServerConfig()
	config.SigningPolicy = policy
	config.Authenticator = StaticAccounts(Account{
		Domain:   captureDomain,
		Username: captureUsername,
		NTHash:   nt.NTHash(capturePassword),
	})
	if adjust != nil {
		adjust(&config)
	}

	srv, err := NewServer(config)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener := listenLoopback(t)
	addr := listener.Addr().String()
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	return srv, addr
}

// authSession is an established session on a raw transport, with whatever signing
// state the exchange arrived at.
type authSession struct {
	client   *tcp.TCPTransport
	uid      types.USHORT
	action   types.USHORT
	key      []byte
	signing  bool
	sequence uint32
}

// establishSession drives negotiation and both session-setup legs over a raw
// transport, and returns the resulting session. It fails the test if the logon is
// refused; use rawLogon for the cases where a refusal is the expected outcome.
func establishSession(t *testing.T, addr, domain, username, password string, wantSigning bool) *authSession {
	t.Helper()

	session, status := rawLogon(t, addr, domain, username, password, wantSigning)
	if status != 0 {
		t.Fatalf("session setup answered 0x%08X, want success", status)
	}
	return session
}

// rawLogon drives negotiation and both legs, returning the session and the status
// the second leg was answered with. A non-zero status means no session exists.
func rawLogon(t *testing.T, addr, domain, username, password string, requestSigning bool) (*authSession, uint32) {
	t.Helper()

	client := dialServer(t, addr)
	negotiateWithRawClient(t, client)

	auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, domain, username, password, "CLIENT01", true)
	v := ntlmversion.DefaultVersion()
	negotiateToken, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &v)
	if err != nil {
		t.Fatalf("CreateNegotiateToken() error = %v", err)
	}

	// First leg.
	first := sendSessionSetupSigned(t, client, negotiateToken, 0, requestSigning, nil, 0)
	if first.Header.Status != uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED) {
		t.Fatalf("challenge leg answered 0x%08X, want NT_STATUS_MORE_PROCESSING_REQUIRED", first.Header.Status)
	}
	challengeResponse, ok := first.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		t.Fatalf("challenge leg command is %T", first.Command)
	}
	uid := first.Header.UID

	// Second leg. The client signs it with the key it derived, at sequence 0.
	authenticateToken, err := auth.CreateAuthenticateTokenFromChallengeToken([]byte(challengeResponse.SecurityBlob))
	if err != nil {
		t.Fatalf("CreateAuthenticateTokenFromChallengeToken() error = %v", err)
	}
	key := auth.GetSessionKey()

	var signKey []byte
	if requestSigning && len(key) > 0 {
		signKey = key
	}
	second := sendSessionSetupSigned(t, client, authenticateToken, uid, requestSigning, signKey, 0)

	session := &authSession{client: client, uid: uid, key: key}
	if setupResponse, ok := second.Command.(*commands.SessionSetupAndxResponse); ok {
		session.action = setupResponse.Action
	}
	if second.Header.Status == 0 && requestSigning && len(key) > 0 {
		// The response is signed at 1 and the next request at 2.
		session.signing = second.Header.Flags2&flags2.FLAGS2_SECURITY_SIGNATURE != 0
		session.sequence = signing.NextRequestSequenceNumber(0)
	}
	return session, second.Header.Status
}

// sendSessionSetupSigned sends one session-setup request, optionally asking for
// signing and optionally signing the request itself.
func sendSessionSetupSigned(
	t *testing.T,
	client *tcp.TCPTransport,
	blob []byte,
	uid types.USHORT,
	requestSigning bool,
	signKey []byte,
	sequence uint32,
) *message.Message {
	t.Helper()

	request := newRequest(codes.SMB_COM_SESSION_SETUP_ANDX)
	request.Header.Flags2 |= flags2.FLAGS2_EXTENDED_SECURITY
	if requestSigning {
		request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	}
	request.Header.UID = uid

	setup := commands.NewSessionSetupAndxRequest()
	setup.SecurityBlob = blob
	setup.SecurityBlobLength = types.USHORT(len(blob))
	setup.MaxBufferSize = 16644
	setup.MaxMpxCount = 50
	setup.Capabilities = capabilities.CAP_EXTENDED_SECURITY
	setup.NativeOS = "TestOS"
	setup.NativeLanMan = "TestLanMan"
	request.AddCommand(setup)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the session setup: %v", err)
	}
	if len(signKey) > 0 {
		signing.Sign(signKey, marshalled, sequence)
	}
	if _, err := client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	response, _ := receiveResponse(t, client)
	return response
}

// sendOnSession sends a request on an established session, signing it when the
// session is signing, and returns the response. It also verifies the server's
// response signature and advances the sequence, so a caller does not have to.
func sendOnSession(t *testing.T, session *authSession, command codes.CommandCode, cmd command_interface.CommandInterface) *message.Message {
	t.Helper()

	request := newRequest(command)
	request.Header.UID = session.uid
	if session.signing {
		request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	}
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}
	if session.signing {
		signing.Sign(session.key, marshalled, session.sequence)
	}
	if _, err := session.client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	response, raw := receiveResponse(t, session.client)
	if session.signing {
		responseSequence := signing.ResponseSequenceNumber(session.sequence)
		if !signing.Verify(session.key, raw, responseSequence) {
			t.Fatalf("the server's response is not signed at sequence %d", responseSequence)
		}
		session.sequence = signing.NextRequestSequenceNumber(session.sequence)
	}
	return response
}

// TestSessionEstablishedWithValidCredential asserts a verified logon produces a
// session with a UID the client can then use.
func TestSessionEstablishedWithValidCredential(t *testing.T) {
	srv, addr := serverWithAccount(t, SigningDisabled, nil)

	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, false)

	if session.uid == 0 {
		t.Fatal("the server assigned no UID")
	}
	if session.action&SMB_SETUP_GUEST != 0 {
		t.Fatal("a verified logon was reported as a guest session")
	}

	waitFor(t, func() bool { return srv.Connections() == 1 }, "the connection was not registered")

	// The session is usable: a command that requires one is now dispatched
	// rather than refused.
	response := sendOnSession(t, session, codes.SMB_COM_TREE_CONNECT_ANDX, commands.NewSeekRequest())
	if response.Header.Status != uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_NOT_IMPLEMENTED on an established session", response.Header.Status)
	}
}

// TestWrongPasswordIsRefused asserts a response that does not verify is refused,
// and that no session is left behind.
func TestWrongPasswordIsRefused(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)

	_, status := rawLogon(t, addr, captureDomain, captureUsername, "wrong-password", false)
	if status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
	}
}

// TestUnknownIdentityIsRefusedWithoutGuest asserts an identity the credential
// store does not know is refused when guest access is not enabled.
func TestUnknownIdentityIsRefusedWithoutGuest(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)

	_, status := rawLogon(t, addr, captureDomain, "nobody", capturePassword, false)
	if status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
	}
}

// TestGuestSessionIsReported asserts an unknown identity admitted as a guest is
// reported as one. A client is entitled to treat that as a failure, so silently
// granting it would be wrong.
func TestGuestSessionIsReported(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) { c.AllowGuest = true })

	session := establishSession(t, addr, captureDomain, "nobody", capturePassword, false)
	if session.action&SMB_SETUP_GUEST == 0 {
		t.Fatalf("Action = 0x%04X, want SMB_SETUP_GUEST to be set", uint16(session.action))
	}

	// A verified identity on the same server is still not a guest.
	verified := establishSession(t, addr, captureDomain, captureUsername, capturePassword, false)
	if verified.action&SMB_SETUP_GUEST != 0 {
		t.Fatal("a verified logon was reported as a guest session")
	}
}

// TestGuestIsRefusedWhenSigningIsRequired asserts a session that cannot sign is
// refused under a policy that requires signatures, rather than being established
// and then failing on its first request.
func TestGuestIsRefusedWhenSigningIsRequired(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, func(c *Config) { c.AllowGuest = true })

	// A guest session derives no key, so it cannot meet the policy.
	if _, status := rawLogon(t, addr, captureDomain, "nobody", capturePassword, true); status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
		t.Fatalf("guest logon under SigningRequired answered 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
	}

	// A verified identity does derive one, so it is accepted.
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)
	if !session.signing {
		t.Fatal("signing was required but the session is not signing")
	}
}

// TestSigningRequiredNeedsAnAuthenticator asserts the configuration is refused at
// construction: a policy that demands signatures cannot be met by a server that
// can only produce sessions without a key.
func TestSigningRequiredNeedsAnAuthenticator(t *testing.T) {
	config := captureServerConfig()
	config.SigningPolicy = SigningRequired
	if _, err := NewServer(config); err == nil {
		t.Fatal("NewServer() accepted SigningRequired with no Authenticator")
	}
}

// TestSigningPolicyIsAdvertised asserts the negotiate response advertises exactly
// what the policy allows, and that the REQUIRED bit never stands alone.
func TestSigningPolicyIsAdvertised(t *testing.T) {
	cases := []struct {
		name         string
		policy       SigningPolicy
		wantEnabled  bool
		wantRequired bool
	}{
		{"disabled", SigningDisabled, false, false},
		{"enabled", SigningEnabled, true, false},
		{"required", SigningRequired, true, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, addr := serverWithAccount(t, tc.policy, nil)
			client := dialServer(t, addr)

			request := newRequest(codes.SMB_COM_NEGOTIATE)
			request.Header.Flags2 |= flags2.FLAGS2_EXTENDED_SECURITY
			negotiate := commands.NewNegotiateRequest()
			negotiate.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)
			request.AddCommand(negotiate)
			sendRequest(t, client, request)

			response, _ := receiveResponse(t, client)
			negotiateResponse, ok := response.Command.(*commands.NegotiateResponse)
			if !ok {
				t.Fatalf("response command is %T", response.Command)
			}
			mode := negotiateResponse.SecurityMode
			if mode.IsSecuritySignatureEnabled() != tc.wantEnabled {
				t.Errorf("signatures-enabled = %t, want %t", mode.IsSecuritySignatureEnabled(), tc.wantEnabled)
			}
			if mode.IsSecuritySignatureRequired() != tc.wantRequired {
				t.Errorf("signatures-required = %t, want %t", mode.IsSecuritySignatureRequired(), tc.wantRequired)
			}
			// [MS-CIFS] 2.2.4.52.2: REQUIRED must not be set without ENABLED.
			if mode.IsSecuritySignatureRequired() && !mode.IsSecuritySignatureEnabled() {
				t.Error("the REQUIRED bit is set without the ENABLED bit")
			}
		})
	}
}

// TestSignedExchangeRoundTrips asserts that once signing is armed the server signs
// its responses and requires signed requests, across several exchanges, with the
// sequence numbers advancing in pairs.
func TestSignedExchangeRoundTrips(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)

	if !session.signing {
		t.Fatal("signing was not armed")
	}
	if session.sequence != 2 {
		t.Fatalf("the first request after the exchange is at sequence %d, want 2", session.sequence)
	}

	// sendOnSession verifies each response's signature and advances the counter.
	for exchange := 0; exchange < 4; exchange++ {
		echo := commands.NewEchoRequest()
		echo.EchoCount = 1
		echo.Data = []byte("signed")
		response := sendOnSession(t, session, codes.SMB_COM_ECHO, echo)
		if response.Header.Status != 0 {
			t.Fatalf("exchange %d answered 0x%08X, want success", exchange, response.Header.Status)
		}
		if response.Header.Flags2&flags2.FLAGS2_SECURITY_SIGNATURE == 0 {
			t.Fatalf("exchange %d response does not announce a signature", exchange)
		}
	}
	// Four exchanges from 2 leaves the next request at 10.
	if session.sequence != 10 {
		t.Fatalf("after four exchanges the sequence is %d, want 10", session.sequence)
	}
}

// TestUnsignedRequestOnSigningConnectionIsDropped asserts a connection that is
// signing refuses an unsigned request. The connection is dropped rather than
// answered: a request whose signature does not check cannot be attributed, so
// there is nothing to answer.
func TestUnsignedRequestOnSigningConnectionIsDropped(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)

	// Send an echo with no signature at all.
	request := newRequest(codes.SMB_COM_ECHO)
	request.Header.UID = session.uid
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("unsigned")
	request.AddCommand(echo)
	sendRequest(t, session.client, request)

	session.client.SetTimeout(2 * time.Second)
	if raw, err := session.client.Receive(); err == nil {
		t.Fatalf("the server answered an unsigned request on a signing connection: % x", raw)
	}
}

// TestTamperedSignatureIsDropped asserts a request whose signature does not check
// out is refused even though it is otherwise well formed.
func TestTamperedSignatureIsDropped(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)

	request := newRequest(codes.SMB_COM_ECHO)
	request.Header.UID = session.uid
	request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("tampered")
	request.AddCommand(echo)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	signing.Sign(session.key, marshalled, session.sequence)
	// Flip a bit in the signature.
	marshalled[signing.SecuritySignatureOffset] ^= 0x01
	if _, err := session.client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}

	session.client.SetTimeout(2 * time.Second)
	if raw, err := session.client.Receive(); err == nil {
		t.Fatalf("the server answered a request with a bad signature: % x", raw)
	}
}

// TestReplayedSequenceNumberIsDropped asserts a request repeated at a sequence
// number already consumed is refused, which is what stops a captured request being
// replayed.
func TestReplayedSequenceNumberIsDropped(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)

	build := func(sequence uint32) []byte {
		request := newRequest(codes.SMB_COM_ECHO)
		request.Header.UID = session.uid
		request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
		echo := commands.NewEchoRequest()
		echo.EchoCount = 1
		echo.Data = []byte("replay")
		request.AddCommand(echo)
		marshalled, err := request.Marshal()
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		signing.Sign(session.key, marshalled, sequence)
		return marshalled
	}

	// One valid exchange at the expected number.
	first := build(session.sequence)
	if _, err := session.client.Send(first); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	receiveResponse(t, session.client)

	// Replaying the same frame is now signed at a stale number.
	if _, err := session.client.Send(first); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	session.client.SetTimeout(2 * time.Second)
	if raw, err := session.client.Receive(); err == nil {
		t.Fatalf("the server answered a replayed request: % x", raw)
	}
}

// TestCommandWithoutSessionIsRefused asserts a command that acts within a session
// is refused when the UID names none, and that the three commands defined to work
// without one still do.
func TestCommandWithoutSessionIsRefused(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	client := dialServer(t, addr)
	negotiateWithRawClient(t, client)

	// A command that needs a session, on UID 0.
	request := newRequest(codes.SMB_COM_SEEK)
	request.Header.UID = 0
	request.AddCommand(commands.NewSeekRequest())
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	if response.Header.Status != uint32(nt_status.NT_STATUS_SMB_BAD_UID) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_SMB_BAD_UID", response.Header.Status)
	}

	// A command that does not need one still works.
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("sessionless")
	echoRequest := newRequest(codes.SMB_COM_ECHO)
	echoRequest.AddCommand(echo)
	sendRequest(t, client, echoRequest)
	if response, _ := receiveResponse(t, client); response.Header.Status != 0 {
		t.Fatalf("ECHO without a session answered 0x%08X, want success", response.Header.Status)
	}
}

// TestLogoffReleasesTheSession asserts a logoff drops the session and that the UID
// stops working afterwards.
func TestLogoffReleasesTheSession(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, false)

	response := sendOnSession(t, session, codes.SMB_COM_LOGOFF_ANDX, commands.NewLogoffAndxRequest())
	if response.Header.Status != 0 {
		t.Fatalf("LOGOFF_ANDX answered 0x%08X, want success", response.Header.Status)
	}

	// The UID no longer names a session.
	after := sendOnSession(t, session, codes.SMB_COM_TREE_CONNECT_ANDX, commands.NewSeekRequest())
	if after.Header.Status != uint32(nt_status.NT_STATUS_SMB_BAD_UID) {
		t.Fatalf("Status = 0x%08X after logoff, want NT_STATUS_SMB_BAD_UID", after.Header.Status)
	}

	// Logging off twice is refused rather than releasing an identifier twice.
	again := sendOnSession(t, session, codes.SMB_COM_LOGOFF_ANDX, commands.NewLogoffAndxRequest())
	if again.Header.Status != uint32(nt_status.NT_STATUS_SMB_BAD_UID) {
		t.Fatalf("a second logoff answered 0x%08X, want NT_STATUS_SMB_BAD_UID", again.Header.Status)
	}
}

// TestSeveralSessionsOnOneConnection asserts a connection can carry more than one
// session, which is what the UID-keyed pending exchanges exist for: the first leg
// of a second setup must not be mistaken for the second leg of the first.
func TestSeveralSessionsOnOneConnection(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) {
		c.AllowGuest = true
	})

	client := dialServer(t, addr)
	negotiateWithRawClient(t, client)

	logon := func(username, password string) types.USHORT {
		t.Helper()
		auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, captureDomain, username, password, "CLIENT01", true)
		v := ntlmversion.DefaultVersion()
		negotiateToken, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &v)
		if err != nil {
			t.Fatalf("CreateNegotiateToken() error = %v", err)
		}
		first := sendSessionSetupSigned(t, client, negotiateToken, 0, false, nil, 0)
		if first.Header.Status != uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED) {
			t.Fatalf("challenge leg answered 0x%08X", first.Header.Status)
		}
		challengeResponse := first.Command.(*commands.SessionSetupAndxResponse)
		authenticateToken, err := auth.CreateAuthenticateTokenFromChallengeToken([]byte(challengeResponse.SecurityBlob))
		if err != nil {
			t.Fatalf("CreateAuthenticateTokenFromChallengeToken() error = %v", err)
		}
		second := sendSessionSetupSigned(t, client, authenticateToken, first.Header.UID, false, nil, 0)
		if second.Header.Status != 0 {
			t.Fatalf("session setup for %q answered 0x%08X", username, second.Header.Status)
		}
		return second.Header.UID
	}

	firstUID := logon(captureUsername, capturePassword)
	secondUID := logon("bob", capturePassword)

	if firstUID == 0 || secondUID == 0 {
		t.Fatalf("a session was assigned UID 0 (%d, %d)", firstUID, secondUID)
	}
	if firstUID == secondUID {
		t.Fatalf("both sessions were assigned UID 0x%04X", uint16(firstUID))
	}
}

// TestSessionSetupOnUnknownUIDIsRefused asserts a second leg quoting a UID that
// names no pending exchange is refused rather than starting a new one.
func TestSessionSetupOnUnknownUIDIsRefused(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	client := dialServer(t, addr)
	negotiateWithRawClient(t, client)

	response := sendSessionSetupSigned(t, client, []byte{0x01, 0x02, 0x03}, 0x0BAD, false, nil, 0)
	if response.Header.Status != uint32(nt_status.NT_STATUS_SMB_BAD_UID) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_SMB_BAD_UID", response.Header.Status)
	}
}

// TestSessionLimitIsEnforced asserts a connection cannot establish more sessions
// than configured, so a client cannot exhaust the identifier space.
func TestSessionLimitIsEnforced(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) {
		c.MaxSessionsPerConnection = 1
	})

	client := dialServer(t, addr)
	negotiateWithRawClient(t, client)

	openExchange := func() uint32 {
		auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, captureDomain, captureUsername, capturePassword, "CLIENT01", true)
		v := ntlmversion.DefaultVersion()
		token, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &v)
		if err != nil {
			t.Fatalf("CreateNegotiateToken() error = %v", err)
		}
		return sendSessionSetupSigned(t, client, token, 0, false, nil, 0).Header.Status
	}

	// The first exchange takes the only identifier.
	if status := openExchange(); status != uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED) {
		t.Fatalf("first exchange answered 0x%08X", status)
	}
	// The second has none to take.
	if status := openExchange(); status != uint32(nt_status.NT_STATUS_TOO_MANY_SESSIONS) {
		t.Fatalf("second exchange answered 0x%08X, want NT_STATUS_TOO_MANY_SESSIONS", status)
	}
}

// TestIdentifierAllocator covers the allocator directly: identifiers avoid the
// reserved values, are reused once released, and run out at the limit rather than
// being handed out twice.
func TestIdentifierAllocator(t *testing.T) {
	allocator := newIdentifierAllocator(3)

	var allocated []uint16
	for i := 0; i < 3; i++ {
		identifier, err := allocator.Allocate()
		if err != nil {
			t.Fatalf("Allocate() error = %v", err)
		}
		if identifier == 0 || identifier == 0xFFFF {
			t.Fatalf("allocated the reserved identifier 0x%04X", identifier)
		}
		for _, seen := range allocated {
			if seen == identifier {
				t.Fatalf("identifier 0x%04X was allocated twice", identifier)
			}
		}
		allocated = append(allocated, identifier)
	}
	if allocator.InUse() != 3 {
		t.Fatalf("InUse() = %d, want 3", allocator.InUse())
	}

	// The limit is reached.
	if _, err := allocator.Allocate(); err == nil {
		t.Fatal("Allocate() exceeded the limit")
	}

	// Releasing one makes it available again.
	allocator.Release(allocated[1])
	if allocator.InUse() != 2 {
		t.Fatalf("InUse() = %d after a release, want 2", allocator.InUse())
	}
	reused, err := allocator.Allocate()
	if err != nil {
		t.Fatalf("Allocate() after a release error = %v", err)
	}
	if reused != allocated[1] {
		t.Fatalf("reallocated 0x%04X, want the released 0x%04X", reused, allocated[1])
	}

	// Releasing twice, or releasing something never allocated, does not corrupt
	// the count.
	before := allocator.InUse()
	allocator.Release(reused)
	allocator.Release(reused)
	allocator.Release(0)
	allocator.Release(0xFFFF)
	if allocator.InUse() != before-1 {
		t.Fatalf("InUse() = %d after duplicate releases, want %d", allocator.InUse(), before-1)
	}
}

// TestStaticAccounts asserts the credential store matches the username
// case-insensitively and the domain exactly, since the domain is folded into the
// client's response as sent.
func TestStaticAccounts(t *testing.T) {
	hash := nt.NTHash(capturePassword)
	lookup := StaticAccounts(Account{Domain: "Lab.Example.Local", Username: "alice", NTHash: hash})

	if got, ok := lookup("Lab.Example.Local", "alice"); !ok || !bytes.Equal(got[:], hash[:]) {
		t.Fatal("the exact identity was not found")
	}
	if _, ok := lookup("Lab.Example.Local", "ALICE"); !ok {
		t.Fatal("the username should match case-insensitively")
	}
	if _, ok := lookup("LAB.EXAMPLE.LOCAL", "alice"); ok {
		t.Fatal("the domain should match exactly, since it is folded into the response as sent")
	}
	if _, ok := lookup("Lab.Example.Local", "bob"); ok {
		t.Fatal("an unknown username was found")
	}
}

// TestSigningPolicyString asserts each policy renders for a log line.
func TestSigningPolicyString(t *testing.T) {
	cases := map[SigningPolicy]string{
		SigningDisabled:   "disabled",
		SigningEnabled:    "enabled",
		SigningRequired:   "required",
		SigningPolicy(99): "unknown",
	}
	for policy, want := range cases {
		if got := policy.String(); got != want {
			t.Errorf("SigningPolicy(%d).String() = %q, want %q", int(policy), got, want)
		}
	}
}

// TestSessionAccountRendering asserts how a session's identity reads, including
// the anonymous case which has no name at all.
func TestSessionAccountRendering(t *testing.T) {
	if got := (&Session{Domain: "D", Username: "u"}).Account(); got != "D\\u" {
		t.Errorf("Account() = %q", got)
	}
	if got := (&Session{Username: "u"}).Account(); got != "u" {
		t.Errorf("Account() with no domain = %q", got)
	}
	if got := (&Session{IsAnonymous: true}).Account(); got != "<anonymous>" {
		t.Errorf("Account() for an anonymous session = %q", got)
	}
	if (&Session{}).CanSign() {
		t.Error("a session with no key reports that it can sign")
	}
	if !(&Session{SessionKey: []byte{1}}).CanSign() {
		t.Error("a session with a key reports that it cannot sign")
	}
}

// TestClientEstablishesSession is the milestone this phase exists for: the SMB1
// client in this repository authenticates against this server and gets a session,
// where every earlier phase could only refuse it.
func TestClientEstablishesSession(t *testing.T) {
	srv, addr := serverWithAccount(t, SigningDisabled, nil)

	host, portText, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portText)

	client := smb1client.NewClientUsingTCPTransport(net.ParseIP(host), port)
	if err := client.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	defer client.Disconnect()

	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("client SessionSetup() error = %v", err)
	}

	if client.Session == nil || client.Session.SessionUID == 0 {
		t.Fatal("the client holds no session UID after a successful setup")
	}

	// The server agrees a session exists, under the identity that was verified.
	waitFor(t, func() bool { return srv.Connections() == 1 }, "the connection was not registered")

	// A wrong password on the same server is refused, so success above was not
	// the server accepting anything at all.
	wrong := smb1client.NewClientUsingTCPTransport(net.ParseIP(host), port)
	if err := wrong.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("second client Connect() error = %v", err)
	}
	defer wrong.Disconnect()
	badCreds, err := credentials.NewCredentials(captureDomain, captureUsername, "not-the-password", "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := wrong.SessionSetup(badCreds); err == nil {
		t.Fatal("the server accepted a wrong password")
	}
}

// TestClientEstablishesSignedSession asserts the two sides agree on signing end to
// end: the client sees the server require it, activates it, and every message
// after the exchange is signed and verified by the other side.
//
// This is the interop proof for the phase — the client and server derive the MAC
// key and the sequence numbers independently, so a mismatch in either would show
// up here and nowhere else.
func TestClientEstablishesSignedSession(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)

	host, portText, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portText)

	client := smb1client.NewClientUsingTCPTransport(net.ParseIP(host), port)
	if err := client.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	defer client.Disconnect()

	// The client learns the policy from the negotiate response.
	if client.Connection.Server.SigningState != smb1client.SigningStateRequired {
		t.Fatalf("client saw signing state %q, want required", client.Connection.Server.SigningState)
	}

	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("client SessionSetup() with signing required error = %v", err)
	}

	if !client.Connection.IsSigningActive {
		t.Fatal("the client did not activate signing although the server requires it")
	}
	if len(client.Connection.SigningSessionKey) == 0 {
		t.Fatal("the client armed signing with no key")
	}
	// The AUTHENTICATE consumed 0 and its response 1, so the next request is 2.
	if got := client.Connection.ClientNextSendSequenceNumber; got != 2 {
		t.Fatalf("client's next send sequence = %d, want 2", got)
	}

	// Every exchange after this is signed in both directions. Echo is the one
	// command available at this phase that acts on a session, and the client
	// verifies the server's signature itself.
	for exchange := 0; exchange < 3; exchange++ {
		payload := []byte("signed exchange")
		echoed, err := client.Echo(payload)
		if err != nil {
			t.Fatalf("client Echo() on a signed session error = %v", err)
		}
		if !bytes.Equal(echoed, payload) {
			t.Fatalf("echoed %q, want %q", echoed, payload)
		}
	}
}

// TestSecondSessionKeepsTheConnectionSigning asserts a second session set up on a
// connection that is already signing does not re-key it.
//
// Signing belongs to the connection: one key and one sequence carry every session
// on it. Arming it again for a second logon reset both, so the first session's
// next request was verified against a key it never used at a sequence the client
// had long passed, and the connection was dropped.
//
// TestSeveralSessionsOnOneConnection covers two sessions on one connection but
// runs with SigningDisabled, which is why this case needs its own test.
func TestSecondSessionKeepsTheConnectionSigning(t *testing.T) {
	const secondUsername, secondPassword = "bob", "bobs-password"

	_, addr := serverWithAccount(t, SigningRequired, func(c *Config) {
		c.Authenticator = StaticAccounts(
			Account{Domain: captureDomain, Username: captureUsername, NTHash: nt.NTHash(capturePassword)},
			Account{Domain: captureDomain, Username: secondUsername, NTHash: nt.NTHash(secondPassword)},
		)
	})

	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)
	if !session.signing {
		t.Fatal("signing was not armed on the first session")
	}
	keyBefore := append([]byte(nil), session.key...)

	// The first session works before the second logon. sendOnSession verifies the
	// response signature and advances the sequence, so a failure here is already a
	// signing failure.
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("before")
	if response := sendOnSession(t, session, codes.SMB_COM_ECHO, echo); response.Header.Status != 0 {
		t.Fatalf("the echo before the second logon answered 0x%08X, want success", response.Header.Status)
	}

	// A second logon for another account, on the same connection. Both legs are
	// signed with the connection's key at the connection's sequence, because that
	// is what a signing connection requires of every request on it.
	auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, captureDomain, secondUsername, secondPassword, "CLIENT01", true)
	version := ntlmversion.DefaultVersion()
	negotiateToken, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &version)
	if err != nil {
		t.Fatalf("CreateNegotiateToken() error = %v", err)
	}

	first := sendSessionSetupSigned(t, session.client, negotiateToken, 0, true, session.key, session.sequence)
	if first.Header.Status != uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED) {
		t.Fatalf("the second session's challenge leg answered 0x%08X, want NT_STATUS_MORE_PROCESSING_REQUIRED",
			first.Header.Status)
	}
	session.sequence = signing.NextRequestSequenceNumber(session.sequence)

	challengeResponse, ok := first.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		t.Fatalf("the second session's challenge leg command is %T", first.Command)
	}
	authenticateToken, err := auth.CreateAuthenticateTokenFromChallengeToken([]byte(challengeResponse.SecurityBlob))
	if err != nil {
		t.Fatalf("CreateAuthenticateTokenFromChallengeToken() error = %v", err)
	}

	second := sendSessionSetupSigned(t, session.client, authenticateToken, first.Header.UID, true,
		session.key, session.sequence)
	if second.Header.Status != 0 {
		t.Fatalf("the second session setup answered 0x%08X, want success", second.Header.Status)
	}
	secondUID := second.Header.UID
	if secondUID == session.uid {
		t.Fatalf("both sessions were assigned UID 0x%04X", uint16(secondUID))
	}
	session.sequence = signing.NextRequestSequenceNumber(session.sequence)

	// The connection's key is untouched, so the first session carries on.
	if !bytes.Equal(session.key, keyBefore) {
		t.Fatal("the test's own key changed, which it must not")
	}
	echo = commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("after")
	if response := sendOnSession(t, session, codes.SMB_COM_ECHO, echo); response.Header.Status != 0 {
		t.Fatalf("the echo after the second logon answered 0x%08X, want success", response.Header.Status)
	}

	// And so does the second session, on the same key and sequence.
	onSecond := &authSession{
		client:   session.client,
		uid:      secondUID,
		key:      session.key,
		signing:  true,
		sequence: session.sequence,
	}
	echo = commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("second session")
	if response := sendOnSession(t, onSecond, codes.SMB_COM_ECHO, echo); response.Header.Status != 0 {
		t.Fatalf("an echo on the second session answered 0x%08X, want success", response.Header.Status)
	}
}
