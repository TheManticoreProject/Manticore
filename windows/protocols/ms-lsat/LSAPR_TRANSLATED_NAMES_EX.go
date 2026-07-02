package mslsat

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_NAMES_EX contains the translated names from a LsarLookupSids2/3
// request ([MS-LSAT] 2.2.23). Names is a [unique] pointer to a conformant array of
// LSAPR_TRANSLATED_NAME_EX sized by Entries.
type LSAPR_TRANSLATED_NAMES_EX struct {
	Entries ndr.DWORD
	Names   []LSAPR_TRANSLATED_NAME_EX `ndr:"unique,size_is=Entries"`
}
