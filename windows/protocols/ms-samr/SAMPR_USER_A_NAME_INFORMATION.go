package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_A_NAME_INFORMATION holds a user's account name ([MS-SAMR]
// 2.2.6.18).
type SAMPR_USER_A_NAME_INFORMATION struct {
	UserName msdtyp.RPC_UNICODE_STRING
}
