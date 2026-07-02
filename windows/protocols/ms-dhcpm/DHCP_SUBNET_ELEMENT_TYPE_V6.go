package msdhcpm

// DHCP_SUBNET_ELEMENT_TYPE_V6 is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_SUBNET_ELEMENT_TYPE_V6 uint16

const (
	Dhcpv6IpRanges         DHCP_SUBNET_ELEMENT_TYPE_V6 = 0
	Dhcpv6ReservedIps      DHCP_SUBNET_ELEMENT_TYPE_V6 = 1
	Dhcpv6ExcludedIpRanges DHCP_SUBNET_ELEMENT_TYPE_V6 = 2
)
