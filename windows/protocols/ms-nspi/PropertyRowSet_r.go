package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// PropertyRowSet_r ([MS-NSPI] 2.2.4) is a counted set of address book rows. ARow is an
// embedded conformant array (its maximum_count is hoisted to the front of the structure),
// not a pointer, so it is tagged conformant rather than unique.
type PropertyRowSet_r struct {
	CRows ndr.DWORD
	ARow  []PropertyRow_r `ndr:"conformant,size_is=CRows"`
}
