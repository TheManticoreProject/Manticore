package mscmpo

// RESOURCE_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-CMPO]).
type RESOURCE_TYPE uint16

const (
	RT_CONNECTIONS RESOURCE_TYPE = 0x00000000
)
