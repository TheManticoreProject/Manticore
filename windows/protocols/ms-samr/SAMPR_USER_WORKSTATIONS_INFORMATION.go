package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_WORKSTATIONS_INFORMATION holds the workstations a user may log on
// from ([MS-SAMR] 2.2.6.12).
type SAMPR_USER_WORKSTATIONS_INFORMATION struct {
	WorkStations msdtyp.RPC_UNICODE_STRING
}
