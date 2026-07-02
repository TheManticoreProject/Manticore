package msdhcpm

// DHCP_OPTION_SCOPE_TYPE6 is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_OPTION_SCOPE_TYPE6 uint16

const (
	DhcpDefaultOptions6  DHCP_OPTION_SCOPE_TYPE6 = 0
	DhcpScopeOptions6    DHCP_OPTION_SCOPE_TYPE6 = 1
	DhcpReservedOptions6 DHCP_OPTION_SCOPE_TYPE6 = 2
	DhcpGlobalOptions6   DHCP_OPTION_SCOPE_TYPE6 = 3
)
