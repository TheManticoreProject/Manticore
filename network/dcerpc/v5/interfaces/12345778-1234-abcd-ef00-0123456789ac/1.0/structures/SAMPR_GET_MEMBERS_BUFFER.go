package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_GET_MEMBERS_BUFFER holds the members of a group ([MS-SAMR] 2.2.7.11).
// Members and Attributes are [unique] pointers to conformant arrays, each sized
// by MemberCount.
type SAMPR_GET_MEMBERS_BUFFER struct {
	MemberCount ndr.DWORD
	Members     []ndr.DWORD `ndr:"unique,size_is=MemberCount"`
	Attributes  []ndr.DWORD `ndr:"unique,size_is=MemberCount"`
}
