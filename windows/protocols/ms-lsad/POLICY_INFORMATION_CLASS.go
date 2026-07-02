package mslsad

// POLICY_INFORMATION_CLASS enumerates the policy information classes that select the
// arm of LSAPR_POLICY_INFORMATION ([MS-LSAD] 2.2.4.1). As an NDR enum it is transmitted
// as a 16-bit unsigned value ([C706] section 14.3.6).
type POLICY_INFORMATION_CLASS uint16

const (
	PolicyAuditLogInformation           POLICY_INFORMATION_CLASS = 1
	PolicyAuditEventsInformation        POLICY_INFORMATION_CLASS = 2
	PolicyPrimaryDomainInformation      POLICY_INFORMATION_CLASS = 3
	PolicyPdAccountInformation          POLICY_INFORMATION_CLASS = 4
	PolicyAccountDomainInformation      POLICY_INFORMATION_CLASS = 5
	PolicyLsaServerRoleInformation      POLICY_INFORMATION_CLASS = 6
	PolicyReplicaSourceInformation      POLICY_INFORMATION_CLASS = 7
	PolicyInformationNotUsedOnWire      POLICY_INFORMATION_CLASS = 8
	PolicyModificationInformation       POLICY_INFORMATION_CLASS = 9
	PolicyAuditFullSetInformation       POLICY_INFORMATION_CLASS = 10
	PolicyAuditFullQueryInformation     POLICY_INFORMATION_CLASS = 11
	PolicyDnsDomainInformation          POLICY_INFORMATION_CLASS = 12
	PolicyDnsDomainInformationInt       POLICY_INFORMATION_CLASS = 13
	PolicyLocalAccountDomainInformation POLICY_INFORMATION_CLASS = 14
	PolicyMachineAccountInformation     POLICY_INFORMATION_CLASS = 15
	PolicyLastEntry                     POLICY_INFORMATION_CLASS = 16
)
