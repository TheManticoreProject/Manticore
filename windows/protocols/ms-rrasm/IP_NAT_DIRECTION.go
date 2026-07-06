package msrrasm

// IP_NAT_DIRECTION is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RRASM]).
type IP_NAT_DIRECTION uint16

const (
	NatInboundDirection  IP_NAT_DIRECTION = 0
	NatOutboundDirection IP_NAT_DIRECTION = 1
)
