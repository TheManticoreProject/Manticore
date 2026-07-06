package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_F_NAME_INFORMATION holds a user's full name ([MS-SAMR]
// 2.2.6.19).
type SAMPR_USER_F_NAME_INFORMATION struct {
	FullName msdtyp.RPC_UNICODE_STRING
}
