package msswn

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WITNESS_INTERFACE_INFO describes one witness-capable interface of the server
// ([MS-SWN] 2.2.2.5), as returned in WITNESS_INTERFACE_LIST by WitnessrGetInterfaceList.
//
// InterfaceGroupName is a fixed 260-WCHAR, null-terminated interface group name. State is
// a USHORT taking the values UNKNOWN (0x0000), AVAILABLE (0x0001), or UNAVAILABLE
// (0x00FF). IPV4 is the interface's IPv4 address (0 if none); IPV6 is a fixed array of
// eight USHORTs holding the IPv6 address (all zero if none). Flags is a bitmask
// (INTERFACE_WITNESS 0x00000001 => this node runs the Witness service; otherwise the
// low bit selects whether IPV4/IPV6 is valid per the spec). All fields are fixed size and
// transmitted inline.
type WITNESS_INTERFACE_INFO struct {
	InterfaceGroupName [260]uint16
	Version            ndr.DWORD
	State              uint16
	IPV4               ndr.DWORD
	IPV6               [8]uint16
	Flags              uint32
}
