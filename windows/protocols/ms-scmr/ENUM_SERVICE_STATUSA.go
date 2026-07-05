package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ENUM_SERVICE_STATUSA ([MS-SCMR] 2.2.10) is the ANSI service-enumeration entry. The two
// names are [unique] pointers to conformant-varying ASCII arrays.
type ENUM_SERVICE_STATUSA struct {
	LpServiceName *ndr.STR `ndr:"unique"`
	LpDisplayName *ndr.STR `ndr:"unique"`
	ServiceStatus SERVICE_STATUS
}
