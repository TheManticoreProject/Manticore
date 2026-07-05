package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_GROUP_ADM_COMMENT_INFORMATION contains group fields ([MS-SAMR] 2.2.5.7). It holds
// the group's administrative comment.
type SAMPR_GROUP_ADM_COMMENT_INFORMATION struct {
	AdminComment dtyp.RPC_UNICODE_STRING
}
