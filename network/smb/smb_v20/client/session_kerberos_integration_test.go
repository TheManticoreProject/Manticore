//go:build integration

// Live integration coverage for the SMB2 Kerberos session setup
// (SessionSetupKerberos) against a real file server / domain controller,
// exercising both SMB 2.x signing (HMAC-SHA256) and SMB 3.1.1 signing
// (AES-128-CMAC) plus AES-GCM encryption. Excluded from the default build by the
// "integration" tag; skipped unless the environment is configured.
//
// Configuration:
//
//	KRB5_TEST_KDC     KDC / domain-controller host or IP (obtains the TGT)
//	KRB5_TEST_REALM   Kerberos realm / AD domain
//	KRB5_TEST_USER    account sAMAccountName
//	KRB5_TEST_PASS    account password
//	KRB5_TEST_TARGET  SMB server FQDN (defaults to KRB5_TEST_KDC). The SPN used is
//	                  cifs/<target>; the transport dials the resolved IP on 445.
//
// Example:
//
//	KRB5_TEST_KDC=10.0.0.10 KRB5_TEST_REALM=EXAMPLE.LOCAL \
//	KRB5_TEST_USER=Administrator KRB5_TEST_PASS='…' \
//	KRB5_TEST_TARGET=dc.example.local \
//	  go test -tags integration -v -run TestLiveSMB ./network/smb/smb_v20/client/
package client

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

type smbEnv struct {
	KDC    string
	Realm  string
	User   string
	Pass   string
	Target string // SMB server FQDN (for the cifs/ SPN)
	IP     net.IP // resolved target address for the transport
}

func requireSMBEnv(t *testing.T) smbEnv {
	t.Helper()
	e := smbEnv{
		KDC:    os.Getenv("KRB5_TEST_KDC"),
		Realm:  os.Getenv("KRB5_TEST_REALM"),
		User:   os.Getenv("KRB5_TEST_USER"),
		Pass:   os.Getenv("KRB5_TEST_PASS"),
		Target: os.Getenv("KRB5_TEST_TARGET"),
	}
	if e.KDC == "" || e.Realm == "" || e.User == "" || e.Pass == "" {
		t.Skip("set KRB5_TEST_KDC/KRB5_TEST_REALM/KRB5_TEST_USER/KRB5_TEST_PASS to run the live SMB Kerberos tests")
	}
	if e.Target == "" {
		e.Target = e.KDC
	}
	addr, err := net.ResolveIPAddr("ip", e.Target)
	if err != nil {
		t.Fatalf("resolve target %q: %v", e.Target, err)
	}
	e.IP = addr.IP
	return e
}

