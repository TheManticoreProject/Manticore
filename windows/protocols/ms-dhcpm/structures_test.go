package msdhcpm

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in. This is the wire-shape acceptance gate for the
// MS-DHCPM NDR structures in the absence of a live DHCP server.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

func binData(b ...byte) DHCP_BINARY_DATA {
	return DHCP_BINARY_DATA{DataLength: ndr.DWORD(len(b)), Data: b}
}

// TestScalarStructs covers the plain fixed-layout structures (no pointers or arrays).
func TestScalarStructs(t *testing.T) {
	roundTrip(t, "DATE_TIME", DATE_TIME{DwLowDateTime: 0xDEADBEEF, DwHighDateTime: 0x01D7F00D})
	roundTrip(t, "DHCP_IP_RANGE", DHCP_IP_RANGE{StartAddress: 0x0A000001, EndAddress: 0x0A0000FE})
	roundTrip(t, "DHCP_IP_CLUSTER", DHCP_IP_CLUSTER{ClusterAddress: 0x0A000000, ClusterMask: 0xFFFFFF00})
	roundTrip(t, "DWORD_DWORD", DWORD_DWORD{DWord1: 1, DWord2: 2})
	roundTrip(t, "DHCP_BOOTP_IP_RANGE", DHCP_BOOTP_IP_RANGE{
		StartAddress: 0x0A000001, EndAddress: 0x0A0000FE, BootpAllocated: 10, MaxBootpAllowed: 50,
	})
}

// TestDHCP_HOST_INFO covers a struct carrying two [unique] wide-string pointers, one
// present and one null.
func TestDHCP_HOST_INFO(t *testing.T) {
	roundTrip(t, "DHCP_HOST_INFO/full", DHCP_HOST_INFO{
		IpAddress:   0x0A000001,
		NetBiosName: wstr("DC01"),
		HostName:    wstr("dc01.contoso.com"),
	})
	roundTrip(t, "DHCP_HOST_INFO/null-strings", DHCP_HOST_INFO{IpAddress: 0x0A000001})
}

// TestDHCP_BINARY_DATA covers a counted byte blob ([size_is] BYTE* Data). DHCP_CLIENT_UID
// is a type alias for the same shape.
func TestDHCP_BINARY_DATA(t *testing.T) {
	roundTrip(t, "DHCP_BINARY_DATA", binData(0x00, 0x0C, 0x29, 0xAB, 0xCD, 0xEF))
	roundTrip(t, "DHCP_BINARY_DATA/empty", DHCP_BINARY_DATA{DataLength: 0, Data: []uint8{}})
	roundTrip(t, "DHCP_CLIENT_UID", DHCP_CLIENT_UID(binData(0x01, 0x02, 0x03)))
}

// TestDHCP_SUBNET_INFO covers a nested struct (embedded DHCP_HOST_INFO), two unique
// strings, and a 16-bit NDR enum field.
func TestDHCP_SUBNET_INFO(t *testing.T) {
	roundTrip(t, "DHCP_SUBNET_INFO", DHCP_SUBNET_INFO{
		SubnetAddress: 0x0A000000,
		SubnetMask:    0xFFFFFF00,
		SubnetName:    wstr("LAN"),
		SubnetComment: wstr("primary scope"),
		PrimaryHost:   DHCP_HOST_INFO{IpAddress: 0x0A000001, HostName: wstr("dc01")},
		SubnetState:   DhcpSubnetEnabled,
	})
}

// TestDHCP_SUBNET_INFO_VQ additionally covers the trailing INT64 reserved fields.
func TestDHCP_SUBNET_INFO_VQ(t *testing.T) {
	roundTrip(t, "DHCP_SUBNET_INFO_VQ", DHCP_SUBNET_INFO_VQ{
		SubnetAddress: 0x0A000000,
		SubnetMask:    0xFFFFFF00,
		SubnetName:    wstr("LAN"),
		PrimaryHost:   DHCP_HOST_INFO{IpAddress: 0x0A000001},
		SubnetState:   DhcpSubnetEnabled,
		QuarantineOn:  1,
		Reserved3:     0x7FFFFFFFFFFFFFFF,
		Reserved4:     -1,
	})
}

// TestDHCP_IP_ARRAY covers a [unique] pointer to a conformant array of DWORDs — the
// shape returned by R_DhcpEnumSubnets.
func TestDHCP_IP_ARRAY(t *testing.T) {
	roundTrip(t, "DHCP_IP_ARRAY", DHCP_IP_ARRAY{
		NumElements: 3,
		Elements:    []ndr.DWORD{0x0A000000, 0x0A010000, 0x0A020000},
	})
	roundTrip(t, "DHCP_IP_ARRAY/empty", DHCP_IP_ARRAY{NumElements: 0, Elements: []ndr.DWORD{}})
}

