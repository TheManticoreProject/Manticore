package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// QUERY_SERVICE_CONFIGA ([MS-SCMR] 2.2.15) is the ANSI form of the service configuration
// record. Each string is a [unique] pointer to a conformant-varying ASCII array.
type QUERY_SERVICE_CONFIGA struct {
	DwServiceType      ndr.DWORD
	DwStartType        ndr.DWORD
	DwErrorControl     ndr.DWORD
	LpBinaryPathName   *ndr.STR `ndr:"unique"`
	LpLoadOrderGroup   *ndr.STR `ndr:"unique"`
	DwTagId            ndr.DWORD
	LpDependencies     *ndr.STR `ndr:"unique"`
	LpServiceStartName *ndr.STR `ndr:"unique"`
	LpDisplayName      *ndr.STR `ndr:"unique"`
}
