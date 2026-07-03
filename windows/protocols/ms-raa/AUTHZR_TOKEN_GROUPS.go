package msraa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// AUTHZR_TOKEN_GROUPS is the array of group SIDs and attributes of a client context
// ([MS-RAA] 2.2.3.9). Groups is the IDL's trailing conformant array member
// (AUTHZR_SID_AND_ATTRIBUTES Groups[]) — NOT a pointer — so it is tagged conformant, not
// unique: NDR hoists its maximum_count to the front of the structure and writes the
// elements in place with no referent id ([C706] 14.3.10, [MS-RPCE] 2.2.4). GroupCount is
// derived from the element count on marshal.
type AUTHZR_TOKEN_GROUPS struct {
	GroupCount ndr.DWORD
	Groups     []AUTHZR_SID_AND_ATTRIBUTES `ndr:"conformant,size_is=GroupCount"`
}
