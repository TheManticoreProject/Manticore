package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_REFERENCED_DOMAIN_LIST contains the domains referenced in a translation
// operation ([MS-LSAT] 2.2.12). Domains is a [unique] pointer to a conformant array of
// LSAPR_TRUST_INFORMATION sized by Entries.
type LSAPR_REFERENCED_DOMAIN_LIST struct {
	Entries    ndr.DWORD
	Domains    []LSAPR_TRUST_INFORMATION `ndr:"unique,size_is=Entries"`
	MaxEntries ndr.DWORD
}
