package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_WORKSTATIONS_INFORMATION holds the workstations a user may log on
// from ([MS-SAMR] 2.2.6.12).
type SAMPR_USER_WORKSTATIONS_INFORMATION struct {
	WorkStations dtyp.RPC_UNICODE_STRING
}
