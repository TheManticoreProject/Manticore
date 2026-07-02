package mscmpo

// SESSION_RANK is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-CMPO]).
type SESSION_RANK uint16

const (
	SRANK_PRIMARY   SESSION_RANK = 0x00000001
	SRANK_SECONDARY SESSION_RANK = 0x00000002
)
