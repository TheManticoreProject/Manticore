package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_GROUP_ADM_COMMENT_INFORMATION contains group fields ([MS-SAMR] 2.2.5.7). It holds
// the group's administrative comment.
type SAMPR_GROUP_ADM_COMMENT_INFORMATION struct {
	AdminComment msdtyp.RPC_UNICODE_STRING
}
