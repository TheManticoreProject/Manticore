package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_502_CONTAINER is the level-502 session enumeration container
// ([MS-SRVS] 2.2.4.21). Buffer is a [unique] pointer to a conformant array of
// SESSION_INFO_502 sized by EntriesRead.
type SESSION_INFO_502_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SESSION_INFO_502 `ndr:"unique,size_is=EntriesRead"`
}
