package mswkst

// NETSETUP_NAME_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-WKST]).
type NETSETUP_NAME_TYPE uint16

const (
	NetSetupUnknown           NETSETUP_NAME_TYPE = 0
	NetSetupMachine           NETSETUP_NAME_TYPE = 1
	NetSetupWorkgroup         NETSETUP_NAME_TYPE = 2
	NetSetupDomain            NETSETUP_NAME_TYPE = 3
	NetSetupNonExistentDomain NETSETUP_NAME_TYPE = 4
	NetSetupDnsMachine        NETSETUP_NAME_TYPE = 5
)
