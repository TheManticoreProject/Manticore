package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_ALIAS_ADM_COMMENT_INFORMATION contains alias fields ([MS-SAMR] 2.2.6.4). It holds
// the alias's administrative comment.
type SAMPR_ALIAS_ADM_COMMENT_INFORMATION struct {
	AdminComment msdtyp.RPC_UNICODE_STRING
}
