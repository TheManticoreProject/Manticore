package msfrs2

// VERSION_CHANGE_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FRS2]).
type VERSION_CHANGE_TYPE uint16

const (
	CHANGE_NOTIFY VERSION_CHANGE_TYPE = 0
	CHANGE_ALL    VERSION_CHANGE_TYPE = 2
)
