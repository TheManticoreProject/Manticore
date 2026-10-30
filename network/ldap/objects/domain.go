package objects

import (
	"fmt"
	"strconv"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
	"github.com/go-ldap/ldap/v3"
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

// IsDomainFunctionalityLevelGreaterThan checks if the domain's functionality level is greater than the specified level.
func (domain *Domain) IsDomainFunctionalityLevelGreaterThan(functionalityLevel ldap_attributes.DomainFunctionalityLevel) (bool, error) {
	domainFunctionalityLevel, err := domain.GetDomainFunctionalityLevel()
	if err != nil {
		return false, err
	}

	return domainFunctionalityLevel > functionalityLevel, nil
}

// IsDomainFunctionalityLevelGreaterThanOrEqualTo checks if the domain's functionality level is greater than or equal to the specified level.
func (domain *Domain) IsDomainFunctionalityLevelGreaterThanOrEqualTo(functionalityLevel ldap_attributes.DomainFunctionalityLevel) (bool, error) {
	domainFunctionalityLevel, err := domain.GetDomainFunctionalityLevel()
	if err != nil {
		return false, err
	}

	return domainFunctionalityLevel >= functionalityLevel, nil
}

// IsDomainFunctionalityLevelLowerThan checks if the domain's functionality level is lower than the specified level.
func (domain *Domain) IsDomainFunctionalityLevelLowerThan(functionalityLevel ldap_attributes.DomainFunctionalityLevel) (bool, error) {
	domainFunctionalityLevel, err := domain.GetDomainFunctionalityLevel()
	if err != nil {
		return false, err
	}

	return domainFunctionalityLevel < functionalityLevel, nil
}

// IsDomainFunctionalityLevelLowerThanOrEqualTo checks if the domain's functionality level is lower than or equal to the specified level.
func (domain *Domain) IsDomainFunctionalityLevelLowerThanOrEqualTo(functionalityLevel ldap_attributes.DomainFunctionalityLevel) (bool, error) {
	domainFunctionalityLevel, err := domain.GetDomainFunctionalityLevel()
	if err != nil {
		return false, err
	}

	return domainFunctionalityLevel <= functionalityLevel, nil
}

// IsDomainFunctionalityLevelEqualTo checks if the domain's functionality level is equal to the specified level.
func (domain *Domain) IsDomainFunctionalityLevelEqualTo(functionalityLevel ldap_attributes.DomainFunctionalityLevel) (bool, error) {
	domainFunctionalityLevel, err := domain.GetDomainFunctionalityLevel()
	if err != nil {
		return false, err
	}

	return domainFunctionalityLevel == functionalityLevel, nil
}

// GetDomainFunctionalityLevel gets the domain's functionality level.
func (domain *Domain) GetDomainFunctionalityLevel() (ldap_attributes.DomainFunctionalityLevel, error) {
	var err error

	query := fmt.Sprintf("(distinguishedName=%s)", ldap.EscapeFilter(domain.DistinguishedName))
	attributes := []string{"msDS-Behavior-Version"}
	results, err := domain.LdapSession.QueryBaseObject(domain.DistinguishedName, query, attributes)
	if err != nil {
		return 0, fmt.Errorf("error querying LDAP: %w", err)
	}

	if len(results) != 0 {
		domainFunctionalityLevel, err := strconv.Atoi(results[0].GetAttributeValue("msDS-Behavior-Version"))
		if err != nil {
			return 0, err
		} else {
			return ldap_attributes.DomainFunctionalityLevel(domainFunctionalityLevel), nil
		}
	} else {
		return 0, fmt.Errorf("domain functionality level not found for domain %s", domain.DistinguishedName)
	}
}
