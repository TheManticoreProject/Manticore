package mslsat

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// LSAPR_REFERENCED_DOMAIN_LIST contains the domains referenced in a translation
// operation ([MS-LSAT] 2.2.12). Domains is a [unique] pointer to a conformant array of
// LSAPR_TRUST_INFORMATION sized by Entries. LSAPR_TRUST_INFORMATION is a type common to
// the lsarpc interface and canonically defined by the MS-LSAD structures package, so it
// is referenced across the package boundary here.
type LSAPR_REFERENCED_DOMAIN_LIST struct {
	Entries    ndr.DWORD
	Domains    []mslsad.LSAPR_TRUST_INFORMATION `ndr:"unique,size_is=Entries"`
	MaxEntries ndr.DWORD
}
