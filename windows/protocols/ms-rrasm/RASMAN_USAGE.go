package msrrasm

// RASMAN_USAGE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RRASM]).
type RASMAN_USAGE uint16

const (
	CALL_NONE            RASMAN_USAGE = 0x00
	CALL_IN              RASMAN_USAGE = 0x01
	CALL_OUT             RASMAN_USAGE = 0x02
	CALL_ROUTER          RASMAN_USAGE = 0x04
	CALL_LOGON           RASMAN_USAGE = 0x08
	CALL_OUT_ONLY        RASMAN_USAGE = 0x10
	CALL_IN_ONLY         RASMAN_USAGE = 0x20
	CALL_OUTBOUND_ROUTER RASMAN_USAGE = 0x40
)
