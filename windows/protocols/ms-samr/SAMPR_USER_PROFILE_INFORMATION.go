package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_PROFILE_INFORMATION holds a user's profile path ([MS-SAMR]
// 2.2.6, ProfileInformation).
type SAMPR_USER_PROFILE_INFORMATION struct {
	ProfilePath msdtyp.RPC_UNICODE_STRING
}
