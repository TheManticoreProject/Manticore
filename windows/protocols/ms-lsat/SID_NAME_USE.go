package mslsat

// SID_NAME_USE enumerates the type of a security principal identified by a SID
// ([MS-LSAT] 2.2.13). As an NDR enum it is transmitted as a 16-bit unsigned value
// ([C706] section 14.3.6).
type SID_NAME_USE uint16

const (
	SidTypeUser           SID_NAME_USE = 1
	SidTypeGroup          SID_NAME_USE = 2
	SidTypeDomain         SID_NAME_USE = 3
	SidTypeAlias          SID_NAME_USE = 4
	SidTypeWellKnownGroup SID_NAME_USE = 5
	SidTypeDeletedAccount SID_NAME_USE = 6
	SidTypeInvalid        SID_NAME_USE = 7
	SidTypeUnknown        SID_NAME_USE = 8
	SidTypeComputer       SID_NAME_USE = 9
	SidTypeLabel          SID_NAME_USE = 10
)
