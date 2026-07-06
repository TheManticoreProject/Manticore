package msrrasm

// RASMAN_STATUS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RRASM]).
type RASMAN_STATUS uint16

const (
	OPEN        RASMAN_STATUS = 0
	CLOSED      RASMAN_STATUS = 1
	UNAVAILABLE RASMAN_STATUS = 2
	REMOVED     RASMAN_STATUS = 3
)
