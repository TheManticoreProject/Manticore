package msrrasm

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MIB_IPFORWARDROW ([MS-RRASM] 2.2.1.2.201). DwForwardType and DwForwardProto are
// each the DWORD arm of a C union (with MIB_IPFORWARD_TYPE / MIB_IPFORWARD_PROTO
// enum overlays); both arms are 4 bytes, so the DWORD arm is modeled directly.
type MIB_IPFORWARDROW struct {
	DwForwardDest      ndr.DWORD
	DwForwardMask      ndr.DWORD
	DwForwardPolicy    ndr.DWORD
	DwForwardNextHop   ndr.DWORD
	DwForwardIfIndex   ndr.DWORD
	DwForwardType      ndr.DWORD
	DwForwardProto     ndr.DWORD
	DwForwardAge       ndr.DWORD
	DwForwardNextHopAS ndr.DWORD
	DwForwardMetric1   ndr.DWORD
	DwForwardMetric2   ndr.DWORD
	DwForwardMetric3   ndr.DWORD
	DwForwardMetric4   ndr.DWORD
	DwForwardMetric5   ndr.DWORD
}
