package mscmpo

// TEARDOWN_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-CMPO]).
type TEARDOWN_TYPE uint16

const (
	TT_FORCE   TEARDOWN_TYPE = 0x00000000
	TT_PROBLEM TEARDOWN_TYPE = 0x00000002
)
