package msdhcpm

// DHCP_FAILOVER_SERVER is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_FAILOVER_SERVER uint16

const (
	PrimaryServer   DHCP_FAILOVER_SERVER = 0x00000000
	SecondaryServer DHCP_FAILOVER_SERVER = 0x00000001
)
