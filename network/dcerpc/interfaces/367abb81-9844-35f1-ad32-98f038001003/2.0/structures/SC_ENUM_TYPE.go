package structures

// SC_ENUM_TYPE is declared [v1_enum] in the IDL, so it is transmitted as a 32-bit
// value ([C706] 14.3.6, [MS-RPCE] 2.2.5.1.6) — not the default 16-bit NDR enum.
type SC_ENUM_TYPE uint32

const (
	SC_ENUM_PROCESS_INFO SC_ENUM_TYPE = 0
)
