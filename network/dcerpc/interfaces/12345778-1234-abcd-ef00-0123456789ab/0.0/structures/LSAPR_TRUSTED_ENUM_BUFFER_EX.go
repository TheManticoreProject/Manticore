package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_ENUM_BUFFER_EX is the result buffer of LsarEnumerateTrustedDomainsEx
// ([MS-LSAD] 2.2.7.24). EnumerationBuffer is a [unique] pointer to a conformant array of
// LSAPR_TRUSTED_DOMAIN_INFORMATION_EX sized by EntriesRead.
type LSAPR_TRUSTED_ENUM_BUFFER_EX struct {
	EntriesRead       ndr.DWORD
	EnumerationBuffer []LSAPR_TRUSTED_DOMAIN_INFORMATION_EX `ndr:"unique,size_is=EntriesRead"`
}
