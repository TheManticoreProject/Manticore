//go:build integration

package rpcpipe

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
)

// TestImpacketCallsOverPipeWriteAndRead has impacket, a third-party client,
// call srvsvc and wkssvc on this server. impacket sends every PDU with
// WRITE_ANDX and reads the answer with READ_ANDX rather than transacting, so the
// test covers that path end to end. It is skipped when python3 cannot import
// impacket.
func TestImpacketCallsOverPipeWriteAndRead(t *testing.T) {
	if err := exec.Command("python3", "-c", "import impacket").Run(); err != nil {
		t.Skipf("impacket is not available: %v", err)
	}

	port := startPipeServer(t)

	for _, fragment := range []string{"0", "40", "64", "100"} {
		out, err := exec.Command("python3", "testdata/impacket_pipe.py", strconv.Itoa(port), "LAB", "alice", "Passw0rd!", fragment).CombinedOutput()
		t.Logf("max fragment %s:\n%s", fragment, out)
		if err != nil || !strings.Contains(string(out), "OK") {
			t.Fatalf("impacket failed with max fragment %s: %v", fragment, err)
		}
		for _, want := range []string{"share files", "share IPC$", "share share299", "computer MANTICORE", "domain LAB"} {
			if !strings.Contains(string(out), want) {
				t.Errorf("max fragment %s: output lacks %q", fragment, want)
			}
		}
	}
}

// startPipeServer serves rpcpipe on IPC$ with 302 shares, enough for the
// NetrShareEnum answer to span many fragments, and returns the port.
func startPipeServer(t *testing.T) int {
	t.Helper()
	srv, err := server.NewServer(server.Config{
		SigningPolicy:   server.SigningEnabled,
		ServerName:      "MANTICORE",
		DomainName:      "LAB",
		DNSComputerName: "manticore.lab.local",
		DNSDomainName:   "lab.local",
		Authenticator:   server.StaticAccounts(server.Account{Domain: "LAB", Username: "alice", NTHash: nt.NTHash("Passw0rd!")}),
	})
	if err != nil {
		t.Fatal(err)
	}
	pipes := New(Options{
		ServerName: "MANTICORE",
		DomainName: "LAB",
		// 300 shares make the NetrShareEnum answer span many fragments, which a
		// message-mode pipe hands over one read at a time.
		Shares: func() []ShareEntry {
			shares := []ShareEntry{{Name: "files", Comment: "test share"}, {Name: "IPC$", Type: 3}}
			for i := 0; i < 300; i++ {
				shares = append(shares, ShareEntry{Name: fmt.Sprintf("share%03d", i), Comment: "padding to span several fragments"})
			}
			return shares
		},
	})
	if err := srv.AddShare(&server.Share{Name: "IPC$", Type: server.ShareTypeNamedPipe, Pipes: pipes}); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	go srv.ListenAndServe(net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	for deadline := time.Now().Add(5 * time.Second); !srv.Listening() && time.Now().Before(deadline); {
		time.Sleep(10 * time.Millisecond)
	}
	t.Cleanup(func() { srv.Close() })

	return port
}

// TestRPCClientEnumeratesManySharesOverTransact has Samba's rpcclient, which
// transacts and then reads the rest, enumerate 302 shares: an answer of many
// fragments delivered one message per read. It is skipped without rpcclient.
func TestRPCClientEnumeratesManySharesOverTransact(t *testing.T) {
	rpcclient, err := exec.LookPath("rpcclient")
	if err != nil {
		t.Skip("rpcclient is not available")
	}
	port := startPipeServer(t)
	out, err := exec.Command(rpcclient, "//127.0.0.1", "-U", "alice%Passw0rd!", "-W", "LAB", "-p", strconv.Itoa(port),
		"-m", "NT1", "--option=client min protocol=NT1", "--option=client max protocol=NT1",
		"-c", "netshareenumall 1").CombinedOutput()
	t.Logf("rpcclient: %v\n%s", err, out)
	for _, want := range []string{"netname: files", "netname: share000", "netname: share299"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output lacks %q", want)
		}
	}
}
