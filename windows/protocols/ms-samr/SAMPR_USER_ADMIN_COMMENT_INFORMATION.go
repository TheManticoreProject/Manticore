package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_ADMIN_COMMENT_INFORMATION holds a user's administrative comment
// ([MS-SAMR] 2.2.6, AdminCommentInformation).
type SAMPR_USER_ADMIN_COMMENT_INFORMATION struct {
	AdminComment msdtyp.RPC_UNICODE_STRING
}
