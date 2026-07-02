package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_SID_ENUM_BUFFER is the set of SIDs to translate in a LsarLookupSids request
// ([MS-LSAT] 2.2.18). SidInfo is a [unique] pointer to a conformant array of
// LSAPR_SID_INFORMATION sized by Entries.
type LSAPR_SID_ENUM_BUFFER struct {
	Entries ndr.DWORD
	SidInfo []LSAPR_SID_INFORMATION `ndr:"unique,size_is=Entries"`
}
