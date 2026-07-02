package msfrs1

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// COMM_PACKET is the primary FRS replication message ([MS-FRS1] 2.2.3.5), transmitted
// as the payload of FrsRpcSendCommPkt. The six leading DWORDs describe the packet; Pkt
// is a [unique] pointer to a size_is(PktLen) byte buffer holding the back-to-back
// COMM_PACKET element structures.
//
// DataName and DataHandle are [ignore] void* fields: per the IDL the pointer itself is
// transmitted but its referent is not, and the spec requires both to be 0. A nil
// [unique] pointer marshals as a NULL referent id (four zero octets), which is exactly
// that wire form, so they are modeled as always-nil unique pointers.
type COMM_PACKET struct {
	Major      ndr.DWORD
	Minor      ndr.DWORD
	CsId       ndr.DWORD
	MemLen     ndr.DWORD
	PktLen     ndr.DWORD
	UpkLen     ndr.DWORD
	Pkt        []uint8 `ndr:"unique,size_is=PktLen"`
	DataName   *uint32 `ndr:"unique"` // [ignore] void*, MUST be 0 (always nil on the wire)
	DataHandle *uint32 `ndr:"unique"` // [ignore] void*, MUST be 0 (always nil on the wire)
}
