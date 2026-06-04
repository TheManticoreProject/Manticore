package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_ENUM_BUFFER is the result buffer of a trusted-domain enumeration
// ([MS-LSAD] 2.2.7.18). Information is a [unique] pointer to a conformant array of
// LSAPR_TRUST_INFORMATION sized by EntriesRead.
type LSAPR_TRUSTED_ENUM_BUFFER struct {
	EntriesRead ndr.DWORD
	Information []LSAPR_TRUST_INFORMATION `ndr:"unique,size_is=EntriesRead"`
}
