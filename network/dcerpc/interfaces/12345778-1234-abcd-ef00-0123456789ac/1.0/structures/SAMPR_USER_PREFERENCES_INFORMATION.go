package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_PREFERENCES_INFORMATION holds user preference attributes
// ([MS-SAMR] 2.2.6.8).
type SAMPR_USER_PREFERENCES_INFORMATION struct {
	UserComment dtyp.RPC_UNICODE_STRING
	Reserved1   dtyp.RPC_UNICODE_STRING
	CountryCode uint16
	CodePage    uint16
}
