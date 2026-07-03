package msraiw

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WINSINTF_BROWSER_INFO_T ([MS-RAIW] 2.2.2.4). PName is a [string] LPBYTE, i.e. a
// NUL-terminated OEM/ASCII byte string, modeled as a [unique] ndr.STR.
type WINSINTF_BROWSER_INFO_T struct {
	DwNameLen ndr.DWORD
	PName     *ndr.STR `ndr:"unique"`
}
