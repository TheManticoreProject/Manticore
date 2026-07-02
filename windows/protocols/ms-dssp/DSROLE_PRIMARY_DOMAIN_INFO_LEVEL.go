package msdssp

// DSROLE_PRIMARY_DOMAIN_INFO_LEVEL is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DSSP]).
type DSROLE_PRIMARY_DOMAIN_INFO_LEVEL uint16

const (
	DsRolePrimaryDomainInfoBasic DSROLE_PRIMARY_DOMAIN_INFO_LEVEL = 1
	DsRoleUpgradeStatus          DSROLE_PRIMARY_DOMAIN_INFO_LEVEL = 2
	DsRoleOperationState         DSROLE_PRIMARY_DOMAIN_INFO_LEVEL = 3
)
