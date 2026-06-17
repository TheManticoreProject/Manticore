package ldap

import (
	"fmt"
	"net"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// dnsProbeTimeout bounds each TCP connection attempt to a domain controller's
// DNS port so a single unresponsive host cannot stall the whole enumeration.
const dnsProbeTimeout = 2 * time.Second

// GetDomainDNSServers retrieves the IP addresses of DNS servers from both read-only domain controllers
// and domain controllers in the LDAP session. It attempts to establish a TCP connection to port 53
// (DNS service) on each hostname found in the domain controllers. If the connection is successful,
// it extracts the IP address of the DNS server and adds it to the list of DNS servers.
//
// A domain controller that is unreachable or not serving DNS is skipped rather
// than treated as a fatal error: the function returns the DNS servers it could
// reach. An error is only returned when the domain controllers themselves cannot
// be listed.
//
// Returns:
// - A slice of strings containing the IP addresses of the reachable DNS servers.
// - An error if the domain controllers could not be listed.
func (ldapSession *Session) GetDomainDNSServers() ([]string, error) {
	dnsServers := []string{}

	probe := func(hostname string) {
		conn, err := net.DialTimeout("tcp", hostname+":53", dnsProbeTimeout)
		if err != nil {
			// Host is unreachable or not serving DNS; skip it.
			return
		}
		defer conn.Close()

		// Get remote address which contains the IP of the connected DNS server.
		// Use net.SplitHostPort so IPv6 addresses ("[::1]:53") are handled correctly.
		host, _, splitErr := net.SplitHostPort(conn.RemoteAddr().String())
		if splitErr != nil {
			return
		}
		dnsServers = append(dnsServers, host)
	}

	// Check in domain controllers
	domainControllersMap, err := ldapSession.GetAllDomainControllers()
	if err != nil {
		return dnsServers, fmt.Errorf("failed to list domain controllers: %w", err)
	}

	for distinguishedName := range domainControllersMap {
		for _, hostname := range domainControllersMap[distinguishedName] {
			probe(hostname)
		}
	}

	// Check in read only domain controllers
	readOnlyDomainControllersMap, err := ldapSession.GetAllReadOnlyDomainControllers()
	if err != nil {
		return dnsServers, fmt.Errorf("failed to list read-only domain controllers: %w", err)
	}

	for distinguishedName := range readOnlyDomainControllersMap {
		for _, hostname := range readOnlyDomainControllersMap[distinguishedName] {
			probe(hostname)
		}
	}

	return dnsServers, nil
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
