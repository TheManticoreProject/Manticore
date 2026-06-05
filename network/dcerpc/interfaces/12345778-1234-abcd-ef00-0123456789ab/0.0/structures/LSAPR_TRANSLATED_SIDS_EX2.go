package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_SIDS_EX2 contains the translated SIDs from a LsarLookupNames3/4
// request ([MS-LSAT] 2.2.27). Sids is a [unique] pointer to a conformant array of
// LSAPR_TRANSLATED_SID_EX2 sized by Entries.
type LSAPR_TRANSLATED_SIDS_EX2 struct {
	Entries ndr.DWORD
	Sids    []LSAPR_TRANSLATED_SID_EX2 `ndr:"unique,size_is=Entries"`
}
