package kerberos

import (
	"fmt"
	"strings"
	"time"

	"github.com/jcmturner/gokrb5/v8/config"
)

// KerberosInit initializes the Kerberos configuration and service principal name for LDAP authentication.
//
// Parameters:
// - fqdnLDAPHost: A string representing the fully qualified domain name of the LDAP server.
// - fqndRealm: A string representing the fully qualified domain name of the realm.
//
// Returns:
// - A string representing the service principal name for LDAP authentication.
// - A pointer to the Kerberos configuration.
func KerberosInit(fqdnLDAPHost, fqndRealm string) (string, *config.Config) {
	servicePrincipalName := fmt.Sprintf("ldap/%s", fqdnLDAPHost)

	fqndRealm = strings.ToUpper(fqndRealm)
	// This is always in uppercase, if not we get the error:
	// error performing GSSAPI bind: [Root cause: KRBMessage_Handling_Error]
	// | KRBMessage_Handling_Error: AS Exchange Error: AS_REP is not valid or client password/keytab incorrect
	// |  | KRBMessage_Handling_Error: CRealm in response does not match what was requested.
	// |  |  | Requested: lab.local;
	// |  |  | Reply: lab.local
	// | 2024/10/08 15:36:16 error querying AD: LDAP Result Code 1 "Operations Error": 000004DC: LdapErr: DSID-0C090A5C,
	// | comment: In order to perform this operation a successful bind must be completed on the connection., data 0, v4563

	krb5Conf := config.New()
	// LibDefaults
	krb5Conf.LibDefaults.AllowWeakCrypto = false
	krb5Conf.LibDefaults.DefaultRealm = fqndRealm
	krb5Conf.LibDefaults.DNSLookupRealm = false
	krb5Conf.LibDefaults.DNSLookupKDC = false
	krb5Conf.LibDefaults.TicketLifetime = time.Duration(24) * time.Hour
	krb5Conf.LibDefaults.RenewLifetime = time.Duration(24*7) * time.Hour
	krb5Conf.LibDefaults.Forwardable = true
	krb5Conf.LibDefaults.Proxiable = true
	krb5Conf.LibDefaults.RDNS = false
	krb5Conf.LibDefaults.UDPPreferenceLimit = 1 // Force use of tcp
	krb5Conf.LibDefaults.DefaultTGSEnctypes = []string{"aes256-cts-hmac-sha1-96", "aes128-cts-hmac-sha1-96", "arcfour-hmac-md5"}
	krb5Conf.LibDefaults.DefaultTktEnctypes = []string{"aes256-cts-hmac-sha1-96", "aes128-cts-hmac-sha1-96", "arcfour-hmac-md5"}
	krb5Conf.LibDefaults.PermittedEnctypes = []string{"aes256-cts-hmac-sha1-96", "aes128-cts-hmac-sha1-96", "arcfour-hmac-md5"}
	krb5Conf.LibDefaults.PermittedEnctypeIDs = []int32{18, 17, 23}
	krb5Conf.LibDefaults.DefaultTGSEnctypeIDs = []int32{18, 17, 23}
	krb5Conf.LibDefaults.DefaultTktEnctypeIDs = []int32{18, 17, 23}
	krb5Conf.LibDefaults.PreferredPreauthTypes = []int{18, 17, 23}

	// Realms
	krb5Conf.Realms = append(krb5Conf.Realms, config.Realm{
		Realm:         fqndRealm,
		AdminServer:   []string{fqdnLDAPHost},
		DefaultDomain: fqndRealm,
		KDC:           []string{fmt.Sprintf("%s:88", fqdnLDAPHost)},
		KPasswdServer: []string{fmt.Sprintf("%s:464", fqdnLDAPHost)},
		MasterKDC:     []string{fqdnLDAPHost},
	})

	// Domain Realm
	krb5Conf.DomainRealm[strings.ToLower(fqndRealm)] = fqndRealm
	krb5Conf.DomainRealm[fmt.Sprintf(".%s", strings.ToLower(fqndRealm))] = fqndRealm

	return servicePrincipalName, krb5Conf
}
