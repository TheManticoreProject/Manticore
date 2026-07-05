package mssamr

// DOMAIN_INFORMATION_CLASS enumerates the domain information classes for the
// SAMPR_DOMAIN_INFO_BUFFER union ([MS-SAMR] 2.2.4.16). As an NDR enum it is transmitted
// as a 16-bit unsigned value ([C706] section 14.3.6). The IDL skips the value 10.
type DOMAIN_INFORMATION_CLASS uint16

const (
	DomainPasswordInformation    DOMAIN_INFORMATION_CLASS = 1
	DomainGeneralInformation     DOMAIN_INFORMATION_CLASS = 2
	DomainLogoffInformation      DOMAIN_INFORMATION_CLASS = 3
	DomainOemInformation         DOMAIN_INFORMATION_CLASS = 4
	DomainNameInformation        DOMAIN_INFORMATION_CLASS = 5
	DomainReplicationInformation DOMAIN_INFORMATION_CLASS = 6
	DomainServerRoleInformation  DOMAIN_INFORMATION_CLASS = 7
	DomainModifiedInformation    DOMAIN_INFORMATION_CLASS = 8
	DomainStateInformation       DOMAIN_INFORMATION_CLASS = 9
	DomainGeneralInformation2    DOMAIN_INFORMATION_CLASS = 11
	DomainLockoutInformation     DOMAIN_INFORMATION_CLASS = 12
	DomainModifiedInformation2   DOMAIN_INFORMATION_CLASS = 13
)
