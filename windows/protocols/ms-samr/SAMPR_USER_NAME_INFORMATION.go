package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_NAME_INFORMATION holds a user's account name and full name
// ([MS-SAMR] 2.2.6.17).
type SAMPR_USER_NAME_INFORMATION struct {
	UserName dtyp.RPC_UNICODE_STRING
	FullName dtyp.RPC_UNICODE_STRING
}