// TestDHCP_CLIENT_INFO covers a rich nested structure: a counted UID, two unique
// strings, an embedded DATE_TIME, and an embedded DHCP_HOST_INFO.
func TestDHCP_CLIENT_INFO(t *testing.T) {
	roundTrip(t, "DHCP_CLIENT_INFO", DHCP_CLIENT_INFO{
		ClientIpAddress:       0x0A000064,
		SubnetMask:            0xFFFFFF00,
		ClientHardwareAddress: binData(0x00, 0x0C, 0x29, 0x11, 0x22, 0x33),
		ClientName:            wstr("workstation-1"),
		ClientComment:         wstr("reserved"),
		ClientLeaseExpires:    DATE_TIME{DwLowDateTime: 0x1234, DwHighDateTime: 0x01D7},
		OwnerHost:             DHCP_HOST_INFO{IpAddress: 0x0A000001, HostName: wstr("dc01")},
	})
}

// TestDHCP_CLIENT_INFO_ARRAY covers a conformant array of [unique] pointers to
// DHCP_CLIENT_INFO (elem=unique).
func TestDHCP_CLIENT_INFO_ARRAY(t *testing.T) {
	roundTrip(t, "DHCP_CLIENT_INFO_ARRAY", DHCP_CLIENT_INFO_ARRAY{
		NumElements: 2,
		Clients: []*DHCP_CLIENT_INFO{
			{ClientIpAddress: 0x0A000064, ClientName: wstr("host-a")},
			{ClientIpAddress: 0x0A000065, ClientName: wstr("host-b")},
		},
	})
}

// TestDHCP_OPTION_DATA_ELEMENT exercises every arm of the DHCP_OPTION_ELEMENT_UNION
// discriminated union, including the two [unique] wide-string arms.
func TestDHCP_OPTION_DATA_ELEMENT(t *testing.T) {
	cases := []struct {
		name string
		el   DHCP_OPTION_DATA_ELEMENT
	}{
		{"byte", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpByteOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpByteOption, ByteOption: 0xAB}}},
		{"word", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpWordOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpWordOption, WordOption: 0xBEEF}}},
		{"dword", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpDWordOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpDWordOption, DWordOption: 0xDEADBEEF}}},
		{"dworddword", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpDWordDWordOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpDWordDWordOption, DWordDWordOption: DWORD_DWORD{DWord1: 1, DWord2: 2}}}},
		{"ip", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpIpAddressOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpIpAddressOption, IpAddressOption: 0x0A000001}}},
		{"string", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpStringDataOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpStringDataOption, StringDataOption: wstr("example.com")}}},
		{"binary", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpBinaryDataOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpBinaryDataOption, BinaryDataOption: binData(1, 2, 3)}}},
		{"encapsulated", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpEncapsulatedDataOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpEncapsulatedDataOption, EncapsulatedDataOption: binData(4, 5)}}},
		{"ipv6", DHCP_OPTION_DATA_ELEMENT{OptionType: DhcpIpv6AddressOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpIpv6AddressOption, Ipv6AddressDataOption: wstr("fe80::1")}}},
	}
	for _, c := range cases {
		roundTrip(t, "DHCP_OPTION_DATA_ELEMENT/"+c.name, c.el)
	}
}

// TestDHCP_OPTION_DATA covers a [unique] conformant array of the union element above.
func TestDHCP_OPTION_DATA(t *testing.T) {
	roundTrip(t, "DHCP_OPTION_DATA", DHCP_OPTION_DATA{
		NumElements: 2,
		Elements: []DHCP_OPTION_DATA_ELEMENT{
			{OptionType: DhcpDWordOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpDWordOption, DWordOption: 42}},
			{OptionType: DhcpStringDataOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpStringDataOption, StringDataOption: wstr("dns")}},
		},
	})
}

// TestDHCP_SUBNET_ELEMENT_DATA_V5 exercises a discriminated union whose arms are
// [unique] pointers to distinct structures (one arm per DHCP_SUBNET_ELEMENT_TYPE).
func TestDHCP_SUBNET_ELEMENT_DATA_V5(t *testing.T) {
	roundTrip(t, "DHCP_SUBNET_ELEMENT_DATA_V5/iprange", DHCP_SUBNET_ELEMENT_DATA_V5{
		ElementType: DhcpIpRanges,
		Element: DHCP_SUBNET_ELEMENT_UNION_V5{
			Tag:     DhcpIpRanges,
			IpRange: &DHCP_BOOTP_IP_RANGE{StartAddress: 0x0A000001, EndAddress: 0x0A0000FE},
		},
	})
	roundTrip(t, "DHCP_SUBNET_ELEMENT_DATA_V5/reserved", DHCP_SUBNET_ELEMENT_DATA_V5{
		ElementType: DhcpReservedIps,
		Element: DHCP_SUBNET_ELEMENT_UNION_V5{
			Tag:        DhcpReservedIps,
			ReservedIp: &DHCP_IP_RESERVATION_V4{ReservedIpAddress: 0x0A000064, BAllowedClientTypes: 1},
		},
	})
	roundTrip(t, "DHCP_SUBNET_ELEMENT_DATA_V5/exclude", DHCP_SUBNET_ELEMENT_DATA_V5{
		ElementType: DhcpExcludedIpRanges,
		Element: DHCP_SUBNET_ELEMENT_UNION_V5{
			Tag:            DhcpExcludedIpRanges,
			ExcludeIpRange: &DHCP_IP_RANGE{StartAddress: 0x0A0000C8, EndAddress: 0x0A0000CF},
		},
	})
}

