package msdhcpm

// DHCP_PROPERTY_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_PROPERTY_TYPE uint16

const (
	DhcpPropTypeByte   DHCP_PROPERTY_TYPE = 0
	DhcpPropTypeWord   DHCP_PROPERTY_TYPE = 1
	DhcpPropTypeDword  DHCP_PROPERTY_TYPE = 2
	DhcpPropTypeString DHCP_PROPERTY_TYPE = 3
	DhcpPropTypeBinary DHCP_PROPERTY_TYPE = 4
)
