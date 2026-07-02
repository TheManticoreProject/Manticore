package msdhcpm

// FSM_STATE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type FSM_STATE uint16

const (
	NO_STATE           FSM_STATE = 0x00000000
	INIT               FSM_STATE = 1
	STARTUP            FSM_STATE = 2
	NORMAL             FSM_STATE = 3
	COMMUNICATION_INT  FSM_STATE = 4
	PARTNER_DOWN       FSM_STATE = 5
	POTENTIAL_CONFLICT FSM_STATE = 6
	CONFLICT_DONE      FSM_STATE = 7
	RESOLUTION_INT     FSM_STATE = 8
	RECOVER            FSM_STATE = 9
	RECOVER_WAIT       FSM_STATE = 10
	RECOVER_DONE       FSM_STATE = 11
)
