package msdnsp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. It is the wire-shape acceptance gate for the MS-DNSP NDR
// structures in the absence of a live DNS server (see the package doc for the spots that
// still need live validation: the switch_is discriminant duplication, the [1]-sentinel
// enumeration lists, and the ASCII-vs-wide string choice per field).
func roundTrip[T any](t *testing.T, name string, in T) []byte {
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
	return raw
}

var sampleGUID = dtyp.GUID{
	Data1: 0x11223344,
	Data2: 0x5566,
	Data3: 0x7788,
	Data4: [8]byte{0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00},
}

// TestIP4Array round-trips a conformant array of IPv4 addresses.
func TestIP4Array(t *testing.T) {
	roundTrip(t, "IP4_ARRAY", IP4_ARRAY{
		AddrCount: 3,
		AddrArray: []ndr.DWORD{0x0100007f, 0x0101a8c0, 0x0201a8c0},
	})
}

// TestDnsAddr pins the fixed-array wire size of DNS_ADDR: [32]CHAR + [8]DWORD = 64 octets,
// no conformance words. A conformant mis-modeling would inflate this.
func TestDnsAddr(t *testing.T) {
	a := DNS_ADDR{DnsAddrUserDword: [8]ndr.DWORD{1, 2, 3, 4, 5, 6, 7, 8}}
	a.MaxSa[0], a.MaxSa[1] = 0x02, 0x00 // AF_INET, port hi
	raw := roundTrip(t, "DNS_ADDR", a)
	if len(raw) != 64 {
		t.Errorf("DNS_ADDR marshaled to %d octets, want 64", len(raw))
	}
}

// TestDnsAddrArray round-trips the DNS_ADDR_ARRAY conformant array of fixed-size structs.
func TestDnsAddrArray(t *testing.T) {
	roundTrip(t, "DNS_ADDR_ARRAY", DNS_ADDR_ARRAY{
		MaxCount:  2,
		AddrCount: 2,
		Family:    2,
		AddrArray: []DNS_ADDR{
			{DnsAddrUserDword: [8]ndr.DWORD{1}},
			{DnsAddrUserDword: [8]ndr.DWORD{2}},
		},
	})
}

// TestDnsRpcBuffer round-trips the counted byte blob (conformant, size_is=DwLength).
func TestDnsRpcBuffer(t *testing.T) {
	roundTrip(t, "DNS_RPC_BUFFER", DNS_RPC_BUFFER{
		DwLength: 5,
		Buffer:   []uint8{0xde, 0xad, 0xbe, 0xef, 0x00},
	})
}

// TestDnsRpcRecord round-trips a resource record with its trailing conformant data buffer.
func TestDnsRpcRecord(t *testing.T) {
	roundTrip(t, "DNS_RPC_RECORD", DNS_RPC_RECORD{
		WDataLength:  4,
		WType:        1, // A
		DwTtlSeconds: 3600,
		Buffer:       []uint8{127, 0, 0, 1},
	})
}

// TestDnsRpcNameAndParam round-trips a struct with a [unique] ASCII string ([string] char*).
func TestDnsRpcNameAndParam(t *testing.T) {
	s := ndr.STR("AllowUpdate")
	roundTrip(t, "DNS_RPC_NAME_AND_PARAM", DNS_RPC_NAME_AND_PARAM{
		DwParam:     1,
		PszNodeName: &s,
	})
}

// TestDnsRpcZoneW2K round-trips a struct carrying a [unique] wide string ([string] wchar_t*).
func TestDnsRpcZoneW2K(t *testing.T) {
	n := ndr.WSTR("example.com")
	roundTrip(t, "DNS_RPC_ZONE_W2K", DNS_RPC_ZONE_W2K{
		PszZoneName: &n,
		Flags:       0x00000001,
		ZoneType:    1,
		Version:     0,
	})
}

// TestDnsRpcZoneListW2K round-trips a conformant array of [unique] pointers to structs —
// the enumeration-buffer shape that exercises per-element referent ids.
func TestDnsRpcZoneListW2K(t *testing.T) {
	a := ndr.WSTR("a.example.com")
	b := ndr.WSTR("b.example.com")
	roundTrip(t, "DNS_RPC_ZONE_LIST_W2K", DNS_RPC_ZONE_LIST_W2K{
		DwZoneCount: 2,
		ZoneArray: []*DNS_RPC_ZONE_W2K{
			{PszZoneName: &a, ZoneType: 1},
			{PszZoneName: &b, ZoneType: 1},
		},
	})
}

