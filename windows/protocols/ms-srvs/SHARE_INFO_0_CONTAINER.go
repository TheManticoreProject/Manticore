package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_0_CONTAINER contains a count and an array of SHARE_INFO_0 entries
// ([MS-SRVS] 2.2.4.32). Buffer is a [unique] [size_is(EntriesRead)] pointer to a
// conformant array.
type SHARE_INFO_0_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SHARE_INFO_0 `ndr:"unique,size_is=EntriesRead"`
}
