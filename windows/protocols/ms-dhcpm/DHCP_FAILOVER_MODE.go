package msdhcpm

// DHCP_FAILOVER_MODE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_FAILOVER_MODE uint16

const (
	LoadBalance DHCP_FAILOVER_MODE = 0x00000000
	HotStandby  DHCP_FAILOVER_MODE = 0x00000001
)
