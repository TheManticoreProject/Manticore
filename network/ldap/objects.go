package ldap

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/ldap/objects"
)

// GetAllDomains retrieves all domain objects from the LDAP directory.
//
// This function performs an LDAP search to find all objects with the objectClass "domain"
// within the domain's distinguished name. It retrieves the distinguished name and objectSid
// attributes for each domain object and constructs a map of Domain objects.
//
// Returns:
//   - A map where the keys are the distinguished names of the domain objects and the values are
//     pointers to Domain objects representing the retrieved domain objects.
func (ldapSession *Session) GetAllDomains() (map[string]*objects.Domain, error) {
	attributes := []string{"distinguishedName", "objectSid"}
	query := "(objectClass=domain)"

	ldapResults, err := ldapSession.QueryWholeSubtree("", query, attributes)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}

	domainsMap := make(map[string]*objects.Domain)

	if len(ldapResults) != 0 {
		for _, entry := range ldapResults {
			DNSName := GetDomainFromDistinguishedName(entry.GetAttributeValue("distinguishedName"))
			DNSName = strings.ToUpper(DNSName)

			NetBIOSName := strings.ToUpper(entry.GetAttributeValue("dc"))

			domain := &objects.Domain{
				DistinguishedName: entry.GetAttributeValue("distinguishedName"),
				NetBIOSName:       NetBIOSName,
				DNSName:           DNSName,
				SID:               ParseSIDFromBytes(entry.GetRawAttributeValue("objectSid")),
			}

			domainsMap[DNSName] = domain
		}
	}

	return domainsMap, nil
}

// GetDomain retrieves the domain object from the LDAP server.
//
// This function searches for the domain object in the LDAP server and returns the domain object.
//
// Parameters:
//   - domain (string): The domain to search for.
//
// Returns:
//   - *objects.Domain: The domain object if found, otherwise nil.
func (ldapSession *Session) GetDomain(domain string) (*objects.Domain, error) {

	query := "(objectClass=domain)"

	attributes := []string{"distinguishedName", "objectSid", "dc"}

	ldapResults, err := ldapSession.QueryWholeSubtree("defaultNamingContext", query, attributes)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}

	if strings.Contains(domain, ".") {
		// FQDN
		for _, entry := range ldapResults {
			DNSName := GetDomainFromDistinguishedName(entry.GetAttributeValue("distinguishedName"))
			DNSName = strings.ToUpper(DNSName)

			NetBIOSName := strings.ToUpper(entry.GetAttributeValue("dc"))

			if DNSName == strings.ToUpper(domain) {
				return &objects.Domain{
					LdapSession: ldapSession,

					DistinguishedName: entry.GetAttributeValue("distinguishedName"),
					NetBIOSName:       NetBIOSName,
					DNSName:           DNSName,
					SID:               ParseSIDFromBytes(entry.GetRawAttributeValue("objectSid")),
				}, nil
			}
		}
	} else {
		// Netbios Name
		for _, entry := range ldapResults {
			DNSName := GetDomainFromDistinguishedName(entry.GetAttributeValue("distinguishedName"))
			DNSName = strings.ToUpper(DNSName)

			NetBIOSName := strings.ToUpper(entry.GetAttributeValue("dc"))

			if NetBIOSName == strings.ToUpper(domain) {
				return &objects.Domain{
					LdapSession: ldapSession,

					DistinguishedName: entry.GetAttributeValue("distinguishedName"),
					NetBIOSName:       NetBIOSName,
					DNSName:           DNSName,
					SID:               ParseSIDFromBytes(entry.GetRawAttributeValue("objectSid")),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no domain found for %s", domain)
}
