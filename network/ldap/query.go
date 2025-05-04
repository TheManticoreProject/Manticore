package ldap

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"

	"github.com/go-ldap/ldap/v3"
)

// Query performs an LDAP search operation based on the provided parameters.
//
// Parameters:
//   - searchBase: A string representing the base DN (Distinguished Name) from which the search should start.
//     Special values "defaultNamingContext", "configurationNamingContext", and "schemaNamingContext"
//     will be replaced with the corresponding values from the root DSE.
//   - query: A string representing the LDAP search filter.
//   - attributes: A slice of strings specifying the attributes to retrieve from the search results.
//   - scope: An integer representing the scope of the search. Valid values are:
//   - ScopeBaseObject: Search only the base object.
//   - ScopeSingleLevel: Search one level under the base object.
//   - ScopeWholeSubtree: Search the entire subtree under the base object.
//
// Returns:
// - A slice of pointers to ldap.Entry objects representing the search results.
//
// The function first determines the appropriate search base by checking if the provided searchBase
// matches any of the special values. If it does, it retrieves the corresponding value from the root DSE.
// It then constructs an LDAP search request with the specified parameters and performs the search.
// If the search is successful, it returns the search results. If an error occurs, it logs a warning
// and returns an empty slice.
func (ldapSession *Session) Query(searchBase string, query string, attributes []string, scope int) ([]*ldap.Entry, error) {
	debug := false

	// Parsing parameters
	if len(searchBase) == 0 {
		searchBase = "defaultNamingContext"
	}
	rootDSE, err := ldapSession.GetRootDSE()
	if err != nil {
		return nil, fmt.Errorf("error fetching RootDSE: %w", err)
	}
	if strings.ToLower(searchBase) == "defaultnamingcontext" {
		if debug {
			logger.Debug(fmt.Sprintf("Using defaultNamingContext %s ...\n", rootDSE.GetAttributeValue("defaultNamingContext")))
		}
		searchBase = rootDSE.GetAttributeValue("defaultNamingContext")
	} else if strings.ToLower(searchBase) == "configurationnamingcontext" {
		if debug {
			logger.Debug(fmt.Sprintf("Using configurationNamingContext %s ...\n", rootDSE.GetAttributeValue("configurationNamingContext")))
		}
		searchBase = rootDSE.GetAttributeValue("configurationNamingContext")
	} else if strings.ToLower(searchBase) == "schemanamingcontext" {
		if debug {
			logger.Debug(fmt.Sprintf("Using schemaNamingContext CN=Schema,%s ...\n", rootDSE.GetAttributeValue("configurationNamingContext")))
		}
		searchBase = fmt.Sprintf("CN=Schema,%s", rootDSE.GetAttributeValue("configurationNamingContext"))
	}

	if (scope != ldap.ScopeBaseObject) && (scope != ldap.ScopeSingleLevel) && (scope != ldap.ScopeWholeSubtree) {
		scope = ldap.ScopeWholeSubtree
	}

	// Specify LDAP search parameters
	// https://pkg.go.dev/gopkg.in/ldap.v3#NewSearchRequest
	searchRequest := ldap.NewSearchRequest(
		// Base DN
		searchBase,
		// Scope
		scope,
		// DerefAliases
		ldap.NeverDerefAliases,
		// SizeLimit
		0,
		// TimeLimit
		0,
		// TypesOnly
		false,
		// Search filter
		query,
		// Attributes to retrieve
		attributes,
		// Controls
		nil,
	)

	// Perform LDAP search
	searchResult, err := ldapSession.connection.SearchWithPaging(searchRequest, 1000)
	if err != nil {
		return nil, fmt.Errorf("error searching LDAP: %w", err)
	}

	return searchResult.Entries, nil
}

// QueryBaseObject performs an LDAP query with a scope of Base Object.
//
// This function executes the specified query on the given search base with a scope of Base Object,
// meaning it searches only the base object itself.
//
// Parameters:
//   - searchBase (string): The base DN (Distinguished Name) from which the search should start.
//   - query (string): The LDAP query to be executed.
//   - attributes ([]string): The list of attributes to retrieve for each entry.
//
// Returns:
//   - []*ldap.Entry: A slice of LDAP entries that match the query within the base object. If no entries are found or
//     if an error occurs, the function returns nil.
//
// Example:
//
//	ldapSession := &Session{}
//	searchBase := "cn=admin,dc=example,dc=com"
//	query := "(objectClass=*)"
//	attributes := []string{"cn", "sn", "mail"}
//	entries := ldapSession.QueryBaseObject(searchBase, query, attributes)
//	for _, entry := range entries {
//	    fmt.Println(entry.GetAttributeValue("cn"))
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the Query method
//     is implemented correctly.
//   - The function logs warnings if the search fails or if no entries are found.
func (ldapSession *Session) QueryBaseObject(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	entries, err := ldapSession.Query(searchBase, query, attributes, ldap.ScopeBaseObject)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}
	return entries, nil
}

