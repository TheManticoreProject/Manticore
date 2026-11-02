//go:build integration

// Live integration coverage for the SMB1 Kerberos session setup
// (SessionSetupKerberos) against a real Windows server with SMB1 enabled and
// signing required. Excluded from the default build by the "integration" tag;
// skipped unless the environment is configured.
//
// Configuration:
//
//	KRB5_TEST_KDC     KDC / domain-controller host or IP (obtains the TGT)
//	KRB5_TEST_REALM   Kerberos realm / AD domain
//	KRB5_TEST_USER    account sAMAccountName
//	KRB5_TEST_PASS    account password
//	KRB5_TEST_TARGET  SMB server FQDN (defaults to KRB5_TEST_KDC). The SPN used is
//	                  cifs/<target>; the transport dials the resolved IP on 445.
//	KRB5_TEST_IP      optional address to dial instead of resolving the target
//
// Example:
//
//	KRB5_TEST_KDC=10.0.0.10 KRB5_TEST_REALM=EXAMPLE.LOCAL \
//	KRB5_TEST_USER=Administrator KRB5_TEST_PASS='…' \
//	KRB5_TEST_TARGET=dc.example.local \
//	  go test -tags integration -v -run TestLiveSMB1_Kerberos ./network/smb/smb_v10/client/
package client

import (
	"net"
	"os"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

type smb1KerberosEnv struct {
	KDC, Realm, User, Pass, Target string
	IP                             net.IP
}

func requireSMB1KerberosEnv(t *testing.T) smb1KerberosEnv {
	t.Helper()
	e := smb1KerberosEnv{
		KDC:    os.Getenv("KRB5_TEST_KDC"),
		Realm:  os.Getenv("KRB5_TEST_REALM"),
		User:   os.Getenv("KRB5_TEST_USER"),
		Pass:   os.Getenv("KRB5_TEST_PASS"),
		Target: os.Getenv("KRB5_TEST_TARGET"),
	}
	if e.KDC == "" || e.Realm == "" || e.User == "" || e.Pass == "" {
		t.Skip("set KRB5_TEST_KDC/KRB5_TEST_REALM/KRB5_TEST_USER/KRB5_TEST_PASS to run the live SMB1 Kerberos tests")
	}
	if e.Target == "" {
		e.Target = e.KDC
	}
	if ip := net.ParseIP(os.Getenv("KRB5_TEST_IP")); ip != nil {
		e.IP = ip
		return e
	}
	addr, err := net.ResolveIPAddr("ip", e.Target)
	if err != nil {
		t.Fatalf("resolve target %q: %v", e.Target, err)
	}
	e.IP = addr.IP
	return e
}

// TestLiveSMB1_Kerberos_Signed authenticates with Kerberos from a password and
// from an NT hash, then performs signed tree connect, open, read and logoff. A
// signing-required server drops the connection on a bad signature and every
// response signature is verified, so the exchange proves the signing key taken
// from the AP-REP.
func TestLiveSMB1_Kerberos_Signed(t *testing.T) {
	e := requireSMB1KerberosEnv(t)
	for _, tc := range []struct{ name, pass, hashes string }{
		{"password", e.Pass, ""},
		{"nthash", "", ":" + nt.NTHashHex(e.Pass)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			creds, err := credentials.NewCredentials(e.Realm, e.User, tc.pass, tc.hashes)
			if err != nil {
				t.Fatalf("NewCredentials: %v", err)
			}
			c := NewClientUsingTCPTransport(e.IP, 445)
			if err := c.Connect(e.IP, 445); err != nil {
				t.Fatalf("Connect/Negotiate: %v", err)
			}
			defer c.Disconnect()

			if err := c.SessionSetupKerberos(creds, e.KDC, "cifs/"+e.Target); err != nil {
				t.Fatalf("SessionSetupKerberos: %v", err)
			}
			if c.Connection.Server.SigningState != SigningStateDisabled && !c.Connection.IsSigningActive {
				t.Fatal("signing not active after Kerberos session setup on a signing server")
			}
			if err := c.TreeConnect("C$"); err != nil {
				t.Fatalf("TreeConnect(C$): %v", err)
			}
			fid, err := c.OpenFile(`\Windows\win.ini`, fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
			if err != nil {
				t.Fatalf("OpenFile: %v", err)
			}
			data, err := c.ReadFile(fid, 0, 4096)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			c.CloseFile(fid)
			if len(data) == 0 {
				t.Error("read nothing from win.ini")
			}
			if err := c.Logoff(); err != nil {
				t.Errorf("Logoff: %v", err)
			}
			t.Logf("[ok] SMB1 Kerberos (%s): signing=%v, read %d bytes", tc.name, c.Connection.IsSigningActive, len(data))
		})
	}
}
