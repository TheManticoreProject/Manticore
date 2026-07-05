package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_PROFILE_INFORMATION holds a user's profile path ([MS-SAMR]
// 2.2.6, ProfileInformation).
type SAMPR_USER_PROFILE_INFORMATION struct {
	ProfilePath dtyp.RPC_UNICODE_STRING
}
