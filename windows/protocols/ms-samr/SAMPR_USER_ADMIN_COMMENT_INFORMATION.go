package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_ADMIN_COMMENT_INFORMATION holds a user's administrative comment
// ([MS-SAMR] 2.2.6, AdminCommentInformation).
type SAMPR_USER_ADMIN_COMMENT_INFORMATION struct {
	AdminComment dtyp.RPC_UNICODE_STRING
}
