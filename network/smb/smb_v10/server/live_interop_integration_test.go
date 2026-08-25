//go:build integration

// Live interoperability coverage for the SMB 1.0 server against a third-party
// client, driven through the client's own command-line tools. Excluded from the
// default build by the "integration" tag.
//
// The unit suite pairs this server with the SMB1 client in this repository, which
// is the fastest way to cover behaviour but has a structural blind spot: the two
// halves were written together, so a wire detail both get wrong agrees with itself
// and every round-trip passes. Everything in this file exists to be driven by an
// implementation that shares none of this one's assumptions.
//
// It found real defects. Among them: a request field decoded with a single-byte
// terminator when the client had negotiated Unicode, so every path arrived
// truncated to one character; transaction parameter blocks located by arithmetic
// that assumed the block ended flush with its last field, which shifted every
// field once a client padded the end of the block; and a signature demanded on the
// message that establishes signing, which no specification asks for and which made
// signing unusable rather than mandatory.
//
// Configuration:
//
//	SMB1_TEST_CLIENT   path to a third-party smbclient-style binary. Required;
//	                   every test skips when it is unset.
//	SMB1_TEST_RPCCLI   path to the matching RPC client binary, for the named-pipe
//	                   tests. Those skip when it is unset.
//	SMB1_TEST_PORT     port to listen on (default 4445). A privileged port is not
//	                   used, so the client must accept a port argument.
//
// The server is started inside the test on a loopback port, so nothing outside the
// process is touched and no lab is needed beyond the client binary itself.
package server

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
)

// The identity the live server accepts.
const (
	liveDomain   = "WORKGROUP"
	liveUsername = "alice"
	livePassword = "Passw0rd!"
	liveShare    = "files"
)

// liveClient returns the configured client binary, or skips.
func liveClient(t *testing.T) string {
	t.Helper()
	path := os.Getenv("SMB1_TEST_CLIENT")
	if path == "" {
		t.Skip("SMB1_TEST_CLIENT is not set; no third-party client to interoperate with")
	}
	return path
}

// liveRPCClient returns the configured RPC client binary, or skips.
func liveRPCClient(t *testing.T) string {
	t.Helper()
	path := os.Getenv("SMB1_TEST_RPCCLI")
	if path == "" {
		t.Skip("SMB1_TEST_RPCCLI is not set; no third-party RPC client to interoperate with")
	}
	return path
}

// livePort returns the port to listen on.
func livePort() int {
	if value := os.Getenv("SMB1_TEST_PORT"); value != "" {
		if port, err := strconv.Atoi(value); err == nil {
			return port
		}
	}
	return 4445
}

// livePipe is a pipe handler that answers a DCE/RPC bind with a bind_ack and
// echoes anything else.
//
// The bind_ack is what makes the named-pipe path testable with a real RPC client:
// without one the client abandons the connection at the bind and never exercises a
// second transaction, a read, or the close.
type livePipe struct {
	mutex sync.Mutex
	calls []string
}

func (p *livePipe) record(call string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.calls = append(p.calls, call)
}

// Calls returns the handler calls seen so far.
func (p *livePipe) Calls() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return append([]string(nil), p.calls...)
}

func (p *livePipe) OpenPipe(name string) error {
	p.record("open:" + name)
	if strings.ToLower(name) != "srvsvc" {
		return fmt.Errorf("pipe %q not found", name)
	}
	return nil
}

func (p *livePipe) ClosePipe(name string) error {
	p.record("close:" + name)
	return nil
}

func (p *livePipe) Transact(name string, input []byte, maxOutput int) ([]byte, bool, error) {
	if answer := rpcBindAck(input); answer != nil {
		p.record("bind:" + name)
		if len(answer) > maxOutput {
			return answer[:maxOutput], true, nil
		}
		return answer, false, nil
	}

	p.record("transact:" + name)
	answer := append([]byte("echo:"), input...)
	if len(answer) > maxOutput {
		return answer[:maxOutput], true, nil
	}
	return answer, false, nil
}

