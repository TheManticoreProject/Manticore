package server

import (
	"bytes"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	ntlmflags "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	ntlmversion "github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// The identity the test client authenticates with. The domain is mixed-case on
// purpose: NTLMv2 folds it into the response exactly as sent, so anything that
// normalizes it produces a response that will not crack back.
const (
	captureDomain   = "Lab.Example.Local"
	captureUsername = "alice"
	capturePassword = "Passw0rd!"
)

// captureServerConfig is the identity the server advertises.
func captureServerConfig() Config {
	return Config{
		ServerName:      "MANTICORE",
		DomainName:      "LAB",
		DNSComputerName: "manticore.lab.example.local",
		DNSDomainName:   "lab.example.local",
		Timeout:         5 * time.Second,
	}
}

// startCaptureServer stands up a server with a capture handler on an ephemeral
// loopback port.
func startCaptureServer(t *testing.T, config CaptureConfig) (*Server, *CaptureHandler, string) {
	t.Helper()

	handler, err := NewCaptureHandler(config)
	if err != nil {
		t.Fatalf("NewCaptureHandler() error = %v", err)
	}

	srv, err := NewServer(captureServerConfig())
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	srv.RegisterHandler(handler)

	listener := listenLoopback(t)
	addr := listener.Addr().String()
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	return srv, handler, addr
}

// authenticateWithClient drives the merged SMB1 client against the server: it
// negotiates and then attempts a session setup, which the server is expected to
// refuse. The refusal is returned rather than failed on.
func authenticateWithClient(t *testing.T, addr, domain, username, password string) error {
	t.Helper()

	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort(%q) error = %v", addr, err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("Atoi(%q) error = %v", portText, err)
	}

	client := smb1client.NewClientUsingTCPTransport(net.ParseIP(host), port)
	if err := client.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	// Connect negotiates as part of connecting, so there is no separate
	// Negotiate call: a second one would be a protocol violation.
	defer client.Disconnect()

	creds, err := credentials.NewCredentials(domain, username, password, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	// The server refuses every logon in this phase, so an error here is the
	// expected outcome and is handed back to the caller.
	return client.SessionSetup(creds)
}

// TestClientNegotiatesWithServer asserts the merged SMB1 client completes
// negotiation against this server and agrees on the dialect and the capabilities
// the server advertises.
func TestClientNegotiatesWithServer(t *testing.T) {
	srv, err := NewServer(captureServerConfig())
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	listener := listenLoopback(t)
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })

	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)

	client := smb1client.NewClientUsingTCPTransport(net.ParseIP(host), port)
	if err := client.Connect(net.ParseIP(host), port); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	// Connect negotiates as part of connecting.
	defer client.Disconnect()

	if got := client.Connection.SelectedDialect; got != dialects.DIALECT_NT_LM_0_12 {
		t.Fatalf("client selected dialect %q, want %q", got, dialects.DIALECT_NT_LM_0_12)
	}
	if client.Connection.Server.MaxBufferSize != DefaultMaxBufferSize {
		t.Fatalf("client saw MaxBufferSize %d, want %d", client.Connection.Server.MaxBufferSize, DefaultMaxBufferSize)
	}
	if client.Connection.MaxMpxCount != DefaultMaxMpxCount {
		t.Fatalf("client saw MaxMpxCount %d, want %d", client.Connection.MaxMpxCount, DefaultMaxMpxCount)
	}
	// The client needs extended security to have been advertised, or it cannot
	// take the NTLMSSP path at all.
	if !client.Connection.Server.Capabilities.HasCapability(0x80000000) {
		t.Fatalf("server capabilities 0x%08X do not include CAP_EXTENDED_SECURITY",
			uint32(client.Connection.Server.Capabilities))
	}
	// Signing must not be advertised: this phase cannot sign.
	if client.Connection.Server.SigningState != smb1client.SigningStateDisabled {
		t.Fatalf("server advertised signing state %q, want disabled", client.Connection.Server.SigningState)
	}
	// The server GUID is generated, so it must not be all zeroes.
	if client.Connection.Server.ServerGUID.ToFormatD() == "00000000-0000-0000-0000-000000000000" {
		t.Fatal("server advertised a zero GUID")
	}
}

