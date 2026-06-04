package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_ACCOUNT_ENUM_BUFFER is the result buffer of an account enumeration ([MS-LSAD]
// 2.2.5.4). Information is a [unique] pointer to a conformant array of
// LSAPR_ACCOUNT_INFORMATION sized by EntriesRead.
type LSAPR_ACCOUNT_ENUM_BUFFER struct {
	EntriesRead ndr.DWORD
	Information []LSAPR_ACCOUNT_INFORMATION `ndr:"unique,size_is=EntriesRead"`
}
