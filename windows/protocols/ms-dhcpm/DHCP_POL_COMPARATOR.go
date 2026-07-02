package msdhcpm

// DHCP_POL_COMPARATOR is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_POL_COMPARATOR uint16

const (
	DhcpCompEqual        DHCP_POL_COMPARATOR = 0
	DhcpCompNotEqual     DHCP_POL_COMPARATOR = 1
	DhcpCompBeginsWith   DHCP_POL_COMPARATOR = 2
	DhcpCompNotBeginWith DHCP_POL_COMPARATOR = 3
	DhcpCompEndsWith     DHCP_POL_COMPARATOR = 4
	DhcpCompNotEndWith   DHCP_POL_COMPARATOR = 5
)