// TestCaptureFromRealClient is the milestone this phase exists for: a real client
// authenticates, the server refuses it, and the response is harvested in a form
// that can be cracked offline.
func TestCaptureFromRealClient(t *testing.T) {
	_, handler, addr := startCaptureServer(t, CaptureConfig{})

	// The server refuses every logon, so the client must fail.
	if err := authenticateWithClient(t, addr, captureDomain, captureUsername, capturePassword); err == nil {
		t.Fatal("client SessionSetup() succeeded, but this phase refuses every logon")
	}

	captured := handler.Credentials()
	if len(captured) != 1 {
		t.Fatalf("captured %d credentials, want 1", len(captured))
	}
	credential := captured[0]

	if credential.Domain != captureDomain {
		t.Fatalf("captured domain = %q, want %q (the case must survive)", credential.Domain, captureDomain)
	}
	if credential.Username != captureUsername {
		t.Fatalf("captured username = %q, want %q", credential.Username, captureUsername)
	}
	if !credential.IsNTLMv2() {
		t.Fatalf("captured a %d-byte NT response, want a NetNTLMv2 one", len(credential.NtResponse))
	}
	if credential.ServerChallenge == ([8]byte{}) {
		t.Fatal("captured a zero server challenge, so the response could not be cracked")
	}
	if credential.RemoteAddr == nil {
		t.Fatal("captured no remote address")
	}
	if credential.Time.IsZero() {
		t.Fatal("captured no timestamp")
	}

	line, mode, err := credential.Hashcat()
	if err != nil {
		t.Fatalf("Hashcat() error = %v", err)
	}
	if mode != ntlmv2.HashcatMode {
		t.Fatalf("mode = %d, want %d for NetNTLMv2", mode, ntlmv2.HashcatMode)
	}

	// The line must be a well-formed mode-5600 record naming this identity and
	// carrying the challenge the server issued.
	fields := strings.Split(line, ":")
	if len(fields) != 6 {
		t.Fatalf("hashcat line has %d fields, want 6: %q", len(fields), line)
	}
	if fields[0] != captureUsername {
		t.Fatalf("line names user %q, want %q", fields[0], captureUsername)
	}
	if fields[2] != captureDomain {
		t.Fatalf("line names domain %q, want %q", fields[2], captureDomain)
	}
	if !strings.EqualFold(fields[3], hex.EncodeToString(credential.ServerChallenge[:])) {
		t.Fatalf("line carries challenge %q, want %x", fields[3], credential.ServerChallenge)
	}
	if len(fields[4]) != 32 {
		t.Fatalf("NTProofStr field is %d hex chars, want 32", len(fields[4]))
	}
	if len(fields[5]) == 0 {
		t.Fatal("the blob field is empty, so the line is not crackable")
	}
}

// TestCapturedResponseCracksBack is the proof that a capture is usable: recomputing
// the response from the password the client used must reproduce the captured
// NTProofStr exactly. If it does not, the material would never crack.
func TestCapturedResponseCracksBack(t *testing.T) {
	_, handler, addr := startCaptureServer(t, CaptureConfig{})

	if err := authenticateWithClient(t, addr, captureDomain, captureUsername, capturePassword); err == nil {
		t.Fatal("client SessionSetup() unexpectedly succeeded")
	}

	captured := handler.Credentials()
	if len(captured) != 1 {
		t.Fatalf("captured %d credentials, want 1", len(captured))
	}
	credential := captured[0]

	// This is what a cracker does: derive the response key from a candidate
	// password and the claimed identity, then verify the response against the
	// challenge and the blob the client sent.
	responseKeyNT := ntlmv2.NTOWFv2(capturePassword, credential.Username, credential.Domain)
	if !ntlmv2.VerifyNTChallengeResponse(responseKeyNT, credential.ServerChallenge, credential.NtResponse) {
		t.Fatal("the captured response does not verify against the password the client used, so it would never crack")
	}

	// A wrong candidate must not verify, or every guess would appear correct.
	wrongKey := ntlmv2.NTOWFv2("not-the-password", credential.Username, credential.Domain)
	if ntlmv2.VerifyNTChallengeResponse(wrongKey, credential.ServerChallenge, credential.NtResponse) {
		t.Fatal("the captured response verified against the wrong password")
	}
}