// QuerySingleLevel performs an LDAP query with a scope of Single Level.
//
// This function executes the specified query on the given search base with a scope of Single Level,
// meaning it searches only the immediate children of the base object, but not the base object itself.
//
// Parameters:
//   - searchBase (string): The base DN (Distinguished Name) from which the search should start.
//   - query (string): The LDAP query to be executed.
//   - attributes ([]string): The list of attributes to retrieve for each entry.
//
// Returns:
//   - []*ldap.Entry: A slice of LDAP entries that match the query within the single level. If no entries are found or
//     if an error occurs, the function returns nil.
//
// Example:
//
//	ldapSession := &Session{}
//	searchBase := "ou=users,dc=example,dc=com"
//	query := "(objectClass=person)"
//	attributes := []string{"cn", "sn", "mail"}
//	entries := ldapSession.QuerySingleLevel(searchBase, query, attributes)
//	for _, entry := range entries {
//	    fmt.Println(entry.GetAttributeValue("cn"))
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the Query method
//     is implemented correctly.
//   - The function logs warnings if the search fails or if no entries are found.
func (ldapSession *Session) QuerySingleLevel(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	entries, err := ldapSession.Query(searchBase, query, attributes, ldap.ScopeSingleLevel)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}
	return entries, nil
}

// QueryWholeSubtree performs an LDAP query with a scope of Whole Subtree.
//
// This function executes the specified query on the given search base with a scope of Whole Subtree,
// meaning it searches the entire subtree, including the base object and all its descendants.
//
// Parameters:
//   - searchBase (string): The base DN (Distinguished Name) from which the search should start.
//   - query (string): The LDAP query to be executed.
//   - attributes ([]string): The list of attributes to retrieve for each entry.
//
// Returns:
//   - []*ldap.Entry: A slice of LDAP entries that match the query within the whole subtree. If no entries are found or
//     if an error occurs, the function returns nil.
//
// Example:
//
//	ldapSession := &Session{}
//	searchBase := "dc=example,dc=com"
//	query := "(objectClass=person)"
//	attributes := []string{"cn", "sn", "mail"}
//	entries := ldapSession.QueryWholeSubtree(searchBase, query, attributes)
//	for _, entry := range entries {
//	    fmt.Println(entry.GetAttributeValue("cn"))
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the Query method
//     is implemented correctly.
//   - The function logs warnings if the search fails or if no entries are found.
func (ldapSession *Session) QueryWholeSubtree(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	entries, err := ldapSession.Query(searchBase, query, attributes, ldap.ScopeWholeSubtree)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}
	return entries, nil
}

// QueryChildren performs an LDAP query with a scope of Children.
//
// This function executes the specified query on the given search base with a scope of Children,
// meaning it searches only the immediate children of the base object.
//
// Parameters:
//   - searchBase (string): The base DN (Distinguished Name) from which the search should start.
//   - query (string): The LDAP query to be executed.
//   - attributes ([]string): The list of attributes to retrieve for each entry.
//
// Returns:
//   - []*ldap.Entry: A slice of LDAP entries that match the query within the immediate children. If no entries are found or
//     if an error occurs, the function returns nil.
//
// Example:
//
//	ldapSession := &Session{}
//	searchBase := "ou=users,dc=example,dc=com"
//	query := "(objectClass=person)"
//	attributes := []string{"cn", "sn", "mail"}
//	entries := ldapSession.QueryChildren(searchBase, query, attributes)
//	for _, entry := range entries {
//	    fmt.Println(entry.GetAttributeValue("cn"))
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the Query method
//     is implemented correctly.
//   - The function logs warnings if the search fails or if no entries are found.
func (ldapSession *Session) QueryChildren(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	entries, err := ldapSession.Query(searchBase, query, attributes, ldap.ScopeChildren)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}
	return entries, nil
}

// QueryAllNamingContexts performs an LDAP query across all naming contexts.
//
// This function retrieves the RootDSE entry to obtain the naming contexts and then performs the specified query
// on each naming context. The results from all naming contexts are combined and returned as a single slice of entries.
//
// Parameters:
//   - query (string): The LDAP query to be executed.
//   - attributes ([]string): The list of attributes to retrieve for each entry.
//   - scope (int): The scope of the query (e.g., ldap.ScopeBaseObject, ldap.ScopeSingleLevel, ldap.ScopeWholeSubtree).
//
// Returns:
//   - []*ldap.Entry: A slice of LDAP entries that match the query across all naming contexts. If no entries are found or
//     if an error occurs, the function returns nil.
//
// Example:
//
//	ldapSession := &Session{}
//	query := "(objectClass=person)"
//	attributes := []string{"cn", "sn", "mail"}
//	scope := ldap.ScopeWholeSubtree
//	entries := ldapSession.QueryAllNamingContexts(query, attributes, scope)
//	for _, entry := range entries {
//	    fmt.Println(entry.GetAttributeValue("cn"))
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the GetRootDSE and Query methods
//     are implemented correctly.
//   - The function logs warnings if the RootDSE entry cannot be retrieved or if no naming contexts are found.
func (ldapSession *Session) QueryAllNamingContexts(query string, attributes []string, scope int) ([]*ldap.Entry, error) {
	// Fetch the RootDSE entry to get the naming contexts
	rootDSE, err := ldapSession.GetRootDSE()
	if err != nil {
		return nil, fmt.Errorf("error fetching RootDSE: %w", err)
	}

	// Retrieve the namingContexts attribute
	namingContexts := rootDSE.GetAttributeValues("namingContexts")
	if len(namingContexts) == 0 {
		return nil, fmt.Errorf("no naming contexts found")
	}

	// Store all entries from all naming contexts
	var allEntries []*ldap.Entry

	// Iterate over each naming context and perform the query
	for _, context := range namingContexts {
		entries, err := ldapSession.Query(context, query, attributes, scope)
		if err != nil {
			return nil, fmt.Errorf("error querying LDAP: %w", err)
		}
		if entries != nil {
			allEntries = append(allEntries, entries...)
		}
	}

	return allEntries, nil
}