// TestDHCP_SEARCH_INFO exercises the client-lookup union: two scalar/blob arms and one
// [unique] wide-string arm (ClientName).
func TestDHCP_SEARCH_INFO(t *testing.T) {
	roundTrip(t, "DHCP_SEARCH_INFO/ip", DHCP_SEARCH_INFO{
		SearchType: DhcpClientIpAddress,
		SearchInfo: DHCP_CLIENT_SEARCH_UNION{Tag: DhcpClientIpAddress, ClientIpAddress: 0x0A000064},
	})
	roundTrip(t, "DHCP_SEARCH_INFO/hwaddr", DHCP_SEARCH_INFO{
		SearchType: DhcpClientHardwareAddress,
		SearchInfo: DHCP_CLIENT_SEARCH_UNION{Tag: DhcpClientHardwareAddress, ClientHardwareAddress: binData(0, 0x0C, 0x29, 1, 2, 3)},
	})
	roundTrip(t, "DHCP_SEARCH_INFO/name", DHCP_SEARCH_INFO{
		SearchType: DhcpClientName,
		SearchInfo: DHCP_CLIENT_SEARCH_UNION{Tag: DhcpClientName, ClientName: wstr("host-1")},
	})
}

// TestDHCP_PROPERTY exercises the property value union across its arms, including the
// [unique] string arm and the binary-blob arm.
func TestDHCP_PROPERTY(t *testing.T) {
	roundTrip(t, "DHCP_PROPERTY/dword", DHCP_PROPERTY{
		ID: DhcpPropIdClientAddressStateEx, Type: DhcpPropTypeDword,
		Value: DHCP_PROPERTY_VALUE_UNION{Tag: DhcpPropTypeDword, DWordValue: 7},
	})
	roundTrip(t, "DHCP_PROPERTY/string", DHCP_PROPERTY{
		ID: DhcpPropIdPolicyDnsSuffix, Type: DhcpPropTypeString,
		Value: DHCP_PROPERTY_VALUE_UNION{Tag: DhcpPropTypeString, StringValue: wstr("contoso.com")},
	})
	roundTrip(t, "DHCP_PROPERTY/binary", DHCP_PROPERTY{
		ID: DhcpPropIdClientAddressStateEx, Type: DhcpPropTypeBinary,
		Value: DHCP_PROPERTY_VALUE_UNION{Tag: DhcpPropTypeBinary, BinaryValue: binData(0xAA, 0xBB)},
	})
}

// TestDHCP_OPTION covers a full option definition: unique strings plus an embedded
// DHCP_OPTION_DATA (default value) and a 16-bit option-type enum.
func TestDHCP_OPTION(t *testing.T) {
	roundTrip(t, "DHCP_OPTION", DHCP_OPTION{
		OptionID:      6,
		OptionName:    wstr("DNS Servers"),
		OptionComment: wstr("rfc2132"),
		DefaultValue: DHCP_OPTION_DATA{
			NumElements: 1,
			Elements: []DHCP_OPTION_DATA_ELEMENT{
				{OptionType: DhcpIpAddressOption, Element: DHCP_OPTION_ELEMENT_UNION{Tag: DhcpIpAddressOption, IpAddressOption: 0x0A000001}},
			},
		},
		OptionType: DhcpArrayTypeOption,
	})
}

// TestDHCP_ATTRIB exercises the server-attribute union, whose discriminant is a
// 4-byte switch_type(unsigned long): case 1 (BOOL) and case 2 (ULONG).
func TestDHCP_ATTRIB(t *testing.T) {
	roundTrip(t, "DHCP_ATTRIB/bool", DHCP_ATTRIB{
		DhcpAttribId:   0x51,
		DhcpAttribType: 1,
		Field:          DHCP_ATTRIB_Field{Tag: 1, DhcpAttribBool: 1},
	})
	roundTrip(t, "DHCP_ATTRIB/ulong", DHCP_ATTRIB{
		DhcpAttribId:   0x52,
		DhcpAttribType: 2,
		Field:          DHCP_ATTRIB_Field{Tag: 2, DhcpAttribUlong: 0xCAFEBABE},
	})
}

// TestDHCP_OPTION_ARRAY covers a [unique] conformant array of DHCP_OPTION.
func TestDHCP_OPTION_ARRAY(t *testing.T) {
	roundTrip(t, "DHCP_OPTION_ARRAY", DHCP_OPTION_ARRAY{
		NumElements: 1,
		Options: []DHCP_OPTION{
			{OptionID: 3, OptionName: wstr("Router"), OptionType: DhcpUnaryElementTypeOption},
		},
	})
}
