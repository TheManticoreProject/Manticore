package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_ALIAS_INFO_0_CONTAINER is the level-0 server alias enumeration
// container ([MS-SRVS] 2.2.4.105). Buffer is a [unique] pointer to a conformant
// array of SERVER_ALIAS_INFO_0 sized by EntriesRead.
type SERVER_ALIAS_INFO_0_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SERVER_ALIAS_INFO_0 `ndr:"unique,size_is=EntriesRead"`
}