// rpcBindAck builds a DCE/RPC bind_ack for a bind PDU, or nil when the input is
// not one ([C706] section 12.6.4.4).
//
// The transfer syntax is echoed from the bind's presentation context, so the
// client sees the syntax it offered accepted rather than one it did not propose.
func rpcBindAck(pdu []byte) []byte {
	const (
		ptypeBind      = 11
		ptypeBindAck   = 12
		bindHeaderSize = 72
	)
	if len(pdu) < bindHeaderSize || pdu[0] != 5 || pdu[2] != ptypeBind {
		return nil
	}

	answer := []byte{5, 0, ptypeBindAck, 0x03, 0x10, 0, 0, 0}
	answer = append(answer, 0, 0)          // frag_length, filled in below
	answer = append(answer, 0, 0)          // auth_length
	answer = append(answer, pdu[12:16]...) // call_id, echoed
	answer = append(answer, pdu[16:18]...) // max_xmit_frag
	answer = append(answer, pdu[18:20]...) // max_recv_frag
	answer = append(answer, 0x34, 0x12, 0, 0)

	// sec_addr: a counted, null-terminated port name padded to a 4-byte boundary.
	port := []byte("\\PIPE\\srvsvc\x00")
	answer = append(answer, byte(len(port)), byte(len(port)>>8))
	answer = append(answer, port...)
	for len(answer)%4 != 0 {
		answer = append(answer, 0)
	}

	// p_result_list: one result, acceptance, echoing the offered syntax.
	answer = append(answer, 1, 0, 0, 0)
	answer = append(answer, 0, 0)          // ack_result = acceptance
	answer = append(answer, 0, 0)          // ack_reason
	answer = append(answer, pdu[52:72]...) // transfer syntax

	answer[8] = byte(len(answer))
	answer[9] = byte(len(answer) >> 8)
	return answer
}

