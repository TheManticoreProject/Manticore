package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_10_CONTAINER is the level-10 session enumeration container
// ([MS-SRVS] 2.2.4.20). Buffer is a [unique] pointer to a conformant array of
// SESSION_INFO_10 sized by EntriesRead.
type SESSION_INFO_10_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SESSION_INFO_10 `ndr:"unique,size_is=EntriesRead"`
}
