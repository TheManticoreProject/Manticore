package msrrasm

// OSPF_FILTER_ACTION is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RRASM]).
type OSPF_FILTER_ACTION uint16

const (
	ACTION_DROP   OSPF_FILTER_ACTION = 0
	ACTION_ACCEPT OSPF_FILTER_ACTION = 1
)
