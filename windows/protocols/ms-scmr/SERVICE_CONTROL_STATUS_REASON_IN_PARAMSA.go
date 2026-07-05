package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVICE_CONTROL_STATUS_REASON_IN_PARAMSA ([MS-SCMR] 2.2.30) carries the ANSI reason and
// comment for RControlServiceExA. PszComment is a [unique] pointer to a conformant-varying
// ASCII array.
type SERVICE_CONTROL_STATUS_REASON_IN_PARAMSA struct {
	DwReason   ndr.DWORD
	PszComment *ndr.STR `ndr:"unique"`
}
