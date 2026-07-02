package msdhcpm

// DHCP_POL_ATTR_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_POL_ATTR_TYPE uint16

const (
	DhcpAttrHWAddr          DHCP_POL_ATTR_TYPE = 0
	DhcpAttrOption          DHCP_POL_ATTR_TYPE = 1
	DhcpAttrSubOption       DHCP_POL_ATTR_TYPE = 2
	DhcpAttrFqdn            DHCP_POL_ATTR_TYPE = 3
	DhcpAttrFqdnSingleLabel DHCP_POL_ATTR_TYPE = 4
)