func (e smbEnv) creds(t *testing.T) *credentials.Credentials {
	t.Helper()
	creds, err := credentials.NewCredentials(e.Realm, e.User, e.Pass, "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	return creds
}

// TestLiveSMB_Kerberos_SMB311_SignedEncrypted negotiates SMB 3.1.1, authenticates
// with Kerberos, performs a signed tree connect, then enables SMB3 encryption and
// performs an AES-GCM-encrypted tree connect.
func TestLiveSMB_Kerberos_SMB311_SignedEncrypted(t *testing.T) {
	e := requireSMBEnv(t)

	c := NewClientUsingTCPTransport(e.IP, 445)
	if err := c.Connect(e.IP, 445); err != nil {
		t.Fatalf("Connect/Negotiate: %v", err)
	}
	defer c.Disconnect()

	if c.Connection.Dialect != dialects.SMB2_DIALECT_3_1_1 {
		t.Fatalf("expected SMB 3.1.1 dialect, negotiated %s", c.Connection.Dialect)
	}

	spn := "cifs/" + e.Target
	if err := c.SessionSetupKerberos(e.creds(t), e.KDC, spn); err != nil {
		t.Fatalf("SessionSetupKerberos: %v", err)
	}
	if !c.Session.SigningActive {
		t.Fatal("signing not active after SMB 3.1.1 Kerberos session setup")
	}

	// Signed tree connect (AES-128-CMAC): the response signature is verified in
	// sendReceive, so a successful connect proves the derived signing key.
	if err := c.TreeConnect("IPC$"); err != nil {
		t.Fatalf("signed TreeConnect(IPC$): %v", err)
	}
	t.Logf("[ok] SMB 3.1.1 signed tree connect to IPC$")

	// Enable SMB3 AES-GCM encryption and perform an encrypted exchange: the tree
	// connect request is wrapped in a TRANSFORM_HEADER and the response is
	// decrypted and authenticated by the AEAD tag.
	if err := c.EnableEncryption(); err != nil {
		t.Fatalf("EnableEncryption: %v", err)
	}
	if !c.IsEncryptionActive() {
		t.Fatal("encryption not active after EnableEncryption")
	}
	if err := c.TreeConnect("IPC$"); err != nil {
		t.Fatalf("encrypted TreeConnect(IPC$): %v", err)
	}
	t.Logf("[ok] SMB 3.1.1 AES-GCM encrypted tree connect to IPC$")
}

// TestLiveSMB_Kerberos_SMB21_Signed negotiates only up to SMB 2.1, authenticates
// with Kerberos, and performs a signed (HMAC-SHA256) tree connect.
//
// Skipped unless KRB5_TEST_SMB21 is set: a modern Windows domain controller
// rejects SMB 2.0.2/2.1 sessions outright (STATUS_USER_SESSION_DELETED — the same
// for NTLM and Kerberos), so this path can only be validated against a member
// server or share that still permits the 2.x dialect family. The 2.x signing key
// derivation itself (session key == HMAC-SHA256 signing key) is covered by the
// package unit tests.
func TestLiveSMB_Kerberos_SMB21_Signed(t *testing.T) {
	e := requireSMBEnv(t)
	if os.Getenv("KRB5_TEST_SMB21") == "" {
		t.Skip("set KRB5_TEST_SMB21=1 (and target a server that permits SMB 2.x) to run the SMB 2.1 signing test; DCs reject 2.x")
	}

	c := NewClientUsingTCPTransport(e.IP, 445)
	if err := c.Transport.Connect(e.IP, 445); err != nil {
		t.Fatalf("transport connect: %v", err)
	}
	defer c.Disconnect()

	if err := negotiate21(c); err != nil {
		t.Fatalf("SMB 2.1 negotiate: %v", err)
	}
	if c.Connection.Dialect != dialects.SMB2_DIALECT_2_1_0 {
		t.Fatalf("expected SMB 2.1 dialect, negotiated %s", c.Connection.Dialect)
	}

	spn := "cifs/" + e.Target
	if err := c.SessionSetupKerberos(e.creds(t), e.KDC, spn); err != nil {
		t.Fatalf("SessionSetupKerberos: %v", err)
	}
	if !c.Session.SigningActive {
		t.Fatal("signing not active after SMB 2.1 Kerberos session setup")
	}

	if err := c.TreeConnect("IPC$"); err != nil {
		t.Fatalf("signed TreeConnect(IPC$): %v", err)
	}
	t.Logf("[ok] SMB 2.1 HMAC-SHA256 signed tree connect to IPC$")
}

// negotiate21 performs an SMB2 NEGOTIATE offering only 2.0.2 and 2.1, forcing the
// server to select the SMB 2.x dialect family (the client's public Negotiate
// always offers up to 3.1.1, which an AD server would prefer). It mirrors the
// non-3.1.1 portion of Negotiate.
func negotiate21(c *Client) error {
	req := commands.NewNegotiateRequest()
	req.AddDialect(dialects.SMB2_DIALECT_2_0_2)
	req.AddDialect(dialects.SMB2_DIALECT_2_1_0)
	req.ClientGuid = c.ClientGuid
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.Capabilities = capabilities.SMB2_GLOBAL_CAP_LARGE_MTU

	// The 2.x dialects carry no pre-auth integrity; keep the buffer non-nil so the
	// (unused for 2.x) hash-folding in the session setup is harmless.
	c.Connection.PreauthIntegrityHashValue = make([]byte, preauthHashLength)

	msg := c.newRequest(req)
	msg.Header.SessionId = 0

	response, err := c.sendReceive(msg, "Negotiate21")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("negotiate failed: 0x%08x", status)
	}
	negotiateResponse, ok := response.Command.(*commands.NegotiateResponse)
	if !ok {
		return fmt.Errorf("unexpected negotiate response command: %T", response.Command)
	}
	c.ApplyNegotiateResponse(negotiateResponse)
	return nil
}
