package msrrasm

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// syntaxes is the set of NDR transfer syntaxes every round-trip is exercised under.
var syntaxes = []ndr.Syntax{ndr.NDR20, ndr.NDR64}

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in — under both NDR transfer syntaxes.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	for _, s := range syntaxes {
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s %v marshal: %v", name, s, err)
		}
		var out T
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s %v unmarshal: %v", name, s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s %v round-trip:\n got %+v\nwant %+v", name, s, out, in)
		}
	}
}

// TestEnumWidths confirms an MS-RRASM enum marshals as a 16-bit NDR enum under NDR20
// ([C706] 14.3.6): 2 octets, not 4.
func TestEnumWidths(t *testing.T) {
	type holder struct{ V ROUTER_INTERFACE_TYPE }
	raw, err := ndr.Marshal(&holder{V: ROUTER_IF_TYPE_FULL_ROUTER})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(raw) != 2 {
		t.Fatalf("ROUTER_INTERFACE_TYPE marshalled to %d bytes, want 2 (NDR enum width)", len(raw))
	}
}

// TestRASDEVICETYPEWidth confirms RASDEVICETYPE is transmitted as 4 octets: its flag
// values (RDT_Tunnel = 0x00010000, ...) exceed the 16-bit NDR enum range, so it is
// modeled with a 32-bit base rather than a 16-bit enum.
func TestRASDEVICETYPEWidth(t *testing.T) {
	if got := uint32(RDT_Broadband); got != 0x00080000 {
		t.Fatalf("RDT_Broadband = 0x%08x, want 0x00080000", got)
	}
	type holder struct{ V RASDEVICETYPE }
	raw, err := ndr.Marshal(&holder{V: RDT_Tunnel})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(raw) != 4 {
		t.Fatalf("RASDEVICETYPE marshalled to %d bytes, want 4", len(raw))
	}
}

// TestIN6_ADDR_Size locks the 16-octet fixed layout of the IN6_ADDR C union.
func TestIN6_ADDR_Size(t *testing.T) {
	in := IN6_ADDR{Byte: [16]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("IN6_ADDR marshalled to %d bytes, want 16", len(raw))
	}
	roundTrip(t, "IN6_ADDR", in)
}

func TestFILETIME_RoundTrip(t *testing.T) {
	roundTrip(t, "FILETIME", FILETIME{DwLowDateTime: 0xDEADBEEF, DwHighDateTime: 0x01D9F00D})
}

func TestMIB_IPSTATS_RoundTrip(t *testing.T) {
	roundTrip(t, "MIB_IPSTATS", MIB_IPSTATS{
		DwForwarding: 1, DwDefaultTTL: 128, DwInReceives: 1000, DwNumRoutes: 42,
	})
}

func TestMIB_TCPROW_RoundTrip(t *testing.T) {
	roundTrip(t, "MIB_TCPROW", MIB_TCPROW{
		DwState: 5, DwLocalAddr: 0x0100007F, DwLocalPort: 445,
		DwRemoteAddr: 0x0A000001, DwRemotePort: 12345,
	})
}

// TestMIB_IPFORWARDROW_RoundTrip exercises the struct whose two C unions were each
// collapsed to their DWORD arm (DwForwardType, DwForwardProto).
func TestMIB_IPFORWARDROW_RoundTrip(t *testing.T) {
	roundTrip(t, "MIB_IPFORWARDROW", MIB_IPFORWARDROW{
		DwForwardDest: 0x0A000000, DwForwardMask: 0x00FFFFFF, DwForwardNextHop: 0x0A000001,
		DwForwardType: 3, DwForwardProto: 2, DwForwardMetric1: 10,
	})
}

// TestINTERFACE_ROUTE_INFO_RoundTrip exercises the struct whose leading C union is
// modeled by its largest (IPv6) arm.
func TestINTERFACE_ROUTE_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "INTERFACE_ROUTE_INFO", INTERFACE_ROUTE_INFO{
		RouteInfo: INTERFACE_ROUTE_INFO_V6{
			DestinationPrefix: IN6_ADDR{Byte: [16]byte{0x20, 0x01}},
			DestPrefixLength:  64,
			NextHopAddress:    IN6_ADDR{Byte: [16]byte{0xfe, 0x80}},
			ValidLifeTime:     3600,
			Flags:             1,
			Metric:            100,
		},
		DwRtInfoIfIndex: 7, DwRtInfoType: 3, DwRtInfoProto: 2, BV4: 0,
	})
}

// TestMIB_OPAQUE_INFO_RoundTrip exercises the opaque MIB container whose payload
// union is modeled by its 8-byte alignment arm.
func TestMIB_OPAQUE_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "MIB_OPAQUE_INFO", MIB_OPAQUE_INFO{DwId: 0x11223344, UllAlign: 0xA1B2C3D4E5F60718})
}

// TestIPX_MIB_INDEX_RoundTrip exercises the C union modeled by its largest arm.
func TestIPX_MIB_INDEX_RoundTrip(t *testing.T) {
	roundTrip(t, "IPX_MIB_INDEX", IPX_MIB_INDEX{
		StaticServicesTableIndex: STATIC_SERVICES_TABLE_INDEX{
			InterfaceIndex: 3, ServiceType: 0x0004, ServiceName: [48]uint8{'S', 'R', 'V'},
		},
	})
}
