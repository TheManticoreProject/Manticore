package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_ALIAS_NAME_INFORMATION contains alias fields ([MS-SAMR] 2.2.6.3). It holds the
// alias's name.
type SAMPR_ALIAS_NAME_INFORMATION struct {
	Name msdtyp.RPC_UNICODE_STRING
}