// TestDnsRpcSkd round-trips the signing-key descriptor, pinning the embedded dtyp.GUID
// (16 octets) — the guard against modeling it on windows/guid.GUID (24 octets under NDR).
func TestDnsRpcSkd(t *testing.T) {
	ksp := ndr.WSTR("Microsoft Software Key Storage Provider")
	roundTrip(t, "DNS_RPC_SKD", DNS_RPC_SKD{
		Guid:                   sampleGUID,
		PwszKeyStorageProvider: &ksp,
		FIsKSK:                 1,
		DwKeyLength:            2048,
	})
}

// TestDnsRpcUtf8StringList round-trips a conformant array of [unique] ASCII string
// pointers ([size_is(dwCount),string] char* pszStrings[]).
func TestDnsRpcUtf8StringList(t *testing.T) {
	a, b := ndr.STR("first.example."), ndr.STR("second.example.")
	roundTrip(t, "DNS_RPC_UTF8_STRING_LIST", DNS_RPC_UTF8_STRING_LIST{
		DwCount:    2,
		PszStrings: []*ndr.STR{&a, &b},
	})
}

// TestDnsRpcUnicodeStringList round-trips a conformant array of [unique] wide string
// pointers ([size_is(dwCount),string] wchar_t* pwszStrings[]).
func TestDnsRpcUnicodeStringList(t *testing.T) {
	a, b := ndr.WSTR("alpha"), ndr.WSTR("beta")
	roundTrip(t, "DNS_RPC_UNICODE_STRING_LIST", DNS_RPC_UNICODE_STRING_LIST{
		DwCount:     2,
		PwszStrings: []*ndr.WSTR{&a, &b},
	})
}

// TestDnsRpcZoneStats round-trips the zone-statistics structure, exercising the fixed
// struct arrays ZoneQueryStats[32] and ZoneTransferStats[2] (each element carrying 8-byte
// ULONG64 members that force intra-array alignment).
func TestDnsRpcZoneStats(t *testing.T) {
	var s DNS_RPC_ZONE_STATS_V1
	s.DwRpcStructureVersion = 1
	s.ZoneQueryStats[0] = DNSSRV_ZONE_QUERY_STATS{
		RecordType: ZONE_STATS_TYPE_RECORD_A, QueriesReceived: 10, QueriesResponded: 9,
	}
	s.ZoneQueryStats[31] = DNSSRV_ZONE_QUERY_STATS{RecordType: ZONE_STATS_TYPE_RECORD_OTHERS}
	s.ZoneTransferStats[0] = DNSSRV_ZONE_TRANSFER_STATS{
		TransferType: ZONE_STATS_TYPE_TRANSFER_AXFR, RequestReceived: 2,
	}
	s.ZoneUpdateStats = DNSSRV_ZONE_UPDATE_STATS{Type: ZONE_STATS_TYPE_UPDATE}
	s.ZoneRRLStats = DNSSRV_ZONE_RRL_STATS{Type: ZONE_STATS_TYPE_RRL}
	roundTrip(t, "DNS_RPC_ZONE_STATS_V1", s)
}

// TestDnssrvRpcUnionValueArm round-trips the union with its lone value arm selected (DWORD,
// case=1): discriminant then the inline 4-octet value, no referent id.
func TestDnssrvRpcUnionValueArm(t *testing.T) {
	roundTrip(t, "DNSSRV_RPC_UNION/Dword", DNSSRV_RPC_UNION{
		Tag:   ndr.DWORD(DNSSRV_TYPEID_DWORD),
		Dword: 0x12345678,
	})
}

// TestDnssrvRpcUnionPointerArm round-trips the union with a [unique] pointer arm selected
// (NameAndParam, case=15): discriminant, a referent id, then the deferred struct.
func TestDnssrvRpcUnionPointerArm(t *testing.T) {
	s := ndr.STR("RpcProtocol")
	roundTrip(t, "DNSSRV_RPC_UNION/NameAndParam", DNSSRV_RPC_UNION{
		Tag:          ndr.DWORD(DNSSRV_TYPEID_NAME_AND_PARAM),
		NameAndParam: &DNS_RPC_NAME_AND_PARAM{DwParam: 5, PszNodeName: &s},
	})
}
