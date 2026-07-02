package msdhcpm

// DHCP_OPTION_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_OPTION_TYPE uint16

const (
	DhcpUnaryElementTypeOption DHCP_OPTION_TYPE = 0
	DhcpArrayTypeOption        DHCP_OPTION_TYPE = 1
)