// TestCaptureCallbackAndFileOutput asserts both reporting sinks receive the
// capture, and that the file output is a hashcat-parsable record labelled with its
// mode.
func TestCaptureCallbackAndFileOutput(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "captured.txt")
	reported := make(chan Credential, 4)

	_, handler, addr := startCaptureServer(t, CaptureConfig{
		OnCredential: func(c Credential) { reported <- c },
		OutputFile:   outputPath,
	})

	if err := authenticateWithClient(t, addr, captureDomain, captureUsername, capturePassword); err == nil {
		t.Fatal("client SessionSetup() unexpectedly succeeded")
	}

	select {
	case credential := <-reported:
		if credential.Username != captureUsername {
			t.Fatalf("callback received %q, want %q", credential.Username, captureUsername)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the OnCredential callback was never called")
	}

	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read the output file: %v", err)
	}
	text := string(contents)
	if !strings.Contains(text, "# hashcat mode 5600") {
		t.Fatalf("output file does not label the mode:\n%s", text)
	}

	want, _, err := handler.Credentials()[0].Hashcat()
	if err != nil {
		t.Fatalf("Hashcat() error = %v", err)
	}
	if !strings.Contains(text, want) {
		t.Fatalf("output file does not contain the rendered line:\n%s", text)
	}
}

// TestCaptureUniquePerUser asserts a retrying client is recorded once, since a
// refused client normally tries again and would otherwise fill the output with
// duplicates.
func TestCaptureUniquePerUser(t *testing.T) {
	_, handler, addr := startCaptureServer(t, CaptureConfig{UniquePerUser: true})

	for attempt := 0; attempt < 3; attempt++ {
		if err := authenticateWithClient(t, addr, captureDomain, captureUsername, capturePassword); err == nil {
			t.Fatal("client SessionSetup() unexpectedly succeeded")
		}
	}
	if got := len(handler.Credentials()); got != 1 {
		t.Fatalf("captured %d credentials for one identity, want 1", got)
	}

	// A different identity is still recorded.
	if err := authenticateWithClient(t, addr, captureDomain, "bob", capturePassword); err == nil {
		t.Fatal("client SessionSetup() unexpectedly succeeded")
	}
	if got := len(handler.Credentials()); got != 2 {
		t.Fatalf("captured %d credentials for two identities, want 2", got)
	}
}

// TestCaptureRecordsEveryAttemptByDefault asserts that without UniquePerUser each
// attempt is kept, which is what a caller watching for a password change wants.
func TestCaptureRecordsEveryAttemptByDefault(t *testing.T) {
	_, handler, addr := startCaptureServer(t, CaptureConfig{})

	for attempt := 0; attempt < 3; attempt++ {
		if err := authenticateWithClient(t, addr, captureDomain, captureUsername, capturePassword); err == nil {
			t.Fatal("client SessionSetup() unexpectedly succeeded")
		}
	}
	if got := len(handler.Credentials()); got != 3 {
		t.Fatalf("captured %d credentials, want 3", got)
	}
}

