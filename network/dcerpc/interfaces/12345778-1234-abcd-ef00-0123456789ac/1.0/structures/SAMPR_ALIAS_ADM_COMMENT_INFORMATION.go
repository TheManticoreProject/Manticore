package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_ALIAS_ADM_COMMENT_INFORMATION contains alias fields ([MS-SAMR] 2.2.6.4). It holds
// the alias's administrative comment.
type SAMPR_ALIAS_ADM_COMMENT_INFORMATION struct {
	AdminComment dtyp.RPC_UNICODE_STRING
}
