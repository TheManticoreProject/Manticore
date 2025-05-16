package ldap

import (
	"fmt"

	goldapv3 "github.com/go-ldap/ldap/v3"
)

// BaseDNExists checks if a given base distinguished name (baseDN) exists in the LDAP directory.
//
// This function performs an LDAP search with a base scope to determine if the specified baseDN exists.
// It constructs an LDAP search request with the provided baseDN and a search filter of "(objectClass=*)",
// and attempts to retrieve the "distinguishedName" attribute. If the search returns an error indicating
// that the baseDN does not exist (LDAPResultNoSuchObject), the function returns false. Otherwise, it returns true.
//
// Parameters:
//   - baseDN (string): The base distinguished name to check for existence in the LDAP directory.
//
// Returns:
//   - bool: True if the baseDN exists, false if it does not exist or if an error occurs.
//
// Example usage:
//
//	ldapSession := &Session{}
//	exists := ldapSession.BaseDNExists("DC=example,DC=com")
//	if exists {
//	    fmt.Println("The baseDN exists in the LDAP directory.")
//	} else {
//	    fmt.Println("The baseDN does not exist in the LDAP directory.")
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the ldap package
//     is correctly imported and used.
func (ldapSession *Session) BaseDNExists(baseDN string) bool {
	// Specify LDAP search parameters
	// https://pkg.go.dev/gopkg.in/ldap.v3#NewSearchRequest
	searchRequest := goldapv3.NewSearchRequest(
		// Base DN
		baseDN,
		// Scope
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
		[]string{"distinguishedName"},
		// Controls
		nil,
	)

	// Perform LDAP search
	_, err := ldapSession.connection.Search(searchRequest)
	if goldapv3.IsErrorWithCode(err, goldapv3.LDAPResultNoSuchObject) {
		return false
	} else {
		return true
	}
}

// GetAllNamingContexts retrieves all naming contexts from the LDAP server.
//
// This function fetches the RootDSE entry from the LDAP server and retrieves the "namingContexts" attribute,
// which contains a list of all naming contexts available in the LDAP directory.
//
// Returns:
//   - A slice of strings containing the naming contexts if found, otherwise nil.
//
// Example usage:
//
//	ldapSession := &Session{}
//	namingContexts := ldapSession.GetAllNamingContexts()
//	if namingContexts != nil {
//	    for _, context := range namingContexts {
//	        fmt.Printf("Naming Context: %s\n", context)
//	    }
//	} else {
//	    fmt.Println("No naming contexts found.")
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the GetRootDSE method
//     is implemented correctly.
func (ldapSession *Session) GetAllNamingContexts() ([]string, error) {
	// Fetch the RootDSE entry
	rootDSE, err := ldapSession.GetRootDSE()
	if err != nil {
		return nil, fmt.Errorf("error fetching RootDSE: %w", err)
	}

	// Retrieve the namingContexts attribute
	namingContexts := rootDSE.GetAttributeValues("namingContexts")
	if len(namingContexts) == 0 {
		return nil, fmt.Errorf("no naming contexts found")
	}

	return namingContexts, nil
}
