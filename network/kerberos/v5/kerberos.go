// Package kerberos provides Kerberos authentication primitives for Active Directory.
// It includes a native client (KerberosClient), ASREPRoast, and a gokrb5-backed
// helper (KerberosInit) used for LDAP GSSAPI binds.
package kerberos

import (
	"fmt"
	"strings"
	"time"

	"github.com/jcmturner/gokrb5/v8/config"
)

// KerberosInit initialises a gokrb5 configuration and returns the LDAP service
// principal name for the given host and realm.
//
// It is used by the LDAP session layer to perform GSSAPI Kerberos binds via
// the gokrb5 library. The native KerberosClient does not depend on this function.
//
// Parameters:
//   - fqdnLDAPHost: Fully qualified domain name (or IP) of the KDC / LDAP server.
//   - fqndRealm:    Kerberos realm (will be uppercased automatically).
//
// Returns the service principal name ("ldap/<fqdnLDAPHost>") and a ready-to-use
// *config.Config.
func KerberosInit(fqdnLDAPHost, fqndRealm string) (string, *config.Config) {
	servicePrincipalName := fmt.Sprintf("ldap/%s", fqdnLDAPHost)

	// Realm must always be upper-cased; a mismatch causes:
	// "CRealm in response does not match what was requested"
	fqndRealm = strings.ToUpper(fqndRealm)

	krb5Conf := config.New()

	// [libdefaults]
	krb5Conf.LibDefaults.AllowWeakCrypto = false
	krb5Conf.LibDefaults.DefaultRealm = fqndRealm
	krb5Conf.LibDefaults.DNSLookupRealm = false
	krb5Conf.LibDefaults.DNSLookupKDC = false
	krb5Conf.LibDefaults.TicketLifetime = 24 * time.Hour
	krb5Conf.LibDefaults.RenewLifetime = 24 * 7 * time.Hour
	krb5Conf.LibDefaults.Forwardable = true
	krb5Conf.LibDefaults.Proxiable = true
	krb5Conf.LibDefaults.RDNS = false
	krb5Conf.LibDefaults.UDPPreferenceLimit = 1 // Force TCP
	krb5Conf.LibDefaults.DefaultTGSEnctypes = []string{
		"aes256-cts-hmac-sha1-96", "aes128-cts-hmac-sha1-96", "arcfour-hmac-md5",
	}
	krb5Conf.LibDefaults.DefaultTktEnctypes = []string{
		"aes256-cts-hmac-sha1-96", "aes128-cts-hmac-sha1-96", "arcfour-hmac-md5",
	}
	krb5Conf.LibDefaults.PermittedEnctypes = []string{
		"aes256-cts-hmac-sha1-96", "aes128-cts-hmac-sha1-96", "arcfour-hmac-md5",
	}
	krb5Conf.LibDefaults.PermittedEnctypeIDs = []int32{18, 17, 23}
	krb5Conf.LibDefaults.DefaultTGSEnctypeIDs = []int32{18, 17, 23}
	krb5Conf.LibDefaults.DefaultTktEnctypeIDs = []int32{18, 17, 23}
	krb5Conf.LibDefaults.PreferredPreauthTypes = []int{18, 17, 23}

	// [realms]
	krb5Conf.Realms = append(krb5Conf.Realms, config.Realm{
		Realm:         fqndRealm,
		AdminServer:   []string{fqdnLDAPHost},
		DefaultDomain: fqndRealm,
		KDC:           []string{fmt.Sprintf("%s:88", fqdnLDAPHost)},
		KPasswdServer: []string{fmt.Sprintf("%s:464", fqdnLDAPHost)},
		MasterKDC:     []string{fqdnLDAPHost},
	})

	// [domain_realm]
	krb5Conf.DomainRealm[strings.ToLower(fqndRealm)] = fqndRealm
	krb5Conf.DomainRealm[fmt.Sprintf(".%s", strings.ToLower(fqndRealm))] = fqndRealm

	return servicePrincipalName, krb5Conf
}
