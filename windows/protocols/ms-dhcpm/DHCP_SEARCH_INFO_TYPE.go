package msdhcpm

// DHCP_SEARCH_INFO_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_SEARCH_INFO_TYPE uint16

const (
	DhcpClientIpAddress       DHCP_SEARCH_INFO_TYPE = 0
	DhcpClientHardwareAddress DHCP_SEARCH_INFO_TYPE = 1
	DhcpClientName            DHCP_SEARCH_INFO_TYPE = 2
)
