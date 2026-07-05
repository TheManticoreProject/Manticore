package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_GET_GROUPS_BUFFER holds the groups a user belongs to ([MS-SAMR]
// 2.2.7.10). Groups is a [unique] pointer to a conformant array of
// GROUP_MEMBERSHIP sized by MembershipCount.
type SAMPR_GET_GROUPS_BUFFER struct {
	MembershipCount ndr.DWORD
	Groups          []GROUP_MEMBERSHIP `ndr:"unique,size_is=MembershipCount"`
}
