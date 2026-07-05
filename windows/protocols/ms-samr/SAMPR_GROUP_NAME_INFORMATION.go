package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_GROUP_NAME_INFORMATION contains group fields ([MS-SAMR] 2.2.5.6). It holds the
// group's name.
type SAMPR_GROUP_NAME_INFORMATION struct {
	Name dtyp.RPC_UNICODE_STRING
}
