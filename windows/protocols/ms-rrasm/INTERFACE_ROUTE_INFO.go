package msrrasm

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// INTERFACE_ROUTE_INFO ([MS-RRASM] 2.2.1.2.6). The leading field is a C union
// overlaying an IPv4 route view (nine DWORDs, 36 bytes) and an IPv6 route view
// (48 bytes). The union carries no on-wire discriminant, so it is modeled by its
// largest arm (the IPv6 view) as a nested struct; consumers reinterpret the bytes
// per address family. The IPv4 view is: dwRtInfoDest, dwRtInfoMask, dwRtInfoPolicy,
// dwRtInfoNextHop, dwRtInfoAge, dwRtInfoNextHopAS, dwRtInfoMetric1..3.
type INTERFACE_ROUTE_INFO struct {
	RouteInfo          INTERFACE_ROUTE_INFO_V6
	DwRtInfoIfIndex    ndr.DWORD
	DwRtInfoType       ndr.DWORD
	DwRtInfoProto      ndr.DWORD
	DwRtInfoPreference ndr.DWORD
	DwRtInfoViewSet    ndr.DWORD
	BV4                ndr.BOOL
}

// INTERFACE_ROUTE_INFO_V6 is the IPv6 (largest, 48-byte) arm of the
// INTERFACE_ROUTE_INFO leading C union ([MS-RRASM] 2.2.1.2.6).
type INTERFACE_ROUTE_INFO_V6 struct {
	DestinationPrefix IN6_ADDR
	DestPrefixLength  ndr.DWORD
	NextHopAddress    IN6_ADDR
	ValidLifeTime     ndr.DWORD
	Flags             ndr.DWORD
	Metric            ndr.DWORD
}
