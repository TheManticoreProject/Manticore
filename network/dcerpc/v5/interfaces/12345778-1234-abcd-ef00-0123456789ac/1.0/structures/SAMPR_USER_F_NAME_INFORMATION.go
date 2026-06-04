package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_F_NAME_INFORMATION holds a user's full name ([MS-SAMR]
// 2.2.6.19).
type SAMPR_USER_F_NAME_INFORMATION struct {
	FullName dtyp.RPC_UNICODE_STRING
}
