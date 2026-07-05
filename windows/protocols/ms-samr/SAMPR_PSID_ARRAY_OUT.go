package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_PSID_ARRAY_OUT is a counted array of SID pointers used as an [out]
// parameter ([MS-SAMR] 2.2.7.6). Sids is a [unique] pointer to a conformant
// array of SAMPR_SID_INFORMATION sized by Count.
type SAMPR_PSID_ARRAY_OUT struct {
	Count ndr.DWORD
	Sids  []SAMPR_SID_INFORMATION `ndr:"unique,size_is=Count"`
}
