package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_NAME_INFORMATION holds a user's account name and full name
// ([MS-SAMR] 2.2.6.17).
type SAMPR_USER_NAME_INFORMATION struct {
	UserName msdtyp.RPC_UNICODE_STRING
	FullName msdtyp.RPC_UNICODE_STRING
}
