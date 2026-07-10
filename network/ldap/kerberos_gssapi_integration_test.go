//go:build integration

// Live integration coverage for the native (stdlib-only) GSSAPI/Kerberos LDAP
// bind against a real domain controller, with the RFC 4752 integrity (signing)
// and confidentiality (sealing) security layers, each followed by a real search.
// Excluded from the default build by the "integration" tag; skipped unless the
// environment is configured.
//
// Configuration:
//
//	KRB5_TEST_KDC     KDC / domain-controller host or IP
//	KRB5_TEST_REALM   Kerberos realm / AD domain
//	KRB5_TEST_USER    account sAMAccountName
//	KRB5_TEST_PASS    account password
//	KRB5_TEST_TARGET  LDAP server FQDN (defaults to KRB5_TEST_KDC). The DC is both
//	                  the LDAP server and the KDC; the SPN is ldap/<target>, so a
//	                  fully-qualified name is required.
//
// Example:
//
//	KRB5_TEST_KDC=10.0.0.10 KRB5_TEST_REALM=EXAMPLE.LOCAL \
//	KRB5_TEST_USER=Administrator KRB5_TEST_PASS='…' \
//	KRB5_TEST_TARGET=dc.example.local \
//	  go test -tags integration -v -run TestLiveLDAP ./network/ldap/
package ldap_test

import (
	"os"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

type ldapEnv struct {
	Realm  string
	User   string
	Pass   string
	Target string
}

func requireLDAPEnv(t *testing.T) ldapEnv {
	t.Helper()
	kdc := os.Getenv("KRB5_TEST_KDC")
	e := ldapEnv{
		Realm:  os.Getenv("KRB5_TEST_REALM"),
		User:   os.Getenv("KRB5_TEST_USER"),
		Pass:   os.Getenv("KRB5_TEST_PASS"),
		Target: os.Getenv("KRB5_TEST_TARGET"),
	}
	if kdc == "" || e.Realm == "" || e.User == "" || e.Pass == "" {
		t.Skip("set KRB5_TEST_KDC/KRB5_TEST_REALM/KRB5_TEST_USER/KRB5_TEST_PASS to run the live LDAP Kerberos tests")
	}
	if e.Target == "" {
		e.Target = kdc
	}
	return e
}

func (e ldapEnv) creds(t *testing.T) *credentials.Credentials {
	t.Helper()
	creds, err := credentials.NewCredentials(e.Realm, e.User, e.Pass, "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	return creds
}

// gssapiBindAndSearch binds to LDAP with GSSAPI/Kerberos using the requested
// security layer (applied via layer), then performs a real search over the
// (signed and/or sealed) connection: it reads the naming contexts from the
// RootDSE and does a base-object search on the default naming context.
func gssapiBindAndSearch(t *testing.T, e ldapEnv, layer func(*ldap.Session)) {
	t.Helper()
	s, err := ldap.NewSession(e.Target, 389, e.creds(t), false, true)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if layer != nil {
		layer(s)
	}
	ok, err := s.Connect()
	if err != nil || !ok {
		t.Fatalf("GSSAPI Connect: ok=%v err=%v", ok, err)
	}
	defer s.Close()

	ncs, err := s.GetAllNamingContexts()
	if err != nil {
		t.Fatalf("GetAllNamingContexts over GSSAPI layer: %v", err)
	}
	if len(ncs) == 0 {
		t.Fatal("no naming contexts returned")
	}

	entries, err := s.QueryBaseObject("defaultNamingContext", "(objectClass=*)", []string{"distinguishedName", "objectSid"})
	if err != nil {
		t.Fatalf("base-object search over GSSAPI layer: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("base-object search returned no entries")
	}
	t.Logf("[ok] GSSAPI bind + search: %d naming contexts, base DN %q", len(ncs), entries[0].DN)
}

// TestLiveLDAP_GSSAPI_Sign binds with the GSSAPI integrity (signing) layer and
// searches; every LDAP PDU after the bind is GSS-signed and verified.
func TestLiveLDAP_GSSAPI_Sign(t *testing.T) {
	e := requireLDAPEnv(t)
	gssapiBindAndSearch(t, e, func(s *ldap.Session) { s.SetGSSAPISigning() })
}

// TestLiveLDAP_GSSAPI_Seal binds with the GSSAPI confidentiality (sealing) layer
// and searches; every LDAP PDU after the bind is GSS-encrypted.
func TestLiveLDAP_GSSAPI_Seal(t *testing.T) {
	e := requireLDAPEnv(t)
	gssapiBindAndSearch(t, e, func(s *ldap.Session) { s.SetGSSAPISealing() })
}

// TestLiveLDAP_GSSAPI_AuthOnly binds with GSSAPI and no security layer (the
// historical default) and searches, confirming the auth-only path also works.
func TestLiveLDAP_GSSAPI_AuthOnly(t *testing.T) {
	e := requireLDAPEnv(t)
	gssapiBindAndSearch(t, e, nil)
}
