package msdhcpm

// DHCP_FORCE_FLAG is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_FORCE_FLAG uint16

const (
	DhcpFullForce     DHCP_FORCE_FLAG = 0
	DhcpNoForce       DHCP_FORCE_FLAG = 1
	DhcpFailoverForce DHCP_FORCE_FLAG = 2
)
