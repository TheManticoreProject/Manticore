package msdhcpm

// DHCP_SUBNET_ELEMENT_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_SUBNET_ELEMENT_TYPE uint16

const (
	DhcpIpRanges          DHCP_SUBNET_ELEMENT_TYPE = 0
	DhcpSecondaryHosts    DHCP_SUBNET_ELEMENT_TYPE = 1
	DhcpReservedIps       DHCP_SUBNET_ELEMENT_TYPE = 2
	DhcpExcludedIpRanges  DHCP_SUBNET_ELEMENT_TYPE = 3
	DhcpIpUsedClusters    DHCP_SUBNET_ELEMENT_TYPE = 4
	DhcpIpRangesDhcpOnly  DHCP_SUBNET_ELEMENT_TYPE = 5
	DhcpIpRangesDhcpBootp DHCP_SUBNET_ELEMENT_TYPE = 6
	DhcpIpRangesBootpOnly DHCP_SUBNET_ELEMENT_TYPE = 7
)