// TestCaptureStatusIsConfigurable asserts the status returned after a capture can
// be chosen, and that it defaults to a logon failure.
func TestCaptureStatusIsConfigurable(t *testing.T) {
	t.Run("default is logon failure", func(t *testing.T) {
		handler, err := NewCaptureHandler(CaptureConfig{})
		if err != nil {
			t.Fatalf("NewCaptureHandler() error = %v", err)
		}
		if handler.config.Status != nt_status.NT_STATUS_LOGON_FAILURE {
			t.Fatalf("default status = %s, want NT_STATUS_LOGON_FAILURE", statusName(handler.config.Status))
		}
	})

	t.Run("configured status reaches the client", func(t *testing.T) {
		_, _, addr := startCaptureServer(t, CaptureConfig{Status: nt_status.NT_STATUS_ACCESS_DENIED})

		client := dialServer(t, addr)
		status := runSessionSetupLegs(t, client)
		if status != uint32(nt_status.NT_STATUS_ACCESS_DENIED) {
			t.Fatalf("second leg answered 0x%08X, want NT_STATUS_ACCESS_DENIED (0x%08X)",
				status, uint32(nt_status.NT_STATUS_ACCESS_DENIED))
		}
	})
}

// TestChallengeLegReportsMoreProcessingRequired asserts the first leg answers with
// a security blob and STATUS_MORE_PROCESSING_REQUIRED, and assigns a UID the client
// echoes on the second leg. A client that does not see that status abandons the
// exchange.
func TestChallengeLegReportsMoreProcessingRequired(t *testing.T) {
	_, _, addr := startCaptureServer(t, CaptureConfig{})
	client := dialServer(t, addr)

	negotiateWithRawClient(t, client)

	request := newRequest(codes.SMB_COM_SESSION_SETUP_ANDX)
	request.Header.Flags2 |= 1 << 11 // FLAGS2_EXTENDED_SECURITY
	request.Header.UID = 0
	setup := commands.NewSessionSetupAndxRequest()
	setup.SecurityBlob = negotiateTokenForTest(t)
	setup.SecurityBlobLength = types.USHORT(len(setup.SecurityBlob))
	setup.MaxBufferSize = 16644
	setup.MaxMpxCount = 50
	// The extended-security layout, carrying a security blob rather than a
	// password, is selected by this capability bit.
	setup.Capabilities = capabilities.CAP_EXTENDED_SECURITY
	setup.NativeOS = "TestOS"
	setup.NativeLanMan = "TestLanMan"
	request.AddCommand(setup)
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	if response.Header.Status != uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED) {
		t.Fatalf("challenge leg answered 0x%08X, want NT_STATUS_MORE_PROCESSING_REQUIRED (0x%08X)",
			response.Header.Status, uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED))
	}
	if response.Header.UID == 0 {
		t.Fatal("the challenge leg assigned no UID")
	}

	setupResponse, ok := response.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		t.Fatalf("response command is %T, want *commands.SessionSetupAndxResponse", response.Command)
	}
	if len(setupResponse.SecurityBlob) == 0 {
		t.Fatal("the challenge leg carried no security blob")
	}
	if len(setupResponse.NativeOS) == 0 || len(setupResponse.NativeLanMan) == 0 {
		t.Fatal("NativeOS or NativeLanMan is empty, which strict clients reject")
	}
}

// TestSessionSetupBeforeNegotiateIsRefused asserts a session setup on a connection
// that has not agreed a dialect is refused.
func TestSessionSetupBeforeNegotiateIsRefused(t *testing.T) {
	_, _, addr := startCaptureServer(t, CaptureConfig{})
	client := dialServer(t, addr)

	request := newRequest(codes.SMB_COM_SESSION_SETUP_ANDX)
	request.Header.Flags2 |= 1 << 11
	setup := commands.NewSessionSetupAndxRequest()
	setup.Capabilities = capabilities.CAP_EXTENDED_SECURITY
	setup.SecurityBlob = []byte{0x01}
	setup.SecurityBlobLength = 1
	request.AddCommand(setup)
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	if response.Header.Status != uint32(nt_status.NT_STATUS_INVALID_SMB) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_INVALID_SMB (0x%08X)",
			response.Header.Status, uint32(nt_status.NT_STATUS_INVALID_SMB))
	}
}

