package msdhcpm

// DHCP_POL_LOGIC_OPER is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_POL_LOGIC_OPER uint16

const (
	DhcpLogicalOr  DHCP_POL_LOGIC_OPER = 0
	DhcpLogicalAnd DHCP_POL_LOGIC_OPER = 1
)
