package ldap

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// dnsLookupTimeout bounds each domain-controller hostname lookup so a broken
// resolver cannot stall the whole enumeration.
const dnsLookupTimeout = 2 * time.Second

type hostnameResolver interface {
	LookupHost(context.Context, string) ([]string, error)
}

func resolveDomainControllerHostnames(resolver hostnameResolver, controllerSets ...map[string][]string) []string {
	hostnameSet := make(map[string]struct{})
	for _, controllers := range controllerSets {
		for _, hostnames := range controllers {
			for _, hostname := range hostnames {
				hostnameSet[hostname] = struct{}{}
			}
		}
	}
	hostnames := make([]string, 0, len(hostnameSet))
	for hostname := range hostnameSet {
		hostnames = append(hostnames, hostname)
	}
	sort.Strings(hostnames)

	addressSet := make(map[string]struct{})
	for _, hostname := range hostnames {
		ctx, cancel := context.WithTimeout(context.Background(), dnsLookupTimeout)
		addresses, err := resolver.LookupHost(ctx, hostname)
		cancel()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			if net.ParseIP(address) != nil {
				addressSet[address] = struct{}{}
			}
		}
	}

	addresses := make([]string, 0, len(addressSet))
	for address := range addressSet {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	return addresses
}

// GetDomainDNSServers retrieves the IP addresses of DNS servers from both read-only domain controllers
// and domain controllers in the LDAP session. It resolves each LDAP-discovered hostname using the
// configured system resolver without opening a connection to the DNS service itself.
//
// A domain controller whose hostname cannot be resolved is skipped rather than
// treated as a fatal error. An error is only returned when the domain controllers
// themselves cannot be listed.
//
// Returns:
// - A slice of strings containing the resolved IP addresses of the DNS servers.
// - An error if the domain controllers could not be listed.
func (ldapSession *Session) GetDomainDNSServers() ([]string, error) {
	domainControllersMap, err := ldapSession.GetAllDomainControllers()
	if err != nil {
		return []string{}, fmt.Errorf("failed to list domain controllers: %w", err)
	}

	readOnlyDomainControllersMap, err := ldapSession.GetAllReadOnlyDomainControllers()
	if err != nil {
		return resolveDomainControllerHostnames(net.DefaultResolver, domainControllersMap), fmt.Errorf("failed to list read-only domain controllers: %w", err)
	}

	return resolveDomainControllerHostnames(net.DefaultResolver, domainControllersMap, readOnlyDomainControllersMap), nil
}

// GetPrincipalDomainController retrieves the DNS hostname of the principal domain controller (PDC) for a given domain.
// It constructs an LDAP query to search for the computer object that represents the PDC by using the primaryGroupID
// attribute and the specified domain name.
//
// Parameters:
// - domainName: A string representing the name of the domain for which the PDC is to be found.
//
// Returns:
// - A string containing the DNS hostname of the principal domain controller. If no PDC is found, an empty string is returned.
func (ldapSession *Session) GetPrincipalDomainController(domainName string) (string, error) {
	// Constructing LDAP query to find the principal domain controller (PDC)
	query := fmt.Sprintf("(&(objectClass=computer)(primaryGroupID=%d)(dnsHostName=*%s*))", ldap_attributes.RID_DOMAIN_GROUP_CONTROLLERS, domainName)
	attributes := []string{"dnsHostName"}

	// Performing LDAP query
	ldapResults, err := ldapSession.QueryWholeSubtree("", query, attributes)
	if err != nil {
		return "", fmt.Errorf("error querying LDAP: %w", err)
	}

	if len(ldapResults) == 0 {
		return "", nil
	} else {
		// Extracting the DNS hostname of the PDC
		pdcHostname := ldapResults[0].GetAttributeValue("dnsHostName")
		return pdcHostname, nil
	}
}

// GetAllDomainControllers retrieves a map of all domain controllers in the LDAP directory.
// It constructs an LDAP query to search for computer objects that represent domain controllers
// by using the userAccountControl attribute to filter for server trust accounts.
//
// Parameters:
// - ldapSession: A pointer to the LDAP session.
//
// Returns:
//   - A map where the keys are the distinguished names of the domain controllers and the values are slices of strings
//     containing the DNS hostnames of the domain controllers.
func (ldapSession *Session) GetAllDomainControllers() (map[string][]string, error) {
	attributes := []string{"distinguishedName", "dnsHostname"}

	query := "(&"
	// Searching for computer accounts
	query += "(objectClass=computer)"
	//
	query += fmt.Sprintf("(userAccountControl:1.2.840.113556.1.4.803:=%d)", ldap_attributes.UAF_SERVER_TRUST_ACCOUNT)
	// With a DNS hostname
	query += "(dnsHostname=*)"
	// Closing the AND
	query += ")"

	ldapResults, err := ldapSession.QueryWholeSubtree("", query, attributes)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}

	domainControllersMap := make(map[string][]string)

	if len(ldapResults) != 0 {
		for _, entry := range ldapResults {
			domainControllersMap[entry.GetAttributeValue("distinguishedName")] = entry.GetEqualFoldAttributeValues("dnsHostname")
		}
	}

	return domainControllersMap, nil
}

// GetAllReadOnlyDomainControllers retrieves a map of all read-only domain controllers (RODCs) in the LDAP directory.
// It constructs an LDAP query to search for computer objects that represent read-only domain controllers
// by using the userAccountControl attribute to filter for partial secrets accounts.
//
// Parameters:
// - ldapSession: A pointer to the LDAP session.
//
// Returns:
//   - A map where the keys are the distinguished names of the read-only domain controllers and the values are slices of strings
//     containing the DNS hostnames of the read-only domain controllers.
func (ldapSession *Session) GetAllReadOnlyDomainControllers() (map[string][]string, error) {
	attributes := []string{"distinguishedName", "dnsHostname"}

	query := "(&"
	// Searching for computer accounts
	query += "(objectClass=computer)"
	//
	query += fmt.Sprintf("(userAccountControl:1.2.840.113556.1.4.803:=%d)", ldap_attributes.UAF_PARTIAL_SECRETS_ACCOUNT)
	// With a DNS hostname
	query += "(dnsHostname=*)"
	// Closing the AND
	query += ")"

	ldapResults, err := ldapSession.QueryWholeSubtree("", query, attributes)
	if err != nil {
		return nil, fmt.Errorf("error querying LDAP: %w", err)
	}
	readOnlyDomainControllersMap := make(map[string][]string)

	if len(ldapResults) != 0 {
		for _, entry := range ldapResults {
			readOnlyDomainControllersMap[entry.GetAttributeValue("distinguishedName")] = entry.GetEqualFoldAttributeValues("dnsHostname")
		}
	}

	return readOnlyDomainControllersMap, nil
}
