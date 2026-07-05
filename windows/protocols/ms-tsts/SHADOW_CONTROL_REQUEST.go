package mststs

// SHADOW_CONTROL_REQUEST selects the level of control a shadow session requests
// ([MS-TSTS] 2.2.1.4, SessEnvRpc.idl). NDR transmits an enum as a 16-bit value.
type SHADOW_CONTROL_REQUEST uint16

const (
	SHADOW_CONTROL_REQUEST_VIEW        SHADOW_CONTROL_REQUEST = 0
	SHADOW_CONTROL_REQUEST_TAKECONTROL SHADOW_CONTROL_REQUEST = 1
	SHADOW_CONTROL_REQUEST_Count       SHADOW_CONTROL_REQUEST = 2
)
