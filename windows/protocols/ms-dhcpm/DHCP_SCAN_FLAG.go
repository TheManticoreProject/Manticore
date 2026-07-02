package msdhcpm

// DHCP_SCAN_FLAG is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_SCAN_FLAG uint16

const (
	DhcpRegistryFix DHCP_SCAN_FLAG = 0
	DhcpDatabaseFix DHCP_SCAN_FLAG = 1
)
