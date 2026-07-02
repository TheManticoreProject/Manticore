package msdhcpm

// DHCP_SUBNET_STATE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_SUBNET_STATE uint16

const (
	DhcpSubnetEnabled          DHCP_SUBNET_STATE = 0
	DhcpSubnetDisabled         DHCP_SUBNET_STATE = 1
	DhcpSubnetEnabledSwitched  DHCP_SUBNET_STATE = 2
	DhcpSubnetDisabledSwitched DHCP_SUBNET_STATE = 3
	DhcpSubnetInvalidState     DHCP_SUBNET_STATE = 4
)