// liveServer starts a server on the configured port with one disk share and one
// pipe share, and returns it with the pipe handler and the port.
func liveServer(t *testing.T, policy SigningPolicy) (*Server, *MemoryFileSystem, *livePipe, int) {
	t.Helper()

	srv, err := NewServer(Config{
		SigningPolicy: policy,
		Authenticator: StaticAccounts(Account{
			Domain:   liveDomain,
			Username: liveUsername,
			NTHash:   nt.NTHash(livePassword),
		}),
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("hello.txt", []byte("the quick brown fox\n")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddDirectory("subdir"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	if err := fs.AddFile("subdir/inner.txt", []byte("inner\n")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := srv.AddShare(&Share{
		Name: liveShare, Type: ShareTypeDisk, FS: fs,
		Security: NewReflectiveSecurityProvider(false),
	}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	pipes := &livePipe{}
	if err := srv.AddShare(&Share{Name: "IPC$", Type: ShareTypeNamedPipe, Pipes: pipes}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	port := livePort()
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	served := make(chan error, 1)
	go func() { served <- srv.ListenAndServe(address) }()

	// Wait for the listener rather than sleeping a fixed time, so a slow start
	// does not produce a spurious connection failure.
	deadline := time.Now().Add(5 * time.Second)
	for !srv.Listening() {
		if time.Now().After(deadline) {
			t.Fatalf("the server did not start listening on %s", address)
		}
		select {
		case err := <-served:
			t.Fatalf("ListenAndServe(%s) returned early: %v", address, err)
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Cleanup(func() { srv.Close() })
	return srv, fs, pipes, port
}

// runClient drives the third-party client with a semicolon-separated command list
// and returns its combined output.
//
// The dialect is pinned to NT1 in both directions: a modern client disables SMB1
// by default, so without this it negotiates nothing this server speaks.
func runClient(t *testing.T, binary string, port int, signing string, script string) string {
	t.Helper()

	arguments := []string{
		fmt.Sprintf("//127.0.0.1/%s", liveShare),
		"-U", liveUsername + "%" + livePassword,
		"-W", liveDomain,
		"-p", strconv.Itoa(port),
		"-m", "NT1",
		"--option=client min protocol=NT1",
		"--option=client max protocol=NT1",
	}
	if signing != "" {
		arguments = append(arguments, "--option=client signing="+signing)
	}
	arguments = append(arguments, "-c", script)

	command := exec.Command(binary, arguments...)
	command.Dir = t.TempDir()
	output, err := command.CombinedOutput()
	t.Logf("%s %v\nexit: %v\n%s", binary, arguments, err, output)
	return string(output)
}

// assertNoFailure asserts the client's output carries no status code, which is how
// these tools report a failure: they print the NT_STATUS and carry on.
//
// Named exceptions are allowed through, for the operations this server answers
// with a defined refusal rather than a result.
//
// It first asserts the client got as far as talking to the server at all. Without
// that this check passes for free whenever the client could not run — a wrong
// binary path or a refused connection produces output with no status code in it,
// which would otherwise read as success.
func assertNoFailure(t *testing.T, output string, allowed ...string) {
	t.Helper()

	if strings.TrimSpace(output) == "" {
		t.Fatal("the client produced no output at all; it probably never ran")
	}
	for _, fatal := range []string{
		"CONNECTION_REFUSED", "Connection to 127.0.0.1 failed",
		"NT_STATUS_LOGON_FAILURE", "session setup failed", "tree connect failed",
		"Unable to initialize messaging context", "Invalid option",
	} {
		if strings.Contains(output, fatal) {
			t.Fatalf("the client never reached a usable session (%q):\n%s", fatal, output)
		}
	}

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "NT_STATUS_") {
			continue
		}
		permitted := false
		for _, allow := range allowed {
			if strings.Contains(line, allow) {
				permitted = true
				break
			}
		}
		if !permitted {
			t.Errorf("the client reported a failure: %s", strings.TrimSpace(line))
		}
	}
}

// TestLiveClientListsAndTransfers is the milestone: a third-party client
// negotiates, authenticates, lists a directory and moves a file in both
// directions.
//
// The listing is asserted by name. That is the point of doing it here rather than
// in the unit suite: the names are written in whichever encoding the request
// declared, and a server that writes them in the other one produces a listing that
// is the right shape and the wrong text — which a round-trip against a client that
// shares the mistake cannot see.
func TestLiveClientListsAndTransfers(t *testing.T) {
	client := liveClient(t)
	_, _, _, port := liveServer(t, SigningDisabled)

	directory := t.TempDir()
	local := filepath.Join(directory, "upload.txt")
	payload := []byte("content that travelled over SMB1\n")
	if err := os.WriteFile(local, payload, 0o600); err != nil {
		t.Fatalf("writing the upload failed: %v", err)
	}

	output := runClient(t, client, port, "", fmt.Sprintf(
		"lcd %s; ls; get hello.txt fetched.txt; put upload.txt uploaded.txt; ls; du",
		directory))

	// No allowance: the whole session must be clean. A client asks about the
	// volume after a listing whether or not anything wanted it, so leaving that
	// unanswered puts an error in every session.
	assertNoFailure(t, output)

	for _, name := range []string{"hello.txt", "subdir", "uploaded.txt"} {
		if !strings.Contains(output, name) {
			t.Errorf("the listing does not name %q; the response encoding is probably wrong:\n%s", name, output)
		}
	}
}

// TestLiveClientFileOperations exercises the commands that carry a name in a
// request: create, rename, delete, and their directory equivalents.
//
// Rename is the one that matters most. It carries two names, and the first one's
// length decides whether an alignment byte precedes the second, so half of all
// renames take a code path the other half does not.
func TestLiveClientFileOperations(t *testing.T) {
	client := liveClient(t)
	_, fs, _, port := liveServer(t, SigningDisabled)

	output := runClient(t, client, port, "",
		"mkdir newdir; ls; rename hello.txt renamed.txt; ls; rmdir newdir; ls")
	assertNoFailure(t, output)

	// Asserted against the backend rather than the client's output: the name the
	// server actually stored is what proves the request decoded correctly.
	if _, err := fs.Stat("renamed.txt"); err != nil {
		t.Errorf("the renamed file is not present as %q: %v", "renamed.txt", err)
	}
	if _, err := fs.Stat("hello.txt"); err == nil {
		t.Error("the original name is still present after a rename")
	}
	if _, err := fs.Stat("newdir"); err == nil {
		t.Error("the directory is still present after rmdir")
	}
}

// TestLiveClientDeleteByExactName covers the search that a client performs before
// deleting: it names one file exactly rather than with a wildcard, which a server
// that treats every search path as a directory refuses.
func TestLiveClientDeleteByExactName(t *testing.T) {
	client := liveClient(t)
	_, fs, _, port := liveServer(t, SigningDisabled)

	output := runClient(t, client, port, "", "rm hello.txt; ls")
	assertNoFailure(t, output)

	if _, err := fs.Stat("hello.txt"); err == nil {
		t.Error("the file is still present after rm; the exact-name search was probably refused")
	}
}

// TestLiveClientSignedSession runs the whole exchange under mandatory signing,
// which validates this implementation's signing against an independent one in both
// directions: the server verifies every request and the client verifies every
// response.
func TestLiveClientSignedSession(t *testing.T) {
	client := liveClient(t)
	srv, _, _, port := liveServer(t, SigningRequired)

	directory := t.TempDir()
	output := runClient(t, client, port, "required", fmt.Sprintf(
		"lcd %s; ls; du; get hello.txt signed.txt", directory))
	assertNoFailure(t, output)

	// A client that could not verify a response says so rather than failing, so the
	// output is checked for it explicitly.
	if strings.Contains(output, "BAD SIG") || strings.Contains(strings.ToLower(output), "bad signature") {
		t.Errorf("the client rejected a response signature:\n%s", output)
	}

	fetched, err := os.ReadFile(filepath.Join(directory, "signed.txt"))
	if err != nil {
		t.Fatalf("the file did not arrive over the signed session: %v", err)
	}
	if !bytes.Equal(fetched, []byte("the quick brown fox\n")) {
		t.Errorf("the file arrived as %q over a signed session", fetched)
	}

	if srv.Config().SigningPolicy != SigningRequired {
		t.Fatal("the server was not configured to require signing")
	}
}

// TestLiveRPCClientBindsOverAPipe is what the named-pipe work exists for: a real
// RPC client completes a DCE/RPC bind over this server and goes on to make a call.
//
// The handler's calls are the assertion. They show the client opening the pipe by
// name, binding, issuing a request on the same handle, and closing it — the whole
// shape of MS-RPC over SMB1, performed by an implementation that knows nothing
// about this one.
func TestLiveRPCClientBindsOverAPipe(t *testing.T) {
	rpcClient := liveRPCClient(t)
	_, _, pipes, port := liveServer(t, SigningRequired)

	command := exec.Command(rpcClient,
		"//127.0.0.1",
		"-U", liveUsername+"%"+livePassword,
		"-W", liveDomain,
		"-p", strconv.Itoa(port),
		"-m", "NT1",
		"--option=client min protocol=NT1",
		"--option=client max protocol=NT1",
		"-c", "srvinfo")
	command.Dir = t.TempDir()
	output, err := command.CombinedOutput()
	t.Logf("%s\nexit: %v\n%s", rpcClient, err, output)

	if strings.Contains(string(output), "BAD SIG") {
		t.Errorf("the RPC client rejected a response signature:\n%s", output)
	}

	calls := pipes.Calls()
	t.Logf("pipe handler calls: %v", calls)

	// The call the whole path exists to carry. Its absence means the client never
	// got as far as a transaction.
	if !containsCall(calls, "bind:srvsvc") {
		t.Fatalf("the RPC client did not complete a bind over the pipe; handler saw %v", calls)
	}
	// A second transaction means the client accepted the bind_ack and considered
	// the binding established, which is the real assertion: a client that had
	// rejected it would have stopped.
	if !containsCall(calls, "transact:srvsvc") {
		t.Errorf("the client bound but made no call afterwards; handler saw %v", calls)
	}
	if !containsCall(calls, "open:srvsvc") {
		t.Errorf("the pipe was never opened by name; handler saw %v", calls)
	}
}

// TestLiveClientSurvivesPipeMisuse asserts a file operation on a pipe handle is
// refused rather than crashing the connection.
//
// A pipe handle has no file behind it, and an RPC client reads one directly when
// its transaction did not return a whole response. That reached an unguarded
// dereference and took the connection down with it.
func TestLiveClientSurvivesPipeMisuse(t *testing.T) {
	rpcClient := liveRPCClient(t)
	srv, _, _, port := liveServer(t, SigningRequired)

	command := exec.Command(rpcClient,
		"//127.0.0.1",
		"-U", liveUsername+"%"+livePassword,
		"-W", liveDomain,
		"-p", strconv.Itoa(port),
		"-m", "NT1",
		"--option=client min protocol=NT1",
		"--option=client max protocol=NT1",
		"-c", "srvinfo")
	command.Dir = t.TempDir()
	output, _ := command.CombinedOutput()
	t.Logf("%s", output)

	// The server must still be serving: a panic is contained to one connection, so
	// the listener surviving is what says nothing fatal happened.
	if !srv.Listening() {
		t.Fatal("the server stopped listening after the RPC exchange")
	}

	// And a plain client must still be able to connect afterwards.
	client := liveClient(t)
	after := runClient(t, client, port, "required", "ls")
	assertNoFailure(t, after)
}

// containsCall reports whether a handler call was seen.
func containsCall(calls []string, want string) bool {
	for _, call := range calls {
		if call == want {
			return true
		}
	}
	return false
}
