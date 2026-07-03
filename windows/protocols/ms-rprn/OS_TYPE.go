package msrprn

// OS_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RPRN]).
type OS_TYPE uint16

const (
	VER_NT_WORKSTATION       OS_TYPE = 0x00000001
	VER_NT_DOMAIN_CONTROLLER OS_TYPE = 0x00000002
	VER_NT_SERVER            OS_TYPE = 0x00000003
)
