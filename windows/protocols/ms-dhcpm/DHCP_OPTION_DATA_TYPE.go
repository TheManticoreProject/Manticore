package msdhcpm

// DHCP_OPTION_DATA_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DHCPM]).
type DHCP_OPTION_DATA_TYPE uint16

const (
	DhcpByteOption             DHCP_OPTION_DATA_TYPE = 0
	DhcpWordOption             DHCP_OPTION_DATA_TYPE = 1
	DhcpDWordOption            DHCP_OPTION_DATA_TYPE = 2
	DhcpDWordDWordOption       DHCP_OPTION_DATA_TYPE = 3
	DhcpIpAddressOption        DHCP_OPTION_DATA_TYPE = 4
	DhcpStringDataOption       DHCP_OPTION_DATA_TYPE = 5
	DhcpBinaryDataOption       DHCP_OPTION_DATA_TYPE = 6
	DhcpEncapsulatedDataOption DHCP_OPTION_DATA_TYPE = 7
	DhcpIpv6AddressOption      DHCP_OPTION_DATA_TYPE = 8
)
