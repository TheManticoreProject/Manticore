package ldap

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type Entry ldap.Entry

const (
	ScopeBaseObject   = ldap.ScopeBaseObject
	ScopeSingleLevel  = ldap.ScopeSingleLevel
	ScopeChildren     = ldap.ScopeChildren
	ScopeWholeSubtree = ldap.ScopeWholeSubtree
)

// GetRootDSE retrieves the Root DSE (Directory Service Entry) from the LDAP server.
// The Root DSE provides information about the LDAP server, including its capabilities,
// naming contexts, and supported controls.
//
// This function performs a base object search with a filter of "(objectClass=*)"
// to retrieve all attributes of the Root DSE.
//
// Returns:
// - A pointer to an ldap.Entry object representing the Root DSE.
// - If an error occurs during the search, the function logs a warning and returns nil.
//
// Example usage:
//
//	rootDSE := ldapSession.GetRootDSE()
//	if rootDSE != nil {
//	    fmt.Println("Root DSE retrieved successfully")
//	} else {
//	    fmt.Println("Failed to retrieve Root DSE")
//	}
func (ldapSession *Session) GetRootDSE() (*ldap.Entry, error) {
	// Specify LDAP search parameters
	// https://pkg.go.dev/gopkg.in/ldap.v3#NewSearchRequest
	searchRequest := ldap.NewSearchRequest(
		// Base DN blank
		"",
		// Scope Base
		ldap.ScopeBaseObject,
		// DerefAliases
		ldap.NeverDerefAliases,
		// SizeLimit
		1,
		// TimeLimit
		0,
		// TypesOnly
		false,
		// Search filter
		"(objectClass=*)",
		// Attributes to retrieve
		[]string{"*"},
		// Controls
		nil,
	)

	// Perform LDAP search
	searchResult, err := ldapSession.connection.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching LDAP: %w", err)
	}

	return searchResult.Entries[0], nil
}
