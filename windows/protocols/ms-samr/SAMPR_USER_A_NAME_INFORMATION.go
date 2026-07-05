package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_A_NAME_INFORMATION holds a user's account name ([MS-SAMR]
// 2.2.6.18).
type SAMPR_USER_A_NAME_INFORMATION struct {
	UserName dtyp.RPC_UNICODE_STRING
}
