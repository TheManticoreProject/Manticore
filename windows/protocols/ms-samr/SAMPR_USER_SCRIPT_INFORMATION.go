package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_SCRIPT_INFORMATION holds a user's logon script path ([MS-SAMR]
// 2.2.6, ScriptInformation).
type SAMPR_USER_SCRIPT_INFORMATION struct {
	ScriptPath msdtyp.RPC_UNICODE_STRING
}
