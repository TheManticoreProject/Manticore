package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ENUM_SERVICE_STATUS_PROCESSA ([MS-SCMR] 2.2.12) is the ANSI extended service-enumeration
// entry. The two names are [unique] pointers to conformant-varying ASCII arrays.
type ENUM_SERVICE_STATUS_PROCESSA struct {
	LpServiceName        *ndr.STR `ndr:"unique"`
	LpDisplayName        *ndr.STR `ndr:"unique"`
	ServiceStatusProcess SERVICE_STATUS_PROCESS
}
