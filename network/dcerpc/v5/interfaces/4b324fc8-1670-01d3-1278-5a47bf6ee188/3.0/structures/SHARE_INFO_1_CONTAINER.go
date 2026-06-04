package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_1_CONTAINER contains a count and an array of SHARE_INFO_1 entries
// ([MS-SRVS] 2.2.4.33). Buffer is a [unique] [size_is(EntriesRead)] pointer to a
// conformant array.
type SHARE_INFO_1_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SHARE_INFO_1 `ndr:"unique,size_is=EntriesRead"`
}
