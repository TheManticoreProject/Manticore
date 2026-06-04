package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_SIDS_EX contains the translated SIDs from a LsarLookupNames2 request
// ([MS-LSAT] 2.2.25). Sids is a [unique] pointer to a conformant array of
// LSAPR_TRANSLATED_SID_EX sized by Entries.
type LSAPR_TRANSLATED_SIDS_EX struct {
	Entries ndr.DWORD
	Sids    []LSAPR_TRANSLATED_SID_EX `ndr:"unique,size_is=Entries"`
}
