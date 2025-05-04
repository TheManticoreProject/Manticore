package ldap

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// FindObjectSIDByRID searches for an LDAP object based on the provided domain and RID (Relative Identifier).
//
// Parameters:
//   - domain: A string representing the domain name where the search should be performed.
//   - RID: An integer representing the Relative Identifier of the object to be found.
//
// Returns:
//   - A string representing the SID (Security Identifier) of the found object. If no object is found, an empty string is returned.
//
// The function first retrieves the domain object using the provided domain name. It then constructs an LDAP query
// to search for the object with the specified RID. The query is created differently for local RIDs and domain RIDs.
// If the RID matches a local RID, the query is constructed using the local RID format. Otherwise, the query is
// constructed using the domain SID and the provided RID.
//
// The function performs an LDAP search using the constructed query and retrieves the distinguished name and object SID
// of the found object. If more than one result is found, a warning is logged. If exactly one result is found, the
// function parses the object SID from the raw attribute value and returns it.
//
// Example usage:
//
//	ldapSession := &Session{}
//	objectSID := ldapSession.FindObjectSIDByRID("example.com", 500)
//	fmt.Println("Found object SID:", objectSID)
func (ldapSession *Session) FindObjectSIDByRID(domain string, RID int) (string, error) {
	objectSID := ""

	domainObject, err := ldapSession.GetDomain(domain)
	if err != nil {
		return "", fmt.Errorf("error fetching domain: %w", err)
	}

	if domainObject != nil {
		// Create query for local RID
		query := ""
		for _, localRID := range ldap_attributes.LocalRIDs {
			if localRID == RID {
				query = fmt.Sprintf("(objectSid=S-1-5-32-%d)", localRID)
				break
			}
		}
		// Create query for other (domain) RID
		if len(query) == 0 {
			query = fmt.Sprintf("(objectSid=%s-%d)", domainObject.SID, RID)
		}

		// Perform LDAP query to find the object
		attributes := []string{"distinguishedName", "objectSid"}
		results, err := ldapSession.QueryWholeSubtree(domainObject.DistinguishedName, query, attributes)
		if err != nil {
			return "", fmt.Errorf("error querying LDAP: %w", err)
		}

		if len(results) > 1 {
			logger.Warn(fmt.Sprintf("Error: More than one result for SID '%s-%d' in the domain '%s'", domainObject.SID, RID, domain))
		} else {
			if len(results) == 1 {
				// One result found
				objectSID = ParseSIDFromBytes(results[0].GetRawAttributeValue("objectSid"))
			}
		}
	}

	return objectSID, nil
}