// TestSecondNegotiateIsRefused asserts renegotiating on one connection is refused,
// since it would silently discard the state built on the first agreement.
func TestSecondNegotiateIsRefused(t *testing.T) {
	_, _, addr := startCaptureServer(t, CaptureConfig{})
	client := dialServer(t, addr)

	negotiateWithRawClient(t, client)

	request := newRequest(codes.SMB_COM_NEGOTIATE)
	negotiate := commands.NewNegotiateRequest()
	negotiate.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)
	request.AddCommand(negotiate)
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	if response.Header.Status != uint32(nt_status.NT_STATUS_INVALID_SMB) {
		t.Fatalf("Status = 0x%08X, want NT_STATUS_INVALID_SMB", response.Header.Status)
	}
}

// TestUnsupportedDialectIsRefusedByIndex asserts a client offering only dialects
// this server does not speak is answered with DialectIndex 0xFFFF rather than an
// error status, which is what lets it fall back.
func TestUnsupportedDialectIsRefusedByIndex(t *testing.T) {
	_, _, addr := startCaptureServer(t, CaptureConfig{})
	client := dialServer(t, addr)

	request := newRequest(codes.SMB_COM_NEGOTIATE)
	negotiate := commands.NewNegotiateRequest()
	negotiate.Dialects.AddDialect(dialects.DIALECT_PC_NETWORK_PROGRAM_1_0)
	negotiate.Dialects.AddDialect(dialects.DIALECT_LANMAN_2_1)
	request.AddCommand(negotiate)
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	if response.Header.Status != 0 {
		t.Fatalf("Status = 0x%08X, want success alongside the refusal index", response.Header.Status)
	}
	negotiateResponse, ok := response.Command.(*commands.NegotiateResponse)
	if !ok {
		t.Fatalf("response command is %T, want *commands.NegotiateResponse", response.Command)
	}
	if negotiateResponse.DialectIndex != noDialectSelected {
		t.Fatalf("DialectIndex = 0x%04X, want 0xFFFF", uint16(negotiateResponse.DialectIndex))
	}
}

// TestDialectSelectionUsesTheClientsIndex asserts the returned index is an offset
// into the list the client offered, not into a list of the server's own. Getting
// this wrong makes the client resolve the answer to the wrong dialect name.
func TestDialectSelectionUsesTheClientsIndex(t *testing.T) {
	offered := dialects.Dialects{Dialects: []string{
		dialects.DIALECT_PC_NETWORK_PROGRAM_1_0,
		dialects.DIALECT_LANMAN_1_0,
		dialects.DIALECT_LANMAN_2_1,
		dialects.DIALECT_NT_LM_0_12,
	}}
	index, selected := selectDialect(offered)
	if selected != dialects.DIALECT_NT_LM_0_12 {
		t.Fatalf("selected %q, want %q", selected, dialects.DIALECT_NT_LM_0_12)
	}
	if index != 3 {
		t.Fatalf("index = %d, want 3 (the position in the client's list)", index)
	}

	if _, selected := selectDialect(dialects.Dialects{Dialects: []string{dialects.DIALECT_LANMAN_2_1}}); selected != "" {
		t.Fatalf("selected %q from an unsupported list, want none", selected)
	}
}

