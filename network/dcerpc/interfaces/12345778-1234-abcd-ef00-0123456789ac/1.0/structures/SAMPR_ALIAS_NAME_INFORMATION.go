package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_ALIAS_NAME_INFORMATION contains alias fields ([MS-SAMR] 2.2.6.3). It holds the
// alias's name.
type SAMPR_ALIAS_NAME_INFORMATION struct {
	Name dtyp.RPC_UNICODE_STRING
}
