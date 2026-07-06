package msrrasm

// FORWARD_ACTION is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RRASM]).
type FORWARD_ACTION uint16

const (
	FORWARD FORWARD_ACTION = 0
	DROP    FORWARD_ACTION = 1
)
