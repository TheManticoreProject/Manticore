package msdssp

// DSROLE_MACHINE_ROLE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DSSP]).
type DSROLE_MACHINE_ROLE uint16

const (
	DsRole_RoleStandaloneWorkstation   DSROLE_MACHINE_ROLE = 0
	DsRole_RoleMemberWorkstation       DSROLE_MACHINE_ROLE = 1
	DsRole_RoleStandaloneServer        DSROLE_MACHINE_ROLE = 2
	DsRole_RoleMemberServer            DSROLE_MACHINE_ROLE = 3
	DsRole_RoleBackupDomainController  DSROLE_MACHINE_ROLE = 4
	DsRole_RolePrimaryDomainController DSROLE_MACHINE_ROLE = 5
)
