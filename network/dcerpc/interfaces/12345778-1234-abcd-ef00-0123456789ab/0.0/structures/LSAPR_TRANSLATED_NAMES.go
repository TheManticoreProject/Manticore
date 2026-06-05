package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_NAMES contains the translated names from a LsarLookupSids request
// ([MS-LSAT] 2.2.21). Names is a [unique] pointer to a conformant array of
// LSAPR_TRANSLATED_NAME sized by Entries.
type LSAPR_TRANSLATED_NAMES struct {
	Entries ndr.DWORD
	Names   []LSAPR_TRANSLATED_NAME `ndr:"unique,size_is=Entries"`
}
