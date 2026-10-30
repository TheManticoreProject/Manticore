package server

import (
	"bytes"
	"encoding/binary"
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
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
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
	// rather than refused by the UID check. SMB_COM_CLOSE_AND_TREE_DISC is
	// recognized and never served — [MS-CIFS] section 2.2.4.45 has it "reserved
	// but not implemented" — so reaching STATUS_NOT_IMPLEMENTED means the
	// dispatch table was consulted, which is the point.
	response := sendOnSession(t, session, codes.SMB_COM_CLOSE_AND_TREE_DISC,
		commands.NewCloseAndTreeDiscRequest())
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

// TestAnonymousIsGovernedByAllowAnonymousAlone asserts a null session is
// admitted by AllowAnonymous without AllowGuest, and refused by AllowGuest
// without AllowAnonymous: the two flags are independent.
func TestAnonymousIsGovernedByAllowAnonymousAlone(t *testing.T) {
	cases := []struct {
		name           string
		allowGuest     bool
		allowAnonymous bool
		want           uint32
	}{
		{"anonymous only", false, true, 0},
		{"guest only", true, false, uint32(nt_status.NT_STATUS_LOGON_FAILURE)},
		{"both", true, true, 0},
		{"neither", false, false, uint32(nt_status.NT_STATUS_LOGON_FAILURE)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) {
				c.AllowGuest = tc.allowGuest
				c.AllowAnonymous = tc.allowAnonymous
			})

			_, status := rawLogon(t, addr, "", "", "", false)
			if status != tc.want {
				t.Fatalf("anonymous logon answered 0x%08X, want 0x%08X", status, tc.want)
			}
		})
	}
}

// TestNamedGuestIsGovernedByAllowGuestAlone asserts an unverifiable named
// identity is still admitted only under AllowGuest: AllowAnonymous does not
// admit it.
func TestNamedGuestIsGovernedByAllowGuestAlone(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) { c.AllowAnonymous = true })

	_, status := rawLogon(t, addr, captureDomain, "nobody", capturePassword, false)
	if status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
	}
}

// TestAnonymousIsRefusedWhenSigningIsRequired asserts a null session admitted by
// AllowAnonymous is still refused under a policy that requires signatures, since
// it derives no key.
func TestAnonymousIsRefusedWhenSigningIsRequired(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, func(c *Config) { c.AllowAnonymous = true })

	if _, status := rawLogon(t, addr, "", "", "", true); status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
		t.Fatalf("anonymous logon under SigningRequired answered 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
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

// sendNtCancelOnSession sends a signed SMB_COM_NT_CANCEL naming the given PID and
// MID, and advances the session's sequence by the one number a cancel consumes.
// Nothing is read back: a cancel has no response of its own.
func sendNtCancelOnSession(t *testing.T, session *authSession, pid uint32, mid uint16) {
	t.Helper()

	request := newRequest(codes.SMB_COM_NT_CANCEL)
	request.Header.UID = session.uid
	request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	request.Header.SetPID(pid)
	request.Header.MID = types.USHORT(mid)
	request.AddCommand(commands.NewNtCancelRequest())

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the cancel: %v", err)
	}
	signing.Sign(session.key, marshalled, session.sequence)
	if _, err := session.client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	// [MS-SMB] section 3.3.5.1: a cancel consumes one number, not two.
	session.sequence++
}

// TestSignedNtCancelConsumesOneSequenceNumber asserts a signed SMB_COM_NT_CANCEL
// advances the connection's sequence by one rather than two, so the request after
// it verifies. [MS-SMB] section 3.3.5.1 increments ServerNextReceiveSequenceNumber
// by one for a cancel because no response is sent for it; a server that advanced
// it by two rejected the next request and dropped the connection.
func TestSignedNtCancelConsumesOneSequenceNumber(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)

	// A cancel naming nothing outstanding is still a signed request, and still
	// consumes its number.
	sendNtCancelOnSession(t, session, 0x00001234, 0x0099)
	if session.sequence != 3 {
		t.Fatalf("the request after the cancel is at sequence %d, want 3", session.sequence)
	}

	// sendOnSession fails the test if the connection was dropped or the response
	// is not signed at the number after the echo's own.
	session.client.SetTimeout(5 * time.Second)
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("after cancel")
	response := sendOnSession(t, session, codes.SMB_COM_ECHO, echo)
	if response.Header.Status != 0 {
		t.Fatalf("the echo after the cancel answered 0x%08X, want success", response.Header.Status)
	}
}

