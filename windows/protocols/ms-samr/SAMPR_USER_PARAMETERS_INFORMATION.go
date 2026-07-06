package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_PARAMETERS_INFORMATION holds a user's Parameters attribute
// ([MS-SAMR] 2.2.6.13).
type SAMPR_USER_PARAMETERS_INFORMATION struct {
	Parameters msdtyp.RPC_UNICODE_STRING
}
