package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_INFO_3 contains the details of an open file, device, or pipe
// ([MS-SRVS] 2.2.4.7). fi3_pathname and fi3_username are [string] wchar_t* fields
// (pointer_default unique); a nil WSTR is a NULL pointer on the wire.
type FILE_INFO_3 struct {
	Fi3Id          ndr.DWORD
	Fi3Permissions ndr.DWORD
	Fi3NumLocks    ndr.DWORD
	Fi3Pathname    ndr.WSTR `ndr:"unique"`
	Fi3Username    ndr.WSTR `ndr:"unique"`
}
