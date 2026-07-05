package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// TSVIP_MAX_ADAPTER_ADDRESS_LENGTH bounds the PhysicalAddress array of TSVIPAddress
// ([MS-TSTS] 2.2.2.5.2, allproc.h).
const TSVIP_MAX_ADAPTER_ADDRESS_LENGTH = 16

// TSVIPAddress describes the IP address leased to a session ([MS-TSTS] 2.2.2.5.2,
// allproc.h _TSVIPAddress).
//
// In the IDL PhysicalAddress is `[length_is(PhysicalAddressLength)] BYTE
// PhysicalAddress[16]` — a fixed-bound NDR *varying* array (offset + actual_count +
// actual_count octets, with no maximum_count). The declarative codec has no
// non-conformant-varying array path (its varying encoder always prefixes a
// maximum_count), and it ignores array tags on a fixed Go array, so the varying framing
// is not emitted here: PhysicalAddress is carried as the 16 fixed octets and
// PhysicalAddressLength gives the valid prefix. Exact wire fidelity for RpcGetSessionIP
// needs a codec enhancement (or an ndr.Marshaler on this type) and is UNVERIFIED.
type TSVIPAddress struct {
	DwVersion             ndr.DWORD
	IPAddress             TSVIP_SOCKADDR
	PrefixOrSubnetMask    ndr.DWORD
	PhysicalAddressLength uint32
	PhysicalAddress       [TSVIP_MAX_ADAPTER_ADDRESS_LENGTH]uint8
	LeaseExpires          ndr.DWORD
	T1                    ndr.DWORD
	T2                    ndr.DWORD
}
