package objects

import (
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/go-ldap/ldap/v3"
)

type LdapSessionInterface interface {
	InitSession(string, int, *credentials.Credentials, bool, bool) error

	Connect() (bool, error)

	ReConnect() (bool, error)

	Close()

	// Query functions
	Query(searchBase string, query string, attributes []string, scope int) ([]*ldap.Entry, error)

	QueryWholeSubtree(searchBase string, query string, attributes []string) ([]*ldap.Entry, error)

	QueryBaseObject(searchBase string, query string, attributes []string) ([]*ldap.Entry, error)

	QuerySingleLevel(searchBase string, query string, attributes []string) ([]*ldap.Entry, error)

	QueryChildren(searchBase string, query string, attributes []string) ([]*ldap.Entry, error)

	// Domain functions
	GetDomain(distinguishedName string) (*Domain, error)
}
