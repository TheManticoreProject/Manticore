package msdhcpm

// DHCP_OPTION_SCOPE_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_OPTION_SCOPE_TYPE uint16

const (
	DhcpDefaultOptions  DHCP_OPTION_SCOPE_TYPE = 0
	DhcpGlobalOptions   DHCP_OPTION_SCOPE_TYPE = 1
	DhcpSubnetOptions   DHCP_OPTION_SCOPE_TYPE = 2
	DhcpReservedOptions DHCP_OPTION_SCOPE_TYPE = 3
	DhcpMScopeOptions   DHCP_OPTION_SCOPE_TYPE = 4
)
