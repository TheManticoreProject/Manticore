package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_DISPLAY_OEM_USER_BUFFER is the result buffer of an OEM (ASCII) user
// display query ([MS-SAMR] 2.2.8.10). Buffer is a [unique] pointer to a conformant array
// sized by EntriesRead.
type SAMPR_DOMAIN_DISPLAY_OEM_USER_BUFFER struct {
	EntriesRead ndr.DWORD
	Buffer      []SAMPR_DOMAIN_DISPLAY_OEM_USER `ndr:"unique,size_is=EntriesRead"`
}
