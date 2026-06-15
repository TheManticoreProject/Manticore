package structures

// LSAPR_POLICY_DOMAIN_INFORMATION is the discriminated union of policy domain
// information classes ([MS-LSAD] 2.2.4.3). The discriminant Class is a
// POLICY_DOMAIN_INFORMATION_CLASS; the wire form is the discriminant followed by the
// single selected arm ([C706] section 14.3.8). Case values follow the enum: QoS=1,
// Efs=2, KerbTicket=3.
type LSAPR_POLICY_DOMAIN_INFORMATION struct {
	Class                            POLICY_DOMAIN_INFORMATION_CLASS       `ndr:"switch,enum"`
	PolicyDomainQualityOfServiceInfo POLICY_DOMAIN_QUALITY_OF_SERVICE_INFO `ndr:"case=1"`
	PolicyDomainEfsInfo              LSAPR_POLICY_DOMAIN_EFS_INFO          `ndr:"case=2"`
	PolicyDomainKerbTicketInfo       POLICY_DOMAIN_KERBEROS_TICKET_INFO    `ndr:"case=3"`
}
