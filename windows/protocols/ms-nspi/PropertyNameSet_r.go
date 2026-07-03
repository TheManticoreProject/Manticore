package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// PropertyNameSet_r ([MS-NSPI] 2.2.6) is a counted set of property names. ANames is an
// embedded conformant array (maximum_count hoisted to the front of the structure), not a
// pointer, so it is tagged conformant rather than unique.
type PropertyNameSet_r struct {
	CNames ndr.DWORD
	ANames []PropertyName_r `ndr:"conformant,size_is=CNames"`
}
