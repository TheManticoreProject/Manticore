package msdhcpm

// DHCP_PROPERTY_ID is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_PROPERTY_ID uint16

const (
	DhcpPropIdPolicyDnsSuffix      DHCP_PROPERTY_ID = 0
	DhcpPropIdClientAddressStateEx DHCP_PROPERTY_ID = 1
)
