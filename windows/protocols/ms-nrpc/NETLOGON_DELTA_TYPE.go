package msnrpc

// NETLOGON_DELTA_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-NRPC]).
type NETLOGON_DELTA_TYPE uint16

const (
	AddOrChangeDomain     NETLOGON_DELTA_TYPE = 1
	AddOrChangeGroup      NETLOGON_DELTA_TYPE = 2
	DeleteGroup           NETLOGON_DELTA_TYPE = 3
	RenameGroup           NETLOGON_DELTA_TYPE = 4
	AddOrChangeUser       NETLOGON_DELTA_TYPE = 5
	DeleteUser            NETLOGON_DELTA_TYPE = 6
	RenameUser            NETLOGON_DELTA_TYPE = 7
	ChangeGroupMembership NETLOGON_DELTA_TYPE = 8
	AddOrChangeAlias      NETLOGON_DELTA_TYPE = 9
	DeleteAlias           NETLOGON_DELTA_TYPE = 10
	RenameAlias           NETLOGON_DELTA_TYPE = 11
	ChangeAliasMembership NETLOGON_DELTA_TYPE = 12
	AddOrChangeLsaPolicy  NETLOGON_DELTA_TYPE = 13
	AddOrChangeLsaTDomain NETLOGON_DELTA_TYPE = 14
	DeleteLsaTDomain      NETLOGON_DELTA_TYPE = 15
	AddOrChangeLsaAccount NETLOGON_DELTA_TYPE = 16
	DeleteLsaAccount      NETLOGON_DELTA_TYPE = 17
	AddOrChangeLsaSecret  NETLOGON_DELTA_TYPE = 18
	DeleteLsaSecret       NETLOGON_DELTA_TYPE = 19
	DeleteGroupByName     NETLOGON_DELTA_TYPE = 20
	DeleteUserByName      NETLOGON_DELTA_TYPE = 21
	SerialNumberSkip      NETLOGON_DELTA_TYPE = 22
)
