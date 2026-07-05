package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_DISPLAY_OEM_GROUP_BUFFER is the result buffer of an OEM (ASCII) group
// display query ([MS-SAMR] 2.2.8.11). Buffer is a [unique] pointer to a conformant array
// sized by EntriesRead.
type SAMPR_DOMAIN_DISPLAY_OEM_GROUP_BUFFER struct {
	EntriesRead ndr.DWORD
	Buffer      []SAMPR_DOMAIN_DISPLAY_OEM_GROUP `ndr:"unique,size_is=EntriesRead"`
}
