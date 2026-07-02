package msdhcpm

// DHCP_SEARCH_INFO_TYPE_V6 is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_SEARCH_INFO_TYPE_V6 uint16

const (
	Dhcpv6ClientIpAddress DHCP_SEARCH_INFO_TYPE_V6 = 0
	Dhcpv6ClientDUID      DHCP_SEARCH_INFO_TYPE_V6 = 1
	Dhcpv6ClientName      DHCP_SEARCH_INFO_TYPE_V6 = 2
)
