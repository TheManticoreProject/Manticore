package msscmr

// SC_ACTION_TYPE is declared [v1_enum] in the IDL, so it is transmitted as a 32-bit
// value ([C706] 14.3.6, [MS-RPCE] 2.2.5.1.6) — not the default 16-bit NDR enum.
type SC_ACTION_TYPE uint32

const (
	SC_ACTION_NONE        SC_ACTION_TYPE = 0
	SC_ACTION_RESTART     SC_ACTION_TYPE = 1
	SC_ACTION_REBOOT      SC_ACTION_TYPE = 2
	SC_ACTION_RUN_COMMAND SC_ACTION_TYPE = 3
)
