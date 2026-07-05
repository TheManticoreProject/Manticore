package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_SCRIPT_INFORMATION holds a user's logon script path ([MS-SAMR]
// 2.2.6, ScriptInformation).
type SAMPR_USER_SCRIPT_INFORMATION struct {
	ScriptPath dtyp.RPC_UNICODE_STRING
}
