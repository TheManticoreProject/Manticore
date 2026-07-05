package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_ENUMERATION_BUFFER holds the result of an enumeration ([MS-SAMR]
// 2.2.3.5). Buffer is a [unique] pointer to a conformant array of
// SAMPR_RID_ENUMERATION sized by EntriesRead.
type SAMPR_ENUMERATION_BUFFER struct {
	EntriesRead ndr.DWORD
	Buffer      []SAMPR_RID_ENUMERATION `ndr:"unique,size_is=EntriesRead"`
}
