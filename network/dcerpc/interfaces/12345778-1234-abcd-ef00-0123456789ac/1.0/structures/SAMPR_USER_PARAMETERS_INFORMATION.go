package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_PARAMETERS_INFORMATION holds a user's Parameters attribute
// ([MS-SAMR] 2.2.6.13).
type SAMPR_USER_PARAMETERS_INFORMATION struct {
	Parameters dtyp.RPC_UNICODE_STRING
}
