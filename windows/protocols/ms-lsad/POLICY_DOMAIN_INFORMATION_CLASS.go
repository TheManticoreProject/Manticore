package mslsad

// POLICY_DOMAIN_INFORMATION_CLASS enumerates the policy domain information classes that
// select the arm of LSAPR_POLICY_DOMAIN_INFORMATION ([MS-LSAD] 2.2.4.2). As an NDR enum
// it is transmitted as a 16-bit unsigned value ([C706] section 14.3.6).
type POLICY_DOMAIN_INFORMATION_CLASS uint16

const (
	PolicyDomainQualityOfServiceInformation POLICY_DOMAIN_INFORMATION_CLASS = 1
	PolicyDomainEfsInformation              POLICY_DOMAIN_INFORMATION_CLASS = 2
	PolicyDomainKerberosTicketInformation   POLICY_DOMAIN_INFORMATION_CLASS = 3
)
