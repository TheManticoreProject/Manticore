package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_SIDS contains the translated SIDs from a LsarLookupNames request
// ([MS-LSAT] 2.2.19). Sids is a [unique] pointer to a conformant array of
// LSA_TRANSLATED_SID sized by Entries.
type LSAPR_TRANSLATED_SIDS struct {
	Entries ndr.DWORD
	Sids    []LSA_TRANSLATED_SID `ndr:"unique,size_is=Entries"`
}