// TestCredentialHashcatRendering asserts a Credential picks the right mode for the
// response it holds, and refuses one that is neither shape.
func TestCredentialHashcatRendering(t *testing.T) {
	challenge := [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}

	t.Run("NetNTLMv2", func(t *testing.T) {
		credential := Credential{
			Domain: "DOMAIN", Username: "user",
			ServerChallenge: challenge,
			LmResponse:      make([]byte, 24),
			NtResponse:      bytes.Repeat([]byte{0xAB}, 60),
		}
		line, mode, err := credential.Hashcat()
		if err != nil {
			t.Fatalf("Hashcat() error = %v", err)
		}
		if mode != ntlmv2.HashcatMode {
			t.Fatalf("mode = %d, want %d", mode, ntlmv2.HashcatMode)
		}
		if !strings.HasPrefix(line, "user::DOMAIN:") {
			t.Fatalf("line = %q", line)
		}
	})

	t.Run("NetNTLMv1", func(t *testing.T) {
		credential := Credential{
			Domain: "DOMAIN", Username: "user",
			ServerChallenge: challenge,
			LmResponse:      bytes.Repeat([]byte{0x11}, 24),
			NtResponse:      bytes.Repeat([]byte{0x22}, 24),
		}
		line, mode, err := credential.Hashcat()
		if err != nil {
			t.Fatalf("Hashcat() error = %v", err)
		}
		if mode != ntlmv1.HashcatMode {
			t.Fatalf("mode = %d, want %d", mode, ntlmv1.HashcatMode)
		}
		// Mode 5500 puts the challenge last.
		fields := strings.Split(line, ":")
		if !strings.EqualFold(fields[len(fields)-1], hex.EncodeToString(challenge[:])) {
			t.Fatalf("mode 5500 line does not end with the challenge: %q", line)
		}
	})

	t.Run("neither shape", func(t *testing.T) {
		credential := Credential{NtResponse: bytes.Repeat([]byte{0x33}, 10)}
		if _, _, err := credential.Hashcat(); err == nil {
			t.Fatal("Hashcat() accepted a 10-byte NT response")
		}
	})

	t.Run("account rendering", func(t *testing.T) {
		if got := (Credential{Domain: "D", Username: "u"}).Account(); got != "D\\u" {
			t.Fatalf("Account() = %q", got)
		}
		if got := (Credential{Username: "u"}).Account(); got != "u" {
			t.Fatalf("Account() with no domain = %q", got)
		}
	})
}

// TestNewCaptureHandlerRejectsUnwritableOutput asserts an output file that cannot
// be opened fails at construction, not on the first capture when a client is
// waiting and the material would be lost.
func TestNewCaptureHandlerRejectsUnwritableOutput(t *testing.T) {
	unwritable := filepath.Join(t.TempDir(), "no-such-directory", "captured.txt")
	if _, err := NewCaptureHandler(CaptureConfig{OutputFile: unwritable}); err == nil {
		t.Fatal("NewCaptureHandler() accepted an unwritable output path")
	}
}

// TestServerConfigDefaults asserts the defaults a client depends on are filled in,
// and that a MaxBufferSize the protocol forbids is refused.
func TestServerConfigDefaults(t *testing.T) {
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	config := srv.Config()

	if config.NativeOS == "" || config.NativeLanMan == "" {
		t.Fatal("NativeOS and NativeLanMan must be defaulted, since strict clients reject empty ones")
	}
	if config.MaxBufferSize != DefaultMaxBufferSize {
		t.Fatalf("MaxBufferSize = %d, want %d", config.MaxBufferSize, DefaultMaxBufferSize)
	}
	if config.MaxMpxCount != DefaultMaxMpxCount {
		t.Fatalf("MaxMpxCount = %d, want %d", config.MaxMpxCount, DefaultMaxMpxCount)
	}
	if config.ServerGUID.ToFormatD() == "00000000-0000-0000-0000-000000000000" {
		t.Fatal("ServerGUID was not generated")
	}

	// [MS-CIFS] 2.2.4.52.2 requires a multiple of 4.
	if _, err := NewServer(Config{MaxBufferSize: 4355}); err == nil {
		t.Fatal("NewServer() accepted a MaxBufferSize that is not a multiple of 4")
	}
}

// listenLoopback opens a Direct TCP listener on an ephemeral loopback port.
func listenLoopback(t *testing.T) transport.Listener {
	t.Helper()
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() error = %v", err)
	}
	return listener
}

// testClientNegotiateFlags are the NTLM options a client offers. They mirror what
// the SMB1 client in this repository sends.
const testClientNegotiateFlags = ntlmflags.NTLMSSP_NEGOTIATE_UNICODE |
	ntlmflags.NTLMSSP_NEGOTIATE_NTLM |
	ntlmflags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
	ntlmflags.NTLMSSP_NEGOTIATE_SIGN |
	ntlmflags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
	ntlmflags.NTLMSSP_NEGOTIATE_128 |
	ntlmflags.NTLMSSP_NEGOTIATE_56 |
	ntlmflags.NTLMSSP_REQUEST_TARGET |
	ntlmflags.NTLMSSP_NEGOTIATE_TARGET_INFO

