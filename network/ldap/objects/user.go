package objects

type User struct {
	// LdapSession is the LDAP session object
	LdapSession LdapSessionInterface
	// DistinguishedName is the distinguished name of the user
	DistinguishedName string
	// sAMAccountName is the sAMAccountName of the user
	SamAccountName string
}
