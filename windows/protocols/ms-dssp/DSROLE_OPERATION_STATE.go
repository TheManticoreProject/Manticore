package msdssp

// DSROLE_OPERATION_STATE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DSSP]).
type DSROLE_OPERATION_STATE uint16

const (
	DsRoleOperationIdle       DSROLE_OPERATION_STATE = 0
	DsRoleOperationActive     DSROLE_OPERATION_STATE = 1
	DsRoleOperationNeedReboot DSROLE_OPERATION_STATE = 2
)
