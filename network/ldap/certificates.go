package ldap

import (
	"fmt"
)

// GetAllCertificates retrieves the distinguished names of all certificate templates from the LDAP server.
// It searches for objects of category 'pKICertificateTemplate' and extracts the 'distinguishedName'
// attribute values from them.
//
// Returns:
// - A slice of strings containing the distinguished names of all certificate templates.
func (ldapSession *Session) GetAllCertificates() ([]string, error) {
	distinguishedNames := []string{}

	query := "(objectClass=pKICertificateTemplate)"
	attributes := []string{"distinguishedName"}

	ldapResults, err := ldapSession.QueryWholeSubtree("configurationNamingContext", query, attributes)
	if err != nil {
		return distinguishedNames, nil
	}

	if len(ldapResults) != 0 {
		for _, entry := range ldapResults {
			distinguishedNames = append(distinguishedNames, entry.GetAttributeValue("distinguishedName"))
		}
	}

	return distinguishedNames, nil
}

// GetNamesOfAllEnabledCertificates retrieves the names of all enabled certificate templates
// from the LDAP server. It searches for objects of category 'pKIEnrollmentService' and
// extracts the 'certificateTemplates' attribute values from them.
//
// Returns:
// - A slice of strings containing the names of all enabled certificate templates.
func (ldapSession *Session) GetNamesOfAllEnabledCertificates() ([]string, error) {
	names := []string{}

	queryPKIEnrollmentService := "(objectCategory=pKIEnrollmentService)"

	attributes := []string{"certificateTemplates"}

	ldapResultPKIEnrollmentService, err := ldapSession.QueryWholeSubtree("configurationNamingContext", queryPKIEnrollmentService, attributes)
	if err != nil {
		return names, fmt.Errorf("error querying LDAP: %w", err)
	}

	// We have found pKIEnrollmentServices
	if len(ldapResultPKIEnrollmentService) != 0 {
		for _, entry := range ldapResultPKIEnrollmentService {
			// Iterating on the enabled certificate templates of the pKIEnrollmentService
			names = append(names, entry.GetEqualFoldAttributeValues("certificateTemplates")...)
		}
	}

	return names, nil
}

// GetDistinguishedNamesOfAllEnabledCertificates retrieves the distinguished names of all enabled
// certificate templates from the LDAP server. It first fetches the names of all enabled certificate
// templates using the GetNamesOfAllEnabledCertificates function, and then queries the LDAP server
// for each certificate template name to get their distinguished names.
//
// Returns:
// - A slice of strings containing the distinguished names of all enabled certificate templates.
func (ldapSession *Session) GetDistinguishedNamesOfAllEnabledCertificates() ([]string, error) {
	distinguishedNames := []string{}
	certificateTemplateNames, err := ldapSession.GetNamesOfAllEnabledCertificates()
	if err != nil {
		return distinguishedNames, fmt.Errorf("error fetching certificate template names: %w", err)
	}

	for _, name := range certificateTemplateNames {
		queryPKICertificateTemplate := "(&"
		queryPKICertificateTemplate += "(objectClass=pKICertificateTemplate)"
		queryPKICertificateTemplate += fmt.Sprintf("(name=%s)", name)
		queryPKICertificateTemplate += ")"

		attributes := []string{"distinguishedName"}

		ldapResultsPKICertificateTemplate, err := ldapSession.QueryWholeSubtree("configurationNamingContext", queryPKICertificateTemplate, attributes)
		if err != nil {
			return distinguishedNames, fmt.Errorf("error querying LDAP: %w", err)
		}

		if len(ldapResultsPKICertificateTemplate) != 0 {
			for _, entry := range ldapResultsPKICertificateTemplate {
				distinguishedNames = append(distinguishedNames, entry.GetAttributeValue("distinguishedName"))
			}
		}
	}

	return distinguishedNames, nil
}
