package objects

import (
	"fmt"
	"strconv"
)

type Domain struct {
	// LdapSession is the LDAP session object
	LdapSession LdapSessionInterface
	// DistinguishedName is the distinguished name of the domain
	DistinguishedName string
	// NetBIOSName is the NetBIOS name of the domain
	NetBIOSName string
	// DNSName is the DNS name of the domain
	DNSName string
	// SID is the SID of the domain
	SID string
}

// IsDomainAtLeast checks if the domain's functionality level is at least the specified level.
//
// This function retrieves the domain object for the given domain name and queries the LDAP server
// to get the "msDS-Behavior-Version" attribute, which represents the domain's functionality level.
// It then compares this value with the provided functionality level.
//
// Parameters:
//   - domain (string): The name of the domain to check.
//   - functionalityLevel (int): The minimum functionality level to check against.
//
// Returns:
//   - bool: True if the domain's functionality level is at least the specified level, false otherwise.
//
// Example:
//
//	ldapSession := &Session{}
//	domain := "example.com"
//	functionalityLevel := 3
//	isAtLeast := ldapSession.IsDomainAtLeast(domain, functionalityLevel)
//	if isAtLeast {
//	    fmt.Println("The domain's functionality level is at least", functionalityLevel)
//	} else {
//	    fmt.Println("The domain's functionality level is less than", functionalityLevel)
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the GetDomain and QueryBaseObject methods
//     are implemented correctly.
//   - The function logs a warning if the "msDS-Behavior-Version" attribute cannot be parsed to an integer.
func (domain *Domain) IsDomainAtLeast(functionalityLevel int) (bool, error) {
	var err error

	domainObject, err := domain.LdapSession.GetDomain(domain.DistinguishedName)
	if err != nil {
		return false, fmt.Errorf("error fetching domain: %w", err)
	}

	if domainObject != nil {
		query := fmt.Sprintf("(distinguishedName=%s)", domainObject.DistinguishedName)
		attributes := []string{"msDS-Behavior-Version"}
		results, err := domain.LdapSession.QueryBaseObject(domainObject.DistinguishedName, query, attributes)
		if err != nil {
			return false, fmt.Errorf("error querying LDAP: %w", err)
		}

		if len(results) != 0 {
			domainFunctionalityLevel, err := strconv.Atoi(results[0].GetAttributeValue("msDS-Behavior-Version"))
			if err != nil {
				return false, err
			} else {
				if domainFunctionalityLevel >= functionalityLevel {
					return true, nil
				} else {
					return false, nil
				}
			}
		} else {
			return false, nil
		}
	}

	return false, err
}