// deferFirstEcho is a Handler that defers the first ECHO it sees and lets every
// later one through to the built-in dispatch.
type deferFirstEcho struct {
	deferred chan deferredRequest
	used     bool
}

func (h *deferFirstEcho) Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
	// Handlers run on the connection's single receive goroutine, so the flag
	// needs no locking.
	if req.Header.Command != codes.SMB_COM_ECHO || h.used {
		return false
	}
	h.used = true
	h.deferred <- deferredRequest{
		responder:  w.Defer(),
		connection: conn,
		pid:        req.Header.GetPID(),
		mid:        uint16(req.Header.MID),
	}
	return true
}

// TestSignedNtCancelOfAPendingRequestKeepsSequencesInStep asserts the full cancel
// exchange on a signing connection: the cancelled request is still answered at the
// response number its own request reserved, and the request after the cancel
// verifies at the number the client reached by counting the cancel as one.
func TestSignedNtCancelOfAPendingRequestKeepsSequencesInStep(t *testing.T) {
	srv, addr := serverWithAccount(t, SigningRequired, nil)
	session := establishSession(t, addr, captureDomain, captureUsername, capturePassword, true)

	handler := &deferFirstEcho{deferred: make(chan deferredRequest, 1)}
	srv.RegisterHandler(handler)

	// The ECHO at sequence 2 is deferred; its answer is owed at sequence 3.
	pendingSequence := session.sequence
	request := newRequest(codes.SMB_COM_ECHO)
	request.Header.UID = session.uid
	request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("cancel me")
	request.AddCommand(echo)
	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the echo: %v", err)
	}
	signing.Sign(session.key, marshalled, pendingSequence)
	if _, err := session.client.Send(marshalled); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	session.sequence = signing.NextRequestSequenceNumber(pendingSequence)

	var deferred deferredRequest
	select {
	case deferred = <-handler.deferred:
	case <-time.After(5 * time.Second):
		t.Fatal("the echo was never deferred")
	}

	// The cancel is signed at 4 and leaves the next request at 5.
	sendNtCancelOnSession(t, session, deferred.pid, deferred.mid)
	select {
	case <-deferred.responder.Cancelled():
	case <-time.After(5 * time.Second):
		t.Fatal("the cancel never reached the deferred request")
	}
	if err := deferred.responder.RespondWithError(nt_status.NT_STATUS_CANCELLED); err != nil {
		t.Fatalf("RespondWithError() error = %v", err)
	}

	session.client.SetTimeout(5 * time.Second)
	cancelled, raw := receiveResponse(t, session.client)
	if nt_status.NT_STATUS(cancelled.Header.Status) != nt_status.NT_STATUS_CANCELLED {
		t.Fatalf("the cancelled request answered 0x%08X, want STATUS_CANCELLED", cancelled.Header.Status)
	}
	if responseSequence := signing.ResponseSequenceNumber(pendingSequence); !signing.Verify(session.key, raw, responseSequence) {
		t.Fatalf("the cancelled request's answer is not signed at sequence %d", responseSequence)
	}

	// The request after the cancel is at 5, and its answer at 6.
	if session.sequence != 5 {
		t.Fatalf("the request after the cancel is at sequence %d, want 5", session.sequence)
	}
	followUp := commands.NewEchoRequest()
	followUp.EchoCount = 1
	followUp.Data = []byte("after cancel")
	response := sendOnSession(t, session, codes.SMB_COM_ECHO, followUp)
	if response.Header.Status != 0 {
		t.Fatalf("the echo after the cancel answered 0x%08X, want success", response.Header.Status)
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

// TestSessionCannotReachAnotherSessionsHandles asserts a session cannot act on a
// tree, a file handle or a search that another session on the same connection
// opened, even when it quotes the owner's TID.
//
// anyTreeFor checks the session owning a TID, but openFor only matched the FID to
// the request's TID, TREE_DISCONNECT removed whatever tree the TID named, and
// FIND_CLOSE2 closed whatever search the SID named. A guest session could close
// another session's files and searches and disconnect its trees.
func TestSessionCannotReachAnotherSessionsHandles(t *testing.T) {
	config := conformanceConfig(SigningDisabled)
	config.AllowGuest = true
	srv, transportEnd := pipedServer(t, config)

	fs := NewMemoryFileSystem("FILES")
	for _, name := range []string{"first.txt", "second.txt"} {
		if err := fs.AddFile(name, []byte("0123456789")); err != nil {
			t.Fatalf("AddFile() error = %v", err)
		}
	}
	if err := srv.AddShare(&Share{Name: fileShareName, Type: ShareTypeDisk, FS: fs}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	// The owner: an authenticated session with a tree, a file and a search open.
	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}
	if err := client.TreeConnect(fileShareName); err != nil {
		t.Fatalf("TreeConnect() error = %v", err)
	}
	fid, err := client.OpenFile("first.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}

	// SearchCount 1 of two entries, and no close flags, so the search stays open.
	findFirst := []byte{}
	findFirst = binary.LittleEndian.AppendUint16(findFirst, 0x0016) // SearchAttributes
	findFirst = binary.LittleEndian.AppendUint16(findFirst, 1)      // SearchCount
	findFirst = binary.LittleEndian.AppendUint16(findFirst, 0)      // Flags
	findFirst = binary.LittleEndian.AppendUint16(findFirst, smbFindFileBothDirectoryInfo)
	findFirst = binary.LittleEndian.AppendUint32(findFirst, 0) // SearchStorageType
	findFirst = append(findFirst, []byte("\\*\x00")...)
	parameters, _, status := sendTrans2(t, client, subcommands.TRANS2_FIND_FIRST2, findFirst, nil)
	if status != 0 || len(parameters) < 2 {
		t.Fatalf("FIND_FIRST2 answered 0x%08X", status)
	}
	sid := binary.LittleEndian.Uint16(parameters[0:2])

	// The intruder: a guest session on the same connection.
	auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, captureDomain, "mallory", "wrong", "CLIENT01", true)
	v := ntlmversion.DefaultVersion()
	negotiateToken, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &v)
	if err != nil {
		t.Fatalf("CreateNegotiateToken() error = %v", err)
	}
	first := sendSessionSetupSigned(t, transportEnd, negotiateToken, 0, false, nil, 0)
	challenge := first.Command.(*commands.SessionSetupAndxResponse)
	authenticateToken, err := auth.CreateAuthenticateTokenFromChallengeToken([]byte(challenge.SecurityBlob))
	if err != nil {
		t.Fatalf("CreateAuthenticateTokenFromChallengeToken() error = %v", err)
	}
	second := sendSessionSetupSigned(t, transportEnd, authenticateToken, first.Header.UID, false, nil, 0)
	if second.Header.Status != 0 {
		t.Fatalf("the guest session setup answered 0x%08X", second.Header.Status)
	}
	ownerUID, guestUID := client.Session.SessionUID, second.Header.UID

	// Every request below quotes the owner's TID from the guest's UID.
	client.Session.SessionUID = guestUID

	queryInfo := []byte{}
	queryInfo = binary.LittleEndian.AppendUint16(queryInfo, uint16(fid))
	queryInfo = binary.LittleEndian.AppendUint16(queryInfo, smbQueryFileBasicInfo)

	closeFile := commands.NewCloseRequest()
	closeFile.FID = types.USHORT(fid)
	closeSearch := commands.NewFindClose2Request()
	closeSearch.SearchHandle = types.USHORT(sid)

	attempts := []struct {
		name string
		send func() uint32
	}{
		{"READ_ANDX", func() uint32 {
			return binary.LittleEndian.Uint32(sendReadAndx(t, client, fid, 10, 0)[5:9])
		}},
		{"TRANS2_QUERY_FILE_INFORMATION", func() uint32 {
			_, _, status := sendTrans2(t, client, subcommands.TRANS2_QUERY_FILE_INFORMATION, queryInfo, nil)
			return status
		}},
		{"CLOSE", func() uint32 {
			status, _ := sendLegacy(t, client, codes.SMB_COM_CLOSE, closeFile)
			return status
		}},
		{"FIND_CLOSE2", func() uint32 {
			status, _ := sendLegacy(t, client, codes.SMB_COM_FIND_CLOSE2, closeSearch)
			return status
		}},
		{"TREE_DISCONNECT", func() uint32 {
			status, _ := sendLegacy(t, client, codes.SMB_COM_TREE_DISCONNECT, commands.NewTreeDisconnectRequest())
			return status
		}},
	}
	for _, attempt := range attempts {
		if status := attempt.send(); status != uint32(nt_status.NT_STATUS_SMB_BAD_TID) {
			t.Errorf("the guest's %s on the owner's tree answered 0x%08X, want NT_STATUS_SMB_BAD_TID",
				attempt.name, status)
		}
	}

	// The owner's tree, file and search all survived the attempts.
	client.Session.SessionUID = ownerUID
	data, err := client.ReadFile(fid, 0, 10)
	if err != nil {
		t.Fatalf("the owner's read after the guest's attempts failed: %v", err)
	}
	if string(data) != "0123456789" {
		t.Errorf("the owner read %q", data)
	}
	if status, _ := sendLegacy(t, client, codes.SMB_COM_FIND_CLOSE2, closeSearch); status != 0 {
		t.Errorf("the owner's FIND_CLOSE2 answered 0x%08X", status)
	}
	if err := client.CloseFile(fid); err != nil {
		t.Errorf("the owner's close failed: %v", err)
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

// TestClientSecondSessionKeepsTheConnectionSigning drives the repository's own
// client through a second logon on a connection that is already signing, then
// uses each session.
//
// The connection's signing key is set by the first authenticated session, and
// every later request, including a later SESSION_SETUP_ANDX, is signed with that
// key at the connection's running sequence ([MS-SMB] 3.2.5.3, [MS-CIFS] 3.2.4.2.4). The
// client sent the second setup's first leg unsigned, signed its AUTHENTICATE leg
// with the new session's key at sequence 0, and then re-keyed the connection and
// reset its sequence, so the server dropped the connection.
func TestClientSecondSessionKeepsTheConnectionSigning(t *testing.T) {
	const secondUsername, secondPassword = "bob", "bobs-password"

	for _, policy := range []SigningPolicy{SigningEnabled, SigningRequired} {
		policy := policy
		t.Run(policy.String(), func(t *testing.T) {
			config := conformanceConfig(policy)
			config.Authenticator = StaticAccounts(
				Account{Domain: captureDomain, Username: captureUsername, NTHash: nt.NTHash(capturePassword)},
				Account{Domain: captureDomain, Username: secondUsername, NTHash: nt.NTHash(secondPassword)},
			)
			srv, transportEnd := pipedServer(t, config)

			fs := NewMemoryFileSystem("FILES")
			if err := fs.AddFile("shared.txt", []byte("0123456789")); err != nil {
				t.Fatalf("AddFile() error = %v", err)
			}
			if err := srv.AddShare(&Share{Name: fileShareName, Type: ShareTypeDisk, FS: fs}); err != nil {
				t.Fatalf("AddShare() error = %v", err)
			}

			client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
			if err := client.Negotiate(); err != nil {
				t.Fatalf("Negotiate() error = %v", err)
			}

			first, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
			if err != nil {
				t.Fatalf("NewCredentials() error = %v", err)
			}
			if err := client.SessionSetup(first); err != nil {
				t.Fatalf("first SessionSetup() error = %v", err)
			}
			if !client.Connection.IsSigningActive {
				t.Fatal("signing was not armed by the first session")
			}
			firstSession := client.Session
			connectionKey := append([]byte(nil), client.Connection.SigningSessionKey...)

			second, err := credentials.NewCredentials(captureDomain, secondUsername, secondPassword, "")
			if err != nil {
				t.Fatalf("NewCredentials() error = %v", err)
			}
			if err := client.SessionSetup(second); err != nil {
				t.Fatalf("second SessionSetup() error = %v", err)
			}
			secondSession := client.Session
			if secondSession == firstSession || secondSession.SessionUID == firstSession.SessionUID {
				t.Fatal("the second logon did not produce a session of its own")
			}
			if !bytes.Equal(client.Connection.SigningSessionKey, connectionKey) {
				t.Fatal("the second session re-keyed the connection")
			}

			// Each session works on the connection afterwards. The server drops the
			// connection on a request that does not verify, and the client checks
			// every response signature, so each call is also a signing check.
			for _, session := range []*smb1client.Session{secondSession, firstSession} {
				client.Session = session
				if err := client.TreeConnect(fileShareName); err != nil {
					t.Fatalf("TreeConnect() on UID 0x%04X error = %v", session.SessionUID, err)
				}
				if _, err := client.OpenFile("shared.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
					fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE); err != nil {
					t.Fatalf("OpenFile() on UID 0x%04X error = %v", session.SessionUID, err)
				}
			}
		})
	}
}

// TestClientSignsWhenServerOnlyEnablesSigning is the regression guard for a
// client that asked for signing and then did not sign. The client sets
// SMB_FLAGS2_SECURITY_SIGNATURE on its session setup whenever the server offers
// signing, which is how a client asks a server that merely enables it to turn it
// on ([MS-SMB] 3.2.5.2). While signing was activated only for a server that
// *required* it, the server armed signing on being asked and the client kept
// sending unsigned requests, so the first request after session setup failed
// verification and the connection was closed.
func TestClientSignsWhenServerOnlyEnablesSigning(t *testing.T) {
	_, client := pipedClient(t, conformanceConfig(SigningEnabled), true)

	if client.Connection.Server.SigningState != smb1client.SigningStateEnabled {
		t.Fatalf("server signing state = %q, want %q",
			client.Connection.Server.SigningState, smb1client.SigningStateEnabled)
	}
	if !client.Connection.IsSigningActive {
		t.Fatal("signing is inactive against a server that enables it")
	}
	if len(client.Connection.SigningSessionKey) == 0 {
		t.Fatal("signing is active with no session key")
	}

	// The connection is what proves it: a request that did not verify would be
	// answered by the server closing the connection rather than by an error.
	payload := []byte("signed")
	echoed, err := client.Echo(payload)
	if err != nil {
		t.Fatalf("Echo() over a signing-enabled connection error = %v", err)
	}
	if !bytes.Equal(echoed, payload) {
		t.Fatalf("Echo() returned %q, want %q", echoed, payload)
	}
}

// TestClientDoesNotSignWhenServerDisablesSigning is the other side of the same
// decision: a server that does not offer signing gets none, so the fix above
// cannot have turned signing on unconditionally.
func TestClientDoesNotSignWhenServerDisablesSigning(t *testing.T) {
	_, client := pipedClient(t, conformanceConfig(SigningDisabled), true)

	if client.Connection.Server.SigningState != smb1client.SigningStateDisabled {
		t.Fatalf("server signing state = %q, want %q",
			client.Connection.Server.SigningState, smb1client.SigningStateDisabled)
	}
	if client.Connection.IsSigningActive {
		t.Error("signing is active against a server that does not offer it")
	}
	if _, err := client.Echo([]byte("unsigned")); err != nil {
		t.Fatalf("Echo() over an unsigned connection error = %v", err)
	}
}