// negotiateWithRawClient completes SMB dialect negotiation over a raw transport,
// for a test that then wants to drive session setup itself.
func negotiateWithRawClient(t *testing.T, client *tcp.TCPTransport) {
	t.Helper()

	request := newRequest(codes.SMB_COM_NEGOTIATE)
	request.Header.Flags2 |= 1 << 11 // FLAGS2_EXTENDED_SECURITY
	negotiate := commands.NewNegotiateRequest()
	negotiate.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)
	request.AddCommand(negotiate)
	sendRequest(t, client, request)

	response, _ := receiveResponse(t, client)
	if response.Header.Status != 0 {
		t.Fatalf("NEGOTIATE answered 0x%08X, want success", response.Header.Status)
	}
}

// negotiateTokenForTest returns the SPNEGO-wrapped NTLM NEGOTIATE a client sends
// on the first session-setup leg.
func negotiateTokenForTest(t *testing.T) []byte {
	t.Helper()

	auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, captureDomain, captureUsername, capturePassword, "CLIENT01", true)
	v := ntlmversion.DefaultVersion()
	token, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &v)
	if err != nil {
		t.Fatalf("CreateNegotiateToken() error = %v", err)
	}
	return token
}

// runSessionSetupLegs drives both session-setup legs over a raw transport and
// returns the status the server answered the second leg with.
func runSessionSetupLegs(t *testing.T, client *tcp.TCPTransport) uint32 {
	t.Helper()

	negotiateWithRawClient(t, client)

	auth := spnego.NewAuthContext(spnego.AuthTypeNTLM, captureDomain, captureUsername, capturePassword, "CLIENT01", true)
	v := ntlmversion.DefaultVersion()
	negotiateToken, err := auth.CreateNegotiateToken(testClientNegotiateFlags, &v)
	if err != nil {
		t.Fatalf("CreateNegotiateToken() error = %v", err)
	}

	// First leg: send the NEGOTIATE, receive the challenge.
	firstResponse := sendSessionSetup(t, client, negotiateToken, 0)
	if firstResponse.Header.Status != uint32(nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED) {
		t.Fatalf("challenge leg answered 0x%08X, want NT_STATUS_MORE_PROCESSING_REQUIRED", firstResponse.Header.Status)
	}
	challengeResponse, ok := firstResponse.Command.(*commands.SessionSetupAndxResponse)
	if !ok {
		t.Fatalf("challenge leg command is %T", firstResponse.Command)
	}

	// Second leg: answer the challenge with an AUTHENTICATE.
	authenticateToken, err := auth.CreateAuthenticateTokenFromChallengeToken([]byte(challengeResponse.SecurityBlob))
	if err != nil {
		t.Fatalf("CreateAuthenticateTokenFromChallengeToken() error = %v", err)
	}

	secondResponse := sendSessionSetup(t, client, authenticateToken, firstResponse.Header.UID)
	return secondResponse.Header.Status
}

// sendSessionSetup sends one session-setup request carrying a security blob and
// returns the response.
func sendSessionSetup(t *testing.T, client *tcp.TCPTransport, blob []byte, uid types.USHORT) *message.Message {
	t.Helper()

	request := newRequest(codes.SMB_COM_SESSION_SETUP_ANDX)
	request.Header.Flags2 |= 1 << 11 // FLAGS2_EXTENDED_SECURITY
	request.Header.UID = uid

	setup := commands.NewSessionSetupAndxRequest()
	setup.SecurityBlob = blob
	setup.SecurityBlobLength = types.USHORT(len(blob))
	setup.MaxBufferSize = 16644
	setup.MaxMpxCount = 50
	// The extended-security layout, carrying a security blob rather than a
	// password, is selected by this capability bit.
	setup.Capabilities = capabilities.CAP_EXTENDED_SECURITY
	setup.NativeOS = "TestOS"
	setup.NativeLanMan = "TestLanMan"
	request.AddCommand(setup)

	sendRequest(t, client, request)
	response, _ := receiveResponse(t, client)
	return response
}
