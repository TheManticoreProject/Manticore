package structures

// LSAPR_POLICY_INFORMATION is the discriminated union of policy information classes
// ([MS-LSAD] 2.2.4.2). The discriminant Class is a POLICY_INFORMATION_CLASS; the wire
// form is the discriminant followed by the single selected arm ([C706] section 14.3.8).
//
// Each arm is a value field carrying the structure for that class. The numeric case
// values follow the POLICY_INFORMATION_CLASS enum order in the IDL. PolicyDnsDomainInfo
// (12) and PolicyDnsDomainInfoInt (13) share the LSAPR_POLICY_DNS_DOMAIN_INFO type but
// are distinct discriminant values, as are PolicyAccountDomainInfo (5) and
// PolicyLocalAccountDomainInfo (14), which share LSAPR_POLICY_ACCOUNT_DOM_INFO.
type LSAPR_POLICY_INFORMATION struct {
	Class                        POLICY_INFORMATION_CLASS       `ndr:"switch,enum"`
	PolicyAuditLogInfo           POLICY_AUDIT_LOG_INFO          `ndr:"case=1"`
	PolicyAuditEventsInfo        LSAPR_POLICY_AUDIT_EVENTS_INFO `ndr:"case=2"`
	PolicyPrimaryDomainInfo      LSAPR_POLICY_PRIMARY_DOM_INFO  `ndr:"case=3"`
	PolicyPdAccountInfo          LSAPR_POLICY_PD_ACCOUNT_INFO   `ndr:"case=4"`
	PolicyAccountDomainInfo      LSAPR_POLICY_ACCOUNT_DOM_INFO  `ndr:"case=5"`
	PolicyServerRoleInfo         POLICY_LSA_SERVER_ROLE_INFO    `ndr:"case=6"`
	PolicyReplicaSourceInfo      LSAPR_POLICY_REPLICA_SRCE_INFO `ndr:"case=7"`
	PolicyModificationInfo       POLICY_MODIFICATION_INFO       `ndr:"case=9"`
	PolicyAuditFullSetInfo       POLICY_AUDIT_FULL_SET_INFO     `ndr:"case=10"`
	PolicyAuditFullQueryInfo     POLICY_AUDIT_FULL_QUERY_INFO   `ndr:"case=11"`
	PolicyDnsDomainInfo          LSAPR_POLICY_DNS_DOMAIN_INFO   `ndr:"case=12"`
	PolicyDnsDomainInfoInt       LSAPR_POLICY_DNS_DOMAIN_INFO   `ndr:"case=13"`
	PolicyLocalAccountDomainInfo LSAPR_POLICY_ACCOUNT_DOM_INFO  `ndr:"case=14"`
	PolicyMachineAccountInfo     LSAPR_POLICY_MACHINE_ACCT_INFO `ndr:"case=15"`
}
