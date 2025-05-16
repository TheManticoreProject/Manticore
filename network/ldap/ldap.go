package ldap

import (
	"fmt"

	goldapv3 "github.com/go-ldap/ldap/v3"
)

type Entry = goldapv3.Entry

type Control = goldapv3.Control

const (
	ScopeBaseObject   = goldapv3.ScopeBaseObject
	ScopeSingleLevel  = goldapv3.ScopeSingleLevel
	ScopeChildren     = goldapv3.ScopeChildren
	ScopeWholeSubtree = goldapv3.ScopeWholeSubtree
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
func (ldapSession *Session) GetRootDSE() (*Entry, error) {
	// Specify LDAP search parameters
	// https://pkg.go.dev/gopkg.in/ldap.v3#NewSearchRequest
	searchRequest := goldapv3.NewSearchRequest(
		// Base DN blank
		"",
		// Scope Base
		goldapv3.ScopeBaseObject,
		// DerefAliases
		goldapv3.NeverDerefAliases,
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
