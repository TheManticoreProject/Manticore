package objects

import "fmt"

type Computer struct {
	// LdapSession is the LDAP session object
	LdapSession LdapSessionInterface
	// DistinguishedName is the distinguished name of the computer
	DistinguishedName string
	// DNSHostname is the DNS hostname of the computer
	DNSHostname []string
}

// GetAllComputers retrieves all computer objects from the LDAP directory.
//
// This function performs an LDAP search to find all objects with the objectClass "computer"
// within the domain's distinguished name. It retrieves the distinguished name and DNS hostname
// attributes for each computer object and constructs a map of Computer objects.
//
// Returns:
//   - A map where the keys are the distinguished names of the computer objects and the values are
//     pointers to Computer objects representing the retrieved computer objects.
//
// Example usage:
//
//	domain := &Domain{LdapSession: ldapSession, DistinguishedName: "DC=example,DC=com"}
//	computers := domain.GetAllComputers()
//	for dn, computer := range computers {
//	    fmt.Printf("Computer DN: %s, DNS Hostname: %v\n", dn, computer.DNSHostname)
//	}
func (domain *Domain) GetAllComputers() (map[string]*Computer, error) {
	attributes := []string{"distinguishedName", "dnsHostname"}

	query := "(objectClass=computer)"

	ldapResults, err := domain.LdapSession.QueryWholeSubtree(domain.DistinguishedName, query, attributes)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}

	computersMap := make(map[string]*Computer)

	if len(ldapResults) != 0 {
		for _, entry := range ldapResults {
			computer := &Computer{
				DistinguishedName: entry.GetAttributeValue("distinguishedName"),
				DNSHostname:       entry.GetEqualFoldAttributeValues("dnsHostname"),
			}

			computersMap[entry.GetAttributeValue("distinguishedName")] = computer
		}
	}

	return computersMap, nil
}
