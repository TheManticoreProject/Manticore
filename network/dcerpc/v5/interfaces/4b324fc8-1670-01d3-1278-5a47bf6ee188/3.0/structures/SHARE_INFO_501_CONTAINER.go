package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_501_CONTAINER contains a count and an array of SHARE_INFO_501 entries
// ([MS-SRVS] 2.2.4.35). Buffer is a [unique] [size_is(EntriesRead)] pointer to a
// conformant array.
type SHARE_INFO_501_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SHARE_INFO_501 `ndr:"unique,size_is=EntriesRead"`
}
