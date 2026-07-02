package msdhcpm

// DHCP_FILTER_LIST_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_FILTER_LIST_TYPE uint16

const (
	Deny  DHCP_FILTER_LIST_TYPE = 0
	Allow DHCP_FILTER_LIST_TYPE = 1
)
