package msdssp

// DSROLE_SERVER_STATE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DSSP]).
type DSROLE_SERVER_STATE uint16

const (
	DsRoleServerUnknown DSROLE_SERVER_STATE = 0
	DsRoleServerPrimary DSROLE_SERVER_STATE = 1
	DsRoleServerBackup  DSROLE_SERVER_STATE = 2
)
