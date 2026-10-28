//go:build integration

// Live interoperability with a third-party client. The unit suite pairs the
// server with this repository's client, which shares its assumptions; this one
// drives smbclient instead.
//
//	SMB2_TEST_CLIENT  path to smbclient. Every test skips when it is unset.
package server_test

import (
	"bytes"
	"fmt"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// smbclient runs one smbclient command line against the live server.
func smbclient(t *testing.T, port int, args ...string) string {
	t.Helper()
	binary := os.Getenv("SMB2_TEST_CLIENT")
	if binary == "" {
		t.Skip("SMB2_TEST_CLIENT is not set; no third-party client to interoperate with")
	}
	args = append([]string{"//127.0.0.1/files", "-p", strconv.Itoa(port), "-U", `LAB\alice%Passw0rd!`}, args...)
	cmd := exec.Command(binary, args...)
	cmd.WaitDelay = 10 * time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("smbclient %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func TestLiveSmbclient(t *testing.T) {
	for _, tc := range []struct {
		name     string
		protocol string
		config   server.Config
		extra    []string
	}{
		{"SMB 2.0.2", "SMB2_02", server.Config{}, nil},
		{"SMB 2.1", "SMB2_10", server.Config{}, nil},
		{"SMB 3.0", "SMB3_00", server.Config{}, nil},
		{"SMB 3.0.2", "SMB3_02", server.Config{}, nil},
		{"SMB 3.1.1", "SMB3_11", server.Config{}, nil},
		{"SMB1 upgrade to 3.1.1", "SMB3_11", server.Config{}, []string{"--option=clientminprotocol=NT1"}},
		{"SMB1 upgrade to 2.0.2", "SMB2_02", server.Config{}, []string{"--option=clientminprotocol=NT1"}},
		{"SMB 2.1 signed", "SMB2_10", server.Config{RequireSigning: true}, []string{"--client-protection=sign"}},
		{"SMB 3.0 signed", "SMB3_00", server.Config{RequireSigning: true}, []string{"--client-protection=sign"}},
		{"SMB 3.1.1 signed", "SMB3_11", server.Config{RequireSigning: true}, []string{"--client-protection=sign"}},
		{"SMB 3.0 encrypted", "SMB3_00", server.Config{RequireEncryption: true}, []string{"--client-protection=encrypt"}},
		{"SMB 3.1.1 encrypted", "SMB3_11", server.Config{RequireEncryption: true}, []string{"--client-protection=encrypt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, port := startServer(t, tc.config)
			local := filepath.Join(t.TempDir(), "upload.bin")
			payload := bytes.Repeat([]byte("manticore"), 20000)
			if err := os.WriteFile(local, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			downloaded := filepath.Join(t.TempDir(), "download.bin")

			args := append([]string{"-m", tc.protocol}, tc.extra...)
			script := fmt.Sprintf("ls; put %s upload.bin; get upload.bin %s; mkdir sub; rename upload.bin sub\\moved.bin; ls sub\\*; del sub\\moved.bin; rmdir sub; allinfo hello.txt", local, downloaded)
			out := smbclient(t, port, append(args, "-c", script)...)

			for _, line := range strings.Split(out, "\n") {
				// Snapshots are not served, and STATUS_NOT_SUPPORTED says so.
				if strings.Contains(line, "NT_STATUS_") && !strings.Contains(line, "shadow copy") {
					t.Errorf("smbclient reported an error: %s", line)
				}
			}
			if !strings.Contains(out, "hello.txt") {
				t.Errorf("listing does not show hello.txt:\n%s", out)
			}
			if !strings.Contains(out, "moved.bin") {
				t.Errorf("listing of sub does not show moved.bin:\n%s", out)
			}
			if got, err := os.ReadFile(downloaded); err != nil || !bytes.Equal(got, payload) {
				t.Errorf("downloaded %d bytes, want %d (%v)", len(got), len(payload), err)
			}
		})
	}
}
