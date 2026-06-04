package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_INFO_2 contains the identifier of an open file, device, or pipe
// ([MS-SRVS] 2.2.4.6, file info level 2).
type FILE_INFO_2 struct {
	Fi2Id ndr.DWORD
}
